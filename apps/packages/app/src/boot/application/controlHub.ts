import { boot } from 'quasar/wrappers';
import { ControlHubApplication } from 'src/application/controlHub';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const settings = new ControlHubApplication();
	setApplication(settings);
});
