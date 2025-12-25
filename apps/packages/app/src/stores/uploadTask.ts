import { defineStore } from 'pinia';
import { TransferItem } from 'src/utils/interface/transfer';

export type FileState = {
	runningTasks: Record<number, TransferItem>;
};

export const useUploadTask = defineStore('uploadTask', {
	state: (): FileState => ({
		runningTasks: {}
	}),
	getters: {},

	actions: {}
});
