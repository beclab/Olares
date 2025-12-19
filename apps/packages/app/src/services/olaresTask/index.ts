import {
	TransferFront,
	TransferItem,
	TransferStatus
} from 'src/utils/interface/transfer';
import {
	OlaresTaskBaseItem,
	OlaresTaskStatus
} from '../abstractions/olaresTask/interface';
import copyer from './copyer';
import cloudUpload from './cloudUpload';
import { CommonFetch } from 'src/api';
import { useTransfer2Store } from 'src/stores/transfer2';
import { useFilesStore } from 'src/stores/files';

class OlaresTaskManager {
	copyerTask = copyer;
	cloudUploadTask = cloudUpload;
	tasks: Record<string, OlaresTaskBaseItem> = {};
	maxRetries = 3;
	Taskmanager: any;

	async doAction(
		item: TransferItem,
		action: 'start' | 'cancel' | 'pause' | 'resume' | 'complete'
	) {
		if (item.phaseTaskId) {
			if (action === 'cancel') {
				const result = await this.cancelTask(item);
				if (result) {
					delete this.tasks[item.phaseTaskId];
				}
				return result;
			} else if (action == 'pause' || action == 'resume') {
				const result = await this.pauseOrResumeTask(item, action);
				return result;
			}
		}

		return true;
	}

	async initialize(node: string) {
		try {
			const res = await this.getTasks(node, [
				OlaresTaskStatus.PENDING,
				OlaresTaskStatus.RUNNING
			]);

			if (res && Array.isArray(res.tasks)) {
				const runningTaskIds = res.tasks.map((task) => task.id);

				for (let i = 0; i < runningTaskIds.length; i++) {
					const id = runningTaskIds[i];

					const transferStore = useTransfer2Store();
					if (
						res.tasks[i].task_type !== 'copy' &&
						res.tasks[i].task_type !== 'move'
					) {
						return;
					}
					if (transferStore.filesCopyTransferMap[id] === undefined) {
						await this.addTask(id, node, res.tasks[i].task_type, {
							type: TransferFront.copy
						});
						continue;
					}

					if (!this.tasks[id]) {
						this.tasks[id] = {
							...res.tasks[i],
							task_type: res.tasks[i].task_type,
							retryCount: 0,
							pending: false,
							transfer_id: transferStore.filesCopyTransferMap[id],
							node: node
						};
					}

					if (this.tasks[id]) {
						await this.updateTask(this.tasks[id], res.tasks[i]);
					}
				}
			}
		} catch (error) {
			console.error('Failed to fetch and filter running tasks:', error);
		}
	}

	async addTask(
		taskId: string,
		node: string,
		task_type: 'copy' | 'upload' | 'move',
		info: any
	) {
		const task = await this.getTask({ id: taskId, node });
		if (!task) {
			return;
		}

		if (!task.task) return;
		let formatTask: OlaresTaskBaseItem | undefined;
		if (task_type == 'copy' || task_type == 'move') {
			formatTask = await this.copyerTask.addCopyToTransfer(task.task, {
				...info,
				node
			});
		} else if (task_type == 'upload') {
			formatTask = await this.cloudUploadTask.configUploadTask(task.task, {
				...info,
				node
			});
		}

		if (formatTask === undefined) return false;
		if (!this.tasks[taskId]) {
			this.tasks[taskId] = {
				...formatTask,
				task_type,
				retryCount: 0
			};
		}

		return true;
	}

	async query() {
		const runningTasks = Object.values(this.tasks);
		for (const task of runningTasks) {
			try {
				if (task && task.pending) {
					continue;
				}
				task.pending = true;
				const res = await this.getTask(task);
				task.pending = false;

				if (res && res.task) {
					if (task.task_type == 'copy' || task.task_type == 'move') {
						const dealTask = await this.copyerTask.setQueryResult({
							...task,
							...res.task
						});
						await this.updateTask(dealTask, res.task);
					} else if (task.task_type == 'upload') {
						const dealTask = await this.cloudUploadTask.setQueryResult({
							...task,
							...res.task
						});
						await this.updateTask(dealTask, res.task);
					}
				} else {
					if (task.retryCount > this.maxRetries) {
						delete this.tasks[task.id];
						const transferStore = useTransfer2Store();
						if (task.transfer_id)
							transferStore.onFileError(
								task.transfer_id,
								task.task_type == 'upload'
									? TransferFront.upload
									: TransferFront.copy
							);
					} else {
						task.retryCount += 1;
						this.tasks[task.id] = task;
					}
				}
			} catch (taskError) {
				console.error(`Failed to refresh task ${task.id}:`, taskError);
			}
		}
	}

	async addTasks(tasks: OlaresTaskBaseItem[]) {
		tasks.forEach((e) => {
			this.tasks[e.id] = e;
		});
	}

	private async updateTask(task: OlaresTaskBaseItem, queryTask: any) {
		if (
			task.status === OlaresTaskStatus.COMPLETED ||
			task.status === OlaresTaskStatus.CANCELED ||
			task.status === OlaresTaskStatus.CANCELLED
		) {
			delete this.tasks[task.id];
		} else if (task.status === 'failed') {
			if (task.retryCount > this.maxRetries) {
				delete this.tasks[task.id];
			} else {
				task.retryCount += 1;
				this.tasks[task.id] = task;
			}
		} else {
			this.tasks[task.id] = task;
		}

		const transferStore = useTransfer2Store();

		if (
			(task.task_type === 'copy' || task.task_type === 'move') &&
			task.transfer_id &&
			!transferStore.filesInDialogMap[task.transfer_id]
		) {
			transferStore.filesInDialog.push(task.transfer_id);

			const transferItem = this.copyerTask.formatCopyToTransfer(
				{
					...queryTask,
					action:
						queryTask.task_type == 'copy'
							? TransferFront.copy
							: TransferFront.move
				},
				task.node,
				TransferStatus.Running,
				queryTask.dst_type
			);

			transferStore.filesInDialogMap[
				transferStore.filesCopyTransferMap[task.id]
			] = {
				...transferItem,
				isPaused: false,
				id: task.transfer_id,
				speed: 0,
				progress: queryTask.progress / 100,
				leftTime: 0,
				bytes: 0
			};
		}

		const runningTasks = Object.values(this.tasks);

		if (runningTasks.length > 0 && !transferStore.isUploadProgressDialogShow) {
			transferStore.isUploadProgressDialogShow = true;
		}
	}

	private async getTask(task: { id: string; node: string }): Promise<
		| {
				code: number;
				msg: string;
				task: any;
		  }
		| undefined
	> {
		const params: Record<string, string> = {};
		params.task_id = task.id;

		const queryParams = Object.entries(params)
			.map(([k, v]) => `${k}=${v}`)
			.join('&');

		try {
			console.log(
				`/api/task/${task.node}/${queryParams ? `?${queryParams}` : ''}`
			);

			const res: {
				code: number;
				msg: string;
				task: any;
			} = await CommonFetch.get(
				`/api/task/${task.node}/${queryParams ? `?${queryParams}` : ''}`,
				{}
			);
			return res;
		} catch (error) {
			console.error('Failed to fetch copy tasks:', error);
			// throw error;
			return undefined;
		}
	}

	private async getTasks(node: string, status: OlaresTaskStatus[]) {
		const params: Record<string, string> = {};
		if (status !== undefined && status.length !== 0)
			params.status = status.join(',');

		const queryParams = Object.entries(params)
			.map(([k, v]) => `${k}=${v}`)
			.join('&');

		try {
			const res: {
				tasks: OlaresTaskBaseItem[];
				code: number;
				msg: string;
			} = await CommonFetch.get(
				`/api/task/${node}/${queryParams ? `?${queryParams}` : ''}`,
				{}
			);
			res.tasks = res.tasks.map((e) => {
				return {
					...e,
					retryCount: 0,
					pending: false,
					node: node,
					task_type: (e as any).action
				};
			});
			return res;
		} catch (error) {
			return { tasks: [], code: 500, msg: 'Failed to fetch tasks' };
		}
	}
	async cancelTask(item: TransferItem) {
		try {
			let node = '';
			if (item.node) {
				node = item.node;
			}
			const url = `/api/task${node.length > 0 ? '/' + node : ''}/?task_id=${
				item.phaseTaskId
			}`;
			await CommonFetch.delete(url);
			return true;
		} catch (error) {
			console.error(`Failed to delete task ${item.phaseTaskId}:`, error);
			throw false;
		}
	}

	async cancelAllTask() {
		const filesStore = useFilesStore();
		if (filesStore.nodes.length == 0) {
			return false;
		}
		try {
			filesStore.nodes.forEach((node) => {
				const url = `/api/task${
					node.name.length > 0 ? '/' + node.name : ''
				}/?all=1`;
				CommonFetch.delete(url);
			});
			return true;
		} catch (error) {
			throw false;
		}
	}

	async pauseOrResumeTask(item: TransferItem, op: 'pause' | 'resume') {
		try {
			let node = '';
			if (item.node) {
				node = item.node;
			}

			const url = `/api/task${node.length > 0 ? '/' + node : ''}/`;
			const params = {
				task_id: item.phaseTaskId,
				op
			};

			await CommonFetch.post(url, undefined, {
				params
			});
			return true;
		} catch (error) {
			console.error(`Failed to delete task ${item.phaseTaskId}:`, error);
			throw false;
		}
	}
}

export default new OlaresTaskManager();
