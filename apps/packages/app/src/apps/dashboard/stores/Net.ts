import { IP_METHOD_OPTION, SystemIFSItem } from './../types/network';
import { defineStore } from 'pinia';
import { getSystemIFS } from '@apps/dashboard/src/network';
import { t } from '@apps/dashboard/boot/i18n';
import { Locker } from '../types/main';

enum IP_METHOD {
	IP_METHOD_AUTO = 'DHCP',
	IP_METHOD_MANUAL = '静态IP',
	IP_METHOD_NONE = '无IP'
}

interface contentListItem {
	name: string;
	value: string;
}
interface NetStore {
	list: Array<
		SystemIFSItem & {
			contentList: contentListItem[];
			contentList2: contentListItem[];
		}
	>;
	loading: boolean;
	locker: Locker;
}
export const useNetStore = defineStore('NetStore', {
	state: (): NetStore => ({
		list: [],
		loading: false,
		locker: undefined
	}),

	getters: {},
	actions: {
		init() {
			this.flowDetection();
			this.getNetList(true);
		},
		flowDetection() {
			this.getNetList(false, false);
		},
		async getNetList(autofresh = false, testConnectivity = true) {
			if (!autofresh) {
				this.loading = true;
			}
			try {
				const params = {
					testConnectivity
				};
				const res = await getSystemIFS(params);
				const resSortData = sortByIsHost(res.data);

				this.list = resSortData.map((item) => ({
					...item,
					contentList: [
						{
							name: t('NET_OP.IP_ACQUISITION_METHOD'),
							value: item.method
								? t(`NET_OP.${IP_METHOD_OPTION[item.method]}`)
								: '-'
						},
						{ name: t('OWNED_NODE'), value: item.hostname },
						{
							name: t('NET_OP.NETWORK_CONFIGURATION'),
							value: `MTU${item.mtu}`
						},
						{ name: t('NET_OP.IPV4_ADDRESS'), value: item.ip },
						{
							name: t('NET_OP.IPV4_SUBNET_MASK'),
							value: item.ipv4Mask
						},
						{ name: t('NET_OP.IPV4_GATEWAY_ADDRESS'), value: item.ipv4Gateway },
						{ name: t('NET_OP.IPV4_DNS'), value: item.ipv4DNS },
						{
							name: t('NET_OP.IPV4_NETWORK_STATUS'),
							value:
								item.ip && item.internetConnected
									? t('NET_OP.CONN')
									: t('NET_OP.DISCONNECT'),
							icon: netStatusInfo(
								!item.ip ? 0 : !item.internetConnected ? 1 : 2
							)
						}
					],
					contentList2: [
						{ name: t('NET_OP.IPV6_ADDRESS'), value: item.ipv6Address },
						{
							name: t('NET_OP.IPV6_SUBNET_MASK'),
							value: item.ipv4Mask
						},
						{ name: t('NET_OP.IPV6_GATEWAY_ADDRESS'), value: item.ipv6Gateway },
						{ name: t('NET_OP.IPV6_DNS'), value: item.ipv6DNS },
						{
							name: t('NET_OP.IPV6_NETWORK_STATUS'),
							value:
								item.ipv6Address && item.ipv6Connectivity
									? t('NET_OP.CONN')
									: t('NET_OP.DISCONNECT'),
							icon: netStatusInfo(
								!item.ipv6Address ? 0 : !item.ipv6Connectivity ? 1 : 2
							)
						}
					]
				}));
				this.refresh();
			} catch (error) {
				this.loading = false;
			}
			this.loading = false;
		},
		refresh() {
			this.clearLocker();
			this.locker = setTimeout(() => {
				this.getNetList(true);
			}, 5000);
		},
		clearLocker() {
			this.locker && clearTimeout(this.locker);
		}
	}
});

function netStatusInfo(statusIndex: number) {
	const networkOptions = [
		{
			name: 'sym_r_do_not_disturb_on',
			color: 'ink-3'
		},
		{
			name: 'sym_r_language',
			color: 'ink-3'
		},
		{
			name: 'sym_r_language',
			color: 'light-blue-default'
		}
	];

	return networkOptions[statusIndex];
}

function sortByIsHost(arr) {
	return arr.sort((a, b) => {
		if (a.isHostIp && !b.isHostIp) return -1;
		if (!a.isHostIp && b.isHostIp) return 1;
		return 0;
	});
}
