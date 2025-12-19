import { api } from '@apps/control-panel-common/src/boot/axios';
import { safeAtob } from '../utils/base64';
import { get, isNil, isEmpty } from 'lodash-es';

const getPath = ({ cluster, namespace }: { [key: string]: string } = {}) => {
	let path = '';
	if (cluster) {
		path += `/klusters/${cluster}`;
	}
	if (namespace) {
		path += `/namespaces/${namespace}`;
	}
	return path;
};

const etchVariables = async (container: any) => {
	const { cluster, namespace, env = [] } = container;
	const requests: any = [];
	const items: any = [];
	env.forEach((item: any) => {
		if ('valueFrom' in item) {
			const { secretKeyRef, configMapKeyRef } = item.valueFrom;

			if (secretKeyRef) {
				items.push({
					name: item.name,
					key: secretKeyRef.key
				});
				requests.push(
					api.get(
						`/api/v1${getPath({ cluster, namespace })}/secrets/${
							secretKeyRef.name
						}`
					)
				);
			}

			if (configMapKeyRef) {
				items.push({
					name: item.name,
					key: configMapKeyRef.key
				});
				requests.push(
					api.get(
						`/api/v1${getPath({ cluster, namespace })}/configmaps/${
							configMapKeyRef.name
						}`
					)
				);
			}
		} else {
			items.push(item);
			requests.push(item.value || '');
		}
	});
	const result = await Promise.all(requests);

	const data = items.map((item: any, index: number) => {
		const value = result[index].data ? result[index].data : result[index];
		if (value.kind === 'Secret' && item.key) {
			return {
				name: item.name,
				value: safeAtob(get(value.data, item.key, ''))
			};
		}

		if (value.kind === 'ConfigMap' && item.key) {
			return {
				name: item.name,
				value: get(value.data, item.key, '')
			};
		}

		return { value: '', ...item };
	});

	container.variables = data;
	return container;
};

export const fetcEnvList = async ({
	cluster,
	namespace,
	containers = [],
	initContainers = []
}: any) => {
	containers = isNil(containers) ? [] : containers;
	initContainers = isNil(initContainers) ? [] : initContainers;

	const mergeContainers = [
		...initContainers.map((item: any) => ({ ...item, type: 'init' })),
		...containers.map((item: any) => ({ ...item, type: 'work' }))
	];
	const data = await Promise.all(
		mergeContainers.map((container) => {
			container.cluster = cluster;
			container.namespace = namespace;
			return etchVariables(container);
		})
	);

	const list = {
		data,
		isLoading: false
	};
	return data;
};
