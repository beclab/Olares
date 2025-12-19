<template>
	<MyPage>
		<MyCard square flat animated title="Detail">
			<template #extra>
				<div class="row">
					<!-- <Terminal :data="containerPath">{{ t('TERMINAL') }}</Terminal> -->
					<!-- <q-btn
            class="q-ml-md"
            no-caps
            outline
            color="primary"
            :label="t('CONTAINER_LOGS')"
            @click="() => (visible = true)"
          /> -->
				</div>
			</template>
			<DetailPage :data="detail"> </DetailPage>
		</MyCard>
		<MyCard square flat :title="t('VOLUMES')">
			<VolumeContainer
				:volumes="volumes"
				:containers="containersTarget"
				:isMultiProject="podDetail.isFedManaged"
			></VolumeContainer>
		</MyCard>
		<ContainerMonitoring :container="container"></ContainerMonitoring>
		<MyCard square flat :title="t('ENVIRONMENT_VARIABLE_PL')" outline>
			<q-list bordered class="rounded-borders">
				<template v-for="(item, index) in envlist" :key="item.name">
					<q-separator v-if="index != 0" />
					<MyExpansion :label="labelFormat(item)" :default-opened="true">
						<MyChipList v-if="item.variables.length" :data="item.variables">
						</MyChipList>
						<div v-else>
							<Empty></Empty>
						</div>
					</MyExpansion>
				</template>
			</q-list>
		</MyCard>
		<q-inner-loading :showing="loading"> </q-inner-loading>
		<MyDialog v-model="visible">
			<Log></Log>
		</MyDialog>
	</MyPage>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import { getPodDetail } from '@apps/control-hub/src/network';
import { useRoute } from 'vue-router';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import { isEmpty, get } from 'lodash';
import { fetcEnvList } from '@apps/control-panel-common/src/containers/env';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import MyChipList from '@apps/control-panel-common/src/containers/MyChipList.vue';
import Log from './Logs.vue';
import MyDialog from '@apps/control-panel-common/src/components/MyQDialog2.vue';
import MyExpansion from '@apps/control-panel-common/src/components/MyExpansion.vue';
import VolumeContainer from '@apps/control-panel-common/src/containers/VolumeContainer.vue';
import { getWorkloadVolumes } from '@apps/control-panel-common/src/utils/workload.js';
import ContainerMonitoring from '@apps/control-panel-common/src/containers/Monitoring/ContainerMonitoring.vue';

interface Props {
	container?: string;
	createTime?: string;
}

const props = withDefaults(defineProps<Props>(), {});

const route = useRoute();
const detail = ref();
const loading = ref(false);
const envlist = ref();
const envDetail = ref();
const visible = ref(false);
const podDetail = ref<any>({});
const volumes = ref([]);

const getResourceInfo = (type: string, detail: any) => {
	const { resources = {} } = detail;
	const resourceType = resources[type];
	console.log('resourceType', resourceType);
	return (
		resourceType &&
		Object.keys(resourceType)
			.map((key) => {
				const isCpu = key === 'cpu';
				const value =
					isCpu && resourceType[key].endsWith('m')
						? parseInt(resourceType[key], 10) / 1000
						: resourceType[key];

				return t(`${key.toUpperCase().replace(/[^A-Z]/g, '_')}_VALUE`, {
					value
				});
			})
			.join('/')
	);
};

const ports = (detail: any) => {
	const { ports = [] } = detail;

	return isEmpty(ports)
		? '-'
		: ports
				.map(
					(port: any) =>
						`${port.name ? `${port.name}:` : ''}${
							port.containerPort || 'None'
						}/${port.protocol}`
				)
				.join(' ');
};
const command = (detail: any) => {
	const { command = [] } = detail;
	return isEmpty(command) ? '-' : command.join(' ');
};

const getAttrs = (detail: any) => {
	// const { cluster, namespace } = this.props.match.params;

	const namespace = detail.namespace;
	const status = detail.state ? Object.keys(detail.state)[0] : '';

	let imagePullPolicy;

	if (detail.imagePullPolicy === 'Always') {
		imagePullPolicy = t('IMAGE_PULL_POLICY_ALWAYS');
	} else if (detail.imagePullPolicy === 'IfNotPresent') {
		imagePullPolicy = t('IMAGE_PULL_POLICY_IFNOTPRESENT');
	} else if (detail.imagePullPolicy === 'Never') {
		imagePullPolicy = t('IMAGE_PULL_POLICY_NEVER');
	} else {
		imagePullPolicy = detail.imagePullPolicy;
	}

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
			value: t(status)
		},
		{
			name: t('IMAGE'),
			value: detail.image
		},
		{
			name: t('IMAGE_ID'),
			value: detail.imageID
		},
		{
			name: t('PORT'),
			value: ports(detail)
		},
		{
			name: t('COMMAND'),
			value: command(detail)
		},
		{
			name: t('RESOURCE_REQUESTS'),
			value: getResourceInfo('requests', detail)
		},
		{
			name: t('RESOURCE_LIMITS'),
			value: getResourceInfo('limits', detail)
		},
		{
			name: t('IMAGE_PULL_POLICY'),
			value: imagePullPolicy
		},
		{
			name: t('RESTART_PL'),
			value: detail.restartCount
		}
	];
};

const containersTarget = computed(() => {
	const { container }: { [key: string]: any } = route.params;

	if (podDetail.value && podDetail.value.containers) {
		return podDetail.value.containers.filter(
			(item: any) => item.name === container
		);
	} else {
		return [];
	}
});

const labelFormat = (item: any) => {
	let label = '';
	label =
		item.type === 'init'
			? t('INIT_CONTAINER_VALUE', { value: item.name })
			: t('CONTAINER_VALUE', { value: item.name });
	return label;
};

const containers = computed(() => {
	const data = envDetail.value || {};

	return [data];
});

const initContainers = computed(() => {
	const data = envDetail.value || {};

	return [data];
});

const fetchEnv = async () => {
	const { namespace, cluster }: { [key: string]: any } = route.params;
	envlist.value = await fetcEnvList({
		namespace: namespace,
		cluster: cluster,
		containers: containers.value,
		initContainers: initContainers.value
	});
};

async function fetchData() {
	const { namespace, name: podName }: Record<string, any> = route.params;
	const containerName = route.params.container || props.container;
	const params = { namespace, podName };
	loading.value = true;
	getPodDetail(params)
		.then((res) => {
			const pod = ObjectMapper.pods(res.data);
			getWorkloadVolumes(pod).then((workloadRes) => {
				podDetail.value = pod;
				volumes.value = workloadRes;
			});
			const data =
				pod.containers.find((item: any) => item.name === containerName) ||
				pod.initContainers.find((item: any) => item.name === containerName);

			data.createTime = get(pod, 'createTime', '');
			data.app = data.app || pod.app;
			// detail.cluster = cluster
			detail.value = getAttrs(data);
			envDetail.value = data;
			fetchEnv();
		})
		.finally(() => {
			loading.value = false;
		});
}

watch([() => route.params.name, () => route.params.container], fetchData);

fetchData();
</script>

<style></style>
