import { PodItem } from '@apps/control-panel-common/src/network/network';
import { defineStore } from 'pinia';
import { getNodesList } from '@apps/control-hub/src/network';

const CURRENT_NODE_KEY = 'current_node';
const CURRENT_NODE_VALUE = localStorage.getItem(CURRENT_NODE_KEY) || '';

interface TerminalStore {
	data: PodItem[];
	terminal_init: boolean;
	terminal_active: boolean;
	current_node: string;
}
export const useTerminalStore = defineStore('terminalStore', {
	state: (): TerminalStore => ({
		data: [],
		terminal_init: false,
		terminal_active: false,
		current_node: CURRENT_NODE_VALUE
	}),
	getters: {
		terminal_show: (state) => {
			return state.terminal_init && state.terminal_active;
		},
		node_options: (state) => {
			return state.data.map((item) => item.metadata.name);
		}
	},
	actions: {
		async init() {
			const params = {
				limit: -1
			};
			return getNodesList(params).then((res) => {
				this.data = res.data.items;
				const isIncludesNodeList = this.data.find(
					(item) => item.metadata.name === CURRENT_NODE_VALUE
				);
				if (!isIncludesNodeList) {
					this.current_node = '';
					localStorage.removeItem(CURRENT_NODE_KEY);
				}
			});
		},
		async terminalPanelStatus(status: boolean) {
			this.terminal_active = status;
		},
		async terminalPanelInit() {
			this.terminal_init = true;
		},
		setCurrentNode(value) {
			this.current_node = value;
			localStorage.setItem(CURRENT_NODE_KEY, value);
		}
	}
});
