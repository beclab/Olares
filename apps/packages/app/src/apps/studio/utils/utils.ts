import { useDevelopingApps } from '../stores/app';
import { ROUTE_NAME } from '../common/router-name';

export async function pushToSystem(app_name, router, value?: string) {
	const store = useDevelopingApps();

	let namespace: string;

	if (value) {
		namespace = value;
	} else {
		const url = window.location.origin;
		// const url = 'https://studio.zhaohuaiyuan.olares.cn/';
		const olaresId =
			url.split('.')[1] == 'local' ? url.split('.')[2] : url.split('.')[1];
		namespace = app_name + '-dev-' + olaresId;
	}

	router.push({
		name: ROUTE_NAME.WORKLOAD,
		params: {
			namespace: namespace,
			id: app_name
		}
	});
}
