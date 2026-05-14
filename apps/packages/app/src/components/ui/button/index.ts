import BtButton from './src/index.vue';
import { App } from 'vue';

export default (app: App) => {
	app.component(BtButton.name!, BtButton);
};
