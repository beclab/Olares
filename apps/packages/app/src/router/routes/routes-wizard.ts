import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/WizardMainLayout.vue'),
		children: [
			{
				path: '',
				component: () => import('pages/Wizard/IndexPage.vue'),
				meta: { requiresNoAuth: true }
			}
		]
	},
	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/Wizard/ErrorNotFound.vue')
	}
];

export default routes;
