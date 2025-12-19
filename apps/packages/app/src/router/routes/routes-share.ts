import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/ShareMainLayout.vue'),
		children: [
			{
				path: '/sharable-link/:share_id?/',
				component: () => import('src/pages/Share/IndexPage.vue'),
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
