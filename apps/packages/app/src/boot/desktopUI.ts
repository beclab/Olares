import { boot } from 'quasar/wrappers';
import BytetradeUi, { BtCustomDialog, BtNotify } from '@bytetrade/ui';
import { Notify } from 'quasar';

export default boot(async ({ app }) => {
	app.use(BytetradeUi);
	BtNotify.init(Notify);
	app.use(BtCustomDialog, {
		defaultOkClass: 'market-global-ok-button'
	});
});
