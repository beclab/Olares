export default [
	{
		path: '/',
		component: () => import('layouts/files/MainLayout.vue'),
		beforeEnter: (
			to: any,
			_from: any,
			next: (arg0?: { path: string } | undefined) => void
		) => {
			if (to.fullPath == '/') {
				return next({ path: '/Files/Home/' });
			}

			next();
		},
		children: [
			{
				path: 'Files/:path*',
				name: 'Files',
				meta: {
					requiresAuth: true
				}
			},
			{
				path: 'External/:path*',
				name: 'External',
				meta: {
					requiresAuth: true
				}
			},
			{
				path: 'Data/:path*',
				name: 'Data',
				meta: {
					requiresAuth: true
				}
			},
			{
				path: 'Cache/:path*',
				name: 'Cache',
				meta: {
					requiresAuth: true
				}
			},

			{
				path: 'Seahub/:path*',
				name: 'Seahub',
				meta: {
					requiresAuth: true
				}
			},
			{
				path: 'Drive/:path*',
				name: 'Drive',
				meta: {
					requiresAuth: true
				}
			},
			{
				path: 'Share/:path*',
				name: 'Share',
				meta: {
					requiresAuth: true
				}
			}
			// {
			// 	path: 'ShareWith/:path*',
			// 	name: 'ShareWith',
			// 	meta: {
			// 		requiresAuth: true
			// 	}
			// },
			// {
			// 	path: '/mobile/home',
			// 	meta: {
			// 		tabIdentify: 'file',
			// 		minimizeApp: 'true'
			// 	},
			// 	component: () => import('pages/Mobile/file/FileRootPage.vue')
			// }
		]
	},
	{
		path: '/',
		component: () => import('layouts/files/MainLayout.vue'),
		children: [
			{
				path: '/repo/:repo',
				meta: {
					tabIdentify: 'file'
				},
				component: () => import('pages/Mobile/file/FilesRepoPage.vue')
			},
			{
				path: '/Files/',
				component: () => import('src/pages/Mobile/file/FileRootPage.vue')
			},
			{
				path: '/files',
				component: () => import('src/pages/Mobile/file/FileRootPage.vue')
			}
		]
	}
];
