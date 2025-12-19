import { defineStore } from 'pinia';
import { date } from 'quasar';
import { app } from '../globals';
import { OrgMenu } from 'src/globals';
import { VaultMenuItem } from '../utils/contact';
import { i18n } from 'src/boot/i18n';

export type DataState = {
	favorites: boolean;
	attachments: boolean;
	recent: boolean;
	tag: string;
	host: string;
	audit: any;
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
	splitter100: boolean;

	verticalPosition: number;
	syncInfo: any;
	menus: any[];

	hideBackground: boolean;

	googleTest: boolean;
	forbiddenUseGoogleTest: boolean;
};

export const useMenuStore = defineStore('menu', {
	state: () => {
		return {
			favorites: false,
			attachments: false,
			recent: false,
			tag: '',
			host: '',
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
			splitter100: false,
			verticalPosition: 0,
			syncInfo: {
				syncing: false,
				lastSyncTime: ''
			},
			menus: [],
			hideBackground: false,
			googleTest: false,
			forbiddenUseGoogleTest: true
		} as DataState;
	},
	getters: {
		terminusActiveMenu(state) {
			return state.terminusMenuCache[state.terminusMenuCache.length - 1];
		}
	},
	actions: {
		updateMenu({ tags, count, orgs, myteamOrg, invites }) {
			this.menus = [];

			let menus: any[] = [];
			const vaultMenus = this.vaultMenus();
			let teamsMenus = this.teamsMenus();
			let invitesMenus = this.invitesMenus();
			// const toolsMenus = this.toolsMenus();
			// const utilityMenus = this.utilityMenus();

			const elChildren = vaultMenus.children;
			for (let j = 0; j < elChildren.length; j++) {
				const elTag: any = elChildren[j];

				if (count) {
					elTag.count = this.transformCount(elTag.key, count);
				}

				if (elTag.key === VaultMenuItem.TAGS) {
					elTag.disable = false;
					if (tags && tags.length > 0) {
						elTag.children = this.transformTag(tags);
					} else {
						elTag.children = [];
						elTag.disable = true;
					}
				}

				if (elTag.key === VaultMenuItem.MYTEAMS) {
					elTag.disable = false;
					if (myteamOrg.length > 0) {
						elTag.children = this.transformOrgs(myteamOrg);
					} else {
						elTag.children = [];
						elTag.disable = true;
					}
				}
			}

			if (orgs && orgs.length > 0) {
				const elChildren: any = teamsMenus.children[0];
				elChildren.label = orgs[0].name;
				elChildren.key = orgs[0].name;
				if (orgs[0].owner.did !== app.account?.did) {
					elChildren.children = elChildren.children.filter(
						(cell) => cell.key !== OrgMenu.INVITES
					);
				}

				for (let j = 0; j < elChildren.children.length; j++) {
					const elTag: any = elChildren.children[j];
					elTag.org_id = orgs[0].id;
				}
			} else {
				teamsMenus = null;
			}

			if (invites && invites.length > 0) {
				invitesMenus.children = this.transformInvites(invites);
			} else {
				invitesMenus = null;
			}

			// if (!isWeb) {
			// 	utilityMenus = null;
			// }

			menus.push(
				vaultMenus,
				teamsMenus,
				invitesMenus
				// toolsMenus,
				// utilityMenus
			);

			menus = menus.filter((menu) => {
				return menu !== null;
			});

			this.menus = menus;
		},

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
					key: el.name,
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
					count: el.count
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
			this.favorites = false;
			this.attachments = false;
			this.recent = false;
			this.tag = '';
			this.host = '';
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
		setFavorites() {
			this.clear();
			this.favorites = true;
		},
		setAttachments() {
			this.clear();
			this.attachments = true;
		},
		setRecent() {
			this.clear();
			this.recent = true;
		},
		setTag(tag: string) {
			this.clear();
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
			await app.synchronize();
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

			this.updateMenu({
				tags: app.tags,
				myteamOrg: myteamMember,
				count: app.count,
				orgs,
				invites: app.authInfo?.invites
			});

			this.syncInfo = {
				syncing: app.state.syncing,
				lastSyncTime: date.formatDate(app.state.lastSync, 'HH:mm:ss')
			};
		},
		updateGoogleTest(value: boolean) {
			if (this.forbiddenUseGoogleTest) {
				return;
			}
			this.googleTest = value;
		},
		vaultMenus: (): any => {
			const defaultMenu = {
				label: i18n.global.t('vault_t.Vault Classification'),
				key: VaultMenuItem.VAULTCLASSIFICATION,
				icon: '',
				children: [
					{
						label: i18n.global.t('vault_t.all_vaults'),
						key: VaultMenuItem.ALLVAULTS,
						icon: 'sym_r_apps'
					},

					{
						label: i18n.global.t('vault_t.recently_used'),
						key: VaultMenuItem.RECENTLYUSED,
						icon: 'sym_r_schedule'
					},
					{
						label: i18n.global.t('favorites'),
						key: VaultMenuItem.FAVORITES,
						icon: 'sym_r_star'
					},
					{
						label: i18n.global.t('attachments'),
						key: VaultMenuItem.ATTACHMENTS,
						icon: 'sym_r_lab_profile'
					},
					{
						label: i18n.global.t('vault_t.my_vault'),
						key: VaultMenuItem.MyVault,
						icon: 'sym_r_frame_person'
					},
					{
						label: i18n.global.t('vault_t.Team Vaults'),
						key: VaultMenuItem.MYTEAMS,
						icon: 'sym_r_groups',
						children: []
					},
					{
						label: i18n.global.t('tags'),
						key: VaultMenuItem.TAGS,
						icon: 'sym_r_more',
						children: []
					}
				]
			};

			if (process.env.PLATFORM == 'MOBILE') {
				const authMenu = {
					label: i18n.global.t('vault_t.authenticator'),
					key: VaultMenuItem.AUTHENTICATOR,
					icon: 'sym_r_encrypted'
				};
				defaultMenu.children.splice(1, 0, authMenu);
			}

			return defaultMenu;
		},

		teamsMenus: (): any => {
			return {
				label: i18n.global.t('vault_t.Teams'),
				key: VaultMenuItem.TEAMS,
				icon: '',
				children: [
					{
						label: VaultMenuItem.TEAMS,
						key: VaultMenuItem.TEAMS,
						icon: 'sym_r_groups',
						children: [
							// {
							// 	label: OrgMenu.DASHBOARD,
							// 	key: OrgMenu.DASHBOARD,
							// 	icon: 'sym_r_dashboard'
							// },
							// {
							// 	label: i18n.global.t('invites'),
							// 	key: OrgMenu.INVITES,
							// 	icon: 'sym_r_share_reviews'
							// },
							{
								label: i18n.global.t('members'),
								key: OrgMenu.MEMBERS,
								icon: 'sym_r_groups'
							},
							{
								label: i18n.global.t('vaults'),
								key: OrgMenu.VAULTES,
								icon: 'sym_r_apps'
							}
							// {
							// 	label: OrgMenu.SETTINGS,
							// 	key: OrgMenu.SETTINGS,
							// 	icon: 'sym_r_settings'
							// }
						]
					}
				]
			};
		},

		invitesMenus: (): any => {
			return {
				label: i18n.global.t('invites'),
				key: VaultMenuItem.INVITES,
				icon: '',
				children: []
			};
		},

		toolsMenus: (): any => {
			return {
				label: i18n.global.t('web.tools.title'),
				key: VaultMenuItem.TOOLS,
				icon: '',
				children: [
					// {
					// 	label: VaultMenuItem.SECURITYREPORT,
					// 	label: VaultMenuItem.SECURITYREPORT,
					// 	icon: 'sym_r_arming_countdown'
					// },
					{
						label: i18n.global.t('password_generators'),
						key: VaultMenuItem.PASSWORDGENERATOR,
						icon: 'sym_r_casino'
					}
				]
			};
		},

		utilityMenus: (): any => {
			return {
				label: i18n.global.t('Utility'),
				key: VaultMenuItem.UTILITY,
				icon: '',
				children: [
					{
						label: i18n.global.t('Lock Screen'),
						key: VaultMenuItem.LOCKSCREEN,
						icon: 'sym_r_lock'
					},
					{
						label: i18n.global.t('setting'),
						key: VaultMenuItem.SETTINGS,
						icon: 'sym_r_settings'
					}
				]
			};
		}
	}
});
