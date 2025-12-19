import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/LoginMainLayout.vue'),
		children: [
			{
				path: '/',
				component: () => import('pages/Login/LoginPage.vue'),
				meta: { requiresNoAuth: true }
			}
		]
	},
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/Login/ErrorNotFound.vue')
	}
];

export default routes;
