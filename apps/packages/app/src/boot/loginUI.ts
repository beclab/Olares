import { boot } from 'quasar/wrappers';
import BytetradeUi, { BtNotify } from '@bytetrade/ui';
import { Notify } from 'quasar';

export default boot(async ({ app }) => {
	app.use(BytetradeUi);
	BtNotify.init(Notify);
});
