import { boot } from 'quasar/wrappers';
import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor';
import * as monaco from 'monaco-editor';
import { loader } from '@guolao/vue-monaco-editor';

export default boot(async ({ app }) => {
	app.use(VueMonacoEditorPlugin, {});

	loader.config({ monaco });
});
