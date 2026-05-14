import { TRANSACTION_PAGE } from '../constant/constants';
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router';

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

export function handleMarketBack(
	router: Router,
	route: RouteLocationNormalizedLoaded,
	options?: { onlyAppRouteUseMenuItem?: boolean }
): boolean {
	if (window.history && window.history.state && window.history.state.back) {
		return false;
	}

	const onlyAppRouteUseMenuItem = options?.onlyAppRouteUseMenuItem ?? false;
	const menuItem =
		onlyAppRouteUseMenuItem && route.name !== TRANSACTION_PAGE.App
			? undefined
			: getMenuItem(route);

	if (menuItem) {
		if (menuItem === 'All') {
			router.replace({ name: TRANSACTION_PAGE.All });
		} else {
			router.replace({
				name: TRANSACTION_PAGE.CATEGORIES,
				params: { categories: menuItem }
			});
		}
		return true;
	}

	router.replace({ name: TRANSACTION_PAGE.All });
	return true;
}
