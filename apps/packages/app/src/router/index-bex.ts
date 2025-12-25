import { route } from 'quasar/wrappers';
import {
	createMemoryHistory,
	createRouter,
	createWebHashHistory,
	createWebHistory
} from 'vue-router';
import routes from './routes-bex';
import { ROUTE_CONST } from './route-const';
import { isChromeExtension } from 'src/utils/bex/link';

/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */

export default route(function (/* { store, ssrContext } */) {
	const createHistory =
		process.env.VUE_ROUTER_MODE === 'history'
			? createWebHistory
			: createWebHashHistory;

	const Router = createRouter({
		scrollBehavior: () => ({ left: 0, top: 0 }),
		routes,

		// Leave this as is and make changes in quasar.conf.js instead!
		// quasar.conf.js -> build -> vueRouterMode
		// quasar.conf.js -> build -> publicPath
		history: createHistory(
			process.env.MODE === 'ssr' ? void 0 : process.env.VUE_ROUTER_BASE
		)
	});

	Router.beforeEach(async (to, from, next) => {
		console.log(
			'beforeEach',
			isChromeExtension(),
			to.path.startsWith('/home'),
			to
		);
		if (isChromeExtension() && to.path.startsWith('/home')) {
			next({ name: ROUTE_CONST.OPTIONS_ACCOUNT });
			return;
		}
		next();
	});

	return Router;
});
