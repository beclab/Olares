import { boot } from 'quasar/wrappers';
import { SettingsApplication } from 'src/application/settings';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const settings = new SettingsApplication();
	setApplication(settings);
});
