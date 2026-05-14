import { PodsMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import { defineStore } from 'pinia';
import { get, isEmpty } from 'lodash';
import { t, i18n } from 'src/boot/control-hub-i18n';
import { getLocalTime } from '@apps/control-panel-common/src/utils';

export const UsePod = defineStore('poddata', {
	state: (): { data: any } => ({
		data: undefined
	}),
	getters: {
		detail: (state) => {
			const _locale = i18n.global.locale.value;
			return getAttrs(state.data);
		}
	},
	actions: {
		setDetail(data: any) {
			const detail = PodsMapper(data);
			this.data = detail;
		}
	}
});

function getAttrs(detail: any = {}) {
	const namespace = detail.namespace;
	if (isEmpty(detail)) return null;

	const { status, restarts, type } = detail.podStatus;

	return [
		// {
		//   name: t('CLUSTER'),
		//   value: cluster,
		// },
		{
			name: t('PROJECT'),
			value: namespace
		},
		{
			name: t('APP'),
			value: detail.app
		},
		{
			name: t('STATUS'),
			value: t(status),
			type: type
		},
		{
			name: t('POD_IP_ADDRESS'),
			value: detail.podIp
		},
		{
			name: t('NODE_NAME'),
			value: detail.node
		},
		{
			name: t('NODE_IP_ADDRESS'),
			value: detail.nodeIp
		},
		{
			name: t('RESTART_PL'),
			value: restarts
		},
		{
			name: t('QOS_CLASS'),
			value: get(detail, 'status.qosClass')
		},
		{
			name: t('CREATION_TIME_TCAP'),
			value: getLocalTime(detail.createTime).format('YYYY-MM-DD HH:mm:ss')
		},
		{
			name: t('CREATOR'),
			value: detail.creator
		}
	];
}
