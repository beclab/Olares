import { registerPlugin } from '@capacitor/core';
import { DropboxAuthPluginInterface } from './definitions';

const DropboxAuth = registerPlugin<DropboxAuthPluginInterface>(
	'DropboxAuthPlugin',
	{
		web: () => import('./web').then((m) => new m.DropboxWeb())
	}
);

export { DropboxAuth };
