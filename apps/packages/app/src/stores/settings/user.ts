import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';
import {
	AccountInfo,
	AccountModifyStatus,
	CreateAccountRequest,
	RestAccountPassword,
	UpdateAccountQuotaRequest,
	UsersParam
} from 'src/constant/global';
import { useAdminStore } from './admin';
import { OLARES_ROLE } from 'src/constant';

export type AccountState = {
	accounts: AccountInfo[];
};

export const useUserStore = defineStore('settingsUser', {
	state: () => {
		return {
			accounts: []
		} as AccountState;
	},

	getters: {
		currentUserInfo(): AccountInfo | undefined {
			const admin = useAdminStore();
			return this.accounts.find((a) => admin.user.name === a.name);
		}
	},

	actions: {
		async create_account(req: CreateAccountRequest) {
			const tokenStore = useTokenStore();
			await axios.post(tokenStore.url + '/api/users', req);
		},

		async update_account_quoto(user: string, body: UpdateAccountQuotaRequest) {
			const tokenStore = useTokenStore();
			await axios.post(tokenStore.url + '/api/users/' + user + '/limits', body);
		},

		async update_account(account: AccountInfo) {
			for (let i = 0; i < this.accounts.length; ++i) {
				if (this.accounts[i].name == account.name) {
					this.accounts[i] = account;
					return;
				}
			}
			this.accounts.push(account);
		},

		async get_accounts() {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(tokenStore.url + '/api/users');
			console.log(data);
			if (data && data.length > 0) {
				for (let i = 0; i < data.length; ++i) {
					this.update_account(data[i]);
				}

				const admins: AccountInfo[] = [];
				const actives: AccountInfo[] = [];
				const noActives: AccountInfo[] = [];
				this.accounts.sort((a, b) => {
					return a.creation_timestamp - b.creation_timestamp;
				});

				this.accounts.forEach((e) => {
					if (
						e.roles.findIndex(
							(role) => role == OLARES_ROLE.ADMIN || role == OLARES_ROLE.OWNER
						) >= 0
					) {
						admins.push(e);
						return;
					}
					if (e.wizard_complete) {
						actives.push(e);
						return;
					}
					noActives.push(e);
				});

				this.accounts = [...admins, ...actives, ...noActives];
			}
		},

		getUserByName(name: string) {
			return this.accounts.find((u) => u.name === name);
		},

		async get_account_info(username: string) {
			return this.accounts.find((account) => account.name == username);
		},

		async update_account_info(username: string) {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				tokenStore.url + '/api/users/' + username
			);

			this.update_account(data);
		},

		async delete_account(username: string) {
			const tokenStore = useTokenStore();
			return await axios.delete(tokenStore.url + '/api/users/' + username, {});
		},

		async removeLocalAccount(username: string) {
			const index = this.accounts.findIndex(
				(account) => account.name == username
			);
			if (index >= 0) {
				this.accounts.splice(index, 1);
			}
		},

		async reset_account_password(req: RestAccountPassword) {
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				tokenStore.url + '/api/users/' + req.username + '/password',
				req
			);
			return data;
		},

		async get_account_status(username: string): Promise<AccountModifyStatus> {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				tokenStore.url + '/api/users/' + username + '/status'
			);
			return data;
		},

		async getLoginrecords(params: UsersParam): Promise<any> {
			const tokenStore = useTokenStore();
			const { user } = params;
			const data: any = await axios.get(
				tokenStore.url + '/api/users/' + user + '/login-records'
			);
			return data;
		}
	}
});
