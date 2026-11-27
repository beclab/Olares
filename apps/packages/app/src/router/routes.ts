import { RouteRecordRaw } from 'vue-router';

let routes: RouteRecordRaw[] = [];

if (process.env.APPLICATION == 'FILES') {
	routes = [...routes, ...require('./routes/routes-files').default];
} else if (process.env.APPLICATION == 'WISE') {
	routes = [...routes, ...require('./routes/routes-wise').default];
} else if (process.env.APPLICATION == 'SETTINGS') {
	routes = [...routes, ...require('./routes/routes-settings').default];
} else if (process.env.APPLICATION === 'WIZARD') {
	routes = [...routes, ...require('./routes/routes-wizard').default];
} else if (process.env.APPLICATION === 'EDITOR') {
	routes = [...routes, ...require('./routes/routes-profile-editor').default];
} else if (process.env.APPLICATION === 'MARKET') {
	routes = [...routes, ...require('./routes/routes-market').default];
} else if (process.env.APPLICATION == 'DASHBOARD') {
	routes = [...require('./routes/routes-dashboard').default];
} else if (process.env.APPLICATION == 'CONTROL_HUB') {
	routes = [...require('./routes/routes-control-hub').default];
} else if (process.env.APPLICATION === 'DESKTOP') {
	routes = [...routes, ...require('./routes/routes-desktop').default];
}

console.log(routes);

export default routes;
