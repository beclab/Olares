import { getNameSpaceServicesList } from '@apps/control-panel-common/src/network';
import { PodsParam } from '@apps/control-panel-common/src/network/network';
import Services from '@apps/control-hub/src/assets/Services.svg';
import { t } from '@apps/control-hub/src/boot/i18n';
import { componentName } from '@apps/control-hub/src/router/const';

const group = {
	label: 'SERVICES',
	value: 'services',
	componentName: componentName.SERVICES
};
export const ServiceOpeneds = [group.value];

export const getServicesData = (params: PodsParam & { namespace: string }) => {
	return getNameSpaceServicesList(params);
};

export const servicesDataFormatter = (res: any, namespace: string) => {
	const data = res.data.items.map((item: any) => {
		return {
			id: item.metadata.uid,
			title: item.metadata.name,
			img: Services,
			route: {
				name: group.componentName,
				params: {
					kind: group.value,
					namespace,
					name: item.metadata.name,
					pods_uid: item.metadata.uid
				}
			}
		};
	});

	const list = data.length
		? [
				{
					title: group.label,
					id: group.value,
					selectable: false,
					children: data
				}
		  ]
		: [];

	return list;
};
