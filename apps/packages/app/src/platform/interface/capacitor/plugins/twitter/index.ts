import { registerPlugin } from '@capacitor/core';
import { TwitterLoginPlugin } from './definitions';

const TwitterLogin = registerPlugin<TwitterLoginPlugin>('TwitterLogin', {
	web: () => import('./web').then((m) => new m.TwitterLoginWeb())
});

export { TwitterLogin };
