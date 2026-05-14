import { BaseWebsocketBean } from '../applications/base';
import { DownloadRecord } from 'src/utils/interface/rss';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import { wiseInsertTransferItem, WiseWSType } from '../public/wise';

export class WiseWebsocketBean extends BaseWebsocketBean {
	private terminusId = '';
	private db = new TransferDatabase();

	initWebSocket(
		data: {
			terminusId?: string;
			url: string;
			login: boolean;
			loginData: any;
			pongRes: (data: any) => boolean;
		},
		statusUpdate: () => void
	): void {
		this.terminusId = data.terminusId || '';
		super.initWebSocket(data, statusUpdate);
	}

	otherTypeMethods(data: {
		type: string;
		data: any;
		_port?: MessagePort;
	}): boolean {
		switch (data.type) {
			case 'addTaskHistory':
				this.insertTransferItem(data.data);
				return true;

			default:
				return false;
		}
	}

	websocketOnMessage(event: MessageEvent): void {
		super.websocketOnMessage(event);
		const { type: messageType, data: messageData } = JSON.parse(event.data);
		if (messageType === WiseWSType.DOWNLOAD_PROCESS) {
			this.insertTransferItem({
				...messageData,
				progress: `${messageData.percent || 0}`,
				id: messageData.task_id
			});
		} else if (messageType === WiseWSType.ENCLOSURE) {
			this.connections.forEach((port) =>
				port.postMessage({
					type: 'message',
					data: {
						type: WiseWSType.ENCLOSURE,
						data: messageData
					}
				})
			);
		} else if (messageType === WiseWSType.SYNC) {
			this.connections.forEach((port) =>
				port.postMessage({
					type: 'message',
					data: {
						type: WiseWSType.SYNC,
						data: messageData
					}
				})
			);
		}
	}

	private insertTransferItem(downloadItem: DownloadRecord) {
		wiseInsertTransferItem(downloadItem, this.db, this.terminusId, (item) => {
			this.connections.forEach((port) =>
				port.postMessage({
					type: 'message',
					data: {
						type: 'transfer',
						item,
						data: downloadItem
					}
				})
			);
		});
	}
}
