import { TRANSACTION_PAGE } from '../constant/constants';
import type {
	RouteLocationNormalizedLoaded,
	RouteLocationRaw,
	Router
} from 'vue-router';

const MARKET_STACK_STATE_KEY = 'marketStack';

function getMarketStack(router: Router): string[] {
	const state = router.options.history.state;
	const raw = state ? state[MARKET_STACK_STATE_KEY] : undefined;
	return Array.isArray(raw)
		? (raw.filter((v) => typeof v === 'string') as string[])
		: [];
}

export function getMarketStackDepth(router: Router): number {
	return getMarketStack(router).length;
}

export function pushWithParent(router: Router, target: RouteLocationRaw): void {
	const marketStack = [
		...getMarketStack(router),
		router.currentRoute.value.fullPath
	];
	if (typeof target === 'string') {
		router.push({
			path: target,
			state: { [MARKET_STACK_STATE_KEY]: marketStack }
		});
		return;
	}
	router.push({ ...target, state: { [MARKET_STACK_STATE_KEY]: marketStack } });
}

function getMenuItem(route: RouteLocationNormalizedLoaded): string | undefined {
	const menuItemQuery = route.query.menuItem;
	const menuItem = Array.isArray(menuItemQuery)
		? menuItemQuery[0]
		: menuItemQuery;
	if (typeof menuItem === 'string' && menuItem.length > 0) {
		return menuItem;
	}
	return undefined;
}

const TOP_LEVEL_MENU_ITEMS = new Set<string>([
	TRANSACTION_PAGE.All,
	TRANSACTION_PAGE.MyTerminus,
	TRANSACTION_PAGE.Search,
	TRANSACTION_PAGE.Update
]);

function resolveByMenuItem(
	route: RouteLocationNormalizedLoaded
): RouteLocationRaw {
	const menuItem = getMenuItem(route);
	if (!menuItem) {
		return { name: TRANSACTION_PAGE.All };
	}
	if (TOP_LEVEL_MENU_ITEMS.has(menuItem)) {
		return { name: menuItem };
	}
	return {
		name: TRANSACTION_PAGE.CATEGORIES,
		params: { categories: menuItem }
	};
}

function resolveByCategoryParam(
	route: RouteLocationNormalizedLoaded,
	key: string
): RouteLocationRaw {
	const raw = route.params[key];
	const categories = Array.isArray(raw) ? raw[0] : raw;
	if (categories && categories !== 'All') {
		return {
			name: TRANSACTION_PAGE.CATEGORIES,
			params: { categories }
		};
	}
	return { name: TRANSACTION_PAGE.All };
}

function resolveByParentMeta(
	route: RouteLocationNormalizedLoaded
): RouteLocationRaw | undefined {
	const parentRouteName = route.meta.parentRouteName as string | undefined;
	if (!parentRouteName) {
		return undefined;
	}

	const target: RouteLocationRaw = { name: parentRouteName };

	const inheritParams = route.meta.inheritParams as string[] | undefined;
	if (inheritParams && inheritParams.length > 0) {
		const params: Record<string, string> = {};
		inheritParams.forEach((key) => {
			const raw = route.params[key];
			const value = Array.isArray(raw) ? raw[0] : raw;
			if (typeof value === 'string') {
				params[key] = value;
			}
		});
		target.params = params;
	}

	if (route.meta.inheritQuery) {
		target.query = { ...route.query };
	}

	return target;
}

export function handleMarketBack(
	router: Router,
	route: RouteLocationNormalizedLoaded,
	options?: { onlyAppRouteUseMenuItem?: boolean }
): boolean {
	const stack = getMarketStack(router);
	if (stack.length > 0) {
		const targetPath = stack[stack.length - 1];
		const resolved = router.resolve(targetPath);
		router.replace({
			path: resolved.path,
			query: resolved.query,
			hash: resolved.hash,
			state: { [MARKET_STACK_STATE_KEY]: stack.slice(0, -1) }
		});
		return true;
	}

	const parentTarget = resolveByParentMeta(route);
	if (parentTarget) {
		router.replace(parentTarget);
		return true;
	}

	let target: RouteLocationRaw;
	switch (route.name) {
		case TRANSACTION_PAGE.List:
			target = resolveByCategoryParam(route, 'categories');
			break;
		case TRANSACTION_PAGE.TOPIC:
			target = resolveByCategoryParam(route, 'category');
			break;
		case TRANSACTION_PAGE.App:
			target = resolveByMenuItem(route);
			break;
		default: {
			const onlyAppRouteUseMenuItem = options?.onlyAppRouteUseMenuItem ?? false;
			target = onlyAppRouteUseMenuItem
				? { name: TRANSACTION_PAGE.All }
				: resolveByMenuItem(route);
			break;
		}
	}

	router.replace(target);
	return true;
}
