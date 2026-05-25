import BtUploader from './src/index.vue';
import { App } from 'vue';

export default (app: App) => {
	app.component(BtUploader.name, BtUploader);
};
