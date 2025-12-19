import hotkeys from 'hotkeys-js';
import {
	HotkeyOptions,
	mergeOptions,
	DEFAULT_SCOPE
} from 'src/directives/hotkeyOptions';

// HotkeyManager.bind({
// 	key: 'ctrl+s',
// 	handler: () => {
// 		console.log('[HotKey] saved！');
// 	},
// 	scope: 'editor'
// });

// HotkeyManager.setScope('editor');

class HotkeyManager {
	private static instance: HotkeyManager;
	private registeredKeys: Map<string, Set<string>>;

	private constructor() {
		this.registeredKeys = new Map();
	}

	public static getInstance(): HotkeyManager {
		if (!HotkeyManager.instance) {
			HotkeyManager.instance = new HotkeyManager();
		}
		return HotkeyManager.instance;
	}

	public bind(options: Partial<HotkeyOptions>): void {
		const mergedOptions = mergeOptions(options);

		if (!mergedOptions.key || !mergedOptions.handler) {
			console.error('[HotKey] Please provide a valid key and handler!');
			return;
		}

		if (
			this.isKeyBound(mergedOptions.key, mergedOptions.scope) &&
			!mergedOptions.single
		) {
			console.error(
				`[HotKey] Key "${mergedOptions.key}" with scope "${
					mergedOptions.scope || 'global'
				}" is already bound, skipping duplicate binding.`
			);
			return;
		}

		const { key, handler, scope, element, keyup, keydown, splitKey, capture } =
			mergedOptions;

		hotkeys(
			key,
			{ scope, element, keyup, keydown, splitKey, capture },
			(event) => {
				event.preventDefault();
				handler(event);
			}
		);

		this.addKeyToRegistry(key, scope);

		if (mergedOptions.single) {
			this.unbind(key, scope);
		}
	}

	private addKeyToRegistry(key: string, scope?: string): void {
		if (!this.registeredKeys.has(key)) {
			this.registeredKeys.set(key, new Set());
		}
		this.registeredKeys.get(key)!.add(scope || DEFAULT_SCOPE);
	}

	private removeKeyFromRegistry(key: string, scope?: string): void {
		if (!this.registeredKeys.has(key)) return;

		const scopes = this.registeredKeys.get(key)!;
		scopes.delete(scope || DEFAULT_SCOPE);

		if (scopes.size === 0) {
			this.registeredKeys.delete(key);
		}
	}

	private isKeyBound(key: string, scope?: string): boolean {
		if (!this.registeredKeys.has(key)) return false;

		const scopes = this.registeredKeys.get(key)!;
		return scopes.has(scope || DEFAULT_SCOPE);
	}

	public unbind(key: string, scope?: string): void {
		if (!this.registeredKeys.has(key)) {
			console.warn(`[HotKey] Key "${key}" is not bound, no need to unbind.`);
			return;
		}

		if (scope) {
			if (!this.isKeyBound(key, scope)) {
				console.warn(
					`[HotKey] Key "${key}" with scope "${scope}" is not bound, no need to unbind.`
				);
				return;
			}

			hotkeys.unbind(key, scope);
			this.removeKeyFromRegistry(key, scope);
		} else {
			const scopes = Array.from(this.registeredKeys.get(key)!);
			scopes.forEach((s) => hotkeys.unbind(key, s));
			this.registeredKeys.delete(key);
		}
	}

	registerHotkeys(
		configMap: Record<string, (event: KeyboardEvent) => void>,
		scopes = [DEFAULT_SCOPE]
	) {
		for (let i = 0; i < scopes.length; i++) {
			Object.entries(configMap).forEach(([key, handler]) => {
				this.bind({
					key,
					handler,
					scope: scopes[i]
				});
			});

			console.log(
				`[HotKey] Registered ${Object.keys(configMap).length} hotkeys in ${
					scopes[i]
				}`
			);
		}
	}

	unregisterHotkeys(
		configMap: Record<string, (event: KeyboardEvent) => void>,
		scopes = [DEFAULT_SCOPE]
	) {
		for (let i = 0; i < scopes.length; i++) {
			Object.entries(configMap).forEach(([key]) => {
				this.unbind(key, scopes[i]);
			});

			console.log(
				`[HotKey] Unregistered ${Object.keys(configMap).length} hotkeys in ${
					scopes[i]
				}`
			);
		}
	}

	public setScope(scope: string): void {
		hotkeys.setScope(scope);
		console.log(`[HotKey] Current scope set to "${scope}"`);
	}

	public getScope(): string {
		return hotkeys.getScope();
	}

	public deleteScope(scope: string, newScope?: string): void {
		this.registeredKeys.forEach((scopes, key) => {
			if (scopes.has(scope)) {
				hotkeys.unbind(key, scope);
				scopes.delete(scope);
			}
		});

		hotkeys.deleteScope(scope);
		console.log(`[HotKey] Scope "${scope}" has been deleted`);

		if (newScope) {
			this.setScope(newScope);
			console.log(`[HotKey] Switched to new scope "${newScope}"`);
		}
	}

	public clearAll(): void {
		hotkeys.unbind();
		this.registeredKeys.clear();
		hotkeys.setScope('all');
		console.log('[HotKey] All hotkeys cleared, scope reset to "all"');
	}

	public logAllKeyCodes(): void {
		console.log('[HotKey] self hotkeys logs');
		console.log(this.registeredKeys);
		console.log('[HotKey] npm hotkeys logs');
		console.log(hotkeys.getAllKeyCodes());
	}
}

export default HotkeyManager.getInstance();
