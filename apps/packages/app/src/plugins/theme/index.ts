import { registerPlugin } from '@capacitor/core';
import { ThemePluginInterface } from './definitions';

const Theme = registerPlugin<ThemePluginInterface>('ThemePlugin', {
	web: () => import('./web').then((m) => new m.ThemePlugin())
});

export { Theme as ThemePlugin };
