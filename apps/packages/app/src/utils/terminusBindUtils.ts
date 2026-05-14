import { useUserStore } from '../stores/user';
import { VaultType, VaultItem } from '@didvault/sdk/src/core';
import { app } from '../globals';

export function getVaultsByType(type: VaultType): VaultItem[] {
	const res: VaultItem[] = [];

	for (const vault of app.state.vaults) {
		for (const item of vault.items) {
			if (item.type === type) {
				res.push(item);
			}
		}
	}
	return res;
}

export function hasVaultByType(type: VaultType): boolean {
	for (const vault of app.state.vaults) {
		for (const item of vault.items) {
			if (item.type === type) {
				return true;
			}
		}
	}
	return false;
}

export function current_user_bind_status(): BIND_STATUS {
	const store = useUserStore();

	if (!store.users || !store.users.items || store.users.items.size == 0) {
		//warning status wrong
		return BIND_STATUS.VC_NONE;
	}

	const user = store.users!.items.get(store.current_id!)!;

	const hasVc = hasVaultByType(VaultType.VC);
	if (!hasVc && !user.name) {
		return BIND_STATUS.VC_NONE;
	}

	if (!user.name) {
		return BIND_STATUS.TERMINUS_NAME_NONE;
	}

	// if (!user.url) {
	// 	return BIND_STATUS.OS_NONE;
	// }

	if (!user.binding && !user.setup_finished) {
		return BIND_STATUS.OS_WAIT_DNS_OK;
	}

	return BIND_STATUS.BIND_OK;
}

export enum BIND_STATUS {
	VC_NONE,
	TERMINUS_NAME_NONE,
	OS_NONE,
	OS_WAIT_DNS_OK,
	BIND_OK
}
