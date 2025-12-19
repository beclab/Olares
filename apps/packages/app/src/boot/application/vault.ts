import { boot } from 'quasar/wrappers';
import { VaultApplication } from 'src/application/vault';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const vault = new VaultApplication();
	setApplication(vault);
});
