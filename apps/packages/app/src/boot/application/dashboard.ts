import { boot } from 'quasar/wrappers';
import { DashboardApplication } from 'src/application/dashboard';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const settings = new DashboardApplication();
	setApplication(settings);
});
