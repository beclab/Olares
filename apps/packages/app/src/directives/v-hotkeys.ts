import { Directive } from 'vue';
import HotkeyManager from './hotkeyManager';
import { HotkeyOptions, mergeOptions } from 'src/directives/hotkeyOptions';

// <q-btn
// v-hotkey="{ handler: save, key: 'ctrl+q' }"
// label="Ctrl+q"
// 	/>
//
// const save = () => {
// 	console.log('saved！');
// }

// hotkeys disabled in input, textarea and selects
// https://github.com/jaywcjlove/hotkeys-js/issues/321

const vHotkey: Directive<HTMLElement, Partial<HotkeyOptions>> = {
	mounted(el, binding) {
		console.log('[HotKey] mounted');

		const options = mergeOptions(binding.value, el);

		if (!options.key || !options.handler) {
			console.error('[HotKey] Please provide a valid key and handler!');
			return;
		}

		HotkeyManager.bind(options);

		(el as any)._hotkey = { key: options.key, scope: options.scope };
	},

	updated(el, binding) {
		const oldValue = binding.oldValue;
		const newValue = binding.value;

		if (oldValue?.key !== newValue.key || oldValue?.scope !== newValue.scope) {
			const { key, scope } = (el as any)._hotkey || {};
			HotkeyManager.unbind(key, scope);

			const options = mergeOptions(newValue, el);
			HotkeyManager.bind(options);

			(el as any)._hotkey = { key: options.key, scope: options.scope };
		}
	},

	beforeUnmount(el) {
		console.log('[HotKey] beforeUnmount');
		const { key, scope } = (el as any)._hotkey || {};

		HotkeyManager.unbind(key, scope);
	}
};

export default vHotkey;
