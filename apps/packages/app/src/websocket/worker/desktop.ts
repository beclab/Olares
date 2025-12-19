import { MessageTopic } from '@bytetrade/core';
import { BaseWebsocketBean } from '../applications/base';
import { desktopInsertNotificationItem } from '../public/desktop';
import { NotificationItem } from 'src/utils/desktop/notification';
import { NotificationDatabase } from 'src/utils/desktop/notificationDB';

export class DesktopWebsocketBean extends BaseWebsocketBean {
	reconnectMaxNum = -1;
	reconnectGapTime = 3000;
	heartFailNum = 1;

	private db = new NotificationDatabase();

	websocketOnMessage(ev: MessageEvent): void {
		super.websocketOnMessage(ev);
		const message = JSON.parse(ev.data);

		if (message.topic == MessageTopic.Notification) {
			const time = new Date().getTime();

			const item: NotificationItem = {
				appName: message.message?.appName ? message.message.appName : undefined,
				createTime: time,
				updateTime: time,
				childrens: [
					{
						title: message.notification.title,
						body: message.notification.body,
						event: message.event,
						createTime: time
					}
				]
			};
			desktopInsertNotificationItem(item, this.db, (item: NotificationItem) => {
				this.connections.forEach((port) =>
					port.postMessage({
						type: 'message',
						data: {
							type: 'notification',
							data: item
						}
					})
				);
			});
			return;
		}
		this.connections.forEach((port) =>
			port.postMessage({
				type: 'message',
				data: {
					type: 'ws',
					data: ev.data
				}
			})
		);
	}

	async onReconnectSuccess() {
		super.onReconnectSuccess();
		this.connections.forEach((port) => {
			port.postMessage({
				type: 'message',
				data: {
					type: 'reconnected'
				}
			});
		});
	}
}
