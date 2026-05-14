import { App } from 'vue';
import { createApp } from 'vue';

import BtLogoLoading from './src/index.vue';

export default (app: App) => {
	app.component(BtLogoLoading.name!, BtLogoLoading);
};
