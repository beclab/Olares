import { RouteRecordRaw } from 'vue-router';

const desktop: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('src/layouts/mobiles/PadLoginLayout.vue'),
		children: [
			{
				path: '/declaration',
				name: 'declaration',
				component: () => import('src/pages/Mobile/login/DeclarationPage.vue')
			},
			{
				path: '/welcome',
				name: 'welcome',
				component: () => import('pages/Electron/import/WelcomePage.vue')
			},
			{
				path: '/setUnlockPassword',
				name: 'setUnlockPassword',
				component: () =>
					import('src/pages/Electron/import/SetUnlockPwdPage.vue')
			},
			{
				path: '/setup/success',
				name: 'setupSuccess',
				meta: {
					emptyUserDisableBack: true
				},
				component: () =>
					import('src/pages/Mobile/login/account/SetupAccount.vue')
			},
			{
				path: '/import_mnemonic',
				name: 'InputMnemonic',
				component: () => import('pages/Electron/import/InputMnemonicPage.vue')
			},
			{
				path: 'connectLoading',
				component: () => import('src/pages/Mobile/connect/ConnectLoading.vue')
			},
			{
				path: 'ConnectTerminus',
				component: () => import('src/pages/Mobile/connect/ConnectTerminus.vue')
			},
			{
				path: 'bind_customer_vc',
				component: () => import('src/pages/Mobile/login/vc/BindCustomerVC.vue')
			},
			{
				path: 'bind_org_vc',
				component: () => import('src/pages/Mobile/login/vc/BindOrgVC.vue')
			},
			{
				path: 'bind_vc_success',
				meta: {
					noReturn: true
				},
				component: () => import('src/pages/Mobile/login/vc/BindVCSuccess.vue')
			},
			{
				path: 'JoinOrganization',
				component: () =>
					import('src/pages/Mobile/login/vc/JoinOrganizationPage.vue')
			},
			{
				path: '/OrganizationBindVC/:domain',
				component: () =>
					import('src/pages/Mobile/login/vc/OriganizationBindVCPage.vue')
			},
			{
				path: '/CloudDomainManage',
				meta: {},
				component: () =>
					import('src/pages/Mobile/cloud/domain/CloudDomainManage.vue')
			},
			{
				path: '/accounts',
				name: 'accounts',
				meta: {
					noReturn: true
				},
				component: () => import('src/pages/Mobile/AccountList.vue')
			},
			{
				path: '/backup_mnemonics',
				name: 'backupMnemonics',
				component: () =>
					import('src/pages/Mobile/setting/BackupMnemonicsPage.vue')
			}
		]
	},
	{
		path: '/unlock',
		component: () =>
			import('src/pages/Electron/unlock/DesktopUnlockLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('pages/Electron/unlock/UnlockPage.vue')
			}
		]
	},
	{
		path: '/',
		component: () => import('layouts/MobileMainLayout.vue'),
		name: 'MobileMainLayout',
		children: [
			{
				path: '/scanQrCode',
				component: () => import('src/pages/Mobile/ScanQRCodePage.vue')
			}
		]
	},
	{
		path: '/',
		component: () => import('layouts/TermipassMainLayout.vue'),
		name: 'TermipassMainLayout',
		children: [
			{
				path: '/transmission',
				component: () => import('pages/Electron/Transfer/TransferPage.vue')
			},
			{
				path: '/systemSettings',
				component: () => import('pages/Electron/SettingsPage/Account.vue')
			},
			{
				path: '/accountCenter',
				component: () => import('pages/Electron/SettingsPage/Account.vue')
			},
			{
				path: '/Files/:path*',
				name: 'Files',
				component: () => import('pages/Files/FilesPage.vue')
			},
			{
				path: '/Seahub/:path*',
				name: 'Seahub',
				component: () => import('pages/Files/FilesPage.vue')
			},
			{
				path: '/Drive/:path*',
				name: 'Drive',
				component: () => import('pages/Files/FilesPage.vue')
			},
			{
				path: 'Data/:path*',
				name: 'Data',
				component: () => import('pages/Files/FilesPage.vue')
			},
			{
				path: 'Cache/:path*',
				name: 'Cache',
				component: () => import('pages/Files/FilesPage.vue')
			},
			{
				path: '/items/',
				component: () => import('pages/Items/ItemsPage.vue')
			},
			{
				path: '/items/:itemid',
				component: () => import('pages/Items/ItemsPage.vue')
			},
			{
				path: '/settings/',
				component: () => import('pages/Web/Settings/IndexPage.vue')
			},
			{
				path: '/invite-recipient/:org_id/:invite_id',
				component: () => import('pages/Items/InviteRecipient.vue')
			},
			{
				path: '/org/:org_mode',
				component: () => import('pages/Web/Orgs/OrgIndexPage.vue')
			},
			{
				path: '/org/:org_mode/:org_type',
				component: () => import('pages/Web/Orgs/OrgIndexPage.vue')
			},
			{
				path: '/settings/:mode',
				component: () => import('pages/Web/Settings/IndexPage.vue')
			},
			{
				path: '/generator',
				component: () => import('pages/Web/Generator.vue')
			},
			{
				path: '/security',
				component: () => import('pages/Web/Security/IndexPage.vue')
			}
		]
	},
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/ErrorNotFound.vue')
	}
];
export default desktop;
