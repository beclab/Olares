<template>
	<q-card :bordered="bordered" flat class="container-item q-px-lg">
		<div
			class="container-item-content row items-center no-wrap q-col-gutter-x-md"
		>
			<div class="col-4">
				<ListItem size="sm" icon-size="lg">
					<template #avatar>
						<q-img :src="dockerIcon" />
						<MyBadge
							class="my-containers-badge"
							:type="status"
							flicker
							outlined
						></MyBadge>
					</template>
					<template #title>
						<div class="row items-center no-wrap">
							<div
								:class="[
									isInit ? 'text-grey-7' : '',
									status !== 'terminated'
										? 'container-item-hover'
										: 'container-disabled'
								]"
								class="ellipsis"
								@click="containerOverviewHandler(container)"
							>
								{{ container.name }}
							</div>
							<q-icon
								v-permission="[route.params.namespace, true]"
								v-if="!hideLog"
								color="blue-default"
								name="sym_r_article"
								flat
								dense
								size="16px"
								class="q-mx-xs"
								@click="logsHandler(container)"
							>
								<q-tooltip anchor="top middle" self="bottom middle">{{
									t('CONTAINER_LOGS')
								}}</q-tooltip>
							</q-icon>
							<div
								v-permission="route.params.namespace"
								style="margin-top: -1px"
							>
								<Terminal
									v-if="status === 'running' && !hideTerminal"
									:data="containerPath(container)"
								>
									<q-tooltip anchor="top middle" self="bottom middle">{{
										t('TERMINAL')
									}}</q-tooltip>
								</Terminal>
							</div>
						</div>
					</template>
					<template #subTitle>
						<template v-if="reason">
							{{ t(reason) }}
						</template>
						<template v-else>
							{{ t('IMAGE_VALUE', { value: container.image }) }}
						</template>
					</template>
				</ListItem>
			</div>
			<div class="col-8">
				<div class="row no-wrap q-col-gutter-md">
					<div class="col-2">
						<ListItem size="sm" icon-size="lg">
							<template #title>{{ $t(status.toUpperCase()) }}</template>
							<template #subTitle>{{ t('STATUS') }}</template>
						</ListItem>
					</div>
					<div class="col-2">
						<ListItem size="sm" icon-size="lg">
							<template #title>
								{{
									isUndefined(container.restartCount)
										? '-'
										: container.restartCount
								}}
							</template>
							<template #subTitle>
								{{
									!isUndefined(container.restartCount) &&
									container.restartCount === 1
										? t('RESTART')
										: t('RESTART_PL')
								}}
							</template>
						</ListItem>
					</div>
					<div class="col-2 ellipsis">
						<ListItem size="sm" icon-size="lg">
							<template #title>{{
								!isEmpty(container.ports) && container.ports.length === 1
									? t('PORT')
									: t('PORT_PL')
							}}</template>
							<template #subTitle> {{ portsFormat(container.ports) }}</template>
						</ListItem>
					</div>

					<div v-if="loading" class="col-12 monitoring-empty monitoring-nomal">
						{{ t('LOADING') }}
					</div>
					<div
						v-else-if="!hasData"
						class="col-12 monitoring-empty monitoring-nomal"
					>
						{{ t('NO_MONITORING_DATA') }}
					</div>
					<template v-else>
						<div
							v-for="(item, index) in getMonitoringCfgs"
							:key="index"
							class="col-3"
						>
							<div :class="[index !== 0 ? 'q-ml-md' : '']">
								<MylineChartMini :data="item" />
							</div>
						</div>
					</template>
				</div>
			</div>
		</div>
	</q-card>
	<MyDialog v-model="visible" :title="currentContainer.name">
		<Logs :container="currentContainer.name" class="full-height"></Logs>
	</MyDialog>

	<MyDialog v-model="visible2" :title="currentContainer.name">
		<ContainerOverview :container="currentContainer.name"></ContainerOverview>
	</MyDialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import MyBadge from '../../components/MyBadge.vue';
import dockerIcon from '../../assets/docker.svg';
import Terminal from '../TerminalDialog.vue';
import { useRoute } from 'vue-router';
import MyDialog from '../../components/Dialog/Dialog.vue';
import Logs from '../Logs.vue';
import MyPage from '../MyPage.vue';
import { isArray, capitalize, isUndefined, isEmpty } from 'lodash-es';
import { getContainerStatus } from '../../utils/status';
import ContainerOverview from './Overview.vue';
import MylineChartMini from '../../components/Charts/MylineChartMini2.vue';
import ListItem from '../../components/MyListItem/ListItem.vue';
import { getAreaChartOps } from '@apps/control-panel-common/src/utils/monitoring';
import { useI18n } from 'vue-i18n';

interface Props {
	metrics: any;
	container: any;
	isInit?: boolean;
	loading: boolean;
	hideTerminal?: boolean;
	hideLog?: boolean;
	bordered?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	hideTerminal: false,
	hideLog: false
});
const { t } = useI18n();
const route = useRoute();
const visible = ref(false);
const visible2 = ref(false);
const currentContainer = ref({ name: '' });

const hasData = computed(() => {
	return getMonitoringCfgs.value.some((item) => item.data[0].length > 0);
});

const portsFormat = (ports: { containerPort: string; protocol: string }[]) => {
	if (isArray(ports)) {
		return ports
			.map((port) => `${port.containerPort}/${port.protocol}`)
			.join(', ');
	}
	return '-';
};

const containerPath = (container: any) => {
	const { namespace, name: podName }: Record<string, any> = route.params;

	return {
		namespace,
		podName,
		container: container.name
	};
};

const logsHandler = (item: any) => {
	visible.value = true;
	currentContainer.value = item;
};

const containerOverviewHandler = (item: any) => {
	if (status !== 'terminated') {
		visible2.value = true;
		currentContainer.value = item;
	}
};

const { status, reason } = getContainerStatus(props.container);

const getMonitoringCfgs = computed(() => {
	if (isEmpty(props.metrics)) return [];
	const data = [
		{
			type: 'cpu',
			title: t('CPU'),
			unitType: 'cpu',
			legend: ['USED'],
			data: [props.metrics.cpu],
			bgColor: 'transparent'
		},
		{
			type: 'memory',
			title: t('MEMORY'),
			unitType: 'memory',
			legend: ['USED'],
			data: [props.metrics.memory],
			bgColor: 'transparent'
		}
	];

	return data.map((item) => getAreaChartOps(item));
});
</script>

<style lang="scss" scoped>
.container-item {
	border-radius: 8px;
	border: 1px solid $separator;
	.container-item-content {
		height: 72px;
	}
	.my-containers-badge {
		position: absolute;
		right: 0;
		bottom: 0;
	}
}

.container-item-hover {
	&:hover {
		color: $primary;
		cursor: pointer;
	}
}

.container-disabled {
	&:hover {
		cursor: not-allowed;
	}
}
.monitoring-empty {
	color: $grey-6;
	font-size: 12px;
}
.monitoring-nomal {
	line-height: 39px;
	vertical-align: bottom;
}
</style>
