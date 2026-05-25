import { App } from 'vue';

import BtSwitch from './src/index.vue';

export default (app: App) => {
	app.component(BtSwitch.name!, BtSwitch);
};
