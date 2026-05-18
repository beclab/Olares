import { App } from 'vue';

import BtMenu from './src/index.vue';

export default (app: App) => {
	app.component(BtMenu.name!, BtMenu);
};
