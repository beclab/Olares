import { RouteRecordRaw } from 'vue-router';

const mobile: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/MobileLoginLayout.vue'),
		children: [
			{
				path: '/declaration',
				name: 'declaration',
				component: () => import('src/pages/Mobile/login/DeclarationPage.vue')
			},
			{
				path: '/welcome',
				name: 'welcome',
				component: () => import('src/pages/Mobile/login/WelcomePage.vue')
			},
			{
				path: '/setUnlockPassword',
				name: 'setUnlockPassword',
				component: () =>
					import('src/pages/Mobile/login/unlock/SetUnlockPwd.vue')
			},
			{
				path: '/unlock',
				name: 'unlock',
				meta: {
					emptyUserDisableBack: true
				},
				component: () => import('src/pages/Mobile/login/unlock/UnlockPage.vue')
			},
			{
				path: '/bind_vc',
				name: 'bindVC',
				component: () => import('src/pages/Mobile/login/vc/BindTerminusVC.vue')
			},
			{
				path: '/bind_customer_vc',
				component: () => import('src/pages/Mobile/login/vc/BindCustomerVC.vue')
			},
			{
				path: '/bind_org_vc',
				component: () => import('src/pages/Mobile/login/vc/BindOrgVC.vue')
			},
			{
				path: '/bind_vc_success',
				meta: {
					noReturn: true
				},
				component: () => import('src/pages/Mobile/login/vc/BindVCSuccess.vue')
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
				path: '/autoFillList',
				meta: {
					exit: true
				},
				component: () =>
					import('src/pages/Mobile/android/AndroidAutoFillList.vue')
			},
			{
				path: '/import_mnemonic',
				name: 'InputMnemonic',
				component: () => import('pages/Mobile/login/account/InputMnemonic.vue')
			},

			{
				path: '/connectLoading',
				name: 'connectLoading',
				component: () => import('src/pages/Mobile/connect/ConnectLoading.vue')
			},
			{
				path: '/Activate/:disableLeave?/:isCreate?',
				meta: {
					noReturn: true
				},
				component: () =>
					import('src/pages/Mobile/connect/activate/ActivateTerminus.vue')
			},
			{
				path: '/ActivateWizard',
				meta: { noReturn: true },
				component: () =>
					import('src/pages/Mobile/connect/activate/ActivateWizard.vue')
			},

			{
				path: '/ResetPassword',
				meta: {},
				component: () =>
					import('src/pages/Mobile/connect/activate/ResetPassword.vue')
			},

			{
				path: '/scan',
				meta: {
					backgroundHide: true
				},
				component: () =>
					import('src/pages/Mobile/connect/activate/ScanPage.vue')
			},
			{
				path: '/scan_local',
				component: () =>
					import('src/pages/Mobile/connect/activate/ScanLocalPage.vue')
			},

			{
				path: '/ConnectTerminus',
				meta: {
					noReturn: true
				},
				component: () => import('src/pages/Mobile/connect/ConnectTerminus.vue')
			},

			{
				path: '/DomainDetail/:base',
				meta: {},
				component: () =>
					import('src/pages/Mobile/cloud/domain/DomainDetail.vue')
			},

			{
				path: '/CloudDomainManage',
				meta: {},
				component: () =>
					import('src/pages/Mobile/cloud/domain/CloudDomainManage.vue')
			},
			{
				path: '/JoinOrganization',
				component: () =>
					import('src/pages/Mobile/login/vc/JoinOrganizationPage.vue')
			},
			{
				path: '/OrganizationBindVC/:domain',
				component: () =>
					import('src/pages/Mobile/login/vc/OriganizationBindVCPage.vue')
			},
			{
				path: '/BindTerminusName',
				meta: {
					noReturn: true
				},
				component: () =>
					import('src/pages/Mobile/login/vc/BindTerminusNamePage.vue')
			},
			{
				path: '/discover/machine',
				component: () =>
					import(
						'src/pages/Mobile/connect/activate/DiscoverLocalMachinesPage.vue'
					)
			},
			{
				path: '/bluetooth/discover/machine',
				component: () =>
					import(
						'src/pages/Mobile/connect/activate/BluetoothConnectMachinesPage.vue'
					)
			},
			{
				path: '/bluetooth/machine/wifi',
				component: () =>
					import(
						'src/pages/Mobile/connect/activate/BluetoothConnectNetworkPage.vue'
					)
			}
		]
	}
];

const mobileExtension: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/PluginMainLayout.vue'),
		children: [
			{
				path: '/home',
				meta: {
					tabIdentify: 'home',
					minimizeApp: 'true'
				},
				component: () => import('pages/Mobile/file/FileRootPage.vue')
			},
			{
				path: '/shard',
				meta: {
					removeShared: 'true'
				},
				component: () => import('src/pages/Mobile/file/FilesShardPage.vue')
			}
		]
	}
];

export default [...mobile, ...mobileExtension];
