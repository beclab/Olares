<template>
	<div class="q-gutter-y-xl">
		<InfoCardRadio :list="options" @change="listchange" :loading="loading">
			<template #network>
				<div class="q-mt-md">
					<table class="info-table-layout text-h6 text-ink-1">
						<tbody>
							<tr>
								<td>
									<div class="text-h6 text-ink-1">
										<q-skeleton
											v-if="netStore.loading"
											type="text"
											width="88px"
										/>
										<template v-else>{{ txRate }}</template>
									</div>
								</td>
								<td>
									<q-icon
										class="icon-tx"
										name="sym_r_arrow_upward_alt"
										size="22px"
										color="positive"
									/>
								</td>
							</tr>
							<tr>
								<td>
									<div class="text-h6 text-ink-1 q-mt-xs">
										<q-skeleton
											v-if="netStore.loading"
											type="text"
											width="88px"
										/>
										<span v-else>{{ rxRate }}</span>
									</div>
								</td>
								<td>
									<q-icon
										class="icon-rx q-mt-xs"
										name="sym_r_arrow_downward_alt"
										size="22px"
										color="negative"
									/>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
			</template>
			<template #fan>
				<div class="q-mt-md">
					<table class="info-table-layout text-h6 text-ink-1">
						<tbody>
							<tr>
								<td>
									<div class="text-h6 text-ink-1 row justify-between">
										<q-skeleton
											v-if="netStore.loading"
											type="text"
											width="88px"
										/>
										<template v-else>
											<span>{{ $t('CPU') }}</span>
											<span class="q-ml-md">{{
												FanStore.data.cpu_fan_speed
											}}</span>
										</template>
									</div>
								</td>
								<td>
									<span class="q-ml-xs">RPM</span>
								</td>
							</tr>
							<tr>
								<td>
									<div class="text-h6 text-ink-1 q-mt-xs">
										<q-skeleton
											v-if="netStore.loading"
											type="text"
											width="88px"
										/>
										<template v-else>
											<span>{{ $t('GPU_OP.GPU') }}</span>
											<span class="q-ml-md">{{
												FanStore.data.gpu_fan_speed
											}}</span>
										</template>
									</div>
								</td>
								<td>
									<div class="q-ml-xs q-mt-xs">RPM</div>
								</td>
							</tr>
						</tbody>
					</table>
				</div>
			</template>
		</InfoCardRadio>
	</div>
</template>

<script setup lang="ts">
import InfoCardRadio from '../../components/InfoCard/InfoCardRadio.vue';
import { InfoCardItemProps } from '@apps/dashboard/src/components/InfoCard/InfoCardItem.vue';
import { ref, watch, computed, onMounted, onBeforeUnmount } from 'vue';
import { getContentOptions, getTabOptions, MetricTypesFormat } from './config';
import { getAreaChartOps } from '@apps/dashboard/src/utils/monitoring';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useResourcesStore } from '@apps/dashboard/src/stores/Resource';
import gpuImg from '@apps/dashboard/src/assets/gpu.svg';
import gpuImgDark from '@apps/dashboard/src/assets/gpu-dark.svg';
import networkImg from '@apps/dashboard/src/assets/network.svg';
import networkImgDark from '@apps/dashboard/src/assets/network-dark.svg';
import fanImg from '@apps/dashboard/src/assets/fan.svg';
import fanImgDark from '@apps/dashboard/src/assets/fan-dark.svg';
import { ROUTE_NAME } from '@apps/dashboard/src/router/const';
import { GPUNodeList, GraphicsListParams } from '@apps/dashboard/src/types/gpu';
import { useInstantVector } from './GPU/config';
import { getGraphicsList, getNodesList } from '@apps/dashboard/src/network';
import { get, round } from 'lodash';
import { useNetStore } from '@apps/dashboard/stores/Net';
import { getThroughput } from '@apps/dashboard/src/utils/memory';
import { useAppDetailStore } from 'src/apps/controlHub/stores/AppDetail';
import { useFanStore } from '@apps/dashboard/stores/Fan';
const resourcesStore = useResourcesStore();
const netStore = useNetStore();
const FanStore = useFanStore();
const appDetail = useAppDetailStore();
const $q = useQuasar();
const { t, locale } = useI18n();
interface Props {
	data: any;
	type: string;
	loading: boolean;
}
const props = withDefaults(defineProps<Props>(), {});

const clusterOptions = ref();
const hasGPU = ref(appDetail.isAdmin);
const gpuNodeList = ref<GPUNodeList['list']>([]);

const netPortTarget = computed(() => {
	const target = netStore.list.find((item) => item.isHostIp);
	if (!target)
		return {
			txRate: '0',
			rxRate: '0',
			iface: undefined
		};
	return target;
});

const txRate = computed(() => {
	const value = get(netPortTarget.value, 'txRate', 0);
	return getThroughput(value);
});

const rxRate = computed(() => {
	const value = get(netPortTarget.value, 'rxRate', 0);
	return getThroughput(value);
});

const listchange = (data: InfoCardItemProps, index: number) => {
	resourcesStore.activeIndex = index;
};

const options = computed(() => {
	const available = gpuNodeList.value.filter(
		(item) => !item.isExternal && item.isSchedulable
	).length;

	const unAvailable = gpuNodeList.value.filter(
		(item) => !item.isExternal && !item.isSchedulable
	).length;
	const defaultGPUOptions: InfoCardItemProps[] = [
		{
			id: 'gpu',
			total: cardGaugeConfig.value[0].total,
			used: cardGaugeConfig.value[0].used,
			name: t('GPU_OP.GPU'),
			active: false,
			unitType: '',
			unit: '',
			img: $q.dark.isActive ? gpuImgDark : gpuImg,
			img_active: '',
			loading: false,
			info: t('GPU_OP.VIDEO_MEMORY_USAGE'),
			route: {
				name: ROUTE_NAME.GPU_LIST
			}
		}
	];

	const defaultNetworkOptions: InfoCardItemProps[] = [
		{
			id: 'network',
			used: '0',
			total: '0',
			name: t('NET_OP.NETWORK'),
			active: false,
			unitType: '',
			unit: '',
			img: $q.dark.isActive ? networkImgDark : networkImg,
			img_active: '',
			loading: false,
			route: {
				name: ROUTE_NAME.NETWORK_DETAIL
			}
		}
	];

	const fanOptions = [
		{
			id: 'fan',
			used: '0',
			total: '0',
			name: t('FAN_OP.FAN'),
			active: false,
			unitType: '',
			unit: '',
			img: $q.dark.isActive ? fanImgDark : fanImg,
			img_active: '',
			loading: false,
			route: {
				name: ROUTE_NAME.FAN_DETAIL
			}
		}
	];
	const defaultOptions = hasGPU.value
		? defaultGPUOptions.concat(defaultNetworkOptions)
		: defaultNetworkOptions;

	const options = clusterOptions.value.concat(defaultOptions);

	return FanStore.isOlaresOneDevice ? options.concat(fanOptions) : options;
});

const gpuData = ref();
const fetchGraphicsList = async () => {
	const params: GraphicsListParams = {
		filters: {},
		pageRequest: {
			sort: 'ASC',
			sortField: 'id'
		}
	};

	const res = await getGraphicsList(params);
	gpuData.value = res.data.list;
};

const coreUtilizedPercentAvg = computed(() => {
	if (!gpuData.value || gpuData.value.length === 0) return 0;
	const coreUtilizedPercent = gpuData.value.map(
		(item) => item.coreUtilizedPercent
	);
	const sum = coreUtilizedPercent.reduce((acc, val) => acc + val, 0);
	const value = round(sum / coreUtilizedPercent.length, 2);
	return value ? value / 100 : 0;
});

const cardGaugeConfig = useInstantVector([
	{
		title: t('GPU_OP.CPU_R'),
		percent: 0,
		query: `avg(sum(hami_memory_used) by (instance)) / 1024`,
		totalQuery: `avg(sum(hami_memory_size) by (instance))/1024`,
		percentQuery: `(avg(sum(hami_memory_used) by (instance)) / 1024)/(avg(sum(hami_memory_size) by (instance))/1024)*100`,
		total: 0,
		used: 0,
		unit: 'GiB'
	}
]);

const checkGpu = async () => {
	try {
		const res = await getNodesList();
		const target = res.data.items.find((item) => {
			const cudaSupported = get(
				item,
				'metadata.labels["gpu.bytetrade.io/cuda-supported"]'
			);
			return cudaSupported === 'true';
		});

		hasGPU.value = !!target;
	} catch (error) {
		hasGPU.value = false;
	}
};

onMounted(() => {
	if (appDetail.isAdmin) {
		checkGpu();
	}
	fetchGraphicsList();
	netStore.init();
	FanStore.init();
});

onBeforeUnmount(() => {
	netStore.clearLocker();
	FanStore.clearLocker();
});
watch(
	[() => props.data, () => $q.dark.isActive, () => locale.value],
	() => {
		const MetricTypes = MetricTypesFormat(props.type);
		clusterOptions.value = getTabOptions(
			props.data,
			MetricTypes,
			0,
			$q.dark.isActive
		);
		resourcesStore.clusterResource = getContentOptions(
			props.data,
			MetricTypes
		).map((item) => getAreaChartOps(item));
	},
	{
		deep: true,
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.info-table-layout {
	border-collapse: collapse;
	td {
		padding: 0px;
	}
}
</style>
