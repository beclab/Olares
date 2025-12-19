<template>
	<div>
		<q-card
			class="pod-item-container q-px-lg bg-background-1"
			flat
			:bordered="bordered"
		>
			<div class="pod-item-content row items-center q-col-gutter-x-md">
				<div class="col-4">
					<ListItem size="sm" icon-size="lg">
						<template #avatar>
							<q-img :src="podStatus" width="32px">
								<q-tooltip
									anchor="top middle"
									self="bottom middle"
									:offset="[10, 10]"
								>
									<strong>{{ data.name }}</strong>
									<div>
										{{
											t('READY_VALUE', {
												readyCount: statusDetail.readyCount,
												total: statusDetail.total
											})
										}}
									</div>
									<div>
										{{ t('STATUS_VALUE', { value: t(status.name) }) }}
									</div>
								</q-tooltip>
							</q-img>
							<MyBadge :type="status.type" class="pod-item-badge" flicker />
						</template>
						<template #title>
							<span class="pod-title" @click="itemClick(data)">
								{{ data.name }}
							</span>
						</template>
						<template #subTitle>
							<StatusReason
								v-if="!isRunning"
								:status="status.type"
								:reason="status.name"
								:data="data"
								type="pod"
							></StatusReason>
							<div v-else class="my-content">
								{{
									t('CREATED_TIME', {
										diff: getLocalTime(data.createTime).format(
											'YYYY-MM-DD HH:mm:ss'
										)
									})
								}}
							</div>
						</template>
					</ListItem>
				</div>
				<div class="col-8 relative-position monitoring-empty">
					<div class="row q-col-gutter-x-md">
						<div class="col-4">
							<ListItem size="sm" icon-size="lg">
								<template #title>
									{{
										!data.node
											? '-'
											: t('NODE_IP', { node: data.node, ip: data.nodeIp })
									}}
								</template>
								<template #subTitle>
									{{ t('NODE') }}
								</template>
							</ListItem>
						</div>

						<div class="col-2">
							<ListItem size="sm" icon-size="lg">
								<template #title>
									{{ data.podIp || '-' }}
								</template>
								<template #subTitle>
									{{ t('POD_IP_ADDRESS_SCAP') }}
								</template>
							</ListItem>
						</div>

						<div class="" v-if="loading">{{ t('LOADING') }}</div>
						<div class="" v-else-if="!hasData">
							{{ t('NO_MONITORING_DATA') }}
						</div>
						<template v-else>
							<div
								v-for="(item, index) in getMonitoringCfgs"
								:key="index"
								class="col-3"
							>
								<div class="">
									<MylineChartMini :data="item" />
								</div>
							</div>
						</template>
					</div>
				</div>
			</div>
		</q-card>
		<q-space />
	</div>
</template>

<script setup lang="ts">
import { get, isEmpty } from 'lodash-es';
import { getLocalTime } from '@apps/control-panel-common/src/utils';
import { computed, ref } from 'vue';
import StatusReason from '../../components/StatusReason.vue';
import MylineChartMini from '../../components/Charts/MylineChartMini2.vue';
import { getAreaChartOps } from '@apps/control-panel-common/src/utils/monitoring';
import { useRoute, useRouter } from 'vue-router';
import podStatus from '../../assets/pod.svg';
import MyBadge from '../../components/MyBadge.vue';
import ListItem from '../../components/MyListItem/ListItem.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	data: any;
	metrics: any;
	loading: boolean;
	scroll?: boolean;
	bordered?: boolean;
}

const emits = defineEmits(['click']);
const props = withDefaults(defineProps<Props>(), {});
const route = useRoute();
const statusDetail = computed(() => getContainerStatus());

const getContainerStatus = () => {
	const containerStatuses = get(props.data, 'status.containerStatuses') || [];

	let readyCount = 0;

	containerStatuses.forEach((item: any) => {
		if (item.ready) {
			readyCount += 1;
		}
	});

	return {
		readyCount,
		total: containerStatuses.length
	};
};

const status = computed(() => {
	const { status, type } = props.data.podStatus;
	return {
		type: type.toLowerCase(),
		name: status
	};
});

const isRunning = computed(
	() => status.value.type === 'running' || status.value.type === 'completed'
);

const getMonitoringCfgs = computed(() => {
	if (isEmpty(props.metrics)) return [];
	const data = [
		{
			type: 'cpu',
			title: t('CPU'),
			unitType: 'cpu',
			legend: [t('USED')],
			data: [props.metrics.cpu],
			bgColor: 'transparent'
		},
		{
			type: 'memory',
			title: t('MEMORY'),
			unitType: 'memory',
			legend: [t('USED')],
			data: [props.metrics.memory],
			bgColor: 'transparent'
		}
	];

	return data.map((item) => getAreaChartOps(item));
});

const hasData = computed(() => {
	return getMonitoringCfgs.value.some((item) => item.data[0].length > 0);
});

const router = useRouter();

const itemClick = (data: any) => {
	emits('click', data);
};
</script>

<style lang="scss" scoped>
.pod-item-container {
	border-radius: 8px;
	border-color: $separator;
	.pod-item-content {
		height: 72px;
	}
	.pod-title {
		&:hover {
			color: $blue-6;
			cursor: pointer;
		}
	}
	.pod-item-badge {
		position: absolute;
		bottom: 0;
		right: 0;
	}
}
.monitoring-empty {
	color: $ink-2;
	font-size: 12px;
}
</style>
