import { boot } from 'quasar/wrappers';
import VueMarkdown from 'vue-markdown-render';

export default boot(async ({ app }) => {
	app.use((app: any) => {
		app.component(VueMarkdown.name, VueMarkdown);
	}, []);
});
