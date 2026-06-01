import {
	IWebSocketBean,
	IWebSocketBeanParam,
	IWebSocketReconnect,
	IWebSocketSend
} from './websocket';
import { WebSocketHeart } from './WebSocketHeart';
import { WebSocketReconnect } from './WebSocketReconnect';
import { WebSocketSend } from './WebSocketSend';
import { WebSocketStatusEnum } from './WebSocketEnum';
import { PlatformAdapter } from './platformAdapter';

/**
 * WebSocket encapsulation class
 * @description Encapsulates heartbeat mechanism and reconnection mechanism
 */
export class WebSocketBean implements IWebSocketBean {
	status: WebSocketStatusEnum = null as any;
	websocket: WebSocket = null as any;
	heart: WebSocketHeart = null as any;
	reconnect: IWebSocketReconnect = null as any;
	sendObj: IWebSocketSend = null as any;
	param: IWebSocketBeanParam;

	constructor(param: IWebSocketBeanParam) {
		this.param = param;
	}

	onopen = async (restart?: boolean) => {
		// Start heartbeat
		this.heart.start();

		// Notify connection success or reconnection success
		this.reconnect.stop();

		// Update status to connected
		this.status = WebSocketStatusEnum.open;

		// Call lifecycle method
		if (this.param.onopen) await this.param.onopen();

		if (restart && this.param.onReconnectSuccess) {
			await this.param.onReconnectSuccess();
		}

		// Notify to send data
		this.sendObj.onopen();
	};

	onmessage = (ev: MessageEvent<any>) => {
		// Call lifecycle method
		if (this.param.onmessage) this.param.onmessage(ev);

		this.heart.onmessage(ev.data);
	};

	onerror = () => {
		// Call lifecycle method
		if (this.param.onerror) this.param.onerror();
		// Destroy objects
		this.close();
		// Start reconnection
		this.reconnect.start();
	};

	start = (param?: IWebSocketBeanParam, restart = false) => {
		// Close the existing connection if already created
		this.close();

		// Use new configuration or existing configuration
		if (param) this.param = param;
		else param = this.param;

		// Create connection
		this.websocket = new WebSocket(param.url, param.protocols);

		// Update status to loading
		this.status = WebSocketStatusEnum.load;

		// Bind connection success event
		this.websocket.onopen = () => this.onopen(restart);
		// Bind message receiving event
		this.websocket.onmessage = this.onmessage;
		// Bind connection error event
		this.websocket.onerror = this.onerror;
		// Bind connection close event
		this.websocket.onclose = this.onerror;

		// Create heartbeat
		this.heart = new WebSocketHeart(this);

		// Create reconnection object if it doesn't already exist
		if (this.reconnect === null) this.reconnect = new WebSocketReconnect(this);

		// Create data sending management object if it doesn't already exist
		if (this.sendObj === null) this.sendObj = new WebSocketSend(this);

		// Listen for window close event, and actively close the WebSocket connection when the window is closed.
		// This prevents the server from throwing exceptions when the connection hasn't been closed properly.
		PlatformAdapter.addEventListener('beforeunload', this.dispose);
	};

	/**
	 * Send data
	 * @param data Data object, Object, Array, or String
	 * @param resend Whether to resend the data after reconnecting
	 */
	send(data: any, resend = false) {
		return this.sendObj?.send(data, resend);
	}

	/**
	 * Destroy the data that needs to be resent
	 * @param sendId
	 */
	offsend = (sendId: string) => {
		this.sendObj?.offsend(sendId);
	};

	/**
	 * Close WebSocket, destroy binding events, heartbeat events, window close events, and update status to closed
	 */
	close = () => {
		if (this.websocket === null) return;
		PlatformAdapter.removeEventListener('beforeunload', this.dispose);
		// Destroy binding events and close WebSocket
		if (this.websocket) {
			this.websocket.onerror = null;
			this.websocket.onmessage = null;
			this.websocket.onclose = null;
			this.websocket.onopen = null;
			this.websocket.close();
			this.websocket = null as any;
		}
		// Destroy heartbeat events
		if (this.heart) {
			this.heart.stop();
			this.heart = null as any;
		}

		// Update status to closed
		this.status = WebSocketStatusEnum.close;
	};

	/**
	 * Destroy all objects
	 */
	dispose = () => {
		this.close();
		if (this.reconnect) {
			this.reconnect.stop();
			this.reconnect = null as any;
		}
		if (this.sendObj) {
			this.sendObj.clear();
			this.sendObj = null as any;
		}
	};
}
