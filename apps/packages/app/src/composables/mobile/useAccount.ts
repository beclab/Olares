import { Platform } from 'quasar';
import { ref, watch } from 'vue';

export interface StateMessage {
	type: 'STATE_CHANGE';
	key: 'stateA';
	value: boolean;
}

export const useStateService = () => {
	const stateA = ref(false);

	if (Platform.is.bex) {
		chrome.runtime.onMessage.addListener((message: StateMessage) => {
			if (message.type === 'STATE_CHANGE' && message.key === 'stateA') {
				stateA.value = message.value;
			}
		});
	}

	const updateStateA = (value: boolean) => {
		stateA.value = value;
		if (Platform.is.bex) {
			chrome.runtime.sendMessage({
				type: 'STATE_CHANGE',
				key: 'stateA',
				value
			});
		}
	};

	return {
		stateA,
		updateStateA
	};
};
