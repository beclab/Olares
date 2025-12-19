import { boot } from 'quasar/wrappers';
import { DesktopApplication } from 'src/application/desktop';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const desktop = new DesktopApplication();
	setApplication(desktop);
});
