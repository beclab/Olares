import { IWebSocketHeart, IWebSocketBean } from './websocket';

/**
 * WebSocket Heartbeat Mechanism
 */
export class WebSocketHeart implements IWebSocketHeart {
	websocketbean: IWebSocketBean;
	heartSend: string;
	heartGet: string;
	heartGapTime: number;
	failNum = 0;
	heartFailNum: number;
	heartRes?: (ev: any) => boolean;

	constructor(websocketbean: IWebSocketBean) {
		this.websocketbean = websocketbean;
		this.heartSend = this.websocketbean.param.heartSend ?? 'heartSend';
		this.heartGet = this.websocketbean.param.heartGet ?? 'heartGet';
		this.heartGapTime = this.websocketbean.param.heartGapTime ?? 30000;
		this.heartFailNum = this.websocketbean.param.heartFailNum ?? 10;
		this.heartRes = this.websocketbean.param.heartRes;
	}

	timer: number = null as any;

	/**
	 * Start the heartbeat mechanism, sending heartbeat messages at regular intervals.
	 * Stops and triggers an error if no response is received after multiple attempts.
	 */
	start = () => {
		if (this.timer !== null) return;
		this.failNum = 0;
		this.timer = setInterval(() => {
			if (this.failNum >= this.heartFailNum) {
				this.stop();
				this.websocketbean.onerror();
				return;
			}
			this.websocketbean.send(this.heartSend);
			this.failNum++;
		}, this.heartGapTime) as any;
	};

	/**
	 * Stop the heartbeat mechanism.
	 */
	stop = () => {
		clearInterval(this.timer);
		this.timer = null as any;
	};

	/**
	 * Reset failure count on receiving a valid heartbeat response.
	 * @param ev - The received message.
	 */
	onmessage = (ev: any) => {
		if (this.heartRes) {
			if (this.heartRes(ev)) {
				this.failNum = 0;
			}
			return;
		}
		const messagePrefix = this.websocketbean.param.messagePrefix ?? '';
		const messageSuffix = this.websocketbean.param.messageSuffix ?? '';
		const heartGetMessage = messagePrefix + this.heartGet + messageSuffix;
		if (ev === heartGetMessage) this.failNum = 0;
	};
}
