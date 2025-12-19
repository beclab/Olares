import { BaseWebsocketBean } from './base';
import { DownloadRecord } from 'src/utils/interface/rss';
import { TransferDatabase } from 'src/utils/interface/transferDB';
import { wiseInsertTransferItem, WiseWSType } from '../public/wise';
import { busEmit } from 'src/utils/bus';

export class LarePassWebsocketBean extends BaseWebsocketBean {
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

	otherTypeMethods(data: { type: string; data: any }) {
		if (data.type === 'addTaskHistory') {
			this.insertTransferItem(data.data);
			return true;
		}
		return false;
	}

	websocketOnMessage(event: MessageEvent): void {
		super.websocketOnMessage(event);
		try {
			const body: any = JSON.parse(event.data);

			if (body.type === WiseWSType.DOWNLOAD_PROCESS) {
				this.insertTransferItem({
					...body.data,
					progress: `${body.data.percent || 0}`,
					id: body.data.task_id
				});

				if (body.type === 'download_process') {
					busEmit('wiseDownloadProcess', body);
				}
			} else {
				// vault message
				busEmit('receiveMessage', body);
			}
		} catch (e) {
			console.error('message error:', e);
		}
	}

	private insertTransferItem(downloadItem: DownloadRecord) {
		wiseInsertTransferItem(downloadItem, this.db, this.terminusId, (item) => {
			busEmit('CloudTransferUpdate', {
				item,
				data: downloadItem
			});
		});
	}
}
