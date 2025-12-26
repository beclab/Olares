import { ThemeDefinedMode } from '@bytetrade/ui';
export interface ThemePluginInterface {
	get(): Promise<{
		theme: ThemeDefinedMode;
	}>;
	set(options: { theme: ThemeDefinedMode }): Promise<void>;

	//only android
	systemIsDark(): Promise<{
		dark: boolean;
	}>;
}
