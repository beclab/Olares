import { route } from 'quasar/wrappers';
import {
	createMemoryHistory,
	createRouter,
	createWebHashHistory,
	createWebHistory
} from 'vue-router';
import routes from './routes';
// import { getUiType } from '../utils/utils';
// import { useBexStore } from '../stores/bex';
// import { SupportRouters } from '../utils/rss-menu';

/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */

// Multiple Vue SPA iframes that all use HTML5 history mode share a single
// joint session history at the browser level (per WHATWG HTML spec). That
// makes `router.go(-1)` in one iframe replay the previous entry of the *tab*,
// which can belong to a sibling iframe — producing cross-iframe navigation
// interference. When we detect that this router instance is running inside an
// iframe, switch to an in-memory history so the iframe never touches the
// shared browser history.
const isRunningInsideIframe = (): boolean => {
	if (typeof window === 'undefined') return false;
	try {
		return window.self !== window.top;
	} catch {
		// Cross-origin parent access throws SecurityError; that itself
		// proves we are inside an iframe.
		return true;
	}
};

const resolveInitialIframePath = (base: string | undefined): string => {
	if (typeof window === 'undefined') return '/';
	const url = new URL(window.location.href);
	const pathto = url.searchParams.get('pathto');
	if (pathto) return pathto;
	let path = url.pathname + url.search + url.hash;
	if (base && base !== '/' && path.startsWith(base)) {
		path = path.slice(base.length) || '/';
		if (!path.startsWith('/')) path = '/' + path;
	}
	return path || '/';
};

const normalizeBase = (base: string | undefined): string => {
	if (!base || base === '/') return '';
	return base.endsWith('/') ? base.slice(0, -1) : base;
};

export default route(function (/* { store, ssrContext } */) {
	const base =
		process.env.MODE === 'ssr' ? void 0 : process.env.VUE_ROUTER_BASE;

	const insideIframe = isRunningInsideIframe();

	const createHistory = insideIframe
		? createMemoryHistory
		: process.env.VUE_ROUTER_MODE === 'history'
		? createWebHistory
		: createWebHashHistory;

	const Router = createRouter({
		scrollBehavior: () => ({ left: 0, top: 0 }),
		routes,

		// Leave this as is and make changes in quasar.conf.js instead!
		// quasar.conf.js -> build -> vueRouterMode
		// quasar.conf.js -> build -> publicPath
		history: createHistory(base)
	});

	if (insideIframe) {
		const initialPath = resolveInitialIframePath(base);
		if (initialPath && initialPath !== '/') {
			const resolvedTarget = Router.resolve(initialPath);
			let unregister: (() => void) | undefined;
			const settle = () => {
				if (unregister) {
					unregister();
					unregister = undefined;
				}
			};

			unregister = Router.beforeEach((to) => {
				if (to.fullPath === resolvedTarget.fullPath) {
					settle();
					return true;
				}
				settle();
				return resolvedTarget.fullPath;
			});
		}

		// Memory history doesn't update the iframe's address bar, which breaks
		// any code that reads window.location.href to construct a "share" /
		// "open in a new tab" URL. Mirror the active route into the iframe's
		// address bar via `replaceState` (NOT pushState — that would
		// re-pollute the joint session history and reintroduce the
		// cross-iframe back/forward interference), and additionally notify
		// the parent of the new URL so its window_infos[i].redirectUrl can
		// be updated. We do the postMessage here, in the central router,
		// because most sub-apps (dashboard, control-hub, settings, ...)
		// never call `bytetrade.observeUrlChange.childPostMessage`
		// themselves — only MarketMainLayout does — so without this
		// explicit message the parent never learns about iframe navigations
		// and any "open in new tab" button falls back to the iframe's root.
		if (typeof window !== 'undefined') {
			const basePath = normalizeBase(base);
			Router.afterEach((to) => {
				const fullUrl = window.location.origin + basePath + to.fullPath;
				try {
					if (window.location.href !== fullUrl) {
						window.history.replaceState(window.history.state, '', fullUrl);
					}
				} catch {
					/* ignore — best-effort URL bar sync */
				}
				try {
					if (window.parent && window.parent !== window) {
						window.parent.postMessage(
							{ type: 'locationHref', message: window.location.href },
							'*'
						);
					}
				} catch {
					/* ignore — best-effort parent notification */
				}
			});
		}
	}

	return Router;
});
