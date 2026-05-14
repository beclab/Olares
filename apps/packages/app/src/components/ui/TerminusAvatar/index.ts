import { App } from 'vue';

import TerminusAvatar from './src/index.vue';

export default (app: App) => {
	app.component(TerminusAvatar.name!, TerminusAvatar);
};
