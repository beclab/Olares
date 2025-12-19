import { boot } from 'quasar/wrappers';
import vHotkey from 'src/directives/v-hotkeys';

// "async" is optional;
// more info on params: https://v2.quasar.dev/quasar-cli/boot-files
export default boot(async ({ app }) => {
	// something to do
	app.directive('hotkey', vHotkey);
});
