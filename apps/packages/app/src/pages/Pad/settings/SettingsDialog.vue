<template>
	<q-dialog
		class="d-creatVault text-center"
		v-model="show"
		persistent
		ref="dialogRef"
	>
		<q-card class="q-dialog-plugin bg-background-3">
			<div class="header row items-center q-px-md">
				<q-btn dense flat icon="close" @click="onDialogCancel" v-close-popup>
					<q-tooltip>{{ t('buttons.close') }}</q-tooltip>
				</q-btn>
				<q-space />
				<div class="text-h5" style="font-weight: 500">
					{{ t('settings.settings') }}
				</div>
				<q-space />
				<q-btn style="display: none"> </q-btn>
			</div>

			<q-scroll-area class="scroll-min-height">
				<q-card-section class="q-px-md q-pt-none">
					<TerminusSetTheme class="q-mt-md item-centent" />
					<q-item class="q-mt-lg item-centent">
						<q-item-section class="q-pa-md">
							<q-item-label class="text-body1 text-weight-bold">{{
								t('language')
							}}</q-item-label>
							<q-item-label>
								<bt-select
									class="q-mt-md"
									v-model="currentLanguage"
									:options="supportLanguages"
									:border="true"
									@update:modelValue="updateLocale"
								/>
							</q-item-label>
						</q-item-section>
					</q-item>
					<q-item class="q-mt-lg item-centent">
						<q-item-section class="q-pa-md">
							<q-item-label class="text-body1 text-weight-bold">{{
								t('account')
							}}</q-item-label>
							<q-item-label class="q-pt-md text-color-sub-title">
								{{ t('settings.account_root_message') }}
							</q-item-label>

							<q-item-label class="text-grey-8">
								<div class="adminBtn q-mt-md" @click="toAccountCenter">
									<q-icon
										name="sym_r_account_circle"
										size="24px"
										class="q-mr-xs"
									/>
									{{ t('settings.account_administration') }}
								</div>
							</q-item-label>
						</q-item-section>
					</q-item>

					<q-item class="q-mt-lg item-centent">
						<q-item-section class="q-pa-md">
							<q-item-label class="text-body1 text-weight-bold">{{
								t('safety')
							}}</q-item-label>
							<q-item-label class="q-pt-md text-color-sub-title">
								{{
									userStore.passwordReseted
										? t('change_local_password')
										: t(
												'This password is only used for unlocking {AppName} on this device',
												{
													AppName: 'LarePass'
												}
										  )
								}}
							</q-item-label>

							<q-item-label class="text-grey-8">
								<div class="adminBtn q-mt-md" @click="toChangePassword">
									{{
										userStore.passwordReseted
											? t('settings.changePassword')
											: t('Set up a password')
									}}
								</div>
							</q-item-label>
							<q-item-label class="q-pt-md text-grey-8">
								<div
									class="checkbox-content row items-center"
									@click="changeAutoLock(!settings.autoLock)"
								>
									<div
										class="checkbox-common row items-center justify-center"
										:class="
											settings.autoLock
												? 'checkbox-selected-green'
												: 'checkbox-unselect'
										"
									>
										<q-icon
											class="text-ink-on-brand"
											size="12px"
											v-if="settings.autoLock"
											name="sym_r_check"
										/>
									</div>
									<div class="text-body2 text-ink-2">
										{{ t('auto_lock_when_you_leave') }}
									</div>
								</div>
							</q-item-label>

							<q-slide-transition>
								<q-item-label
									v-if="settings.autoLock"
									class="text-grey-8 row items-center justify-between lock-slider"
									:class="!settings.autoLock ? 'hideSlider' : ''"
								>
									<span>10 {{ t('min') }}</span>
									<q-slider
										v-model="settings.lockTime"
										:min="10"
										:max="3 * 24 * 60"
										:step="5"
										label
										:label-value="formatMinutesTime(settings.lockTime)"
										color="yellow"
										style="flex: 1"
										class="q-mx-sm"
										label-text-color="color-title"
										@change="changeAutoLockDelay"
									/>
									<span>3 {{ t('time.days') }}</span>
								</q-item-label>
							</q-slide-transition>

							<q-item-label class="q-pt-md text-color-sub-title">
								{{ t('autolock.reminderTitle') }}
							</q-item-label>
						</q-item-section>
					</q-item>
					<q-item class="q-mt-md item-centent">
						<q-item-section class="q-pa-md">
							<q-item-label class="text-body1 text-weight-bold">{{
								t('about')
							}}</q-item-label>
							<q-item-label class="text-grey-8 q-pt-md">
								{{ t('current_version') }}: {{ settings.appVersion }}
							</q-item-label>
						</q-item-section>
					</q-item>
				</q-card-section>
			</q-scroll-area>
		</q-card>
	</q-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import UserManagentDialog from '../account/UserManagentDialog.vue';
import DialogResetPassword from '../../Electron/SettingsPage/DialogResetPassword.vue';
import { reactive } from 'vue';
import { app } from '../../../globals';
import { formatMinutesTime } from '../../../utils/utils';
import { getPlatform } from '@didvault/sdk/src/core';
import TerminusSetTheme from '../../../components/common/TerminusSetTheme.vue';
import { useUserStore } from '../../../stores/user';
import BtSelect from '../../../components/base/BtSelect.vue';
import { i18n } from '../../../boot/i18n';
import { SupportLanguageType, supportLanguages } from '../../../i18n';

const { dialogRef, onDialogCancel } = useDialogPluginComponent();

const show = ref(true);

const { t } = useI18n();
const $q = useQuasar();
const userStore = useUserStore();

const toAccountCenter = () => {
	show.value = false;
	$q.dialog({
		component: UserManagentDialog
	});
	return;
};

const toChangePassword = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}
	$q.dialog({
		component: DialogResetPassword,
		componentProps: {
			title: t('settings.changePassword'),
			navigation: t('cancel')
		}
	});
};

const settings = reactive({
	automatically: true,
	content: true,
	display: true,
	transmissionrKeep: true,
	transmissionDefaultLocation: true,
	autoLock: app.settings.autoLock,
	lockTime: app.settings.autoLockDelay,
	weakPassword: app.account?.settings.securityReport.weakPasswords,
	repeatPassword: app.account?.settings.securityReport.reusedPasswords,
	leakPassword: app.account?.settings.securityReport.compromisedPaswords,
	expiredPassword: app.account?.settings.securityReport.expiredItems,
	downloadLocation: '',
	appVersion: ''
});

const changeAutoLockDelay = (value: any) => {
	app.setSettings({ autoLockDelay: value });
};
const changeAutoLock = (value: any) => {
	settings.autoLock = value;
	app.setSettings({ autoLock: value });
};

const configVersion = async () => {
	if (!$q.platform.is.nativeMobile) {
		return;
	}
	settings.appVersion = (await getPlatform().getDeviceInfo()).appVersion;
};

const currentLanguage = ref(userStore.locale || i18n.global.locale.value);

const updateLocale = async (language: SupportLanguageType) => {
	if (language) {
		await userStore.updateLanguageLocale(language);
	}
};

configVersion();
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 580px;
		height: 680px;
		border-radius: 12px;

		.header {
			height: 64px;
			width: 100%;
		}
		.scroll-min-height {
			height: min(calc(100% - 64px), 616px);
		}

		.current-user {
			padding: 4px 8px;
			border-radius: 4px;
			text-align: center;
			border: 1px solid $blue-4;
			color: $blue-4;
		}

		.account {
			background: $background-2;
			border-radius: 12px;

			.header {
				display: flex;
				flex-direction: row;
				align-items: center;
				justify-content: space-between;

				.users {
					width: 40px;
					height: 40px;
					border-radius: 20px;
					overflow: hidden;
					margin-left: 10px;
				}

				.info {
					flex: 1;
					margin-left: 10px;
					margin-right: 20px;
					overflow: hidden;
					text-align: left;

					.name {
						color: $ink-1;
					}

					.did {
						color: $sub-title;
						word-break: break-all;
					}
				}

				.delete {
					border: 1px solid $blue-4;
					color: $blue-4;
					padding: 4px 8px;
					border-radius: 6px;
				}
			}
		}

		.item-centent {
			background: $background-2;
			border-radius: 12px;
			width: 100%;
			text-align: left;

			.checkbox-content {
				width: 100%;
				height: 30px;
				.checkbox-common {
					width: 16px;
					height: 16px;
					margin-right: 10px;
					border-radius: 4px;
				}

				.checkbox-unselect {
					border: 1px solid $separator-2;
				}

				.checkbox-selected-yellow {
					background: $yellow-default;
				}
				.checkbox-selected-green {
					background: $positive;
				}
			}
			.adminBtn {
				border: 1px solid $yellow;
				background-color: $yellow-1;
				display: inline-block;
				color: $sub-title;
				padding: 6px 12px;
				border-radius: 8px;
				cursor: pointer;

				&:hover {
					background-color: $yellow-3;
				}
			}

			.lock-slider {
				height: 60px;
				transition: height 0.5s;
				min-height: 0 !important;
				padding-top: 0px !important;
				padding-bottom: 0px !important;
			}

			.hideSlider {
				height: 0 !important;
			}
		}
	}
}
</style>
