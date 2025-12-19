/* eslint-disable @typescript-eslint/no-empty-function */
/*
 * This file is part of KubeSphere Console.
 * Copyright (C) 2019 The KubeSphere Console Authors.
 *
 * KubeSphere Console is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * KubeSphere Console is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with KubeSphere Console.  If not, see <https://www.gnu.org/licenses/>.
 */

import { get } from 'lodash';
// import { getWebSocketProtocol } from 'utils/index.ts';

export const getWebSocketProtocol = (protocol) => {
	if (protocol.startsWith('https')) {
		return 'wss';
	}
	return 'ws';
};

const readyStates = ['connecting', 'open', 'closing', 'closed'];
const defaultOptions = {
	reopenLimit: 5,
	onopen() {},
	onmessage() {},
	onclose() {},
	onerror() {}
};

export default class SocketClient {
	static composeEndpoint = (socketUrl, suffix = '') => {
		const re = /(\w+?:\/\/)?([^\\?]+)/;
		const matchParts = String(socketUrl).match(re);
		return `${getWebSocketProtocol(window.location.protocol)}://${
			matchParts[2]
		}${suffix}`;
	};

	constructor(endpoint, options = {}) {
		this.endpoint = endpoint;
		this.options = Object.assign(defaultOptions, options);

		if (!this.endpoint) {
			throw Error(`invalid websocket endpoint: ${this.endpoint}`);
		}
		this.reopenCount = 0;
		this.setUp();
	}

	getSocketState(readyState) {
		if (readyState === undefined) {
			readyState = this.client.readyState;
		}

		return readyStates[readyState];
	}

	initClient() {
		const subProto = get(this.options, 'subProtocol');

		if (!this.client) {
			this.client = new WebSocket(this.endpoint, subProto);
		}

		if (this.client && this.client.readyState > 1) {
			this.client.close();
			this.client = new WebSocket(this.endpoint, subProto);
		}

		return this.client;
	}

	attachEvents() {
		const { onopen, onmessage, onclose, onerror } = this.options;

		this.client.onopen = (ev) => {
			// if socket will close, try to keep alive
			if (!this.immediately && this.reopenCount < this.options.reopenLimit) {
				this.timer = setTimeout(
					this.setUp.bind(this),
					1000 * 2 ** this.reopenCount
				);
				this.reopenCount++;
			}
			onopen && onopen(ev);
		};

		this.client.onmessage = (message) => {
			let { data } = message;
			if (typeof data === 'string') {
				try {
					data = JSON.parse(data);
				} catch (e) {
					//
				}
			}

			onmessage && onmessage(data);
		};

		this.client.onclose = (ev) => {
			onclose && onclose(ev);
		};

		this.client.onerror = (ev) => {
			console.error('socket error: ', ev);
			onerror && onerror(ev);
		};
	}

	send(data) {
		return this.client.send(data);
	}

	close(val) {
		val && (this.immediately = true);
		this.client.close();
		!!this.timer && clearTimeout(this.timer);
	}

	setUp() {
		this.initClient();
		this.attachEvents();
	}
}
