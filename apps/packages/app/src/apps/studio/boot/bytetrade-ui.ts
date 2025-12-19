import { boot } from 'quasar/wrappers';
import BytetradeUi, { BtNotify, BtDialog, BtCustomDialog } from '@bytetrade/ui';
import { Notify, Dialog } from 'quasar';

export default boot(({ app }) => {
	app.use(BytetradeUi);
	BtNotify.init(Notify);
	BtDialog.init(Dialog);
	app.use(BtCustomDialog, {
		defaultOkClass: 'my-global-ok-button'
	});
});

export { BytetradeUi };
