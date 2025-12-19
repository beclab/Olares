<template>
	<MyPage>
		<ReplicaCard
			:detail="detail"
			enableScale
			:module="$route.params.kind"
			@change="replicaChange"
			style="display: none"
		/>
		<MyCard no-content-gap square flat :title="t('PODS')">
			<template #extra>
				<div class="row items-center">
					<QButtonStyle>
						<q-btn
							flat
							dense
							no-caps
							:label="t('REVISION_RECORDS')"
							@click="() => (visible = true)"
						/>
					</QButtonStyle>
					<q-separator vertical inset spaced />
					<Refresh dense @click="() => PodContainerRef.refresh()" />
				</div>
			</template>
			<PodContainer
				:detail="detail"
				:routePushFunction="routePushHandler"
				ref="PodContainerRef"
				@podChange="updateDetail"
			></PodContainer>
		</MyCard>
		<MyCard no-content-gap square flat animated :title="t('Ports')">
			<PortsTable :ports="ports"></PortsTable>
		</MyCard>
		<MyCard no-content-gap square flat :title="t('ENVIRONMENT_VARIABLE_PL')">
			<EnvironmentsLayout :detail="detail"></EnvironmentsLayout>
		</MyCard>
		<Metadata v-if="detail" :detail="detail" :data="detail"></Metadata>
		<MyCard no-content-gap square flat animated :title="t('EVENT_PL')">
			<template #extra>
				<Refresh @click="refreshEvent" />
			</template>
			<div style="position: relative">
				<Event q-mx-lg :detail="detail" ref="childComponentRef"></Event>
			</div>
		</MyCard>
		<MyDialog v-model="visible" persistent :title="$t('REVISION_RECORDS')">
			<RevisionControl
				:detail="detail"
				ref="RevisionControlRef"
			></RevisionControl>
		</MyDialog>
	</MyPage>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { onMounted, ref, watch, computed, toRef, onBeforeUnmount } from 'vue';
import {
	getWorkloadsControler,
	patchWorkloadsControler
} from '@apps/control-hub/src/network';
import { get, has, isEmpty, isNil, uniqBy, flatten } from 'lodash';
import { safeAtob } from '@apps/control-panel-common/src/utils/base64';
import axios from 'axios';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import { joinSelector } from '@apps/control-hub/src/utils';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import Event from '@apps/control-panel-common/src/containers/Event.vue';
import Metadata from '@apps/control-panel-common/src/containers/Metadata.vue';
import ReplicaCard from './ReplicaCard.vue';
import RevisionControl from './RevisionControl/IndexPage.vue';
import MyDialog from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';
import PodContainer from '@apps/control-panel-common/src/containers/PodsList/PodContainer.vue';
import PortsTable from '@apps/control-panel-common/src/containers/PortsTable.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import Refresh from '@apps/control-panel-common/src/components/Refresh.vue';
import EnvironmentsLayout from '@apps/control-panel-common/src/containers/EnvironmentsLayout.vue';
import { componentName } from '@apps/control-hub/src/router/const';
interface Props {
	module?: string;
}

const loading = ref(false);
const route = useRoute();
const router = useRouter();
const envDetail = ref();
const detail = ref();
const params = ref();
const variables = ref();
const childComponentRef = ref();
const visible = ref(false);
const PodContainerRef = ref();
const props = withDefaults(defineProps<Props>(), {});

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
						`api/v1${getPath({ cluster, namespace })}/secrets/${
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
						`api/v1${getPath({ cluster, namespace })}/configmaps/${
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
	return data;
};

const fetchEnv = async () => {
	const {
		namespace,
		kind,
		pods_name: name
	}: { [key: string]: any } = route.params;
	loading.value = true;
	updateDetail({});
	const { data } = await getWorkloadsControler(namespace, kind, name);
	updateDetail(data);
	loading.value = false;

	params.value = getParams('default');
};

function updateDetail(data: any) {
	const { namespace, kind }: { [key: string]: any } = route.params;
	envDetail.value = data;
	// eslint-disable-next-line @typescript-eslint/ban-ts-comment
	// @ts-ignore
	detail.value = ObjectMapper[kind](data);
}

const getParams = (cluster: string) => {
	const {
		name,
		namespace,
		kind: _kind,
		selector,
		_originData
	} = detail.value || {};
	const kind = _kind || get(_originData, 'kind', '');

	let result: any = {};

	result.namespace = namespace;

	switch (kind) {
		case 'PVC':
			result.pvcName = name;
			break;
		case 'Node':
			result.nodeName = name;
			break;
		case 'Namespace':
			result.namespace = name;
			break;
		case 'Service':
		case 'IPPool':
			result.labelSelector = joinSelector(selector);
			break;
		default:
			result.ownerKind = kind === 'Deployment' ? 'ReplicaSet' : kind;
			result.labelSelector = joinSelector(selector);
	}

	if (has(result, 'labelSelector') && isEmpty(selector)) {
		result = {};
	}
	return result;
};

function replicaChange(value: number) {
	const params = { spec: { replicas: value } };
	const {
		namespace,
		kind,
		pods_name: name
	}: { [key: string]: any } = route.params;
	patchWorkloadsControler(namespace, kind, name, params).finally(() => {
		loading.value = false;
	});
}

function refreshEvent() {
	childComponentRef.value && childComponentRef.value.fetchData();
}

const ports = computed(() => {
	let containers = get(detail.value, 'containers');
	containers = isNil(containers) ? [] : containers;
	const ports = uniqBy(
		flatten(
			containers.map((container: any) =>
				isEmpty(container.ports) ? [] : container.ports
			)
		),
		'name'
	);
	return ports;
});

const routePushHandler = (data) => {
	router.push({
		name: componentName.WORKLOAD_PODS,
		params: {
			...route.params,
			name: data.name,
			uid: data.uid,
			node: data.node,
			createTime: data.createTime
		}
	});
};

watch(
	() => route.params.pods_uid,
	async (newId) => {
		fetchEnv();
	}
);

onMounted(() => {
	fetchEnv();
});

defineExpose({ replicaChange });
</script>
<style lang="scss" scoped>
.environments-wrapper {
	border-radius: 8px;
	border: 1px solid $separator;
}
</style>
