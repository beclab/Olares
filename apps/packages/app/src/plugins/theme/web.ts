import { WebPlugin } from '@capacitor/core';
import { ThemePluginInterface } from './definitions';
import { ThemeDefinedMode } from '@bytetrade/ui';

import { getUserStorage } from 'src/application/platform/userStorage';

export class ThemePlugin extends WebPlugin implements ThemePluginInterface {
	systemIsDark(): Promise<{ dark: boolean }> {
		throw new Error('Method not implemented.');
	}
	async get(): Promise<{ theme: ThemeDefinedMode }> {
		const theme = await getUserStorage().getItem('theme');

		if (theme == undefined) {
			return {
				theme: ThemeDefinedMode.LIGHT
			};
		}
		return {
			theme
		};
	}
	async set(options: { theme: ThemeDefinedMode }) {
		await getUserStorage().setItem('theme', options.theme);
	}
}
