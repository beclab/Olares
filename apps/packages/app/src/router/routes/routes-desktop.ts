import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/DesktopMainLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('@apps/desktop/IndexRouter.vue'),
				meta: { requiresAuth: true }
			}
		]
	},
	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('@apps/desktop/ErrorNotFound.vue')
	}
];

export default routes;
