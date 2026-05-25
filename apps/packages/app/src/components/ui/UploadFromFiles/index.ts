import { App } from 'vue';

import UploadFromFiles from './src/index.vue';

export default (app: App) => {
	app.component(UploadFromFiles.name!, UploadFromFiles);
};
