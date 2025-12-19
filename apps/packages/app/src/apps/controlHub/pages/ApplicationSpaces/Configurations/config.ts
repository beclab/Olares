import {
	getConfigmaps,
	getSecrets,
	getServiceaccounts
} from '@apps/control-panel-common/src/network';
import { PodsParam } from '@apps/control-panel-common/src/network/network';
import { useRoute } from 'vue-router';
import Secrets from '@apps/control-hub/src/assets/Secrets.svg';
import Configmaps from '@apps/control-hub/src/assets/Configmaps.svg';
import ServiceAccounts from '@apps/control-hub/src/assets/ServiceAccounts.svg';
import { componentName } from '@apps/control-hub/src/router/const';
import { t } from '@apps/control-hub/src/boot/i18n';

const data = [
	{
		label: 'SECRET_PL',
		value: 'secret',
		icon: Secrets,
		componentName: componentName.SECRETS
	},
	{
		label: 'CONFIGMAP_PL',
		value: 'configmap',
		icon: Configmaps,
		componentName: componentName.CONFIGMAPS
	},
	{
		label: 'SERVICE_ACCOUNT_PL',
		value: 'service-account',
		icon: ServiceAccounts,
		componentName: componentName.SERVICE_ACCOUNTS
	}
];

export const configurationsOpeneds = data.map((item) => item.value);

export const getConfigurationsData = (
	params: PodsParam & { namespace: string }
) => {
	return Promise.all([
		getSecrets(params),
		getConfigmaps(params),
		getServiceaccounts(params)
	]);
};

export const configurationsDataFormatter = (res: any, namespace: string) => {
	let dataTemp = data.map((child, index) => ({
		title: child.label,
		id: child.value,
		selectable: false,
		children: res[index].data.items.map((item: any) => {
			return {
				id: item.metadata.uid,
				title: item.metadata.name,
				img: child.icon,
				lazy: false,
				route: {
					name: child.componentName,
					params: {
						kind: child.value,
						namespace,
						name: item.metadata.name,
						pods_uid: item.metadata.uid
					}
				}
			};
		})
	}));
	dataTemp = dataTemp.filter((item) => item.children.length > 0);
	return dataTemp;
};
