import { defineStore } from 'pinia';

import { DriveType } from '../utils/interface/files';
import { TransferFront, TransferItem } from 'src/utils/interface/transfer';

import Taskmanager from '../services/olaresTask';

export type FileState = {
	// runningTasks: Record<string, TaskItem>;
	// runningTasksRefreshInterval: any;
};

export const useFilesCopyStore = defineStore('filesCopy', {
	state: (): FileState => ({}),
	getters: {},

	actions: {
		async initialize(node: string) {
			await Taskmanager.initialize(node);
		},

		async createTask({
			taskId,
			type,
			node,
			dst_drive_type,
			action,
			origin_id
		}: {
			taskId: string;
			type: TransferFront;
			node: string;
			dst_drive_type: DriveType;
			action: 'copy' | 'move';
			origin_id?: number;
		}) {
			Taskmanager.addTask(taskId, node, action, {
				type,
				dst_drive_type,
				origin_id
			});
		},

		async deleteCopyTasks(item: TransferItem): Promise<any> {
			return Taskmanager.doAction(item, 'cancel');
		}
	}
});
