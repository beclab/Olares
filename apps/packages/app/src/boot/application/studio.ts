import { boot } from 'quasar/wrappers';
import { StudioApplication } from 'src/application/studio';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const settings = new StudioApplication();
	setApplication(settings);
});
