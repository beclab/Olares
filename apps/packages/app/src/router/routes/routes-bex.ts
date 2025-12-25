import { RouteRecordRaw } from 'vue-router';

const bex: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/PluginMainLayout.vue'),
		children: [
			{
				path: 'home',
				meta: {
					tabIdentify: 'home'
				},
				component: () => import('pages/Plugin/indexPage.vue')
			},
			{
				path: 'items',
				meta: {
					tabIdentify: 'vault'
				},
				component: () => import('pages/Mobile/vault/ItemIndex.vue')
			},
			{
				path: 'items/:itemid',
				component: () => import('pages/Mobile/vault/ItemView.vue')
			},
			{
				path: 'setting',
				component: () => import('pages/Mobile/setting/SettingIndex.vue')
			}
		]
	},
	{
		path: '/inject/google/search',
		component: () => import('pages/Plugin/GoogleSearch.vue')
	},
	{
		path: '/inject/youtube/search',
		component: () => import('pages/Plugin/YoutebuSearch.vue')
	}
];

export default bex;
