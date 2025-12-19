import { RouteRecordRaw } from 'vue-router';
import { componentName } from './const';

export const podRoute = (pre: string) => ({
	path: `${pre}/pods/:node?/:name?`,
	component: () =>
		import('@apps/control-hub/src/pages/NamespacePods/IndexPage.vue'),
	children: [
		{
			path: `${pre}/pods/overview/:node/:namespace/:name/:createTime?`,
			component: () => import('@apps/control-hub/src/pages/Pods/Overview2.vue')
		},
		{
			path: `${pre}/pods/metadata/:node/:namespace/:name`,
			component: () => import('@apps/control-hub/src/pages/Pods/Metadata.vue')
		},
		{
			path: `${pre}/pods/containers/overview/:node/:namespace/:name/:container?`,
			component: () =>
				import('@apps/control-hub/src/pages/Containers/Overview.vue')
		},
		{
			path: `${pre}/pods/environments/:node/:namespace/:name`,
			component: () =>
				import('@apps/control-hub/src/pages/Pods/Environments.vue')
		},
		{
			path: `${pre}/pods/events/:node/:namespace/:name`,
			component: () => import('@apps/control-hub/src/pages/Pods/Events.vue')
		}
	]
});

const namespaceRoute = {
	path: '/namespace',
	component: () =>
		import('@apps/control-hub/src/pages/Namespace/IndexPage.vue'),
	children: [
		{
			path: '/namespace/:namespace/pods',
			name: 'podsList',
			component: () =>
				import('@apps/control-hub/src/pages/Applications/PodList.vue'),
			children: [
				{
					path: '/namespace/pods/overview/:node/:namespace/:name/:createTime',
					component: () =>
						import('@apps/control-hub/src/pages/Applications/ContainerList.vue')
				}
			]
		}
	]
};

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('@apps/control-hub/layouts/MainLayout.vue'),
		redirect: '/application-spaces',
		children: [
			{ ...namespaceRoute },
			{
				path: '/storages',
				component: () =>
					import('@apps/control-hub/src/pages/Storages/IndexPage.vue')
			},
			{
				path: '/jobs',
				component: () =>
					import('@apps/control-hub/src/pages/Jobs/IndexPage.vue'),
				children: [
					...podRoute('/jobs').children,
					{
						path: '/jobs/job/:namespace/:jobName/:jobUid',
						component: () =>
							import('@apps/control-hub/src/pages/Jobs/JobsDetails.vue'),
						name: componentName.JOB_DETAILS,
						children: [
							{
								path: '/job-inner/:jobName/:jobUid/pods/overview/:node/:namespace/:name/:createTime?',
								component: () =>
									import('@apps/control-hub/src/pages/Jobs/Pods.vue')
							}
						]
					},
					{
						path: '/jobs/cronjob/:namespace/:name/:jobUid',
						component: () =>
							import('@apps/control-hub/src/pages/Jobs/CronJobsDetails.vue'),
						meta: {
							module: 'cronjobs'
						}
					},
					{
						path: '/jobs/workloads/:kind/:namespace/:name/container/:container',
						component: () =>
							import('@apps/control-hub/src/pages/Containers/Overview.vue')
					}
				]
			},
			{
				path: '/nodes',
				component: () =>
					import('@apps/control-hub/src/pages/Nodes/IndexPage.vue'),
				children: [
					...podRoute('/nodes').children,
					{
						path: '/nodes/detail/:node',
						component: () =>
							import('@apps/control-hub/src/pages/Nodes/NodeDetail.vue')
					}
				]
			},
			{
				path: '/root',
				component: () =>
					import('@apps/control-hub/src/pages/Pods/IndexPage.vue'),
				children: [
					{
						path: '/root/pods/overview/:node/:namespace/:name/:createTime?',
						component: () =>
							import('@apps/control-hub/src/pages/Pods/Overview3.vue')
					}
				]
			},
			{
				path: '/application-spaces',
				component: () =>
					import('@apps/control-hub/src/pages/ApplicationSpaces/IndexPage.vue'),
				meta: {
					workloadActionHide: true
				},
				children: [
					{
						path: 'workloads/:namespace',
						name: 'Workloads',
						component: () =>
							import(
								'@apps/control-hub/src/pages/ApplicationSpaces/Workloads/Workloads.vue'
							),
						meta: {
							index: 0
						},
						children: [
							{
								path: '/application-spaces/workloads/:kind/:namespace/:pods_name/:pods_uid/:node/:name/:uid/:createTime/pods_overview',
								component: () =>
									import('@apps/control-hub/src/pages/Pods/Overview2.vue'),
								name: componentName.WORKLOAD_PODS
							},
							{
								path: '/application-spaces/workloads/:kind/:namespace/:pods_name/:pods_uid/:createTime?',
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
								path: '/application-spaces/workloads/:kind/:namespace/:name/container/:container',
								component: () =>
									import('@apps/control-hub/src/pages/Containers/Overview.vue')
							},
							{
								path: '/application-spaces/:kind/:namespace/:name/:pods_uid/services_overview',
								name: componentName.SERVICES,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Services/Detail.vue'
									),
								children: [
									{
										path: '/application-spaces/:kind/:namespace/:pods_name/:pods_uid/:node/:name/:uid/:createTime/services_pods_overview',
										component: () =>
											import('@apps/control-hub/src/pages/Pods/Overview4.vue'),
										name: componentName.SERVICES_PODS
									},
									{
										path: '/application-spaces/:kind/:namespace/:pods_name/:pods_uid/:createTime/services_pods_overview2',
										name: componentName.SERVICES_PODS2,
										component: () =>
											import(
												'@apps/control-hub/src/pages/ApplicationSpaces/Services/PodsData.vue'
											)
									}
								]
							},
							{
								path: '/application-spaces/configurations/:kind/:namespace/:name/:pods_uid/secrets_overview',
								name: componentName.SECRETS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/Secrets.vue'
									)
							},
							{
								path: '/application-spaces/configurations/:kind/:namespace/:name/:pods_uid/configmaps_overview',
								name: componentName.CONFIGMAPS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/Configmaps.vue'
									)
							},
							{
								path: '/application-spaces/configurations/:kind/:namespace/:name/:pods_uid/service-accounts_overview',
								name: componentName.SERVICE_ACCOUNTS,
								component: () =>
									import(
										'@apps/control-hub/src/pages/ApplicationSpaces/Configurations/ServiceAccounts.vue'
									)
							}
						]
					}
				]
			},
			{
				path: 'network-policies',
				component: () =>
					import('@apps/control-hub/src/pages/NetworkPolicies/IndexPage.vue'),
				children: [
					{
						path: 'detail/:namespace/:name/:uid',
						component: () =>
							import(
								'@apps/control-hub/src/pages/NetworkPolicies/PoliciesDetail.vue'
							)
					}
				]
			},
			{
				path: 'customresources',
				component: () =>
					import('@apps/control-hub/src/pages/Customresources/IndexPage.vue'),
				children: [
					{
						path: 'detail/:group/:version/:module/:uid',
						component: () =>
							import('@apps/control-hub/src/pages/Customresources/Detail.vue')
					}
				]
			},

			{
				path: 'site-middleware',
				component: () =>
					import('@apps/control-hub/src/pages/Middleware/IndexPage.vue'),
				children: [
					{
						path: '/site-middleware/db/:type',
						component: () =>
							import('@apps/control-hub/src/pages/Middleware/Overview.vue')
					}
				]
			},
			{
				path: '/terminal/:node?',
				name: componentName.TERMINAL_APP,
				component: () =>
					import('@apps/control-hub/src/pages/Terminal/IndexPage.vue')
			}
		]
	},
	{
		path: '/container/logs/v2/:namespace/:name/:container',
		component: () =>
			import('@apps/control-panel-common/src/containers/Logs.vue')
	},
	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('@apps/control-hub/pages/ErrorNotFound.vue')
	}
];

export default routes;
