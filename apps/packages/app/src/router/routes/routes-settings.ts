import { RouteRecordRaw } from 'vue-router';
import { MENU_TYPE } from 'src/constant';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/settings/MainLayout.vue'),
		children: [
			{
				path: '',
				name: MENU_TYPE.Root,
				component: () => import('src/pages/settings/RootPage.vue')
			},
			{
				path: '/person',
				component: () => import('src/pages/settings/Person/IndexPage.vue')
			},
			{
				path: '/olares_space',
				component: () => import('src/pages/settings/Person/OlaresSpacePage.vue')
			},
			{
				path: 'loginHistory/:name?',
				name: 'loginHistory',
				component: () => import('pages/settings/Person/LoginHistoryPage.vue')
			},
			{
				path: '/active_session',
				component: () =>
					import('src/pages/settings/Person/VaultActivieSession.vue')
			},
			{
				path: '/device/:deviceId',
				component: () =>
					import('src/pages/settings/Person/DeviceDetailPage.vue')
			},
			{
				path: '/sso_token',
				component: () => import('src/pages/settings/Person/SSOToken.vue')
			},
			{
				path: '/authority',
				component: () => import('src/pages/settings/Person/AuthorityPage.vue')
			},
			{
				path: '/hardware',
				component: () => import('src/pages/settings/Person/HardwarePage.vue')
			},
			{
				path: '/hardware/:command',
				component: () =>
					import('src/pages/settings/Person/HardwareCommandPage.vue')
			},
			{
				path: '/user',
				name: MENU_TYPE.Users,
				component: () => import('src/pages/settings/Account/IndexPage.vue')
			},
			{
				path: '/user/info/:name?',
				component: () =>
					import('src/pages/settings/Account/pages/UserInfoPage.vue')
			},
			{
				path: '/appearance',
				name: MENU_TYPE.Appearance,
				component: () => import('src/pages/settings/Appearance/IndexPage.vue')
			},
			{
				path: '/knowledge',
				component: () => import('src/pages/settings/Knowledge/IndexPage.vue')
			},
			{
				path: '/application',
				name: MENU_TYPE.Application,
				component: () => import('src/pages/settings/Application/IndexPage.vue')
			},
			{
				path: '/application/info/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationDetailPage.vue'
					)
			},
			{
				path: '/application/ports/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationPortsPage.vue'
					)
			},
			{
				path: '/application/secret/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationSecretPage.vue'
					)
			},
			{
				path: '/application/environment/manager',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationEnvironmentPage.vue'
					)
			},
			{
				path: '/application/acl/:name',
				name: 'appAcl',
				component: () =>
					import('src/pages/settings/Application/pages/ApplicationAclPage.vue')
			},

			{
				path: '/application/entrance/:name/:entrance',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationEntrancePage.vue'
					)
			},

			{
				path: '/application/domain/:name/:entrance/:shared?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationDomainPage.vue'
					)
			},
			{
				path: '/application/permission/detail',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationPermissionPage.vue'
					)
			},

			{
				path: '/integration',
				name: MENU_TYPE.Integration,
				component: () => import('src/pages/settings/Integration/IndexPage.vue')
			},
			{
				path: '/integration/login/space',
				component: () =>
					import('src/pages/settings/Integration/Space/LoginPage.vue')
			},
			{
				path: '/integration/detail/space',
				component: () =>
					import('src/pages/settings/Integration/AccountDetail.vue')
			},
			{
				path: '/integration/cookie',
				component: () =>
					import(
						'src/pages/settings/Integration/pages/CookieManagementPage.vue'
					)
			},
			{
				path: '/integration/cookie/:mainDomain',
				component: () =>
					import('src/pages/settings/Integration/pages/CookieDetailsPage.vue')
			},
			{
				path: '/integration/detail/space/:address',
				component: () =>
					import('src/pages/settings/Integration/TerminusSpaceNFTPage.vue')
			},
			{
				path: '/integration/common/detail/:type/:name',
				component: () =>
					import('src/pages/settings/Integration/IntegrationDetailPage.vue')
			},
			{
				path: '/integration/add',
				component: () =>
					import('src/pages/settings/Integration/pages/IntegrationAddPage.vue')
			},
			{
				path: '/integration/accountList',
				component: () =>
					import('src/pages/settings/Integration/pages/IntegrationListPage.vue')
			},
			{
				path: '/integration/account/add',
				component: () =>
					import(
						'src/pages/settings/Integration/pages/AwsAddIntegrationPage.vue'
					)
			},

			{
				path: '/vpn',
				name: MENU_TYPE.VPN,
				component: () => import('src/pages/settings/Vpn/VPNPage.vue')
			},
			{
				path: '/vpn/active_headscale',
				component: () => import('src/pages/settings/Vpn/HeadScale.vue')
			},

			{
				path: '/network',
				name: MENU_TYPE.Network,
				component: () => import('src/pages/settings/Network/NetworkPage.vue')
			},
			{
				path: '/network/reverse_proxy',
				component: () =>
					import('src/pages/settings/Network/ReverseProxyPage.vue')
			},
			{
				path: '/network/host',
				component: () => import('src/pages/settings/Network/HostPage.vue')
			},
			{
				path: '/gpu',
				name: MENU_TYPE.GPU,
				component: () => import('src/pages/settings/GPU/GPUPage.vue')
			},
			{
				path: '/video',
				name: MENU_TYPE.Video,
				component: () => import('src/pages/settings/Video/IndexPage.vue')
			},
			{
				path: '/video/hardwareAcceleration',
				component: () =>
					import('src/pages/settings/Video/pages/HardwareAccelerationPage.vue')
			},
			{
				path: '/video/encodingScheme',
				component: () =>
					import('src/pages/settings/Video/pages/EncodingSchemePage.vue')
			},
			{
				path: '/video/transcodingSettings',
				component: () =>
					import('src/pages/settings/Video/pages/TranscodingSettingsPage.vue')
			},
			{
				path: '/video/audioTranscoding',
				component: () =>
					import('src/pages/settings/Video/pages/AudioTranscodingPage.vue')
			},
			{
				path: '/video/encodingQuality',
				component: () =>
					import('src/pages/settings/Video/pages/EncodingQualityPage.vue')
			},
			{
				path: '/video/others',
				component: () => import('src/pages/settings/Video/pages/OthersPage.vue')
			},
			{
				path: '/video/optionsSelect/:type',
				component: () =>
					import('src/pages/settings/Video/pages/MobileOptionsSelectPage.vue')
			},
			{
				path: '/search',
				name: MENU_TYPE.Search,
				component: () => import('src/pages/settings/Search/SearchPage.vue')
			},
			{
				path: '/search/file',
				component: () => import('src/pages/settings/Search/FileSearch.vue')
			},
			{
				path: '/backup',
				name: MENU_TYPE.Backup,
				component: () => import('src/pages/settings/Backup2/BackupPage.vue')
			},
			{
				path: '/restore',
				name: MENU_TYPE.Restore,
				component: () => import('src/pages/settings/Backup2/RestorePage.vue')
			},
			{
				path: '/backup/:backupId/:snapshotId',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupSnapshotDetail.vue')
			},
			{
				path: '/backup/:backupId',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupDetail.vue')
			},
			{
				path: '/backup/create_backup/:backup_type/:backup_path?',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupNew.vue')
			},
			{
				path: '/developer',
				name: MENU_TYPE.Developer,
				component: () => import('src/pages/settings/Developer/IndexPage.vue')
			},
			{
				path: '/developer/log',
				component: () =>
					import('src/pages/settings/Developer/pages/LogPage.vue')
			},
			{
				path: '/developer/mirror',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorPage.vue')
			},
			{
				path: '/developer/mirror/endpoint',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorEndpointPage.vue')
			},
			{
				path: '/developer/images',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorImagesPage.vue')
			},
			{
				path: '/developer/environment',
				component: () =>
					import('src/pages/settings/Developer/pages/SystemEnvironmentPage.vue')
			},
			// {
			// 	path: '/backup/restore_existing_backup/:backupId?/:snapshotId?',
			// 	component: () =>
			// 		import('src/pages/settings/Backup2/pages/RestoreExistingBackup.vue')
			// },
			// {
			// 	path: '/backup/restore_custom_url',
			// 	component: () =>
			// 		import('src/pages/settings/Backup2/pages/RestoreCustomURL.vue')
			// },
			{
				path: '/backup/restoreOptions/:type',
				component: () =>
					import('src/pages/settings/Backup2/pages/MultiRestoreOptions.vue')
			},
			{
				path: '/backup/restore/:restoreId',
				component: () =>
					import('src/pages/settings/Backup2/pages/RestoreDetail.vue')
			},
			{
				path: '/ns',
				component: () => import('pages/settings/Notification/IndexPage.vue'),
				meta: {}
			},
			{
				path: '/job',
				component: () => import('pages/settings/Notification/Job/JobIndex.vue'),
				meta: {}
			},
			{
				path: '/job/:id',
				component: () =>
					import('pages/settings/Notification/Job/JobDetail.vue'),
				meta: {}
			},
			{
				path: 'sender',
				component: () =>
					import('pages/settings/Notification/Sender/SenderIndex.vue'),
				meta: {}
			},
			{
				path: '/sender/create',
				component: () =>
					import('pages/settings/Notification/Sender/SenderTemplate.vue'),
				meta: {}
			},
			{
				path: '/recipients',
				component: () =>
					import('pages/settings/Notification/Recipients/RecipientsIndex.vue'),
				meta: {}
			},
			{
				path: '/recipients/:id',
				component: () =>
					import('pages/settings/Notification/Recipients/RecipientsDetail.vue'),
				meta: {}
			},
			{
				path: '/notify',
				component: () =>
					import('pages/settings/Notification/Notify/NotifyIndex.vue'),
				meta: {}
			},
			{
				path: '/notify/:id',
				component: () =>
					import('pages/settings/Notification/Notify/NotifyRule.vue'),
				meta: {}
			},
			{
				path: '/template',
				component: () =>
					import('pages/settings/Notification/Template/TemplateIndex.vue'),
				meta: {}
			},
			{
				path: '/template/:id',
				component: () =>
					import('pages/settings/Notification/Template/TemplateContent.vue'),
				meta: {}
			}
		]
	},

	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/settings/ErrorNotFound.vue')
	}
];

// if (Platform.is.mobile) {
// 	routes.push({
// 		path: '',
// 		component: () => import('layouts/MainLayout.vue'),
// 		children: [
// 			{
// 				path: '',
// 				component: () => import('src/pages/Mobile/RootPage.vue')
// 			}
// 		]
// 	});
// } else {
// 	routes.push({
// 		path: '',
// 		component: () => import('layouts/MainLayout.vue'),
// 		children: [
// 			{
// 				path: '',
// 				component: () => import('src/pages/Person/IndexPage.vue')
// 			}
// 		]
// 	});
// }

export default routes;
