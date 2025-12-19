import hotkeys from 'hotkeys-js';

/**
 * https://github.com/Mitscherlich/vue-use-hotkeys
 */

// HotKeys understands the following modifiers: ⇧, shift, option, ⌥, alt, ctrl, control, command, and ⌘.

// The following special keys can be used for shortcuts: backspace, tab, clear, enter, return, esc, escape, space, up, down, left, right, home, end, pageup, pagedown, del, delete, f1 through f19, num_0 through num_9, num_multiply, num_add, num_enter, num_subtract, num_decimal, num_divide.

// ⌘ Command() ⌃ Control ⌥ Option(alt) ⇧ Shift ⇪ Caps Lock(Capital) fn Does not support fn ↩︎ return/Enter space

// Define HotkeyOptions with default values
export interface HotkeyOptions {
	key: string; // Key combination, e.g., 'ctrl+s'
	handler: (event: KeyboardEvent) => void; // Callback function when the hotkey is triggered
	scope?: string; // Scope, defaults to current scope
	element?: HTMLElement; // Element to bind the event to, defaults to global
	keyup?: boolean; // Trigger on key release, default is false
	keydown?: boolean; // Trigger on key press, default is true
	splitKey?: string; // Separator for key combinations, default is '+'
	capture?: boolean; // Trigger during capture phase, default is false
	single?: boolean; // Allow only one callback, default is false
}

export const DEFAULT_SCOPE = 'all';

// Function to merge options with defaults
export function mergeOptions(
	options: Partial<HotkeyOptions>,
	defaultElement?: HTMLElement
): HotkeyOptions {
	if (!options.key) {
		throw new Error('[HotKey] "key" is required in HotkeyOptions.');
	}
	if (!options.handler) {
		throw new Error('[HotKey] "handler" is required in HotkeyOptions.');
	}

	return {
		key: options.key,
		handler: options.handler,
		scope: options.scope || DEFAULT_SCOPE,
		element: options.element || defaultElement,
		keyup: options.keyup || false,
		keydown: options.keydown || true,
		splitKey: options.splitKey || '+',
		capture: options.capture || false,
		single: options.single || false
	};
}
