import { defineStore } from 'pinia';
import axios from 'axios';
import { useTokenStore } from './token';
import { AccountInfo, UserInfo } from 'src/constant/global';
import { OLARES_ROLE } from 'src/constant';
import { Cookies } from 'quasar';

export type MeState = {
	user: UserInfo;
};

function extractUserPayload(payload: any): UserInfo | null {
	if (!payload || typeof payload !== 'object') {
		return null;
	}

	const looksLikeEnvelope =
		'code' in payload &&
		'data' in payload &&
		payload.data &&
		typeof payload.data === 'object';
	const looksLikeUser =
		typeof payload.name === 'string' ||
		typeof payload.owner_role === 'string' ||
		typeof payload.terminusName === 'string';

	if (looksLikeUser && !looksLikeEnvelope) {
		return payload as UserInfo;
	}

	if (looksLikeEnvelope) {
		return extractUserPayload(payload.data);
	}

	return null;
}

export const useMeStore = defineStore('me', {
	state: () => {
		return {
			user: {
				created_user: '',
				is_ephemeral: false,
				name: '',
				owner_role: '',
				terminusName: '',
				wizard_complete: true,
				zone: '',
				imgContentMode: 'fill'
			}
		} as MeState;
	},
	getters: {
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
		islocal(): boolean {
			return window.location.hostname.endsWith('olares.local');
		}
	},

	actions: {
		saveUserName(userName: string) {
			this.user.name = userName;
		},

		getAppLocalUrl(id: string): string {
			const owner =
				(this.user.terminusName || '').split('@')[0] || this.user.name;
			return id + '.' + owner + '.' + 'olares.local';
		},

		async init() {
			if (this.user.owner_role) {
				return;
			}
			await this.get_user_info();
		},

		async get_user_info() {
			const url = window.location.origin + '/api/backend/v1/user-info';
			const data: any = await axios.get(url);
			const userPayload = extractUserPayload(data);
			if (userPayload) {
				this.user = { ...this.user, ...userPayload };
			} else {
				console.warn(
					'[meStore.get_user_info] unrecognized payload shape, keep current user untouched',
					data && typeof data === 'object' ? Object.keys(data) : data
				);
			}
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
		},
		async logout() {
			try {
				const tokenStore = useTokenStore();
				await axios.post(tokenStore.url + '/api/logout');
				Cookies.remove('auth_token');
			} catch (e) {
				return e;
			}
		}
	}
});
