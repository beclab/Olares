import { useMiddlewareStore } from '@apps/control-hub/stores/Middleware';
import { ref, computed } from 'vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
import { useTerminalStore } from '@apps/control-hub/src/stores/TerminalStore';
import { capitalize } from 'lodash';
import { useSplitMenu } from '@apps/control-panel-common/src/stores/menu';
import { MIDDLEWARE_ICONS } from '@apps/control-panel-common/src/network/middleware';

export interface Breadcrumb {
	title: string;
	icon?: string;
	img?: string;
}
export const breadcrumbs = ref<Array<Breadcrumb>>([]);

export const updateBreadcrumbs = (data: Breadcrumb, init = false) => {
	if (init) {
		breadcrumbs.value = [];
	}
	breadcrumbs.value.push(data);
};

export const options = computed(() => [
	{
		key: 'application-spaces',
		label: t('BROWSE'),
		icon: 'sym_r_dvr',
		link: '/application-spaces'
	},
	{
		key: 'namespace',
		label: t('NAMESPACE'),
		icon: 'sym_r_markunread_mailbox',
		link: '/namespace'
	},
	{
		key: 'root',
		label: t('PODS'),
		icon: 'sym_r_deployed_code',
		link: '/root'
	}
]);

export const options2 = computed(() => [
	{
		key: 'storages',
		label: t('STORAGES'),
		icon: 'sym_r_hard_drive',
		link: '/storages'
	},
	{
		key: 'network-policies',
		label: t('NETWORKS'),
		icon: 'sym_r_sensors',
		link: '/network-policies'
	},
	{
		key: 'jobs',
		label: t('JOBS'),
		icon: 'sym_r_work',
		link: '/jobs'
	},
	{
		key: 'customresources',
		label: t('CRD_PL'),
		icon: 'sym_r_package_2',
		link: '/customresources'
	}
]);

export const options3 = computed(() => {
	const useMiddleware = useMiddlewareStore();
	const appDetail = useAppDetailStore();

	if (appDetail.isAdmin) {
		return useMiddleware.list.map((item) => ({
			key: item.type,
			label: t(capitalize(item.type)),
			icon: MIDDLEWARE_ICONS[item.type] || 'sym_r_dns',
			link: `/site-middleware/db/${item.type}`
		}));
	} else {
		useMiddleware.clearLocker();
		return [];
	}
});

export const options4 = computed(() => {
	const terminalStore = useTerminalStore();
	const data = [
		{
			key: 'terminal',
			label: t('OLARIS_TERMINAL'),
			icon: 'sym_r_terminal',
			link: `/terminal/${terminalStore.current_node || ''}`
		}
	];
	return data;
});

const optionAll = computed(() => {
	const appDetailsStore = useAppDetailStore();

	return appDetailsStore.isDemo
		? [...options.value, ...options2.value]
		: [
				...options.value,
				...options2.value,
				...options3.value,
				...options4.value
		  ];
});
export const active = ref(optionAll.value[0].key);

export const updateKey = (key) => {
	const splitMenu = useSplitMenu();

	splitMenu.changeStatus(key);
	active.value = key;
};

export const updateKeyFirstOption = () => {
	const target = optionAll.value[0];
	updateKey(target.key);
};
export const currentItem = computed(() =>
	optionAll.value.find((item) => item.key === active.value)
);

export const breadcrumbMap = {
	...options
};
