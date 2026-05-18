import { IWebSocketBean, IWebSocketSend } from './websocket';
import { WebSocketStatusEnum } from './WebSocketEnum';

const isObject = (val: any): any => val !== null && typeof val === 'object';

/**
 * WebSocket data sending management
 */
export class WebSocketSend implements IWebSocketSend {
	websocketbean: IWebSocketBean;

	sendPrefix: string;
	sendSuffix: string;

	constructor(websocketbean: IWebSocketBean) {
		this.websocketbean = websocketbean;
		this.sendPrefix = this.websocketbean.param.sendPrefix ?? '';
		this.sendSuffix = this.websocketbean.param.sendSuffix ?? '';
	}

	/**
	 * Temporary sending management object
	 */
	sendTemp: { tag: string; data: any; resend: boolean; sendId?: string }[] = [];

	/**
	 * Resend ID
	 */
	sendId = 1000;

	/**
	 * Get resend ID
	 * @returns
	 */
	getSendId = () => {
		this.sendId++;
		return this.sendId + '';
	};

	/**
	 * Cache data identifier
	 */
	tag = '___senTemp';

	/**
	 * Resend data management
	 */
	sendMap: { [key: string]: any } = {};

	/**
	 * Send data
	 * @param data Data object, Object, Array, or String
	 * @param resend Whether the data needs to be resent
	 */
	send(data: any, resend = false) {
		if (this.websocketbean.status === WebSocketStatusEnum.open) {
			let sendId: string = null as any;

			// First, check if it is cached data to be sent. If so, retrieve the data and resend status.
			if (isObject(data)) {
				if (data.tag === this.tag) {
					resend = data.resend;
					// If resend is true, sendId must exist
					if (resend) sendId = data.sendId;
					data = data.data;
				}
			}

			// If the data needs to be resent, save it
			if (resend) {
				if (sendId === null) sendId = this.getSendId();
				this.sendMap[sendId] = data;
			}

			// Check if the data is an object or array, then convert it to a string
			if (isObject(data) || Array.isArray(data)) {
				data = JSON.stringify(data);
			}

			// Send data
			this.websocketbean.websocket.send(
				this.sendPrefix + data + this.sendSuffix
			);

			// If the data needs to be resent, return the sendId
			return resend ? sendId : true;
		} else {
			let sendId: string = null as any;

			if (isObject(data)) {
				// If it is cached data to be sent, do nothing
				if (data.tag === this.tag) return false;
			}

			// If not connected, store in temporary cache and send after connection
			const sendTempItem: any = {
				tag: this.tag,
				data,
				resend
			};

			if (resend) {
				sendId = this.getSendId();
				sendTempItem.sendId = sendId;
			}
			this.sendTemp.push(sendTempItem);
			return resend ? sendId : false;
		}
	}

	/**
	 * Destroy resend data information
	 * @param sendId
	 */
	offsend = (sendId: string) => {
		this.sendMap[sendId] = undefined;
		delete this.sendMap[sendId];
	};

	/**
	 * Notify when the connection is open
	 */
	onopen = () => {
		// Process resend data
		Object.keys(this.sendMap).forEach((key) => {
			if (this.sendMap[key] !== undefined) this.send(this.sendMap[key]);
		});

		// Process temporary data
		for (let i = this.sendTemp.length - 1; i >= 0; i--) {
			const item = this.sendTemp[i];
			const sendStatus = this.send(item);
			if (sendStatus !== false) {
				this.sendTemp.splice(i, 1);
			}
			console.log('=== resend ==>' + sendStatus);
			console.log(item);
			console.log('<=====');
		}
	};

	/**
	 * Clear all cached data
	 */
	clear = () => {
		this.sendMap = {};
		this.sendTemp = [];
	};
}
