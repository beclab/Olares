import { App } from 'vue';

import BtLoading from './src/index.vue';

export default (app: App) => {
	app.component(BtLoading.name!, BtLoading);
};
