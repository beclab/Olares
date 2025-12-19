import { boot } from 'quasar/wrappers';
import { WizardApplication } from 'src/application/wizard';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const wizard = new WizardApplication();
	setApplication(wizard);
});
