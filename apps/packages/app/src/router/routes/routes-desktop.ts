import { RouteRecordRaw } from 'vue-router';

function redirectLaunchpadIfColdStart(
	to: { path: string },
	from: { matched: unknown[] }
) {
	if (to.path === '/launchpad' && from.matched.length === 0) {
		return '/';
	}
}

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/DesktopMainLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('@apps/desktop/IndexRouter.vue'),
				meta: { requiresAuth: true }
			},
			{
				path: 'launchpad',
				component: () => import('@apps/desktop/IndexRouter.vue'),
				meta: { requiresAuth: true },
				beforeEnter: redirectLaunchpadIfColdStart
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
