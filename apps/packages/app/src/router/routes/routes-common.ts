import { RouteRecordRaw } from 'vue-router';

const common: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/MainLayout.vue'),
		children: [
			{
				path: 'items/',
				component: () => import('pages/Items/ItemsPage.vue')
			},
			{
				path: 'items/:itemid',
				component: () => import('pages/Items/ItemsPage.vue')
			},
			{
				path: 'settings/',
				component: () => import('pages/Web/Settings/IndexPage.vue')
			},
			{
				path: 'invite-recipient/:org_id/:invite_id',
				component: () => import('pages/Items/InviteRecipient.vue')
			},
			{
				path: 'org/:org_mode',
				component: () => import('pages/Web/Orgs/OrgIndexPage.vue')
			},
			{
				path: 'org/:org_mode/:org_type',
				component: () => import('pages/Web/Orgs/OrgIndexPage.vue')
			},
			{
				path: 'settings/:mode',
				component: () => import('pages/Web/Settings/IndexPage.vue')
			},
			{
				path: 'generator/',
				component: () => import('pages/Web/Generator.vue')
			},
			{
				path: 'security/',
				component: () => import('pages/Web/Security/IndexPage.vue')
			}
		]
	},
	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/ErrorNotFound.vue')
	}
];

export default common;
