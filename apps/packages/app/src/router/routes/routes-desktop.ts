import { RouteRecordRaw } from 'vue-router';
import { useApplicationStore } from 'src/stores/desktop/app';

function redirectLaunchpadIfColdStart(
	to: { path: string },
	from: { matched: unknown[] }
) {
	if (to.path === '/launchpad' && from.matched.length === 0) {
		return '/';
	}
}

function gateHomeFromLaunchpadShake(
	_to: unknown,
	from: { path: string }
): boolean | undefined {
	if (from.path !== '/launchpad') return;
	const appStore = useApplicationStore();
	if (!appStore.launchpadShakeMode) return;
	appStore.launchpadShakeMode = false;
	return false;
}

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/DesktopMainLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('@apps/desktop/IndexRouter.vue'),
				meta: { requiresAuth: true },
				beforeEnter: gateHomeFromLaunchpadShake
			},
			{
				path: 'launchpad',
				component: () => import('@apps/desktop/IndexRouter.vue'),
				meta: { requiresAuth: true, exit: true },
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
