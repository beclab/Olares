import { defineStore } from 'pinia';
import { ControllerType } from '../utils/controllerContextUtil';

export const useBexStore = defineStore('controller', {
	state: () => ({
		controller: null as any as ControllerType
	}),

	actions: {
		setController(controllerType: ControllerType) {
			this.controller = controllerType;
		}
	}
});
