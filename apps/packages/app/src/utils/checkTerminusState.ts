import { i18n } from '../boot/i18n';

export enum UserStatus {
	Intranet = 1,
	Internet,
	VPNConnecting,
	OfflineMode,
	Offline,
	RequiresVpn,
	TokenInvalid,
	VaultTokenInvalid,
	Reactivation
}

export const InOfflineText = () => {
	return i18n.global.t('in_offline_mode');
};
export const InOfflineModeCode = UserStatus.OfflineMode;

export enum UserStatusActive {
	normal,
	active,
	error
}

export interface UserStatusInfo {
	status: UserStatus;
	isError: UserStatusActive;
	icon: string;
	title: string;
	description?: string;
	descriptionEx?: string;
}

export enum ConnectActionType {
	Passed = 1,
	Connect = 2,
	Offline = 3,
	ReLogin = 4,
	ReActivate = 5,
	ReConnect = 6,
	OpenTailScale = 7,
	Activate = 8
}

export interface PingResult {
	pinged: boolean;
	passed: boolean;
}
