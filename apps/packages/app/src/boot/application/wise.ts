import { boot } from 'quasar/wrappers';
import { WiseApplication } from 'src/application/wise';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const wise = new WiseApplication();
	setApplication(wise);
});
