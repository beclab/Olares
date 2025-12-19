import { boot } from 'quasar/wrappers';
import BytetradeUi, { BtNotify, BtDialog, BtCustomDialog } from '@bytetrade/ui';
import { Notify, Dialog } from 'quasar';

// "async" is optional;
// more info on params: https://v2.quasar.dev/quasar-cli/boot-files
export default boot(async ({ app }) => {
	// something to do
	BtNotify.init(Notify);
	BtDialog.init(Dialog);
	app.use(BytetradeUi);
	app.use(BtCustomDialog, {
		defaultOkClass: 'profile-global-ok-button'
	});
});
