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
					{{ t('user_management_center') }}
				</div>
				<q-space />
				<q-btn style="display: none"> </q-btn>
			</div>

			<q-card-section class="q-px-md q-pt-none">
				<q-item class="account">
					<q-item-section class="header">
						<div class="users">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="40" />
						</div>

						<div class="info column">
							<div class="name text-body1">
								{{ userStore.current_user?.local_name }}
							</div>
							<div class="did">@{{ userStore.current_user?.domain_name }}</div>
						</div>

						<div
							class="delete cursor-pointer text-body3"
							@click="deleteAccount"
						>
							{{ t('delete_account') }}
						</div>
					</q-item-section>
				</q-item>
				<q-item class="q-mt-md accountCentent">
					<q-item-section class="q-pa-md">
						<div class="row items-center justify-between">
							<div class="text-ink-1 text-h6">
								{{
									!userStore.currentUserBackup
										? t('mnemonic_backup')
										: t('mnemonic_phrase')
								}}
							</div>
							<div
								class="row items-center justify-end"
								v-if="!userStore.currentUserBackup"
							>
								<q-icon
									name="sym_r_gpp_maybe"
									size="20px"
									class="text-negative"
								/>
								<div class="text-negative q-ml-xs">
									{{ t('not_backed_up') }}
								</div>
							</div>
						</div>
						<q-item-label class="q-mt-sm">
							<div class="q-mt-xs text-ink-3 text-subtitle3">
								{{ t('mnemonic_phrase') }}
							</div>
							<div
								class="q-mt-xs mnemonics_wrap column items-center justify-center"
							>
								<terminus-mnemonics-component
									:readonly="true"
									:show-title="false"
									:is-backup="false"
									:is-copy="true"
									:is-paste="false"
									:mnemonics="current_user?.mnemonic"
								/>
								<div class="mnemonics_login" v-if="encrypting">
									<q-icon
										name="sym_r_visibility_off"
										class="text-ink-on-brand"
										size="26px"
									/>
									<div class="q-mt-md q-ml-md q-mr-md content">
										{{ t('back_up_your_mnemonic_phrase_immediately_to_safe') }}
									</div>
									<TerminusEnterBtn
										@sure-action="startBackUp"
										class="q-mt-lg"
										:title="
											$q.platform.is.bex || userStore.currentUserBackup
												? 'Click to view'
												: 'Start Backup'
										"
									/>
								</div>
							</div>
						</q-item-label>
					</q-item-section>
				</q-item>
			</q-card-section>
		</q-card>
	</q-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
import TerminusMnemonicsComponent from '../../../components/common/TerminusMnemonicsComponent.vue';
import DialogLogin from '../../Electron/SettingsPage/DialogLogin.vue';
import TerminusEnterBtn from '../../../components/common/TerminusEnterBtn.vue';
import DialogDelete from '../../Electron/SettingsPage/DialogDelete.vue';
import { app, clearSenderUrl, setSenderUrl } from '../../../globals';

const { dialogRef, onDialogCancel } = useDialogPluginComponent();

const show = ref(true);

const router = useRouter();
const userStore = useUserStore();
const current_user = ref(userStore.current_user);
const { t } = useI18n();
const $q = useQuasar();

const encrypting = ref(true);

const startBackUp = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}

	if ($q.platform.is.bex || userStore.currentUserBackup) {
		$q.dialog({
			component: DialogLogin
		}).onOk(() => {
			encrypting.value = false;
		});
	} else {
		show.value = false;
		router.push({ path: '/backup_mnemonics' });
	}
};

const deleteAccount = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	$q.dialog({
		component: DialogDelete,
		componentProps: {
			title: t('delete_account'),
			message: t('delete_account_message'),
			navigation: t('cancel')
		}
	}).onOk(() => {
		show.value = false;
		_deleteAccount();
	});
};

const _deleteAccount = async () => {
	userStore.userUpdating = true;

	await userStore.removeUser(userStore.current_id!);
	await app.lock();

	if (userStore.users?.items.size == 0) {
		router.replace({
			name: 'setupSuccess'
		});
	} else {
		await userStore.users!.unlock(userStore.password!);
		const user = userStore.current_user;
		const mneminicItem = userStore.current_mnemonic;
		if (user) {
			if (user.setup_finished) {
				setSenderUrl({
					url: user.vault_url
				});
			} else {
				clearSenderUrl();
			}
			await app.load(user.id);
			if (mneminicItem) {
				await app.unlock(mneminicItem.mnemonic);
			}
		}
		if (userStore.current_user) {
			if (userStore.current_user.name) {
				router.replace('/connectLoading');
			} else {
				router.replace('/BindTerminusName');
			}
		}
	}
	userStore.userUpdating = false;
};
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 580px;
		border-radius: 12px;

		.header {
			height: 64px;
			width: 100%;
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

		.accountCentent {
			background: $background-2;
			border-radius: 12px;
			width: 100%;
			text-align: left;
		}
		.mnemonics_wrap {
			width: 100%;
			border-radius: 8px;
			position: relative;

			.mnemonics_login {
				width: 100%;
				height: 100%;
				position: absolute;
				top: 0;
				left: 0;

				backdrop-filter: blur(8px);
				background: $dimmed-background;
				border-radius: 8px;
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				padding: 0 20px;
				color: $ink-on-brand;

				.content {
					text-align: center;
				}
			}
		}
	}
}
</style>
