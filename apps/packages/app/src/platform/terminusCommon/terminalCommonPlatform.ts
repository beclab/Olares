/* eslint-disable @typescript-eslint/no-unused-vars */
import { SubAppPlatform } from '../subAppPlatform';
import { useUserStore } from 'src/stores/user';
import {
	AuthType,
	cloneItemTemplates,
	DeviceInfo,
	ErrorCode,
	Field,
	FieldType,
	ItemTemplate,
	SSHPassworResetType,
	UserItem,
	VaultType
} from '@didvault/sdk/src/core';

import { app } from 'src/globals';
import { RouteLocationNormalizedLoaded, Router } from 'vue-router';
import {
	homeMounted as commonHomeMounted,
	homeUnMounted as commonHomeUnMounted
} from '../homeLayoutCommon';

import { QVueGlobals } from 'quasar';
import UserStatusCommonDialog from '../../components/userStatusDialog/UserStatusCommonDialog.vue';
import TerminusTipDialog from '../../components/dialog/TerminusTipDialog.vue';
import TerminusDesktopTipDialog from '../../components/dialog/TerminusDesktopTipDialog.vue';
import TerminusBexTipDialog from '../../components/dialog/TerminusBexTipDialog.vue';
import { busOn, busEmit, busOff, NetworkErrorMode } from '../../utils/bus';
import { useDeviceStore } from '../../stores/device';
import { useScaleStore } from '../../stores/scale';
import { useMenuStore } from '../../stores/menu';
import { date } from 'quasar';
import { MessageTopic, TermiPassDeviceInfo } from '@bytetrade/core';
import { MessageBody } from '@bytetrade/core';
import SignMessageDialog from '../../layouts/dialog/SignMessageDialog.vue';
import { i18n } from '../../boot/i18n';
import { useTermipassStore } from 'src/stores/termipass';
import { TermiPassStatus } from 'src/utils/termipassState';
import DialogResetPassword from '../../pages/Electron/SettingsPage/DialogResetPassword.vue';
import AccessViaBrowserDialog from '../../components/wizard/AccessViaBrowserDialog.vue';
import { supportLanguages, languagesShort } from 'src/i18n';
import { displayAppServices } from '../../utils/platform';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { watch } from 'vue';
import { useFilesStore } from 'src/stores/files';
import { useFilesCopyStore } from 'src/stores/files-copy';

interface UploadDeviceInfoDevice {
	time: number;
	uploading: boolean;
	info: TermiPassDeviceInfo | undefined;
}

export class TerminusCommonPlatform extends SubAppPlatform {
	route: RouteLocationNormalizedLoaded | undefined;
	router: Router | undefined;
	quasar: QVueGlobals | undefined;
	userLastUploadDeviceRecord: Record<string, UploadDeviceInfoDevice> = {};

	signMessagesList: MessageBody[] = [];
	currentMessage: MessageBody | undefined = undefined;
	dealMssagesListTimer: NodeJS.Timer | undefined;

	sshPasswordResetShowMap: Record<string, boolean> = {};

	isClient = true;

	getQuasar() {
		return this.quasar;
	}

	async appLoadPrepare(data: any): Promise<void> {
		super.appLoadPrepare(data);

		if (data.route) {
			this.route = data.route;
		}
		if (data.router) {
			this.router = data.router;
		}

		if (data.quasar) {
			this.quasar = data.quasar;
		}

		busOn('account_update', async () => {
			const userStore = useUserStore();
			if (userStore.current_user?.setup_finished) {
				userStore.resetCurrentUserData();
				setTimeout(() => {
					busEmit('device_update');
				}, 5000);
				const abilitiesStore = useAppAbilitiesStore();
				abilitiesStore.init();
			}
		});

		busOn('device_update', async () => {
			const userStore = useUserStore();
			const deviceStore = useDeviceStore();
			if (!userStore.current_id) {
				return;
			}
			if (!this.userLastUploadDeviceRecord[userStore.current_id]) {
				this.userLastUploadDeviceRecord[userStore.current_id] = {
					time: 0,
					uploading: false,
					info: undefined
				};
			}
			const termipassInfo = await deviceStore.getTermiPassInfo();
			const oldTermipassInfo =
				this.userLastUploadDeviceRecord[userStore.current_id].info;
			if (
				oldTermipassInfo &&
				deviceStore.compareTerminuPassInfo(oldTermipassInfo, termipassInfo)
			) {
				this.userLastUploadDeviceRecord[userStore.current_id].time =
					new Date().getTime();
				return;
			}
			if (this.userLastUploadDeviceRecord[userStore.current_id].uploading) {
				return;
			}

			this.userLastUploadDeviceRecord[userStore.current_id].uploading = true;
			const status = await userStore.updateDeviceInfo(termipassInfo);
			this.userLastUploadDeviceRecord[userStore.current_id].uploading = false;
			if (status) {
				this.userLastUploadDeviceRecord[userStore.current_id].time =
					new Date().getTime();
				this.userLastUploadDeviceRecord[userStore.current_id].info =
					termipassInfo;
			}
		});

		busOn('signMessage', (body: MessageBody) => {
			this.signMessagesList.push(body);
			this.autoDealSignMessagesList();
		});

		busOn('cancel_sign', (messageId?: string) => {
			if (!messageId) {
				return;
			}
			const index = this.signMessagesList.findIndex(
				(e) => e.message?.id == messageId
			);
			if (index >= 0) {
				this.signMessagesList.splice(index, 1);
			}
			if (this.currentMessage?.message?.id == messageId) {
				this.currentMessage == undefined;
			}
		});

		busOn('receiveMessage', (message: any | string) => {
			const body: any =
				typeof message == 'string' ? JSON.parse(message) : message;
			if (body.topic == MessageTopic.Data) {
				if (
					body.event == 'vault.account.update' ||
					body.event == 'vault.org.update'
				) {
					if (!app.state.locked) {
						app.synchronize();
					}
				}
			} else if (
				body.eventType == 'vault.account.update' ||
				body.eventType == 'vault.org.update'
			) {
				if (!app.state.locked) {
					app.synchronize();
				}
			}
		});

		busOn('configPassword', () => {
			this.quasar
				?.dialog({
					component: this.quasar?.platform.is.mobile
						? UserStatusCommonDialog
						: process.env.APPLICATION_SUB_IS_BEX
						? TerminusBexTipDialog
						: TerminusDesktopTipDialog,
					componentProps: {
						title: i18n.global.t('Security Tips'),
						message: i18n.global.t(
							'Set up a local security password for LarePass to prevent unauthorized access and secure your private data.'
						),
						addSkip: true,
						btnTitle: i18n.global.t('Go to Set Password'),
						confirmBtnTitle: i18n.global.t('Go to Set Password'),
						skipTitle: i18n.global.t('skip'),
						showCancel: true
					}
				})
				.onOk(() => {
					if (
						!this.isPad &&
						(this.quasar?.platform.is.mobile || process.env.IS_BEX)
					) {
						this.router?.push({
							path: '/change_pwd'
						});
					} else {
						this.quasar?.dialog({
							component: DialogResetPassword,
							componentProps: {
								title: i18n.global.t('Set up a password'),
								navigation: i18n.global.t('cancel')
							}
						});
					}
				});
		});

		busOn('terminus_actived', () => {
			busEmit('account_update', true);
			const userStore = useUserStore();
			if (userStore.connected && userStore.current_user?.localMachine) {
				this.quasar
					?.dialog({
						component: AccessViaBrowserDialog
					})
					.onDismiss(() => {
						const userStore = useUserStore();
						if (userStore.current_user?.localMachine) {
							userStore.current_user.localMachine = '';
							userStore.save();
						}
					});
			}
		});

		busOn('receiveMessage', (body) => {
			const data = body.data;
			if (body.eventType === 'appStateUpdate') {
				const app = {
					[data.app]: data.data
				};
				const appAbilitiesStore = useAppAbilitiesStore();
				appAbilitiesStore.data = { ...appAbilitiesStore.data, ...app };
				appAbilitiesStore.busEmitAbilityUpdate();
			}
		});

		const userStore = useUserStore();
		if (userStore.locale) {
			i18n.global.locale.value = userStore.locale;
		} else {
			const locale = navigator.language;
			console.log('locale ====>', locale);
			if (languagesShort[locale]) {
				i18n.global.locale.value = languagesShort[locale] as any;
			} else if (supportLanguages.find((e) => e.value == locale) != undefined) {
				i18n.global.locale.value = locale as any;
			}
		}

		const filesStore = useFilesStore();
		watch(
			() => filesStore.nodes,
			(newName) => {
				if (newName && newName.length > 0) {
					const filesCopyStore = useFilesCopyStore();
					filesStore.nodes.forEach((node) => {
						filesCopyStore.initialize(node.name);
					});
				}
			},
			{ immediate: true }
		);
	}

	async appRedirectUrl(
		redirect: any,
		_currentRoute: RouteLocationNormalizedLoaded
	) {
		const userStore = useUserStore();
		await userStore.load();
		if (!userStore.isBooted) {
			if (this.quasar?.platform.is.android && !displayAppServices()) {
				redirect({ path: '/declaration' });
			} else {
				redirect({ path: '/welcome' });
			}
			return;
		}

		if (!userStore.current_user) {
			redirect({ path: '/setup/success' });
			return;
		}

		if (userStore.current_user.name) {
			redirect({ path: '/connectLoading' });
		} else {
			redirect({ path: '/BindTerminusName' });
		}
		//
	}

	stateUpdate() {
		if (app.state._errors.length > 0) {
			const INVALID_SESSION = app.state._errors.find(
				(e) => e.code == ErrorCode.INVALID_SESSION
			)
				? true
				: false;
			const SERVER_ERROR = app.state._errors.find(
				(e) => e.code == ErrorCode.SERVER_ERROR
			)
				? true
				: false;

			if (SERVER_ERROR) {
				busEmit('network_error', {
					type: NetworkErrorMode.vault,
					error: ErrorCode.SERVER_ERROR
				});
			} else if (INVALID_SESSION) {
				busEmit('network_error', {
					type: NetworkErrorMode.vault,
					error: ErrorCode.INVALID_SESSION
				});
			} else {
				busEmit('network_error', {
					type: NetworkErrorMode.vault,
					error: app.state._errors[0].code
				});
			}
			app.state._errors = [];
		}

		const menuStore = useMenuStore();
		menuStore.updateMenuInfo();
	}

	async homeMounted(): Promise<void> {
		busOn('appSubscribe', this.stateUpdate);
		commonHomeMounted();

		const menuStore = useMenuStore();
		menuStore.syncInfo = {
			syncing: app.state.syncing || false,
			lastSyncTime: date.formatDate(app.state.lastSync, 'HH:mm:ss')
		};

		const termipassState = useTermipassStore();
		termipassState.state.publicActions.resetCheckEnable(true);
	}

	async setSSHPasswordReset() {
		const userStore = useUserStore();
		const new_user: UserItem = userStore.users!.items.get(
			userStore.current_id!
		)!;
		new_user.owner_ssh_is_default = SSHPassworResetType.RESETED;
		userStore.users!.items.update(new_user);
		await userStore.save();
	}

	async homeUnMounted(): Promise<void> {
		commonHomeUnMounted();
		busOff('appSubscribe', this.stateUpdate);
		const termipassState = useTermipassStore();
		termipassState.state.publicActions.resetCheckEnable(false);
	}

	readonly platformAuthType = AuthType.PublicKey;

	supportsPlatformAuthenticator() {
		return this.biometricKeyStore.isSupported();
	}

	async getDeviceInfo(): Promise<DeviceInfo> {
		const userStore = useUserStore();
		const info = await super.getDeviceInfo();
		if (userStore.locale) {
			info.locale = userStore.locale.split('-')[0];
		} else {
			info.locale = i18n.global.locale.value.split('-')[0];
		}
		return info;
	}

	async userStatusUpdateAction() {
		const termipassStore = useTermipassStore();

		if (termipassStore.totalStatus?.status == TermiPassStatus.Offline) {
			this.quasar?.dialog({
				component: this.quasar?.platform.is.mobile
					? UserStatusCommonDialog
					: TerminusDesktopTipDialog,
				componentProps: {
					title: i18n.global.t('tips'),
					message: i18n.global.t('user_current_status.networkOnLine.message'),
					btnTitle: i18n.global.t('i_got_it'),

					confirmBtnTitle: i18n.global.t('i_got_it')
				}
			});
			return;
		}

		if (termipassStore.totalStatus?.status == TermiPassStatus.Reactivation) {
			this.quasar?.dialog({
				component: this.quasar?.platform.is.mobile
					? UserStatusCommonDialog
					: TerminusDesktopTipDialog,
				componentProps: {
					title: i18n.global.t('tips'),
					message: i18n.global.t(
						this.quasar?.platform.is.mobile
							? 'user_current_status.reactivation.mobile_message'
							: 'user_current_status.reactivation.message'
					),
					isReactive: this.quasar?.platform.is.mobile,
					btnTitle: i18n.global.t('ignore_it'),
					confirmBtnTitle: i18n.global.t('i_got_it')
				}
			});
			return;
		}

		if (termipassStore.totalStatus?.status == TermiPassStatus.TokenInvalid) {
			this.quasar
				?.dialog({
					component: this.quasar?.platform.is.mobile
						? UserStatusCommonDialog
						: TerminusDesktopTipDialog,
					componentProps: {
						title: i18n.global.t('tips'),
						message: i18n.global.t('user_current_status.token_invalid.message'),
						btnTitle: i18n.global.t('i_got_it'),
						confirmBtnTitle: i18n.global.t('i_got_it')
					}
				})
				.onOk(() => {
					this.router?.push({
						path: '/ConnectTerminus'
					});
				});
			return;
		}

		if (
			termipassStore.totalStatus?.status == TermiPassStatus.VaultTokenInvalid
		) {
			this.quasar
				?.dialog({
					component: this.quasar?.platform.is.mobile
						? UserStatusCommonDialog
						: TerminusDesktopTipDialog,
					componentProps: {
						title: i18n.global.t('tips'),
						message: i18n.global.t('user_current_status.srp_invalid.message'),
						btnTitle: i18n.global.t('i_got_it'),
						confirmBtnTitle: i18n.global.t('i_got_it')
					}
				})
				.onOk(() => {
					this.router?.push({
						path: '/ConnectTerminus'
					});
				});
			return;
		}

		if (termipassStore.totalStatus?.status == TermiPassStatus.RequiresVpn) {
			const scaleStore = useScaleStore();
			this.quasar
				?.dialog({
					component: this.quasar?.platform.is.mobile
						? TerminusTipDialog
						: TerminusDesktopTipDialog,
					componentProps: {
						title: i18n.global.t(
							'user_current_status.requires_vpn.dialog_title'
						),
						message: i18n.global.t('user_current_status.requires_vpn.message')
					}
				})
				.onOk(async () => {
					scaleStore.start();
				});
			return;
		}
	}

	private async autoDealSignMessagesList() {
		if (this.signMessagesList.length == 0 && this.dealMssagesListTimer) {
			clearInterval(this.dealMssagesListTimer);
			this.dealMssagesListTimer = undefined;
			this.currentMessage = undefined;
			return;
		}

		if (!this.dealMssagesListTimer) {
			this.dealMssagesListTimer = setInterval(async () => {
				if (this.signMessagesList.length == 0) {
					this.autoDealSignMessagesList();
					return;
				}

				this.dealSignMessage();
			}, 1000);
		}
	}
	private async dealSignMessage() {
		if (this.currentMessage) {
			return;
		}
		const userStore = useUserStore();

		const message = this.signMessagesList[0];
		if (
			message.terminusName &&
			message.terminusName != userStore.current_user?.name
		) {
			notifyFailed(
				i18n.global.t(
					"The account you're logged into on this app doesn't match the one on the page. Please switch to the correct account."
				)
			);
			this.signMessagesList.splice(0, 1);
			return;
		}

		if (
			message &&
			message.message &&
			message.message.data &&
			message.message.data.did &&
			message.message.data.did != userStore.current_id
		) {
			this.signMessagesList.splice(0, 1);
			notifyFailed(
				i18n.global.t(
					"The account you're logged into on this app doesn't match the one on the page. Please switch to the correct account."
				)
			);
			return;
		}

		this.currentMessage = message;
		if (!(await userStore.unlockFirst())) {
			this.currentMessage = undefined;
			this.signMessagesList.splice(0, 1);
			return;
		}

		const index = this.signMessagesList.findIndex(
			(e) => e == this.currentMessage
		);

		this.quasar
			?.dialog({
				component: SignMessageDialog,
				componentProps: {
					body: this.currentMessage
				}
			})
			.onDismiss(() => {
				if (index >= 0) {
					this.signMessagesList.splice(index, 1);
				}
				this.currentMessage = undefined;
			});
	}
}
