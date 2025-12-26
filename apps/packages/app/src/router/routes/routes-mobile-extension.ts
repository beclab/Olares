import { RouteRecordRaw } from 'vue-router';

const mobileExtension: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/MobileMainLayout.vue'),
		children: [
			{
				path: '/home',
				meta: {
					tabIdentify: 'file',
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

export default mobileExtension;
