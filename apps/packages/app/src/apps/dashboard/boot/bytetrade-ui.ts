import { boot } from 'quasar/wrappers';
import BtUI from '@bytetrade/ui';

export default boot(({ app }) => {
	app.use(BtUI);
});
