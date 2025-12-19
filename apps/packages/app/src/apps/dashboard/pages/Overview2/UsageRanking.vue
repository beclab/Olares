<template>
	<q-card flat class="full-height transparent">
		<ColumnFlex>
			<template #header>
				<div class="row items-center justify-between q-py-lg">
					<div class="text-h5 text-ink-1">{{ $t('USAGE_RANKING') }}</div>
					<div
						class="text-h6 text-light-blue-default cursor-pointer"
						@click="routeToApplication"
					>
						{{ $t('MORE') }}
					</div>
				</div>
			</template>

			<div class="full-height" :class="[vertical ? 'column' : 'row']">
				<div
					v-for="(list, index) in monitoringData"
					:key="index"
					style="flex: 1"
					:class="[
						index === 0 && vertical ? 'q-mb-xl' : '',
						index === 0 && !vertical ? 'q-mr-xl' : ''
					]"
				>
					<q-card
						flat
						bordered
						class="q-px-xl q-pb-lg rounded-borders-lg full-height"
					>
						<ColumnFlex>
							<template #header>
								<div class="my-spliter q-mb-md">
									<div class="row item-center justify-between q-py-lg">
										<div class="text-h6 text-ink-1">
											{{ t(list.title, { count: 5 }) }}
										</div>
									</div>
								</div>
							</template>
							<bt-scroll-area class="full-height">
								<div class="q-gutter-y-md">
									<ApplicationItem
										v-for="(item, index) in list.data.slice(0, 5)"
										:key="item.id"
										:data="item"
										:index="index + 1"
										:loading="loading"
									></ApplicationItem>
								</div>
							</bt-scroll-area>
						</ColumnFlex>
					</q-card>
				</div>
			</div>
		</ColumnFlex>
	</q-card>
</template>

<script setup lang="ts">
import { getResult } from '@apps/control-panel-common/src/containers/Monitoring/config';
import { get, orderBy, sortBy, unionBy, isEmpty } from 'lodash';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import { getMetrics, getNamespaces } from '@apps/dashboard/src/network';
import {
	getLastMonitoringData,
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import ApplicationItem from './ApplicationItem.vue';
import ColumnFlex from './components/ColumnFlex.vue';
import { getValue } from './config';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import { useRouter } from 'vue-router';
import { fetchWorkloadsMetrics } from '../Applications2/config';
import { useI18n } from 'vue-i18n';
import { useAppList } from '@apps/dashboard/src/stores/AppList';

const { t } = useI18n();
const appList = useAppList();

interface Props {
	vertical?: boolean;
}

withDefaults(defineProps<Props>(), {});
const router = useRouter();

const appDetail = useAppDetailStore();
const userNamespace = `user-space-${appDetail.user.username}`;

const loadingApps = new Array(19).fill({});
const loadingData = [
	{
		title: 'TOP_COUNT_CPU_USAGE_APPS',
		data: loadingApps
	},
	{
		title: 'TOP_COUNT_MEMORY_USAGE_APPS',
		data: loadingApps
	}
];
const apps = ref(loadingApps);
const loading = ref(true);
const monitoringData = ref(loadingData);
const fetchData = (showLoading = true) => {
	const data = appList.appsWithNamespace;
	if (showLoading) {
		loading.value = true;
	}
	apps.value = data.map((item) => ({
		...item,
		isSystem: item.namespace === userNamespace
	}));
	fetchWorkloadsMetrics(apps.value, userNamespace)
		.then((data) => {
			monitoringData.value = [
				{
					title: 'TOP_COUNT_CPU_USAGE_APPS',
					data: data.cpu_usage
				},
				{
					title: 'TOP_COUNT_MEMORY_USAGE_APPS',
					data: data.memory_usage
				}
			];
		})
		.finally(() => {
			loading.value = false;
		});
};

const getTabOptions = (results: any) => {
	const data = getResult(results);

	const result = {
		cpu_usage: get(data, `${MetricTypes.cpu_usage}.data.result`, []).map(
			(item: any) => ({
				name: item.metric.pod.split('-').slice(0, -2).join('-'),
				value: getValue(item)
			})
		),
		memory_usage: get(data, `${MetricTypes.memory_usage}.data.result`, []).map(
			(item: any) => ({
				name: item.metric.pod.split('-').slice(0, -2).join('-'),
				value: getValue(item)
			})
		)
	};

	return result;
};

const getTabOptions2 = (results: any) => {
	const data = getResult(results);

	const result = {
		cpu_usage: get(
			data,
			`${NamespaceMetricTypes.cpu_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getValue(item)
		})),
		memory_usage: get(
			data,
			`${NamespaceMetricTypes.memory_usage}.data.result`,
			[]
		).map((item: any) => ({
			name: item.metric.namespace,
			value: getValue(item)
		}))
	};

	return result;
};

const MetricTypes = {
	cpu_usage: 'pod_cpu_usage',
	memory_usage: 'pod_memory_usage_wo_cache'
};
const NamespaceMetricTypes = {
	cpu_usage: 'namespace_cpu_usage',
	memory_usage: 'namespace_memory_usage_wo_cache'
};

const formatResult = (res: any, namespaceRes: any) => {
	const data1 = getTabOptions(res.data.results);
	const data2 = getTabOptions2(namespaceRes.data.results);
	const cpu_usage_total = data1.cpu_usage
		.concat(data2.cpu_usage)
		.map((item: any) => {
			const app = apps.value.find(
				(app) => app.namespace === item.name || app.deployment === item.name
			);

			if (app) {
				const unit = getSuitableUnit(item.value, 'cpu');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app,
					value: parseFloat(item.value)
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const cpu_usage = unionBy(
		orderBy(cpu_usage_total, ['value', 'title'], ['desc', 'asc']),
		'title'
	);

	const memory_usage_total = data1.memory_usage
		.concat(data2.memory_usage)
		.map((item: any) => {
			const app = apps.value.find(
				(app) => app.namespace === item.name || app.deployment === item.name
			);
			if (app) {
				const unit = getSuitableUnit(item.value, 'memory');
				const used = getValueByUnit(item.value, unit);
				return {
					unit,
					used,
					...item,
					...app,
					value: parseFloat(item.value)
				};
			}
			return false;
		})
		.filter((item: any) => item);
	const memory_usage = unionBy(
		orderBy(memory_usage_total, ['value', 'title'], ['desc', 'asc']),
		'title'
	);
	monitoringData.value = [
		{
			title: 'CPU',
			data: cpu_usage
		},
		{
			title: 'Memory',
			data: memory_usage
		}
	];
};

const routeToApplication = () => {
	router.push({
		path: '/applications'
	});
};

watch(
	() => appList.appsWithNamespace,
	(newData) => {
		if (isEmpty(newData)) {
			loading.value = true;
		} else {
			fetchData(false);
		}
	},
	{
		immediate: true,
		deep: true
	}
);
</script>

<style lang="scss" scoped>
.my-spliter {
	border-bottom: 1px solid $separator !important;
}
</style>
