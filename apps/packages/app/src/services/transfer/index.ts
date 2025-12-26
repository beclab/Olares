import { Platform } from 'quasar';
import {
	TransferActionInterface,
	TransferClient as TransferClientInterface,
	TransferClientService
} from '../abstractions/transfer/interface';
import { CommonTransfer } from './common';
import { ElectronTransfer } from './electron';
import { MobileTransfer } from './mobile';
import { getAppPlatform } from 'src/application/platform';
import { useUserStore } from 'src/stores/user';
import * as process from 'process';
import { WiseTransfer } from './wise';
import { useDeviceStore } from 'src/stores/device';
import {
	TransferFront,
	TransferItem,
	TransferItemInMemory
} from 'src/utils/interface/transfer';

class TransferClient implements TransferClientInterface {
	client: TransferClientService;

	waitAddSubtasks: Record<
		number,
		{
			finished: boolean;
			subtasks: TransferItem[];
			offset?: number;
		}
	> = {};

	async prepare(
		item: TransferItem,
		offset = 0
	): Promise<
		| undefined
		| {
				child?: TransferItem[];
				finished: boolean;
				callback?: (
					addStatus: boolean,
					ids: number[],
					identifys?: string[]
				) => void;
		  }
	> {
		if (!item.id || !this.waitAddSubtasks[item.id]) {
			return undefined;
		}

		const task = this.waitAddSubtasks[item.id];

		const callback = (
			addStatus: boolean,
			ids: number[],
			uniqueIdentifiers?: string[]
		) => {
			const _task = this.waitAddSubtasks[item.id!];

			let transfer = this.client.downloader;
			if (item.front == TransferFront.upload) {
				transfer = this.client.uploader;
			}
			if (addStatus && transfer['addSubtasksSuccess']) {
				transfer['addSubtasksSuccess'](
					offset,
					item.id!,
					ids,
					uniqueIdentifiers
				);
			}
			if (offset + ids.length == _task.subtasks.length && _task.finished) {
				delete this.waitAddSubtasks[item.id!];
			}
		};
		const nOffset = offset - (task.offset || 0);
		const items =
			task.subtasks.length > nOffset ? task.subtasks.slice(nOffset) : [];

		return {
			child: items,
			finished: task.finished,
			callback
		};
	}

	constructor() {
		if (process.env.APPLICATION === 'WISE') {
			this.client = new WiseTransfer();
		} else {
			if (Platform.is.electron) {
				this.client = new ElectronTransfer();
			} else if (Platform.is.nativeMobile) {
				this.client = new MobileTransfer();
			} else {
				this.client = new CommonTransfer();
			}
		}
		if (this.client.clouder) {
			this.client.clouder.init();
		}
	}
	async doAction(
		item: TransferItem,
		action: 'start' | 'cancel' | 'pause' | 'resume' | 'complete'
	): Promise<boolean> {
		let transfer: TransferActionInterface | undefined = undefined;
		if (item.front == TransferFront.upload) {
			transfer = this.client.uploader;
		}
		if (item.front == TransferFront.download) {
			transfer = this.client.downloader;
		}
		if (item.front == TransferFront.cloud && this.client.clouder) {
			transfer = this.client.clouder;
		}

		if ([TransferFront.copy, TransferFront.move].includes(item.front)) {
			transfer = this.client.copier;
		}

		if (transfer) {
			return await transfer[action](item);
		}
		return false;
	}

	async getTransferInfo(item: TransferItem) {
		let transfer = this.client.downloader;
		if (item.front == TransferFront.upload) {
			transfer = this.client.uploader;
		}
		return await transfer.getTransferInfo(item);
	}

	async restartEnable(item: TransferItem) {
		let transfer = this.client.downloader as TransferActionInterface;
		if (item.front == TransferFront.upload) {
			transfer = this.client.uploader;
		} else if (item.front == TransferFront.cloud) {
			if (!this.client.clouder) {
				return false;
			}
			transfer = this.client.clouder;
		}

		return await transfer.restartEnable(item);
	}

	async autoRetry(item: TransferItemInMemory) {
		const retryCount = item.retryCount || 0;

		if ([TransferFront.copy, TransferFront.move].includes(item.front)) {
			return false;
		}

		if (this.client.errorRetryNumber > retryCount) {
			if (process.env.APPLICATION !== 'FILES') {
				this.doAction(item, 'pause');
			}
			item.retryCount = retryCount + 1;
			setTimeout(async () => {
				const deviceStore = useDeviceStore();
				if (!deviceStore.transferEnable()) {
					return false;
				}
				if (!item.isPaused) this.doAction(item, 'resume');
			}, Math.pow(2, item.retryCount) * 1000);
			return true;
		}
		return false;
	}

	getUserId() {
		if (getAppPlatform().isClient) {
			const userStore = useUserStore();
			return userStore.current_id;
		}
		return undefined;
	}
}
export default new TransferClient();
