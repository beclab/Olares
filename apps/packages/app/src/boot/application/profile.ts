import { boot } from 'quasar/wrappers';
import { setApplication } from '../../application/base';
import { profileApplication } from '../../application/profile';
export default boot(async () => {
	const profile = new profileApplication();
	setApplication(profile);
});
