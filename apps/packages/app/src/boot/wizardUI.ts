import { boot } from 'quasar/wrappers';
import BytetradeUi from '@bytetrade/ui';

export default boot(async ({ app }) => {
	app.use(BytetradeUi);
});
