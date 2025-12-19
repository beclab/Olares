import { boot } from 'quasar/wrappers';
import { LoginApplication } from 'src/application/login';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const login = new LoginApplication();
	setApplication(login);
});
