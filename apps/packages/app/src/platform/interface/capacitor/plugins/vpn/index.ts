import { registerPlugin } from '@capacitor/core';
import { TailScalePluginInterface } from './definitions';

const TailscalePlugin =
	registerPlugin<TailScalePluginInterface>('TailScalePlugin');

export { TailscalePlugin };
