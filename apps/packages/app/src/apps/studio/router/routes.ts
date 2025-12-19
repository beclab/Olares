import { ROUTE_NAME } from '@apps/studio/src/common/router-name';
import { RouteRecordRaw } from 'vue-router';
import {
	componentName,
	componentMeta
} from '@apps/control-hub/src/router/const';
import { STUDIO_ROUTE_NAME } from '@apps/studio/src/common/router-name';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('@apps/studio/layouts/MainLayout.vue'),
		beforeEnter: (to, _from, next) => {
			if (to.fullPath == '/') {
				return next({ path: '/home' });
			}
			next();
		},
		children: [
			{
				path: '/home',
				component: () => import('@apps/studio/pages/HomePage.vue')
			},
			{
				path: '/controlhub-layout',
				component: () =>
					import('@apps/studio/src/layouts/ControlHubLayout.vue'),
				children: [
					{
						path: '/app/:id/workloads/:namespace/overview',
						name: ROUTE_NAME.WORKLOAD,
						meta: {
							app: componentMeta.STUDIO,
							headerHide: true,
							workloadActionHide: true
						},
						component: () =>
							import('@apps/studio/src/pages/WorkloadsWrapper.vue'),
						children: [
							{
								path: '/app/:id/workloads/:kind/:namespace/:pods_name/:pods_uid/:node/:name/:uid/:createTime/pods_overview',
								component: () =>
									import('@apps/control-hub/src/pages/Pods/Overview2.vue'),
								name: componentName.WORKLOAD_PODS
							},
							{
								path: '/app/:id/workloads/:kind/:namespace/:pods_name/:pods_uid/:createTime?',
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Workloads/Detail.vue'
									),
								name: componentName.WORKLOAD_POD_TOP,
								meta: {
									index: 1
								}
							},
							{
								path: '/app/:id/workloads/:kind/:namespace/:name/container/:container',
								component: () =>
									import('@apps/control-hub/src/pages/Containers/Overview.vue')
							},
							{
								path: '/app/:id/:kind/:namespace/:name/:pods_uid/services_overview',
								name: componentName.SERVICES,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Services/Detail.vue'
									),
								children: [
									{
										path: '/app/:id/:kind/:namespace/:pods_name/:pods_uid/:node/:name/:uid/:createTime/services_pods_overview',
										component: () =>
											import('@apps/control-hub/src/pages/Pods/Overview4.vue'),
										name: componentName.SERVICES_PODS
									},
									{
										path: '/app/:id/:kind/:namespace/:pods_name/:pods_uid/:createTime/services_pods_overview2',
										name: componentName.SERVICES_PODS2,
										component: () =>
											import(
												'@apps/control-hub/src/pages/ApplicationSpaces/Services/PodsData.vue'
											)
									}
								]
							},
							{
								path: '/app/:id/configurations/:kind/:namespace/:name/:pods_uid/secrets_overview',
								name: componentName.SECRETS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/Secrets.vue'
									)
							},
							{
								path: '/app/:id/configurations/:kind/:namespace/:name/:pods_uid/configmaps_overview',
								name: componentName.CONFIGMAPS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/Configmaps.vue'
									)
							},
							{
								path: '/app/:id/configurations/:kind/:namespace/:name/:pods_uid/service-accounts_overview',
								name: componentName.SERVICE_ACCOUNTS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/ServiceAccounts.vue'
									)
							}
						]
					},
					{
						path: '/app/:id/:path*',
						name: STUDIO_ROUTE_NAME.APP_Edit,
						component: () => import('../pages/ApplicationPage.vue')
					}
				]
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

export default routes;
