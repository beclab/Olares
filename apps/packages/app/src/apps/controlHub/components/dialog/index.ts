import { h } from 'vue';
import { Dialog } from 'quasar';
import { QDialogOptions } from 'quasar/dist/types';

import IndexDialog from './src/IndexDialog.vue';
interface DialogPropsType extends QDialogOptions {
	component?: any;
	platform?: 'web' | 'mobile';
	icon?: string;
	titAlign?: string;
	okText?: string;
	okLoading?: boolean;
	okStyle?: any;
	prompt?: any;
}

const BtDialog = {
	currentDialog: <any>null,
	dialog: Dialog,

	init(dialog?: Dialog): void {
		if (!dialog) {
			return;
		}
		this.dialog = dialog;
	},

	async show(props: DialogPropsType) {
		if (!this.dialog) {
			this.init(Dialog);
		}
		let renderContent: any = null;
		renderContent = () => {
			return h(
				IndexDialog,
				{
					platform: props.platform,
					title: props.title,
					icon: props.icon,
					titAlign: props.titAlign,
					ok: props.ok,
					cancel: props.cancel,
					okLoading: props.okLoading,
					okStyle: props.okStyle,
					prompt: props.prompt
				},
				{
					default: () =>
						props.message
							? h('div', props.message)
							: h('div', h(props.component))
				}
			);
		};

		if (this.currentDialog) {
			this.currentDialog = null;
		}
		return new Promise((resolve, reject) => {
			this.currentDialog = this.dialog
				.create({
					...props,
					component: renderContent
				})
				.onOk((res) => {
					resolve(res);
				})
				.onCancel(() => {
					resolve(false);
				});
		});
	}
};

export default BtDialog;
