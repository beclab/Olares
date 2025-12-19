import { getPlatform } from '@didvault/sdk/src/core';
import { TerminusCommonPlatform } from '../../platform/terminusCommon/terminalCommonPlatform';
import { useTermipassStore } from '../../stores/termipass';
import { UserStatusActive } from '../../utils/checkTerminusState';

export function useTerminusUserStatus(emit: any) {
	const termipassStore = useTermipassStore();
	const itemClick = () => {
		if (termipassStore.totalStatus?.isError != UserStatusActive.error) {
			if (emit) emit('superAction');
			return;
		}

		const platform = getPlatform() as TerminusCommonPlatform;
		platform.userStatusUpdateAction();
	};

	const configIconClass = (active: UserStatusActive) => {
		if (active == UserStatusActive.error) {
			return 'red';
		}
		if (active == UserStatusActive.normal) {
			return 'grey';
		}
		return 'green';
	};

	const configTitleClass = (active: UserStatusActive) => {
		if (active == UserStatusActive.error) {
			return 'status-error';
		}
		if (active == UserStatusActive.normal) {
			return 'status-normal';
		}
		return 'status-active';
	};

	return {
		itemClick,
		configIconClass,
		configTitleClass,
		termipassStore,
		UserStatusActive
	};
}
