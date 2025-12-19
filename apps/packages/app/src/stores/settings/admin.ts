import { defineStore } from 'pinia';
import axios from 'axios';
import { useTokenStore } from './token';
import { AccountInfo, UserInfo } from 'src/constant/global';
import {
	OlaresInfo,
	DefaultOlaresInfo,
	TermiPassDeviceInfo,
	SSOToken
} from '@bytetrade/core';
import { MENU_TYPE, MenuItem, OLARES_ROLE, useMenuItem } from 'src/constant';

export type UserSate = {
	terminus: OlaresInfo;
	user: UserInfo;
	devices: TermiPassDeviceInfo[];
	thisDevice: TermiPassDeviceInfo | null;
};

export interface SSOResult {
	sso: SSOToken;
	termiPass: TermiPassDeviceInfo;
}

export const useAdminStore = defineStore('admin', {
	state: () => {
		return {
			terminus: DefaultOlaresInfo,
			user: {
				created_user: '',
				is_ephemeral: false,
				name: '',
				owner_role: '',
				terminusName: '',
				wizard_complete: true,
				zone: '',
				imgContentMode: 'fill'
			},
			devices: [],
			thisDevice: null
		} as UserSate;
	},
	getters: {
		olaresId(): string {
			return this.terminus.olaresId || this.terminus.terminusName;
		},
		olares_device_id(): string {
			return this.terminus.id || this.terminus.terminusId;
		},
		olaresd(): boolean {
			return this.terminus.olaresd == '1' || this.terminus.terminusd === '1';
		},
		isNormal(): boolean {
			return (
				this.user.owner_role != undefined &&
				this.user.owner_role === OLARES_ROLE.NORMAL
			);
		},
		isAdmin(): boolean {
			return (
				this.user.owner_role != undefined &&
				(this.user.owner_role === OLARES_ROLE.ADMIN ||
					this.user.owner_role === OLARES_ROLE.OWNER)
			);
		},
		isOwner(): boolean {
			return (
				this.user.owner_role != undefined &&
				this.user.owner_role === OLARES_ROLE.OWNER
			);
		},
		menus(): (MenuItem | undefined)[][] {
			if (
				this.user.owner_role != undefined &&
				(this.user.owner_role === OLARES_ROLE.ADMIN ||
					this.user.owner_role === OLARES_ROLE.OWNER)
			) {
				return [
					[
						useMenuItem(MENU_TYPE.Users),
						useMenuItem(MENU_TYPE.Appearance),
						useMenuItem(MENU_TYPE.Application),
						useMenuItem(MENU_TYPE.Integration)
					],
					[
						useMenuItem(MENU_TYPE.VPN),
						useMenuItem(MENU_TYPE.Network),
						useMenuItem(MENU_TYPE.GPU),
						useMenuItem(MENU_TYPE.Video),
						useMenuItem(MENU_TYPE.Search)
					],
					[
						useMenuItem(MENU_TYPE.Backup),
						useMenuItem(MENU_TYPE.Restore),
						useMenuItem(MENU_TYPE.Developer)
					]
				];
			} else {
				return [
					[
						useMenuItem(MENU_TYPE.Appearance),
						useMenuItem(MENU_TYPE.Application),
						useMenuItem(MENU_TYPE.Integration)
					],
					[useMenuItem(MENU_TYPE.VPN)],
					[useMenuItem(MENU_TYPE.Video)],
					[useMenuItem(MENU_TYPE.Search)],
					[
						useMenuItem(MENU_TYPE.Backup),
						useMenuItem(MENU_TYPE.Restore),
						useMenuItem(MENU_TYPE.Developer)
					]
				];
			}
		}
	},

	actions: {
		saveUserName(userName: string) {
			this.user.name = userName;
		},

		async get_user_info() {
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				tokenStore.url + '/api/backend/v1/user-info'
			);

			this.user = data;
		},
		async get_vault_devices(): Promise<any> {
			const tokenStore = useTokenStore();

			const data: any = await axios.get(tokenStore.url + '/api/device/vault');

			return data;
		},
		async get_sso(): Promise<SSOResult[]> {
			const tokenStore = useTokenStore();

			const data: SSOResult[] = await axios.get(
				tokenStore.url + '/api/device/sso'
			);

			return data;
		},
		async revoke_token(token: SSOResult) {
			const tokenStore = useTokenStore();
			await axios.delete(
				`${tokenStore.url}/api/device/sso/${token.termiPass.sso}`
			);
		},
		isCurrentAccount(targetAccount: AccountInfo): boolean {
			if (!targetAccount?.name) {
				return false;
			}
			if (!this.user?.name) {
				return false;
			}
			return this.user.name === targetAccount.name;
		},
		canManageAccount(targetAccount: AccountInfo): boolean {
			if (!targetAccount) {
				return false;
			}

			if (!this.user?.owner_role) {
				return false;
			}

			if (this.user.name === targetAccount.name) {
				return false;
			}

			if (!!process.env.DEMO) {
				return false;
			}

			const targetHighestRole = this.getHighestRole(targetAccount.roles);
			if (!targetHighestRole) {
				return false;
			}

			const permissionLevel = {
				[OLARES_ROLE.OWNER]: 3,
				[OLARES_ROLE.ADMIN]: 2,
				[OLARES_ROLE.NORMAL]: 1
			};

			return (
				permissionLevel[this.user.owner_role] >
				permissionLevel[targetHighestRole]
			);
		},

		getHighestRole(roles: string[] = []): OLARES_ROLE | null {
			if (!roles.length) {
				return null;
			}

			const permissionLevel = {
				[OLARES_ROLE.OWNER]: 3,
				[OLARES_ROLE.ADMIN]: 2,
				[OLARES_ROLE.NORMAL]: 1
			};

			const sortedRoles = roles.sort((a, b) => {
				const levelA = permissionLevel[a as OLARES_ROLE] || 0;
				const levelB = permissionLevel[b as OLARES_ROLE] || 0;
				return levelB - levelA;
			});

			return sortedRoles[0] as OLARES_ROLE;
		}
	}
});
