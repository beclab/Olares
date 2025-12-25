export default [
	{
		path: '/',
		component: () => import('layouts/FilesMainLayout.vue'),
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
			// }
		]
	}
];
