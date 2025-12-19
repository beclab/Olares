import { boot } from 'quasar/wrappers';
import BtUI from '@bytetrade/ui';
import { BtDialog, BtCustomDialog } from '@bytetrade/ui';
import { Dialog } from 'quasar';

export default boot(({ app }) => {
	BtDialog.init(Dialog);
	app.use(BtCustomDialog, {
		defaultOkClass: 'controlhub-global-ok-button'
	});
	app.use(BtUI);
});
