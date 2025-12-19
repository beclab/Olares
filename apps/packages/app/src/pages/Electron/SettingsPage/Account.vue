<template>
	<div class="wrap_account">
		<q-scroll-area ref="scrollAreaRef" style="height: calc(100vh - 40px)">
			<div class="row items-center q-mt-sm q-mb-lg q-pt-sm">
				<q-icon
					class="q-ml-md q-mr-xs"
					name="sym_r_account_circle"
					size="24px"
				/>
				<q-toolbar-title class="q-pl-none text-body2 text-weight-bold">{{
					t('account')
				}}</q-toolbar-title>
				<q-avatar
					class="q-mr-lg cursor-pointer"
					icon="sym_r_close"
					@click="goBack"
				/>
			</div>

			<q-list class="settingList">
				<q-item>
					<q-item-section class="header">
						<div class="users">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="40" />
						</div>

						<div class="info">
							<div class="name terminus-text-ellipsis text-body1">
								{{ currentUser?.local_name }}
							</div>
							<div class="did terminus-text-ellipsis">
								@{{ userStore.current_user?.domain_name }}
							</div>
						</div>

						<div
							class="delete cursor-pointer text-body3"
							@click="deleteAccount"
						>
							{{ t('delete_account') }}
						</div>
					</q-item-section>
				</q-item>

				<q-item class="q-mt-md">
					<q-item-section>
						<q-item-label
							class="text-body1 text-weight-bold q-pl-sm q-ml-xs q-mb-md"
						>
							{{ t('mnemonic_backup') }}
						</q-item-label>
						<q-item-label class="text-ink-1 q-pl-sm q-ml-xs">
							<div class="q-mb-xs text-grey-7">{{ $t('mnemonic_phrase') }}</div>
							<DesktopMnemonics></DesktopMnemonics>
						</q-item-label>
					</q-item-section>
				</q-item>
			</q-list>
		</q-scroll-area>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import { app, setSenderUrl, clearSenderUrl } from '../../../globals';
import { useUserStore } from '../../../stores/user';
import { useMenuStore } from '../../../stores/menu';
import DialogDelete from './DialogDelete.vue';
import { useI18n } from 'vue-i18n';
import DesktopMnemonics from './DesktopMnemonics.vue';

export default defineComponent({
	name: 'AccountPage',
	components: { DesktopMnemonics },
	setup(_props: any, context: any) {
		const $q = useQuasar();
		const userStore = useUserStore();
		const menuStore = useMenuStore();
		const router = useRouter();

		const { t } = useI18n();

		const currentUser = userStore.users?.items.get(userStore.current_id || '');

		const goBack = (to: string) => {
			if (to === '-1') {
				context.emit('setMode', '1');
			} else {
				menuStore.popTerminusMenuCache();
				router.back();
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
				_deleteAccount();
			});
		};

		const _deleteAccount = async () => {
			userStore.userUpdating = true;

			if ($q.platform.is.electron) {
				window.electron.api.files.removeCurrentAccount({
					url: userStore.getModuleSever('files'),
					username: userStore.current_user?.local_name + '',
					authToken: ''
				});
			}

			await userStore.removeUser(userStore.current_id!);
			await app.lock();

			if (userStore.users?.items.size == 0) {
				router.push({
					name: 'welcome'
				});
			} else {
				await userStore.users!.unlock(userStore.password!);
				const user = userStore.current_user;
				const mneminicItem = userStore.current_mnemonic;
				if (user && mneminicItem) {
					$q.loading.show();
					if (user.setup_finished) {
						setSenderUrl({
							url: user.auth_url
						});
					} else {
						clearSenderUrl();
					}
					await app.load(user.id);
					await app.unlock(mneminicItem.mnemonic);
					$q.loading.hide();
				} else {
					console.error('user not found');
				}
				router.replace('/connectLoading');
			}
			userStore.userUpdating = false;
		};

		return {
			currentUser,
			goBack,
			deleteAccount,
			userStore,
			t
		};
	}
});
</script>

<style scoped lang="scss">
.wrap_account {
	width: 100%;
	height: 100%;
	background: $background-1;
	border-radius: 12px;
}

.settingList {
	width: 500px;

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

			.name {
				color: $ink-1;
			}

			.did {
				color: $ink-3;
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
</style>
