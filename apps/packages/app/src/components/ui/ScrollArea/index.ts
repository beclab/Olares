import { App } from 'vue';

import BtScrollArea from './src/index.vue';

export default (app: App) => {
	app.component(BtScrollArea.name!, BtScrollArea);
};
