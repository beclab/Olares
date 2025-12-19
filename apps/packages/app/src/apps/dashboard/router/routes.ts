import { RouteRecordRaw } from 'vue-router';
import { ROUTE_NAME } from './const';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('@apps/dashboard/layouts/MainLayout.vue'),
		redirect: '/overview',
		children: [
			{
				path: '/overview',
				name: ROUTE_NAME.OVERVIEW,
				component: () =>
					import('@apps/dashboard/src/pages/Overview2/IndexPage.vue'),
				children: [
					{
						path: '/physical-resources/cluster',
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/PhysicalResource/IndexPage.vue')
					},
					{
						path: '/resources/:type',
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/ResourcePage.vue')
					},
					{
						path: '/overview/resources/:type',
						name: ROUTE_NAME.PHYSICAL_RESOURCE_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/ResourcePage.vue')
					},
					{
						path: '/overview/gpu/list',
						name: ROUTE_NAME.GPU_LIST,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/GPU/IndexPage.vue'),
						children: [
							{
								path: '/overview/gpu/:uuid/detail',
								name: ROUTE_NAME.GPUS_DETAILS,
								meta: {
									parentRouteName: ROUTE_NAME.GPU_LIST
								},
								component: () =>
									import(
										'@apps/dashboard/src/pages/Overview2/GPU/GPUsDetails.vue'
									)
							},
							{
								path: '/overview/task/:name/:pod_uid/detail',
								name: ROUTE_NAME.TASKS_DETAILS,
								meta: {
									parentRouteName: ROUTE_NAME.GPU_LIST
								},
								component: () =>
									import(
										'@apps/dashboard/src/pages/Overview2/GPU/TasksDetails.vue'
									)
							}
						]
					},
					{
						path: '/overview/network/detail',
						name: ROUTE_NAME.NETWORK_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import(
								'@apps/dashboard/src/pages/Overview2/Network/IndexPage.vue'
							)
					},
					{
						path: '/overview/cpu/detail',
						name: ROUTE_NAME.CPU_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/CPU/IndexPage.vue')
					},
					{
						path: '/overview/memory/detail',
						name: ROUTE_NAME.MEMORY_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/Memory/IndexPage.vue')
					},
					{
						path: '/overview/disk/detail',
						name: ROUTE_NAME.DISK_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/Disk/IndexPage.vue')
					},
					{
						path: '/overview/pods/detail',
						name: ROUTE_NAME.PODS_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/Pods/IndexPage.vue')
					},
					{
						path: '/overview/fan/detail',
						name: ROUTE_NAME.FAN_DETAIL,
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Overview2/Fan/IndexPage.vue')
					},
					{
						path: '/physical-resources/:node',
						meta: {
							parentRouteName: ROUTE_NAME.OVERVIEW
						},
						component: () =>
							import('@apps/dashboard/src/pages/Nodes/NodeMonitoring.vue')
					}
				]
			},
			{
				path: 'nodes',
				component: () => import('@apps/dashboard/src/pages/Nodes/IndexPage.vue')
			},
			{
				path: 'logs/:namespace/:name/:container',
				component: () => import('@apps/dashboard/src/pages/Logs/LogDetail.vue')
			},
			{
				path: 'applications',
				component: () =>
					import('@apps/dashboard/src/pages/Applications2/IndexPage.vue'),
				children: [
					{
						path: '/applications/:namespace/pods',
						name: 'podsList',
						component: () =>
							import('@apps/dashboard/src/pages/Applications/PodList.vue'),
						children: [
							{
								path: '/applications/pods/overview/:node/:namespace/:name/:createTime',
								component: () =>
									import(
										'@apps/dashboard/src/pages/Applications/ContainerList.vue'
									)
							}
						]
					}
				]
			},
			{
				path: '/other',
				component: () => import('@apps/dashboard/src/pages/Other/IndexPage.vue')
			}
			// {
			// 	path: '/analytics',
			// 	component: () =>
			// 		import('@apps/dashboard/src/pages/Analytics/IndexPage.vue'),
			// 	children: [
			// 		{
			// 			path: '/analytics/details/:websiteId',
			// 			component: () =>
			// 				import('@apps/dashboard/src/pages/Analytics/WebsiteDetails.vue')
			// 		}
			// 	]
			// }
		]
	},
	{
		path: '/container/logs/v2/:namespace/:name/:container',
		component: () =>
			import('@apps/control-panel-common/src/containers/Logs.vue')
	},
	{
		path: '/container/logs/:kind/:deployment/:container',
		component: () =>
			import('@apps/control-panel-common/src/containers/Logs.vue')
	},
	{
		path: '/:catchAll(.*)*',
		component: () => import('@apps/dashboard/src/pages/ErrorNotFound.vue')
	}
];

export default routes;
