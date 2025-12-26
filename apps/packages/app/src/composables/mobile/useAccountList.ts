import { useRoute, useRouter } from 'vue-router';
import { useUserStore } from 'src/stores/user';
import { UserItem } from '@didvault/sdk/src/core';
import { app, clearSenderUrl, resetAPP, setSenderUrl } from 'src/globals';
import { computed, ref } from 'vue';
import { getDID } from 'src/did/did-key';
import { getUiType } from 'src/utils/utils';
import { useBexStore } from 'src/stores/bex';
import { useApproval } from 'src/pages/Mobile/wallet/approval';
import { sendUnlock } from 'src/utils/bexFront';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useLarepassWebsocketManagerStore } from 'src/stores/larepassWebsocketManager';
import TerminusTipDialog from 'src/components/dialog/TerminusTipDialog.vue';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { getNativeAppPlatform } from 'src/application/platform';

export const useAccountList = () => {
	const $router = useRouter();
	const route = useRoute();
	const userStore = useUserStore();
	const bexStore = useBexStore();
	const approvalUserIdRef = ref('');
	const $q = useQuasar();
	const { t } = useI18n();

	if (process.env.IS_BEX) {
		bexStore.controller.getApproval().then((approval) => {
			if (approval) {
				approvalUserIdRef.value = approval.data.params.requestDidKey as string;
			}
		});
	}

	const choose = async (id: string) => {
		if (selectIds.value !== null) {
			return;
		}
		if (id == userStore.current_id!) {
			if ($q.platform.is.nativeMobile) {
				getNativeAppPlatform().hookBackAction();
				return;
			}
			$router.back();
			return;
		}
		const user: UserItem = userStore.users!.items.get(id)!;
		userStore.userUpdating = true;

		await app.lock(false);
		await userStore.setCurrentID(user.id);

		if (user.setup_finished) {
			setSenderUrl({
				url: user.vault_url
			});
		} else {
			clearSenderUrl();
		}

		resetAPP();

		await app.load(user.id);
		const mnemonicItem = userStore.current_mnemonic;
		if (mnemonicItem) {
			await app.unlock(mnemonicItem.mnemonic);
		}

		const UIType = getUiType(route);
		if (UIType.isNotification) {
			const { resolveApproval } = useApproval($router);
			sendUnlock();
			if (mnemonicItem) {
				const selectedDidKey = await getDID(mnemonicItem.mnemonic);
				resolveApproval({ selectedDidKey });
			}
			return;
		}

		if (process.env.IS_BEX && userStore.current_id) {
			await bexStore.controller.changeAccount(userStore.current_id);
		}

		if (userStore.current_user) {
			if (userStore.current_user.name) {
				$router.replace('/connectLoading');
			} else {
				$router.replace('/BindTerminusName');
			}
		}
		userStore.userUpdating = false;
	};

	const addAccount = () => {
		$router.push({
			name: 'setupSuccess'
		});
	};

	const terminusSelect = ref();
	const intoCheckedMode = () => {
		if (terminusSelect.value) {
			terminusSelect.value.intoCheckedMode();
		}
	};

	const selectIds = ref<null | string[]>(null);
	const showSelectMode = (value: string[] | null) => {
		selectIds.value = value;
	};

	const totalUsersIds = computed(() => {
		const items: any[] = [];
		if (!userStore.users?.items) {
			return [];
		}
		for (const item of userStore.users.items) {
			items.push({
				id: item.id,
				selectedEnable: (id: string) => {
					return userStore.userIsBackup(id);
				}
			});
		}
		return items;
	});

	const handleSelectAll = () => {
		if (terminusSelect.value) {
			terminusSelect.value.toggleSelectAll();
		}
	};

	const handleClose = () => {
		if (terminusSelect.value) {
			terminusSelect.value.handleClose();
		}
	};

	const handleRemove = () => {
		removeAccounts();
	};

	const removeAccounts = async () => {
		if (selectIds.value == null || selectIds.value.length == 0) {
			return;
		}
		if (!(await userStore.unlockFirst())) {
			return;
		}
		$q.dialog({
			component: TerminusTipDialog,
			componentProps: {
				title: t('delete_account'),
				navigation: t('cancel'),
				position: t('delete')
			}
		}).onOk(async () => {
			const ids = selectIds.value;
			const currentId = userStore.current_id;
			await userStore.removeUsers(ids!);

			terminusSelect.value.handleClose();
			selectIds.value = null;

			if (userStore.current_id && userStore.current_id == currentId) {
				return;
			}

			await app.lock();

			const socketStore = useLarepassWebsocketManagerStore();
			socketStore.dispose();

			if (userStore.users?.items.size == 0) {
				$router.replace({
					name: 'setupSuccess'
				});
				if (process.env.IS_BEX) {
					await bexStore.controller.changeAccount('');
				}
			} else {
				await userStore.users!.unlock(userStore.password!);
				const user = userStore.current_user;
				if (user) {
					if (user.setup_finished) {
						setSenderUrl({
							url: user.vault_url
						});
					} else {
						clearSenderUrl();
					}
					await app.load(user.id);
					if (userStore.current_mnemonic?.mnemonic) {
						await app.unlock(userStore.current_mnemonic.mnemonic);
					}
				}
				socketStore.restart();
				if (process.env.IS_BEX && userStore.current_id) {
					await bexStore.controller.changeAccount(userStore.current_id);
				}
				if (userStore.current_user) {
					if (userStore.current_user.name) {
						$router.replace('/connectLoading');
					} else {
						$router.replace('/BindTerminusName');
					}
				}
			}
			userStore.userUpdating = false;
		});
	};

	const itemOnUnableSelect = (id: string) => {
		const user = userStore.users?.items.get(id);
		if (!user) {
			return;
		}
		notifyFailed(
			t(
				'Cannot delete the account. You must back up your mnemonic phrase first to proceed.'
			)
		);
	};

	return {
		approvalUserIdRef,
		selectIds,
		totalUsersIds,
		terminusSelect,
		choose,
		addAccount,
		intoCheckedMode,
		showSelectMode,
		handleSelectAll,
		handleClose,
		handleRemove,
		itemOnUnableSelect,
		t
	};
};
