export type SocketStore = {
	// Connection status
	isConnected: boolean;
	// Message content
	message: string;
	// Reconnect error
	reconnectError: boolean;
	// Heartbeat message sending time
	heartBeatInterval: number;
	// Heartbeat timer
	heartBeatTimer: number;
};

export type socketType = {
	$connect: () => void;
};
