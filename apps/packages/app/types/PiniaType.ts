export type SocketStore = {
	isConnected: boolean;
	reconnectError: boolean;
	heartBeatInterval: number;
	heartBeatTimer: number;
};
