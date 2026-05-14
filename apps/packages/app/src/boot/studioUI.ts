import { boot } from 'quasar/wrappers';
import BtUI from '@bytetrade/ui';
import BytetradeUi, { BtNotify, BtDialog, BtCustomDialog } from '@bytetrade/ui';
import { Notify, Dialog } from 'quasar';

export default boot(({ app }) => {
	BtNotify.init(Notify);
	BtDialog.init(Dialog);
	app.use(BytetradeUi);
	app.use(BtCustomDialog, {
		defaultOkClass: 'my-global-ok-button'
	});
	app.use(BtUI);
});
