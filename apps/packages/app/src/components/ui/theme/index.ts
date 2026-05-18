import { App } from 'vue';

import BtTheme from './src/IndexPage.vue';

export default (app: App) => {
	app.component(BtTheme.name!, BtTheme);
};
