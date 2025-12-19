import { RouteRecordRaw } from 'vue-router';

const web: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/WebLoginLayout.vue'),
		children: [
			{
				path: 'binding',
				component: () => import('pages/Web/BindingPage.vue')
			},
			{
				path: 'import_mnemonic',
				component: () => import('pages/Web/InputMnemonicPage.vue')
			},
			{
				path: 'setUnlockPassword',
				component: () => import('pages/Web/SetUnlockPwdPage.vue')
			},
			{
				path: 'unlock',
				component: () => import('pages/Web/UnlockPage.vue')
			}
		]
	},
	{
		path: '/error',
		component: () => import('src/pages/ErrorNotFound.vue')
	}
];
export default web;
