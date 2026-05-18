import { boot } from 'quasar/wrappers';
import BaseComponents from '../components/base';
import BytetradeUi, {
	BtNotify,
	BtDialog,
	BtCustomDialog,
	TerminusAvatar
} from '@bytetrade/ui';
import VueLazyload from 'vue-lazyload';
import { Notify, Dialog } from 'quasar';

export default boot(async ({ app }) => {
	// something to do
	app.use(BytetradeUi);
	app.use(VueLazyload);
	app.use(BaseComponents);
	app.use(BtCustomDialog, {
		defaultOkClass: 'files-global-ok-button'
	});
	app.provide('defaultCacheAvatar', true);
	BtNotify.init(Notify);
	BtDialog.init(Dialog);
});
