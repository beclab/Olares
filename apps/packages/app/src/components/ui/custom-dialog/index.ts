import { App } from 'vue';

import BtCustomDialog from './src/IndexPage.vue';

export default (app: App, options: BtCustomDialogOptions = {}) => {
	const { defaultOkClass = 'default-ok-button' } = options;
	app.provide('defaultOkClass', defaultOkClass);
	app.component(BtCustomDialog.name!, BtCustomDialog);
};

export interface BtCustomDialogOptions {
	defaultOkClass?: string;
}
