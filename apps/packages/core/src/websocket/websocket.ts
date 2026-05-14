import { WebSocketStatusEnum } from './WebSocketEnum';
export interface IWebSocketBean {
	/**
	 * Connection status
	 */
	status: WebSocketStatusEnum;

	/**
	 * WebSocket object
	 */
	websocket: WebSocket;

	/**
	 * Heartbeat object
	 */
	heart: IWebSocketHeart;

	/**
	 * Reconnection object
	 */
	reconnect: IWebSocketReconnect;

	/**
	 * Sending object
	 */
	sendObj: IWebSocketSend;

	/**
	 * Parameter information
	 */
	param: IWebSocketBeanParam;

	/**
	 * Close the old connection and create a new connection
	 * @param param
	 * @returns
	 */
	start: (param?: IWebSocketBeanParam, restart?: boolean) => void;

	/**
	 * Send data
	 * @param data Data object, Object, Array, or String
	 * @param resend Whether to resend the data after reconnecting
	 */
	send(data: any, resend?: boolean): string | boolean;

	/**
	 * Destroy the data that needs to be resent
	 * @param sendId
	 */
	offsend: (sendId: string) => void;

	/**
	 * Bind exception handling
	 */
	onerror: () => void;

	/**
	 * Close the WebSocket, destroy binding events, heartbeat events, window close events, and update the status to closed
	 */
	close: () => void;

	/**
	 * Destroy all objects
	 */
	dispose: () => void;
}

/**
 * Parameter information
 */
export interface IWebSocketBeanParam {
	/**
	 * Connection URL
	 */
	url: string;

	/**
	 * Prefix for sending messages, default is empty
	 */
	sendPrefix?: string;

	/**
	 * Suffix for sending messages, default is empty
	 */
	sendSuffix?: string;

	/**
	 * Prefix for receiving messages, default is empty
	 */
	messagePrefix?: string;

	/**
	 * Suffix for receiving messages, default is empty
	 */
	messageSuffix?: string;

	/**
	 * Lifecycle - called first after establishing a connection
	 */
	onopen?: () => Promise<any>;

	/**
	 * Lifecycle - called first after receiving data
	 */
	onmessage?: (ev: MessageEvent<any>) => any;

	/**
	 * Lifecycle - called first after closing or encountering a connection exception
	 */
	onerror?: () => void;

	/**
	 * Lifecycle - called first when reconnection starts
	 */
	onreconnect?: () => void;

	// Reconnection parameter list

	/**
	 * Maximum number of reconnection attempts, default is 10
	 */
	reconnectMaxNum?: number;

	/**
	 * Reconnection interval time, default is 30000
	 */
	reconnectGapTime?: number;

	/**
	 * Initial gradient interval time for reconnection, default is 1000
	 */
	gradientReconnectStartTime?: number;

	/**
	 * Maximum gradient interval time for reconnection, default is 60000
	 */
	gradientReconnectMaxTime?: number;

	/**
	 * Whether gradient reconnection is required, default is true
	 */
	isReconnectGradientEnabled?: boolean;

	/**
	 * Whether reconnection is required, default is false
	 */
	needReconnect?: boolean;

	/**
	 * Notification of successful reconnection
	 */
	onReconnectSuccess?: () => Promise<void>;

	/**
	 * Notification of reconnection failure
	 */
	onReconnectFailure?: () => void;

	// Heartbeat parameter list

	/**
	 * Heartbeat sending content, default is 'heartSend'
	 */
	heartSend?: string;

	/**
	 * Heartbeat receiving content, default is 'heartGet'
	 */
	heartGet?: string;

	/**
	 * Heartbeat sending interval time, default is 30000
	 */
	heartGapTime?: number;

	/**
	 * Maximum number of heartbeat failures, default is 10
	 */
	heartFailNum?: number;

	heartRes?: (ev: any) => boolean;

	protocols?: string | string[];
}

/**
 * Heartbeat
 */
export interface IWebSocketHeart {
	/**
	 * Heartbeat sending content, default is 'heartSend'
	 */
	heartSend: string;

	/**
	 * Heartbeat receiving content, default is 'heartGet'
	 */
	heartGet: string;

	/**
	 * Heartbeat sending interval time, default is 30000
	 */
	heartGapTime: number;

	/**
	 * Number of heartbeat failures
	 */
	failNum: number;

	/**
	 * Maximum number of heartbeat failures, default is 10
	 */
	heartFailNum: number;

	/**
	 * WebSocketBean object
	 */
	websocketbean: IWebSocketBean;

	/**
	 * Handle received heartbeat information
	 * @param ev
	 * @returns
	 */
	onmessage: (ev: any) => any;

	heartRes?: (ev: any) => boolean;
}

/**
 * Reconnection
 */
export interface IWebSocketReconnect {
	/**
	 * Status of whether reconnection is enabled
	 */
	status: boolean;
	/**
	 * WebSocketBean object
	 */
	websocketbean: IWebSocketBean;

	/**
	 * Current number of reconnection attempts
	 */
	num: number;

	/**
	 * Maximum number of reconnection attempts, default is 10
	 */
	reconnectMaxNum: number;

	/**
	 * Reconnection interval time
	 */
	reconnectGapTime: number;

	/**
	 * Start attempting reconnection
	 */
	start: () => void;

	/**
	 * Stop reconnection
	 */
	stop: () => void;
}

/**
 * Data sending management
 */
export interface IWebSocketSend {
	/**
	 * WebSocketBean object
	 */
	websocketbean: IWebSocketBean;

	/**
	 * Prefix for sending messages
	 */
	sendPrefix: string;

	/**
	 * Suffix for sending messages
	 */
	sendSuffix: string;

	/**
	 * Send data
	 * @param data Data object, Object, Array, or String
	 * @param resend
	 */
	send(data: any, resend?: boolean): string | boolean;

	/**
	 * Destroy the data that needs to be resent
	 * @param sendId
	 */
	offsend: (sendId: string) => void;

	/**
	 * Notify connection opened
	 */
	onopen: () => void;

	/**
	 * Clear all cached data
	 */
	clear: () => void;
}
