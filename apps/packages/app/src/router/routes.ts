import { RouteRecordRaw } from 'vue-router';
import { isPad } from 'src/utils/platform';

let routes: RouteRecordRaw[] = [];

if (process.env.APPLICATION == 'VAULT') {
	routes = [
		...routes,
		...require('./routes/routes-common').default,
		...require('./routes/routes-web').default
	];
} else if (process.env.APPLICATION == 'FILES') {
	routes = [...routes, ...require('./routes/routes-files').default];
} else if (process.env.APPLICATION == 'WISE') {
	routes = [...routes, ...require('./routes/routes-wise').default];
} else if (process.env.APPLICATION == 'SETTINGS') {
	routes = [...routes, ...require('./routes/routes-settings').default];
} else if (process.env.APPLICATION === 'LOGIN') {
	routes = [...routes, ...require('./routes/routes-login').default];
} else if (process.env.APPLICATION === 'WIZARD') {
	routes = [...routes, ...require('./routes/routes-wizard').default];
} else if (process.env.APPLICATION === 'EDITOR') {
	routes = [...routes, ...require('./routes/routes-profile-editor').default];
} else if (process.env.APPLICATION === 'PREVIEW') {
	routes = [...routes, ...require('./routes/routes-profile-preview').default];
} else if (process.env.APPLICATION === 'MARKET') {
	routes = [...routes, ...require('./routes/routes-market').default];
} else if (process.env.APPLICATION == 'DASHBOARD') {
	routes = [...require('./routes/routes-dashboard').default];
} else if (process.env.APPLICATION == 'CONTROL_HUB') {
	routes = [...require('./routes/routes-control-hub').default];
} else if (process.env.APPLICATION == 'STUDIO') {
	routes = [...require('./routes/routes-studio').default];
} else if (process.env.APPLICATION == 'SHARE') {
	routes = [...require('./routes/routes-share').default];
} else if (process.env.APPLICATION == 'LAREPASS') {
	if (process.env.PLATFORM == 'MOBILE') {
		if (isPad()) {
			routes = [...routes, ...require('./routes/routes-pad').default];
		} else if (process.env.DEV_PLATFORM_BEX) {
			routes = [...require('./routes-bex').default];
		} else {
			routes = [
				...routes,
				...require('./routes/routes-mobile').default,
				...require('./routes/routes-mobile-common').default,
				...require('./routes/routes-mobile-extension').default
			];
		}
	} else if (process.env.PLATFORM == 'DESKTOP') {
		routes = [
			...routes,
			...require('./routes/routes-larepass-desktop').default
		];
	}
} else if (process.env.APPLICATION === 'DESKTOP') {
	routes = [...routes, ...require('./routes/routes-desktop').default];
}

console.log(routes);

export default routes;
