import { useUserStore } from '../stores/user';

export function current_user_bind_status(): BIND_STATUS {
	const user = useUserStore().current_user;
	if (!user?.name) {
		return BIND_STATUS.VC_NONE;
	}

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
