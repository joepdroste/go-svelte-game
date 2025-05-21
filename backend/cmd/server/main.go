package main

import (
	"fmt"
	"game-server/internal/config"
	"game-server/internal/game"
	"game-server/internal/server"
	"log"
	"time"
)

func initializeGame() (*config.Config, *game.World, error) {
	fmt.Println("Initializing game...")
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}
	fmt.Printf("Configuration loaded: ServerPort=%s, MapWidth=%d, MapHeight=%d\n", cfg.ServerPort, cfg.MapWidth, cfg.MapHeight)

	world := game.NewWorld(cfg.MapWidth, cfg.MapHeight)
	fmt.Printf("Game world initialized with %d x %d tiles.\n", world.Width, world.Height)
	fmt.Println("Map:")
	fmt.Println(world.String())

	return cfg, world, nil
}

func main() {
	fmt.Println("Starting game server...")

	cfg, world, err := initializeGame()
	if err != nil {
		log.Fatalf("Failed during game initialization: %v", err)
	}

	log.Printf("Game initialized. Handing off to server module to listen on port %s.", cfg.ServerPort)

	// debugPrint(world)

	server.Start(world, cfg.ServerPort) // Start the server
}

func debugPrint(world *game.World) {
	fmt.Print("\x1b[2J")         // Clear screen
	fmt.Print("\x1b[?25l")       // Hide cursor
	defer fmt.Print("\x1b[?25h") // Restore cursor

	var frameCount int
	var fps int
	lastFPSUpdate := time.Now()

	for {
		frameCount++

		now := time.Now()
		if now.Sub(lastFPSUpdate) >= time.Second {
			fps = frameCount
			frameCount = 0
			lastFPSUpdate = now
		}

		// Draw output
		fmt.Print("\x1b[H") // Move to top-left
		fmt.Printf("FPS: %d\n", fps)
		fmt.Printf("Frames (this second): %d\n", frameCount)
		fmt.Println(world.String())
	}
}
