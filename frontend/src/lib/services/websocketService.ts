import { writable, type Writable } from 'svelte/store';
import type { GenericMessage } from '$lib/protocol/messages';

const WEBSOCKET_URL = 'ws://localhost:8080/ws';

export interface WebSocketService {
	connect: () => void;
	disconnect: () => void;
	sendMessage: <T>(type: string, payload: T) => void;
	isConnected: Writable<boolean>;
	lastError: Writable<string | null>;
	onMessage: <T>(messageType: string, callback: (payload: T) => void) => () => void;
}

function createWebSocketService(): WebSocketService {
	let socket: WebSocket | null = null;
	const isConnected = writable(false);
	const lastError = writable<string | null>(null);
	const messageListeners = new Map<string, Set<(payload: any) => void>>();

	function connect() {
		if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
			console.log('WebSocket already connected or connecting.');
			return;
		}

		socket = new WebSocket(WEBSOCKET_URL);
		console.log('Attempting WebSocket connection...');

		socket.onopen = () => {
			console.log('WebSocket connection established.');
			isConnected.set(true);
			lastError.set(null);
		};

		socket.onmessage = (event) => {
			console.log('RAW WebSocket data received:', event.data); // <<<--- ADD THIS LINE
			try {
				const genericMessage = JSON.parse(event.data as string) as GenericMessage;
				console.log('Parsed genericMessage:', genericMessage); // <<<--- ADD THIS LINE TOO
				
				const listeners = messageListeners.get(genericMessage.type);
				if (listeners) {
					listeners.forEach(callback => {
						try {
							callback(genericMessage.payload);
						} catch (listenerError) {
							console.error(`Error in listener for message type ${genericMessage.type}:`, listenerError);
							// Optionally set lastError here too, with more specific info
						}
					});
				} else {
					// console.warn(`No listeners for message type: ${genericMessage.type}`);
				}
			} catch (error) {
				console.error('Error parsing top-level message or dispatching to listeners:', error);
				lastError.set('Error processing received message.');
			}
		};

		socket.onerror = (event) => // Changed from error to event
        {
			console.error('WebSocket error:', event);
			// Cast event to ErrorEvent to access message, or just log generic error
			const errorMsg = event instanceof ErrorEvent ? event.message : 'WebSocket connection error.';
			lastError.set(errorMsg);
			isConnected.set(false); // Usually onclose will also set this
		};

		socket.onclose = (event) => {
			console.log('WebSocket connection closed:', event.reason, `Code: ${event.code}`);
			isConnected.set(false);
			socket = null; // Clear the socket instance
			// Optionally, implement auto-reconnect logic here
		};
	}

	function disconnect() {
		if (socket && socket.readyState === WebSocket.OPEN) {
			socket.close(1000, 'Client initiated disconnect'); // 1000 is normal closure
		}
		socket = null;
		isConnected.set(false);
	}

	function sendMessage<T>(type: string, payload: T) {
		if (socket && socket.readyState === WebSocket.OPEN) {
			const message: GenericMessage<T> = { type, payload };
			try {
				socket.send(JSON.stringify(message));
			} catch (error) {
				console.error("Error sending message:", error);
				lastError.set("Error sending message.");
			}
		} else {
			console.warn('WebSocket not connected. Message not sent:', type, payload);
			lastError.set('WebSocket not connected. Cannot send message.');
		}
	}

	function onMessage<T>(messageType: string, callback: (payload: T) => void): () => void {
		if (!messageListeners.has(messageType)) {
			messageListeners.set(messageType, new Set());
		}
		const listeners = messageListeners.get(messageType)!;
		listeners.add(callback);

		// Return an unsubscribe function
		return () => {
			listeners.delete(callback);
			if (listeners.size === 0) {
				messageListeners.delete(messageType);
			}
		};
	}

	return {
		connect,
		disconnect,
		sendMessage,
		isConnected,
		lastError,
		onMessage,
	};
}

export const websocketService = createWebSocketService();