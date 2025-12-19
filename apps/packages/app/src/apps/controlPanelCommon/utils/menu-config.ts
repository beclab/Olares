import { compact } from 'lodash';
import { RouteLocationNormalizedLoaded, useRoute, useRouter } from 'vue-router';

interface MenuIdIdFormatParams {
	layout?: string;
	namespace?: string;
	name?: string;
	container?: string;
	layout_pod?: string;
	layout_containers?: string;
}

export function menuIdFormat(
	params: MenuIdIdFormatParams,
	route: RouteLocationNormalizedLoaded
) {
	const defaultParams = route.params || {};
	const newParams = { ...defaultParams, ...params };
	const data = [
		newParams.layout,
		newParams.layout_pod,
		newParams.layout_containers,
		newParams.namespace,
		newParams.name,
		newParams.container
	];

	console.log('layout_pod', newParams);
	return compact(data).join('-');
}
