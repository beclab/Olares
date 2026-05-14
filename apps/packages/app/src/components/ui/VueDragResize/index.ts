import VueDragResize from './src/vue-drag-resize.vue';
import { App } from 'vue';

export default (app: App) => {
	app.component('VueDragResize', VueDragResize);
};
