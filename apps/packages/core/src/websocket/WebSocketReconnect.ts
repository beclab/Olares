import { IWebSocketReconnect, IWebSocketBean } from './websocket';

/**
 * WebSocket reconnection mechanism and data resending mechanism
 */
export class WebSocketReconnect implements IWebSocketReconnect {
	/**
	 * Connection status
	 */
	status: boolean;

	/**
	 * WebSocketBean object
	 */
	websocketbean: IWebSocketBean;

	/**
	 * Current reconnection attempts
	 */
	num = 0;

	/**
	 * Maximum reconnection attempts
	 */
	reconnectMaxNum = 10;

	/**
	 * Reconnection interval time
	 */
	reconnectGapTime = 30000;

	/**
	 * Initial gradient reconnection interval time
	 */
	gradientReconnectStartTime = 1000;

	/**
	 * Maximum gradient reconnection interval time
	 */
	gradientReconnectMaxTime = 60000;

	/**
	 * Whether gradient reconnection is enabled
	 */
	isReconnectGradientEnabled = true;

	constructor(websocketBean: IWebSocketBean) {
		this.websocketbean = websocketBean;
		this.status = websocketBean.param.needReconnect ?? false;
		this.reconnectMaxNum = this.websocketbean.param.reconnectMaxNum ?? 10;
		this.reconnectGapTime = this.websocketbean.param.reconnectGapTime ?? 30000;
		this.gradientReconnectStartTime =
			this.websocketbean.param.gradientReconnectStartTime ?? 1000;
		this.gradientReconnectMaxTime =
			this.websocketbean.param.gradientReconnectMaxTime ?? 60000;
		this.isReconnectGradientEnabled =
			this.websocketbean.param.isReconnectGradientEnabled ?? true;
	}

	timer: number = null as any;

	/**
	 * Start reconnection attempts
	 */
	start = () => {
		if (!this.status || this.reconnectMaxNum === 0) return;
		if (this.timer !== null) return;
		this.num = 0;
		if (this.websocketbean.param.onreconnect) {
			this.websocketbean.param.onreconnect();
		}
		this.scheduleReconnect();
	};

	/**
	 * Stop reconnection attempts
	 */
	stop = () => {
		if (!this.status) return;
		clearTimeout(this.timer);
		this.timer = null as any;
	};

	/**
	 * Perform reconnection
	 */
	tryReconnect = () => {
		this.websocketbean.start(undefined, true);
		this.num++;
	};

	/**
	 * Perform gradient reconnection
	 */
	scheduleReconnect = () => {
		if (this.reconnectMaxNum > 0 && this.num >= this.reconnectMaxNum) {
			if (this.websocketbean.param.onReconnectFailure) {
				this.websocketbean.param.onReconnectFailure();
			}
			this.stop();
			return;
		}

		let interval = this.reconnectGapTime;
		if (this.isReconnectGradientEnabled) {
			interval = Math.min(
				this.gradientReconnectStartTime * Math.pow(2, this.num),
				this.gradientReconnectMaxTime
			);
		}

		console.log(
			`Attempt ${this.num + 1} to reconnect, interval: ${interval} ms`
		);
		console.log(
			`[core.ws.reconnect] schedule reconnect: attempt=${
				this.num + 1
			}, interval=${interval}, max=${this.reconnectMaxNum}`
		);

		this.timer = setTimeout(() => {
			this.timer = null as any;
			console.log(
				`[core.ws.reconnect] timer fired: next attempt=${this.num + 1}`
			);
			this.scheduleReconnect();
		}, interval) as any;

		try {
			console.log(
				`[core.ws.reconnect] tryReconnect now: attempt=${this.num + 1}`
			);
			this.tryReconnect();
		} catch (error) {
			console.error('websocket reconnect attempt failed unexpectedly', error);
		}
	};
}
