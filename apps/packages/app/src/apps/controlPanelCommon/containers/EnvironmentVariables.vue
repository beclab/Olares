<template>
	<div v-for="item in envlist" :key="item.name">
		<q-expansion-item :label="labelFormat(item)">
			<MyChipList v-if="item.env && item.env.length > 0" :data="item.variables">
			</MyChipList>
			<div v-else>
				<Empty></Empty>
			</div>
		</q-expansion-item>
		<q-separator spaced inset />
	</div>
	<q-inner-loading :showing="loading"> </q-inner-loading>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { onMounted, ref, watch, computed } from 'vue';
import { getDeploymentsDetail } from '../network';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
import { get, isEmpty, isNil } from 'lodash';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { safeAtob } from '../utils/base64';
import axios from 'axios';
import MyCard from '@apps/control-panel-common/src/components/MyCard.vue';
import Empty from '../components/Empty.vue';
import MyChipList from './MyChipList.vue';
import { ObjectMapper } from '../utils/object.mapper';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	module?: string;
}

let loading = ref(false);
const route = useRoute();
const envDetail = ref();
const envlist = ref();
const variables = ref();

const props = withDefaults(defineProps<Props>(), {});

const labelFormat = (item: any) => {
	let label = '';
	label =
		item.type === 'init'
			? t('INIT_CONTAINER_VALUE', { value: item.name })
			: t('CONTAINER_VALUE', { value: item.name });
	return label;
};

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
					axios.get(
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
					axios.get(
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
		const value = result[index];

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

	variables.value = data;

	container.variables = data;
	return container;
};

const containers = computed(() => {
	const data = envDetail.value || {};
	const { spec, containers = [] } = data;

	if (props.module === 'containers') return [data];

	if (!isEmpty(containers)) return containers;
	if (!isEmpty(spec)) return get(spec, 'template.spec.containers', []);

	return [];
});

const initContainers = computed(() => {
	const data = envDetail.value || {};
	const { spec, initContainers: initContainersData = [] } = data;

	if (props.module === 'containers') return [data];

	if (!isEmpty(initContainersData)) return initContainersData;
	if (!isEmpty(spec)) return get(spec, 'template.spec.initContainers', []);

	return [];
});

const fetchList2 = async ({
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

const fetchData = async () => {
	const { namespace, cluster }: { [key: string]: any } = route.params;
	envlist.value = await fetchList2({
		namespace: namespace,
		cluster: cluster,
		containers: containers.value,
		initContainers: initContainers.value
	});
};

const fetchEnv = () => {
	const { namespace, name, container }: { [key: string]: any } = route.params;
	loading.value = true;
	getDeploymentsDetail(namespace, name)
		.then((res) => {
			// envDetail.value = ObjectMapper.pods(res.data);

			const pod = ObjectMapper.pods(res.data);
			const detail =
				pod.containers.find((item: any) => item.name === container) ||
				pod.initContainers.find((item: any) => item.name === container);
			detail.createTime = get(pod, 'createTime', '');
			detail.app = detail.app || pod.app;

			envDetail.value = detail;

			fetchData();
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(
	() => route.params,
	async (newId) => {
		fetchEnv();
	}
);
onMounted(() => {
	fetchEnv();
});
</script>
