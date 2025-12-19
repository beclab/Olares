import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('src/layouts/profile/ProfileEditorLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('src/pages/profile/block/BlockIndexPage.vue')
			},
			{
				path: 'block/:id',
				name: 'blockEditor',
				component: () => import('src/pages/profile/block/BlockEditorPage.vue')
			}
		]
	},
	{
		path: '/preview',
		component: () => import('src/layouts/profile/ProfilePreviewLayout.vue')
	},
	{
		path: '/avatar',
		component: () => import('src/layouts/profile/AvatarChoosePage.vue')
	},

	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('src/pages/ErrorNotFound.vue')
	}
];

export default routes;
