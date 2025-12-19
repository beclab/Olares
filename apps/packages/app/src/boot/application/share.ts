import { boot } from 'quasar/wrappers';
import { ShareApplication } from 'src/application/share';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const files = new ShareApplication();
	setApplication(files);
});
