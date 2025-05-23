package server

import (
	"encoding/json"
	"fmt"
	"game-server/internal/game"
	"game-server/internal/protocol"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	player *game.Player
	world  *game.World
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	world      *game.World
}

func NewHub(world *game.World) *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		world:      world,
	}
}

func (h *Hub) Run() {
	log.Println("Hub started...")
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered: %s (Player ID: %s). Total clients: %d", client.conn.RemoteAddr(), client.player.GetID(), len(h.clients))

			h.world.Mu.Lock()
			mapData := NewS2C_MapData(h.world)

			var playersData []protocol.S2C_PlayerData
			for _, p := range h.world.Players {
				playersData = append(playersData, NewS2C_PlayerData(p))
			}

			var monstersData []protocol.S2C_MonsterData
			for _, m := range h.world.Monsters {
				monstersData = append(monstersData, NewS2C_MonsterData(m))
			}
			h.world.Mu.Unlock()

			initialStatePayload := protocol.S2C_InitialStatePayload{
				PlayerID: client.player.GetID(),
				Map:      mapData,
				Players:  playersData,
				Monsters: monstersData,
			}
			initialStateMsg := protocol.GenericMessage{
				Type:    protocol.S2C_MessageTypeInitialState,
				Payload: initialStatePayload,
			}
			jsonInitialMsg, err := json.Marshal(initialStateMsg)
			if err != nil {
				log.Printf("Error marshaling initial state for player %s: %v", client.player.GetID(), err)
			} else {
				select {
				case client.send <- jsonInitialMsg:
					log.Printf("Sent initial state to player %s", client.player.GetID())
				default:
					log.Printf("Failed to send initial state to player %s: send channel blocked/closed.", client.player.GetID())
				}
			}

			playerJoinedPayload := protocol.S2C_PlayerJoinedPayload{
				S2C_PlayerData: NewS2C_PlayerData(client.player),
			}
			playerJoinedMsg := protocol.GenericMessage{
				Type:    protocol.S2C_MessageTypePlayerJoined,
				Payload: playerJoinedPayload,
			}
			jsonPlayerJoinedMsg, err := json.Marshal(playerJoinedMsg)
			if err != nil {
				log.Printf("Error marshaling player joined message for %s: %v", client.player.GetID(), err)
			} else {
				h.broadcast <- jsonPlayerJoinedMsg
				log.Printf("Scheduled broadcast for player joined: %s", client.player.GetID())
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				playerIDToBroadcast := client.player.GetID()

				delete(h.clients, client)
				close(client.send)
				h.world.RemovePlayer(playerIDToBroadcast)
				log.Printf("Client unregistered: Player ID: %s. Player removed. Total clients: %d", playerIDToBroadcast, len(h.clients))

				playerLeftPayload := protocol.S2C_PlayerLeftPayload{
					ID: playerIDToBroadcast,
				}
				playerLeftMsg := protocol.GenericMessage{
					Type:    protocol.S2C_MessageTypePlayerLeft,
					Payload: playerLeftPayload,
				}
				jsonPlayerLeftMsg, err := json.Marshal(playerLeftMsg)
				if err != nil {
					log.Printf("Error marshaling player left message for %s: %v", playerIDToBroadcast, err)
				} else {
					h.broadcast <- jsonPlayerLeftMsg
					log.Printf("Scheduled broadcast for player left: %s", playerIDToBroadcast)
				}
			}

		case message := <-h.broadcast:
			numClients := len(h.clients)
			if numClients == 0 {
				continue
			}

			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					log.Printf("Client %s send buffer full or slow during broadcast. Removing from broadcast.", c.conn.RemoteAddr())
					playerID := c.player.GetID()
					delete(h.clients, c)
					close(c.send)
					log.Printf("Forcefully removed client %s (Player %s) from broadcast recipients due to slow send.", c.conn.RemoteAddr(), playerID)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		log.Printf("Hub's main broadcast channel is full during Hub.Broadcast(). Message dropped.")
	}
}

func (c *Client) processIncomingMessage(genericMsg protocol.GenericMessage) {
	switch genericMsg.Type {
	case protocol.C2S_MessageTypeMove:
		var movePayload protocol.C2S_MovePayload

		payloadBytes, err := json.Marshal(genericMsg.Payload)
		if err != nil {
			log.Printf("Player %s: Error re-marshaling C2S_MovePayload from generic payload: %v", c.player.GetID(), err)
			return
		}
		if err := json.Unmarshal(payloadBytes, &movePayload); err != nil {
			log.Printf("Player %s: Error unmarshaling C2S_MovePayload: %v", c.player.GetID(), err)
			return
		}

		log.Printf("Player %s attempting move: dx=%d, dy=%d", c.player.GetID(), movePayload.DX, movePayload.DY)

		var moved bool
		var engagedMonster *game.Monster
		var playerCurrentX, playerCurrentY int

		c.world.Mu.Lock()

		moved, engagedMonster = c.player.Move(movePayload.DX, movePayload.DY, c.world)

		playerCurrentX = c.player.GetX()
		playerCurrentY = c.player.GetY()

		if engagedMonster != nil {
			c.player.IsInCombat = true
			c.player.CombatTargetID = engagedMonster.GetID()
			engagedMonster.IsInCombat = true
			engagedMonster.CombatTargetID = c.player.GetID()

			log.Printf("Combat initiated: Player %s vs Monster %s (%s)", c.player.GetID(), engagedMonster.GetID(), engagedMonster.Name)

			combatInitiatedPayload := protocol.S2C_CombatInitiatedPayload{
				PlayerID:  c.player.GetID(),
				MonsterID: engagedMonster.GetID(),
				PlayerX:   playerCurrentX,
				PlayerY:   playerCurrentY,
				MonsterX:  engagedMonster.GetX(),
				MonsterY:  engagedMonster.GetY(),
			}
			combatMsg := protocol.GenericMessage{
				Type:    protocol.S2C_MessageTypeCombatInitiated,
				Payload: combatInitiatedPayload,
			}
			jsonCombatMsg, marshalErr := json.Marshal(combatMsg)
			if marshalErr != nil {
				log.Printf("Player %s: Error marshaling S2C_CombatInitiated message: %v", c.player.GetID(), marshalErr)
			} else {
				c.hub.Broadcast(jsonCombatMsg)
			}
		}

		c.world.Mu.Unlock()

		if moved {
			log.Printf("Player %s successfully moved to (%d, %d)", c.player.GetID(), playerCurrentX, playerCurrentY)

			moveUpdatePayload := protocol.S2C_EntityMovedPayload{
				ID:         c.player.GetID(),
				EntityType: protocol.EntityTypePlayer,
				X:          playerCurrentX,
				Y:          playerCurrentY,
			}
			msg := protocol.GenericMessage{
				Type:    protocol.S2C_MessageTypeEntityMoved,
				Payload: moveUpdatePayload,
			}
			jsonMsg, marshalErr := json.Marshal(msg)
			if marshalErr != nil {
				log.Printf("Player %s: Error marshaling S2C_EntityMoved message: %v", c.player.GetID(), marshalErr)
			} else {
				c.hub.Broadcast(jsonMsg)
			}
		} else if engagedMonster == nil {

			log.Printf("Player %s move (dx=%d, dy=%d) was invalid and no combat initiated. Current pos: (%d,%d)", c.player.GetID(), movePayload.DX, movePayload.DY, playerCurrentX, playerCurrentY)
		}
	case protocol.C2S_MessageTypeAttack:
		if !c.player.IsInCombat || c.player.CombatTargetID == "" {
			log.Printf("Player %s sent attack command but is not in combat or has no target.", c.player.GetID())
			return
		}

		var attackPayload protocol.C2S_AttackPayload
		payloadBytes, err := json.Marshal(genericMsg.Payload)
		if err != nil {
			log.Printf("Player %s: Error re-marshaling C2S_AttackPayload: %v", c.player.GetID(), err)
			return
		}
		if err := json.Unmarshal(payloadBytes, &attackPayload); err != nil {
			log.Printf("Player %s: Error unmarshaling C2S_AttackPayload: %v", c.player.GetID(), err)
			return
		}

		if attackPayload.TargetID != c.player.CombatTargetID {
			log.Printf("Player %s attacked target %s, but current combat target is %s.", c.player.GetID(), attackPayload.TargetID, c.player.CombatTargetID)
			return
		}

		log.Printf("Player %s attacking Monster %s", c.player.GetID(), attackPayload.TargetID)

		var defeatedMonsterID string

		c.world.Mu.Lock()

		monster, monsterExists := c.world.Monsters[attackPayload.TargetID]
		if !monsterExists || monster == nil || !monster.IsInCombat || monster.CombatTargetID != c.player.GetID() {
			log.Printf("Player %s attack failed: Monster %s not valid or not in combat with player.", c.player.GetID(), attackPayload.TargetID)
			c.world.Mu.Unlock()
			return
		}

		damageDealt := c.player.Attack - monster.Defense
		if damageDealt < 0 {
			damageDealt = 0
		}
		isMonsterDefeated := monster.TakeDamage(damageDealt)

		log.Printf("Player %s dealt %d damage to Monster %s. Monster HP: %d/%d.",
			c.player.GetID(), damageDealt, monster.GetID(), monster.CurrentHP, monster.MaxHP)

		playerAttackCombatUpdate := protocol.S2C_CombatUpdatePayload{
			AttackerID:         c.player.GetID(),
			DefenderID:         monster.GetID(),
			DamageDealt:        damageDealt,
			DefenderCurrentHP:  monster.CurrentHP,
			IsDefenderDefeated: isMonsterDefeated,
		}

		if isMonsterDefeated {
			log.Printf("Monster %s was defeated by Player %s!", monster.GetID(), c.player.GetID())
			defeatedMonsterID = monster.GetID()

			c.player.IsInCombat = false
			c.player.CombatTargetID = ""
		}

		c.world.Mu.Unlock()

		combatUpdateMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypeCombatUpdate, Payload: playerAttackCombatUpdate}
		jsonCombatUpdateMsg, errCU := json.Marshal(combatUpdateMsg)
		if errCU == nil {
			c.hub.Broadcast(jsonCombatUpdateMsg)
		} else {
			log.Printf("Error marshaling player attack combat update: %v", errCU)
		}

		if isMonsterDefeated {
			entityRemovedPayload := protocol.S2C_EntityRemovedPayload{
				ID:         defeatedMonsterID,
				EntityType: protocol.EntityTypeMonster,
			}
			entityRemovedMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypeEntityRemoved, Payload: entityRemovedPayload}
			jsonEntityRemovedMsg, errER := json.Marshal(entityRemovedMsg)
			if errER == nil {
				c.hub.Broadcast(jsonEntityRemovedMsg)
			} else {
				log.Printf("Error marshaling entity removed: %v", errER)
			}

			c.world.RemoveMonster(defeatedMonsterID)

			var playerLeveledUp bool
			var statUpdatePayload protocol.S2C_PlayerStatUpdatePayload

			c.world.Mu.Lock()

			xpGained := monster.XPValue

			playerLeveledUp = c.player.GainXP(xpGained)

			statUpdatePayload = protocol.S2C_PlayerStatUpdatePayload{
				PlayerID:      c.player.GetID(),
				Level:         c.player.Level,
				XP:            c.player.XP,
				XPToNextLevel: c.player.XPToNextLevel,
				MaxHP:         c.player.MaxHP,
				CurrentHP:     c.player.CurrentHP,
				Attack:        c.player.Attack,
				Defense:       c.player.Defense,
			}
			c.world.Mu.Unlock()

			playerStatMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypePlayerStatUpdate, Payload: statUpdatePayload}
			jsonPlayerStatMsg, errPSU := json.Marshal(playerStatMsg)
			if errPSU == nil {
				c.hub.Broadcast(jsonPlayerStatMsg)
				log.Printf("Player %s stats updated and broadcasted. Leveled up: %t", c.player.GetID(), playerLeveledUp)
			} else {
				log.Printf("Error marshaling player stat update after monster defeat: %v", errPSU)
			}
		} else {
			log.Printf("Monster %s (HP: %d/%d) survived. Retaliating...", monster.GetID(), monster.CurrentHP, monster.MaxHP)

			var monsterAttackCombatUpdate protocol.S2C_CombatUpdatePayload
			var playerStatUpdateForDefeat *protocol.S2C_PlayerStatUpdatePayload

			c.world.Mu.Lock()

			monsterDamageDealt := monster.Attack - c.player.Defense
			if monsterDamageDealt < 0 {
				monsterDamageDealt = 0
			}
			isPlayerDefeated := c.player.TakeDamage(monsterDamageDealt)

			log.Printf("Monster %s dealt %d damage to Player %s. Player HP: %d/%d.",
				monster.GetID(), monsterDamageDealt, c.player.GetID(), c.player.CurrentHP, c.player.MaxHP)

			monsterAttackCombatUpdate = protocol.S2C_CombatUpdatePayload{
				AttackerID:         monster.GetID(),
				DefenderID:         c.player.GetID(),
				DamageDealt:        monsterDamageDealt,
				DefenderCurrentHP:  c.player.CurrentHP,
				IsDefenderDefeated: isPlayerDefeated,
			}

			if isPlayerDefeated {
				log.Printf("Player %s was defeated by Monster %s!", c.player.GetID(), monster.GetID())
				c.player.ResetToLevel1()
				monster.IsInCombat = false
				monster.CombatTargetID = ""

				playerStatUpdateForDefeat = &protocol.S2C_PlayerStatUpdatePayload{
					PlayerID:      c.player.GetID(),
					Level:         c.player.Level,
					XP:            c.player.XP,
					XPToNextLevel: c.player.XPToNextLevel,
					MaxHP:         c.player.MaxHP,
					CurrentHP:     c.player.CurrentHP,
					Attack:        c.player.Attack,
					Defense:       c.player.Defense,
				}
			}
			c.world.Mu.Unlock()

			monsterAttackMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypeCombatUpdate, Payload: monsterAttackCombatUpdate}
			jsonMonsterAttackMsg, errMA := json.Marshal(monsterAttackMsg)
			if errMA == nil {
				c.hub.Broadcast(jsonMonsterAttackMsg)
			} else { /* log */
			}

			if playerStatUpdateForDefeat != nil {
				playerStatMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypePlayerStatUpdate, Payload: *playerStatUpdateForDefeat}
				jsonPlayerStatMsg, errPSU := json.Marshal(playerStatMsg)
				if errPSU == nil {
					c.hub.Broadcast(jsonPlayerStatMsg)
				} else { /* log */
				}
			}
		}
	case protocol.C2S_MessageTypeUsePotion:
		log.Printf("Player %s attempting to use a potion.", c.player.GetID())

		var actualHealAmount int
		var notificationMsg string

		c.world.Mu.Lock()
		if c.player.CurrentHP <= 0 {
			log.Printf("Player %s cannot use potion, is defeated.", c.player.GetID())
			notificationMsg = "You are defeated and cannot use a potion."
		} else if c.player.CurrentHP >= c.player.MaxHP {
			log.Printf("Player %s is already at full health.", c.player.GetID())
			notificationMsg = "You are already at full health."

		} else {
			potionHealAmount := 30
			actualHealAmount = c.player.Heal(potionHealAmount)
			if actualHealAmount > 0 {
				notificationMsg = fmt.Sprintf("You healed for %d HP.", actualHealAmount)
			} else {
				notificationMsg = "You feel no different."
			}
		}

		statUpdatePayload := protocol.S2C_PlayerStatUpdatePayload{
			PlayerID:      c.player.GetID(),
			Level:         c.player.Level,
			XP:            c.player.XP,
			XPToNextLevel: c.player.XPToNextLevel,
			MaxHP:         c.player.MaxHP,
			CurrentHP:     c.player.CurrentHP,
			Attack:        c.player.Attack,
			Defense:       c.player.Defense,
		}
		c.world.Mu.Unlock()

		playerStatMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypePlayerStatUpdate, Payload: statUpdatePayload}
		jsonPlayerStatMsg, errPSU := json.Marshal(playerStatMsg)
		if errPSU == nil {
			c.hub.Broadcast(jsonPlayerStatMsg)
		} else {
			log.Printf("Error marshaling player stat update after potion use: %v", errPSU)
		}

		if notificationMsg != "" {
			notificationPayload := protocol.S2C_NotificationPayload{
				Message: notificationMsg,
				Level:   "info",
			}
			if actualHealAmount > 0 {
				notificationPayload.Level = "success"
			}

			notifyMsg := protocol.GenericMessage{Type: protocol.S2C_MessageTypeNotification, Payload: notificationPayload}
			jsonNotifyMsg, errNotify := json.Marshal(notifyMsg)
			if errNotify == nil {
				select {
				case c.send <- jsonNotifyMsg:
				default:
					log.Printf("Failed to send potion use notification to %s: channel full/closed", c.player.GetID())
				}
			} else {
				log.Printf("Error marshaling potion notification: %v", errNotify)
			}
		}
	default:
		log.Printf("Player %s: Received unknown message type '%s'", c.player.GetID(), genericMsg.Type)
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("readPump: Client %s (Player %s) disconnected, connection closed.", c.conn.RemoteAddr(), c.player.GetID())
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("readPump error for player %s: %v", c.player.GetID(), err)
			} else {
				log.Printf("readPump: Player %s WebSocket closed: %v", c.player.GetID(), err)
			}
			break
		}

		var genericMsg protocol.GenericMessage
		if err := json.Unmarshal(rawMessage, &genericMsg); err != nil {
			log.Printf("Error unmarshaling message from player %s: %v. Message: %s", c.player.GetID(), err, string(rawMessage))
			continue
		}

		c.processIncomingMessage(genericMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("writePump: Client %s (Player %s) disconnected, connection closed.", c.conn.RemoteAddr(), c.player.GetID())
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Printf("writePump: Hub closed send channel for player %s.", c.player.GetID())
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("writePump error (WriteMessage) for player %s: %v", c.player.GetID(), err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("writePump error (Ping) for player %s: %v", c.player.GetID(), err)
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	log.Printf("Client connected: %s", conn.RemoteAddr())

	playerID := fmt.Sprintf("player-%d", rand.Intn(10000))
	var startX, startY int
	hub.world.Mu.Lock()
	for {
		fmt.Printf("looking")
		sx := rand.Intn(hub.world.Width-2) + 1
		sy := rand.Intn(hub.world.Height-2) + 1
		if hub.world.IsWalkable(sx, sy) && !hub.world.IsOccupiedInternal(sx, sy) {
			startX = sx
			startY = sy
			fmt.Printf("found start position")
			break
		}
	}

	player := game.NewPlayer(playerID, startX, startY)
	hub.world.AddPlayer(player)
	hub.world.Mu.Unlock()

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		player: player,
		world:  hub.world,
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	log.Printf("Player %s created and client pumps started for %s.", player.GetID(), conn.RemoteAddr())
}

func Start(world *game.World, port string) {
	hub := NewHub(world)
	world.SetHubBroadcaster(hub)

	world.SpawnInitialMonsters(5)

	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Game server is running. Connect via WebSocket on /ws.")
	})
	log.Printf("HTTP server listening on :%s, WebSocket endpoint on /ws", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
