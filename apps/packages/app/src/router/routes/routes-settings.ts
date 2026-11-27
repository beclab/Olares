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
				name: 'person',
				component: () => import('src/pages/settings/Person/IndexPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Root }
			},
			{
				path: '/olares_space',
				component: () =>
					import('src/pages/settings/Person/OlaresSpacePage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: 'loginHistory/:name?',
				name: 'loginHistory',
				component: () => import('pages/settings/Person/LoginHistoryPage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/active_session',
				component: () =>
					import('src/pages/settings/Person/VaultActivieSession.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/device/:deviceId',
				component: () =>
					import('src/pages/settings/Person/DeviceDetailPage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/sso_token',
				component: () => import('src/pages/settings/Person/SSOToken.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/authority',
				component: () => import('src/pages/settings/Person/AuthorityPage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/feedback',
				component: () => import('src/pages/settings/Person/FeedbackPage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/acknowledgment',
				component: () =>
					import('src/pages/settings/Person/AcknowledgmentPage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/hardware',
				name: 'hardware',
				component: () => import('src/pages/settings/Person/HardwarePage.vue'),
				meta: { parentRouteName: 'person' }
			},
			{
				path: '/hardware/:command',
				component: () =>
					import('src/pages/settings/Person/HardwareCommandPage.vue'),
				meta: { parentRouteName: 'hardware' }
			},
			{
				path: '/user',
				name: MENU_TYPE.Users,
				component: () => import('src/pages/settings/Account/IndexPage.vue')
			},
			{
				path: '/user/info/:name?',
				component: () =>
					import('src/pages/settings/Account/pages/UserInfoPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Users }
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
				name: 'applicationInfo',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationDetailPage.vue'
					),
				meta: { parentRouteName: MENU_TYPE.Application }
			},
			{
				path: '/application/ports/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationPortsPage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/application/secret/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationSecretPage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/application/environment/manager/:name?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationEnvironmentPage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/application/acl/:name',
				name: 'appAcl',
				component: () =>
					import('src/pages/settings/Application/pages/ApplicationAclPage.vue'),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},

			{
				path: '/application/entrance/:name/:entrance',
				name: 'applicationEntrance',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationEntrancePage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/application/domain/:name/:entrance/policies',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationEntrancePoliciesPage.vue'
					),
				meta: {
					parentRouteName: 'applicationEntrance',
					inheritParams: ['name', 'entrance']
				}
			},
			{
				path: '/application/domain/:name/:entrance/:shared?',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationDomainPage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/application/:name/provider',
				component: () =>
					import(
						'src/pages/settings/Application/pages/ApplicationProviderPage.vue'
					),
				meta: { parentRouteName: 'applicationInfo', inheritParams: ['name'] }
			},
			{
				path: '/integration',
				name: MENU_TYPE.Integration,
				component: () => import('src/pages/settings/Integration/IndexPage.vue')
			},
			{
				path: '/integration/login/space',
				component: () =>
					import('src/pages/settings/Integration/Space/LoginPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Integration }
			},
			{
				path: '/integration/detail/space',
				name: 'integrationSpaceDetail',
				component: () =>
					import('src/pages/settings/Integration/AccountDetail.vue'),
				meta: { parentRouteName: MENU_TYPE.Integration }
			},
			{
				path: '/integration/cookie',
				name: 'integrationCookie',
				component: () =>
					import(
						'src/pages/settings/Integration/pages/CookieManagementPage.vue'
					),
				meta: { parentRouteName: MENU_TYPE.Integration }
			},
			{
				path: '/integration/cookie/:mainDomain',
				component: () =>
					import('src/pages/settings/Integration/pages/CookieDetailsPage.vue'),
				meta: { parentRouteName: 'integrationCookie' }
			},
			{
				path: '/integration/detail/space/:address',
				component: () =>
					import('src/pages/settings/Integration/TerminusSpaceNFTPage.vue'),
				meta: { parentRouteName: 'integrationSpaceDetail' }
			},
			{
				path: '/integration/common/detail/:type/:name',
				component: () =>
					import('src/pages/settings/Integration/IntegrationDetailPage.vue'),
				meta: { parentRouteName: 'integrationAccountList' }
			},
			{
				path: '/integration/add',
				name: 'integrationAdd',
				component: () =>
					import('src/pages/settings/Integration/pages/IntegrationAddPage.vue'),
				meta: { parentRouteName: 'integrationAccountList' }
			},
			{
				path: '/integration/accountList',
				name: 'integrationAccountList',
				component: () =>
					import(
						'src/pages/settings/Integration/pages/IntegrationListPage.vue'
					),
				meta: { parentRouteName: MENU_TYPE.Integration }
			},
			{
				path: '/integration/account/add',
				component: () =>
					import(
						'src/pages/settings/Integration/pages/AwsAddIntegrationPage.vue'
					),
				meta: { parentRouteName: 'integrationAdd' }
			},
			{
				path: '/integration/smbAccount',
				component: () =>
					import('src/pages/settings/Integration/pages/SMBAccountPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Integration }
			},
			{
				path: '/vpn',
				name: MENU_TYPE.VPN,
				component: () => import('src/pages/settings/Vpn/VPNPage.vue')
			},
			{
				path: '/vpn/active_headscale',
				component: () => import('src/pages/settings/Vpn/HeadScale.vue'),
				meta: { parentRouteName: MENU_TYPE.VPN }
			},

			{
				path: '/network',
				name: MENU_TYPE.Network,
				component: () => import('src/pages/settings/Network/NetworkPage.vue')
			},
			{
				path: '/network/reverse_proxy',
				component: () =>
					import('src/pages/settings/Network/ReverseProxyPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Network }
			},
			{
				path: '/network/host',
				component: () => import('src/pages/settings/Network/HostPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Network }
			},
			{
				path: '/network/overlay_gateway',
				component: () =>
					import('src/pages/settings/Network/OverlayGatewayPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Network }
			},
			{
				path: '/gpu',
				name: MENU_TYPE.GPU,
				component: () => import('src/pages/settings/GPU/AcceleratorPage.vue')
			},
			{
				path: '/gpu/manage/:node',
				component: () => import('src/pages/settings/GPU/ManageNodePage.vue'),
				meta: { parentRouteName: MENU_TYPE.GPU }
			},
			{
				path: '/video',
				name: MENU_TYPE.Video,
				component: () => import('src/pages/settings/Video/IndexPage.vue')
			},
			{
				path: '/video/hardwareAcceleration',
				component: () =>
					import('src/pages/settings/Video/pages/HardwareAccelerationPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/encodingScheme',
				component: () =>
					import('src/pages/settings/Video/pages/EncodingSchemePage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/transcodingSettings',
				component: () =>
					import('src/pages/settings/Video/pages/TranscodingSettingsPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/audioTranscoding',
				component: () =>
					import('src/pages/settings/Video/pages/AudioTranscodingPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/encodingQuality',
				component: () =>
					import('src/pages/settings/Video/pages/EncodingQualityPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/others',
				component: () =>
					import('src/pages/settings/Video/pages/OthersPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/video/optionsSelect/:type',
				component: () =>
					import('src/pages/settings/Video/pages/MobileOptionsSelectPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Video }
			},
			{
				path: '/search',
				name: MENU_TYPE.Search,
				component: () => import('src/pages/settings/Search/SearchPage.vue')
			},
			{
				path: '/search/file',
				name: 'searchFile',
				component: () => import('src/pages/settings/Search/FileSearch.vue'),
				meta: { parentRouteName: MENU_TYPE.Search }
			},
			{
				path: '/search/failed-files',
				component: () =>
					import('src/pages/settings/Search/SearchFailureListPage.vue'),
				meta: { parentRouteName: 'searchFile' }
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
				path: '/restore/restoreOptions/:type',
				component: () =>
					import('src/pages/settings/Backup2/pages/MultiRestoreOptions.vue'),
				meta: { parentRouteName: MENU_TYPE.Restore }
			},
			{
				path: '/restore/task/:restoreId',
				component: () =>
					import('src/pages/settings/Backup2/pages/RestoreDetail.vue'),
				meta: { parentRouteName: MENU_TYPE.Restore }
			},
			{
				path: '/backup/:backupId/:snapshotId',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupSnapshotDetail.vue'),
				meta: { parentRouteName: 'backupDetail' }
			},
			{
				path: '/backup/:backupId',
				name: 'backupDetail',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupDetail.vue'),
				meta: { parentRouteName: MENU_TYPE.Backup }
			},
			{
				path: '/backup/create_backup/:backup_type/:backup_path?',
				component: () =>
					import('src/pages/settings/Backup2/pages/BackupNew.vue'),
				meta: { parentRouteName: MENU_TYPE.Backup }
			},
			{
				path: '/developer',
				name: MENU_TYPE.Developer,
				component: () => import('src/pages/settings/Developer/IndexPage.vue')
			},
			{
				path: '/developer/log',
				component: () =>
					import('src/pages/settings/Developer/pages/LogPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Developer }
			},
			{
				path: '/developer/mirror',
				name: 'developerMirror',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Developer }
			},
			{
				path: '/developer/mirror/endpoint',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorEndpointPage.vue'),
				meta: { parentRouteName: 'developerMirror' }
			},
			{
				path: '/developer/images',
				component: () =>
					import('src/pages/settings/Developer/pages/MirrorImagesPage.vue'),
				meta: { parentRouteName: MENU_TYPE.Developer }
			},
			{
				path: '/developer/environment',
				component: () =>
					import(
						'src/pages/settings/Developer/pages/SystemEnvironmentPage.vue'
					),
				meta: { parentRouteName: MENU_TYPE.Developer }
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
				path: '/ns',
				name: 'notification',
				component: () => import('pages/settings/Notification/IndexPage.vue'),
				meta: {}
			},
			{
				path: '/job',
				name: 'notificationJob',
				component: () => import('pages/settings/Notification/Job/JobIndex.vue'),
				meta: { parentRouteName: 'notification' }
			},
			{
				path: '/job/:id',
				component: () =>
					import('pages/settings/Notification/Job/JobDetail.vue'),
				meta: { parentRouteName: 'notificationJob' }
			},
			{
				path: 'sender',
				name: 'notificationSender',
				component: () =>
					import('pages/settings/Notification/Sender/SenderIndex.vue'),
				meta: { parentRouteName: 'notification' }
			},
			{
				path: '/sender/create',
				component: () =>
					import('pages/settings/Notification/Sender/SenderTemplate.vue'),
				meta: { parentRouteName: 'notificationSender' }
			},
			{
				path: '/recipients',
				name: 'notificationRecipients',
				component: () =>
					import('pages/settings/Notification/Recipients/RecipientsIndex.vue'),
				meta: { parentRouteName: 'notification' }
			},
			{
				path: '/recipients/:id',
				component: () =>
					import('pages/settings/Notification/Recipients/RecipientsDetail.vue'),
				meta: { parentRouteName: 'notificationRecipients' }
			},
			{
				path: '/notify',
				name: 'notificationNotify',
				component: () =>
					import('pages/settings/Notification/Notify/NotifyIndex.vue'),
				meta: { parentRouteName: 'notification' }
			},
			{
				path: '/notify/:id',
				component: () =>
					import('pages/settings/Notification/Notify/NotifyRule.vue'),
				meta: { parentRouteName: 'notificationNotify' }
			},
			{
				path: '/template',
				name: 'notificationTemplate',
				component: () =>
					import('pages/settings/Notification/Template/TemplateIndex.vue'),
				meta: { parentRouteName: 'notification' }
			},
			{
				path: '/template/:id',
				component: () =>
					import('pages/settings/Notification/Template/TemplateContent.vue'),
				meta: { parentRouteName: 'notificationTemplate' }
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
