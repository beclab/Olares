import { defineStore } from 'pinia';
import { date } from 'quasar';
import { app } from '../globals';
import { OrgMenu } from 'src/globals';
import { VaultMenuItem } from '../utils/contact';
import { i18n } from 'src/boot/i18n';
import { getAppPlatform } from 'src/application/platform';
import { Org, TagInfo } from '@didvault/sdk/src/core';

export type DataState = {
	audit: any;
	tag: string;
	vaultId: string;
	currentItem: string;
	leftDrawerOpen: boolean;
	rightDrawerOpen: boolean;
	dialogShow: boolean;
	useSafeArea: boolean;
	isEdit: boolean;

	org_id: string;
	org_dashboard: boolean;
	org_members: boolean;
	org_vaults: boolean;
	org_settings: boolean;
	org_invites: boolean;
	org_mode_id: string;

	terminusMenuCache: string[];
	cacheMax: number;

	splitterModel: number;

	verticalPosition: number;
	syncInfo: any;

	hideBackground: boolean;

	googleTest: boolean;

	forbiddenUseGoogleTest: boolean;

	vaultCount: {
		favorites: number;
		attachments: number;
		recent: number;
		total: number;
		report: number;
		myvault: number;
		authenticator: number;
	};

	orgs: Org[];

	myteamMembers: {
		id: string;
		name: string;
		revision?: string;
		count: number;
	}[];

	tags: TagInfo[];
};

export const useMenuStore = defineStore('menu', {
	state: () => {
		return {
			tag: '',
			audit: null,
			vaultId: '',
			currentItem:
				process.env.PLATFORM === 'MOBILE'
					? VaultMenuItem.AUTHENTICATOR
					: VaultMenuItem.ALLVAULTS,
			leftDrawerOpen: false,
			rightDrawerOpen: false,
			dialogShow: false,
			useSafeArea: true,
			isEdit: false,

			org_id: '',
			org_dashboard: false,
			org_members: false,
			org_vaults: false,
			org_settings: false,
			org_invites: false,
			org_mode_id: '',

			terminusMenuCache: <string[]>[],
			cacheMax: 10,

			splitterModel: 50,
			verticalPosition: 0,
			syncInfo: {
				syncing: false,
				lastSyncTime: ''
			},
			hideBackground: false,
			googleTest: false,
			forbiddenUseGoogleTest: true,
			vaultCount: {
				favorites: 0,
				attachments: 0,
				recent: 0,
				total: 0,
				report: 0,
				myvault: 0,
				authenticator: 0
			},
			orgs: [],
			myteamMembers: [],
			tags: []
		} as DataState;
	},
	getters: {
		terminusActiveMenu(state) {
			return state.terminusMenuCache[state.terminusMenuCache.length - 1];
		},
		vaults(state) {
			const defaultMenu = {
				label: i18n.global.t('vault_t.Vault Classification'),
				key: VaultMenuItem.VAULTCLASSIFICATION,
				icon: '',
				children: [
					{
						label: i18n.global.t('vault_t.all_vaults'),
						key: VaultMenuItem.ALLVAULTS,
						icon: 'sym_r_apps',
						count: state.vaultCount.total
					},

					{
						label: i18n.global.t('vault_t.recently_used'),
						key: VaultMenuItem.RECENTLYUSED,
						icon: 'sym_r_schedule',
						count: state.vaultCount.recent
					},
					{
						label: i18n.global.t('favorites'),
						key: VaultMenuItem.FAVORITES,
						icon: 'sym_r_star',
						count: state.vaultCount.favorites
					},
					{
						label: i18n.global.t('attachments'),
						key: VaultMenuItem.ATTACHMENTS,
						icon: 'sym_r_lab_profile',
						count: state.vaultCount.attachments
					},
					{
						label: i18n.global.t('vault_t.my_vault'),
						key: VaultMenuItem.MyVault,
						icon: 'sym_r_frame_person',
						count: state.vaultCount.myvault
					},
					{
						label: i18n.global.t('vault_t.Team Vaults'),
						key: VaultMenuItem.MYTEAMS,
						icon: 'sym_r_groups',
						children: this.myteamMembers.map((member) => {
							return {
								label: member.name,
								key: member.id,
								icon: 'sym_r_deployed_code',
								count: member.count,
								vaultId: member.id
							};
						}),
						disable: this.myteamMembers.length === 0
					},
					{
						label: i18n.global.t('tags'),
						key: VaultMenuItem.TAGS,
						icon: 'sym_r_more',
						children: this.tags.map((tag) => {
							return {
								label: tag.name,
								key: tag.name,
								icon: 'sym_r_sell',
								count: tag.count
							};
						}),
						disable: this.tags.length === 0
					}
				]
			};

			if (process.env.PLATFORM == 'MOBILE') {
				const authMenu = {
					label: i18n.global.t('vault_t.authenticator'),
					key: VaultMenuItem.AUTHENTICATOR,
					icon: 'sym_r_encrypted',
					count: app.count.authenticator
				};
				defaultMenu.children.splice(1, 0, authMenu);
			}
			return defaultMenu;
		},
		teamVautls(state) {
			const teamsMenus = {
				label: i18n.global.t('vault_t.Teams'),
				key: VaultMenuItem.TEAMS,
				icon: '',
				children: [
					{
						label: VaultMenuItem.TEAMS,
						key: VaultMenuItem.TEAMS,
						icon: 'sym_r_groups',
						children:
							this.orgs.length == 0
								? []
								: [
										{
											label: i18n.global.t('members'),
											key: OrgMenu.MEMBERS,
											icon: 'sym_r_groups',
											org_id: this.orgs[0].id
										},
										{
											label: i18n.global.t('vaults'),
											key: OrgMenu.VAULTES,
											icon: 'sym_r_apps',
											org_id: this.orgs[0].id
										}
								  ]
					}
				]
			};
			return teamsMenus;
		},
		menus() {
			if (this.orgs.length == 0) {
				return [this.vaults];
			}
			return [this.vaults, this.teamVautls];
		}
	},
	actions: {
		transformInvites(invites: any) {
			const newInvites: any[] = [];
			for (let i = 0; i < invites.length; i++) {
				const el = invites[i];
				const tagObj = {
					label: el.orgName,
					key: el.orgName,
					id: el.id,
					orgId: el.orgId,
					icon: 'sym_r_mail'
				};
				newInvites.push(tagObj);
			}
			return newInvites;
		},

		transformOrgs(myteamOrg: any) {
			const newVaults: any[] = [];
			for (let k = 0; k < myteamOrg.length; k++) {
				const el = myteamOrg[k];
				const tagObj = {
					label: el.name,
					key: el.id,
					icon: 'sym_r_deployed_code',
					count: el.count,
					vaultId: el.id
				};
				newVaults.push(tagObj);
			}
			return newVaults;
		},

		transformTag(tags: any) {
			const newTags: any[] = [];
			for (let k = 0; k < tags.length; k++) {
				const el = tags[k];
				const tagObj = {
					label: el.name,
					key: el.name,
					icon: 'sym_r_sell',
					count: el.count,
					tagId: el.name
				};
				newTags.push(tagObj);
			}
			return newTags;
		},

		transformApporg(apporgs: any) {
			const newApporgs: any[] = [];
			for (let k = 0; k < apporgs.length; k++) {
				const el = apporgs[k];
				const tagObj = {
					label: el.name,
					key: el.name,
					icon: 'sym_r_sell',
					count: el.count
				};
				newApporgs.push(tagObj);
			}
			return newApporgs;
		},

		transformCount(key: string, count: any) {
			switch (key) {
				case VaultMenuItem.ALLVAULTS:
					return count.total;
					break;

				case VaultMenuItem.AUTHENTICATOR:
					return count.authenticator;
					break;

				case VaultMenuItem.RECENTLYUSED:
					return count.recent;
					break;

				case VaultMenuItem.FAVORITES:
					return count.favorites;
					break;

				case VaultMenuItem.ATTACHMENTS:
					return count.attachments;
					break;

				case VaultMenuItem.MyVault:
					return count.myvault;
					break;

				default:
					break;
			}
		},

		clear() {
			this.tag = '';
			this.audit = null;
			this.vaultId = '';

			this.org_id = '';
			this.org_dashboard = false;
			this.org_members = false;
			this.org_vaults = false;
			this.org_settings = false;
			this.org_invites = false;
			this.org_mode_id = '';
			this.currentItem =
				process.env.PLATFORM === 'MOBILE'
					? VaultMenuItem.AUTHENTICATOR
					: VaultMenuItem.ALLVAULTS;
		},
		setTag(tag: string) {
			// this.clear();
			this.tag = tag;
		},
		changeItemMenu(vaultId = '') {
			if (vaultId) {
				this.vaultId = vaultId;
			}
		},

		selectOrgMenu(org_id: string, mode: OrgMenu) {
			this.clear();
			this.org_id = org_id;
			if (mode == OrgMenu.DASHBOARD) {
				this.org_dashboard = true;
			} else if (mode == OrgMenu.MEMBERS) {
				this.org_members = true;
			} else if (mode == OrgMenu.VAULTES) {
				this.org_vaults = true;
			} else if (mode == OrgMenu.SETTINGS) {
				this.org_settings = true;
			} else if (mode == OrgMenu.INVITES) {
				this.org_invites = true;
			}
		},

		changeSafeArea(safeArea: boolean) {
			this.useSafeArea = safeArea;
		},

		pushTerminusMenuCache(menuName: string) {
			if (this.terminusMenuCache.length >= this.cacheMax) {
				this.terminusMenuCache.shift();
			}
			this.terminusMenuCache.push(menuName);
		},

		popTerminusMenuCache() {
			if (this.terminusMenuCache.length <= 1) {
				return false;
			}
			this.terminusMenuCache.pop();
		},

		setSplitterModel(value: number) {
			this.splitterModel = value;
		},

		updateHideBackground(hideBackground: boolean) {
			this.hideBackground = hideBackground;
		},

		async handleSync() {
			await getAppPlatform().vaultSync();
		},

		updateMenuInfo() {
			const orgs = app.orgs.filter((org) => org.isAdmin(app.account!));
			const myteamMember: any =
				app.orgs && app.orgs.length > 0
					? app.orgs[0].getVaultsForMember(app.account!)
					: [];
			for (let i = 0; i < myteamMember.length; i++) {
				const el = myteamMember[i];
				el.count = app.getVault(el.id)?.items.size;
			}

			this.vaultCount = app.count;

			this.myteamMembers = myteamMember;

			this.orgs = orgs;

			this.tags = app.tags;

			this.syncInfo = {
				syncing: app.state.syncing,
				lastSyncTime: date.formatDate(app.state.lastSync, 'HH:mm:ss')
			};
		}
	}
});
