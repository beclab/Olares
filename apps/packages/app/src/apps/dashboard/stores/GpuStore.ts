import { defineStore } from 'pinia';
import { getGraphicsList } from '@apps/dashboard/src/network/gpu';
import {
	Graphics,
	SettingsGPUItem,
	TaskItem,
	TaskItemWithShareMode
} from '@apps/dashboard/src/types/gpu';
import { uniqBy } from 'lodash';

interface GpuStore {
	gpuList: Graphics[];
	taskList: TaskItemWithShareMode[];
	gpuListCache: Graphics[];
	taskListCache: TaskItem[];
	deviceIds: Array<{ label: string; value: string }>;
}
export const useGpuStore = defineStore('GpuStore', {
	state: (): GpuStore => ({
		gpuList: [],
		taskList: [],
		deviceIds: [],
		gpuListCache: [],
		taskListCache: []
	}),

	getters: {
		gpuSelectOptions(state) {
			return state.gpuList.map((item) => ({
				label: item.nodeName,
				value: item.nodeName
			}));
		},
		gpuNodeListOptions(state) {
			return uniqBy(state.gpuListCache, 'nodeName');
		},
		gpuTypeListOptions(state) {
			return uniqBy(state.gpuListCache, 'type');
		}
	},
	actions: {
		updateGpuList(data, init) {
			this.gpuList = data;
			if (init) {
				this.gpuListCache = data;
			}
		},
		updateTaskList(data, init) {
			this.taskList = data;
			if (init) {
				this.taskListCache = data;
			}
		},
		getDeviceIds() {
			const params = {
				filters: {}
			};

			getGraphicsList(params).then((res) => {
				this.deviceIds = res.data.list.map((item) => ({
					label: item.uuid,
					value: item.uuid
				}));
			});
		}
	}
});
