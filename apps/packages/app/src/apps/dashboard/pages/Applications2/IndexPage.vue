<template>
	<QSectionStyle class="q-ml-sm">
		<div class="action-container row items-center q-gutter-x-md q-mr-xxl">
			<QInputStyle>
				<q-input
					v-model="name"
					type="search"
					outlined
					debounce="500"
					:placeholder="t('SEARCH')"
					clearable
					:disable="loading"
					style="width: 240px"
				>
					<template v-slot:prepend>
						<q-icon name="search" color="ink-2" size="xs" />
					</template>
				</q-input>
			</QInputStyle>
			<BtSelect
				v-model="selected"
				:options="options"
				dense
				outlined
				:disable="loading"
				@update:model-value="fetchData"
			/>
			<SortButtom
				@change="sortHandler"
				:default-value="sort"
				:disable="loading"
			></SortButtom>
			<QButtonStyle>
				<q-btn
					:icon="isRunning ? 'stop' : 'play_arrow'"
					dense
					outline
					color="ink-2"
					@click="toggleRunning"
					:disable="loading"
				>
					<q-tooltip
						><span class="text-body3 text-ink-tooltip">{{
							isRunning
								? $t('stop_autoplay_metrics_refresh')
								: $t('start_autoplay_metrics_refresh')
						}}</span></q-tooltip
					>
				</q-btn>
			</QButtonStyle>
		</div>
	</QSectionStyle>
	<MyGridLayout col-width="510px" gap="xl" :contentLength="sortApps.length">
		<MyCard v-for="(item, itemIndex) in sortApps" :key="item.id">
			<div class="row items-center q-pb-xl">
				<q-skeleton v-if="loading" type="rect" width="32px" height="32px" />
				<div v-else class="row no-wrap relative-position flex-gap-x-md">
					<div
						v-for="(child, n) in item.entrances"
						:key="child.id"
						class="relative-position"
						:style="`margin-left: ${n ? -30 : 0}px;z-index:${n + 1}`"
					>
						<MyAvatarImgVue
							:src="item.icon || child.icon"
							:loading="loading"
							:outlined="item.entrances.length > 1"
							@click="EntranceClickHandler(child, itemIndex)"
						></MyAvatarImgVue>
					</div>
				</div>
				<div
					class="row items-center justify-between text-h4 text-ink-1 q-ml-md"
					style="flex: 1"
				>
					<div class="row items-center">
						<q-skeleton v-if="loading" type="text" width="80px" />
						<span v-else> {{ item.currentEntrance.title }}</span>
						<div
							v-if="authLevelFilter(item.currentEntrance.authLevel)"
							class="q-px-md q-py-xs bg-background-3 rounded-borders-lg q-ml-lg"
						>
							<div class="text-subtitle3 text-positive">
								{{ authLevelFilter(item.currentEntrance.authLevel) }}
							</div>
						</div>
					</div>
					<div class="row items-center" v-if="item.state">
						<MyBadge :type="item.state"></MyBadge>
						<span class="text-subtitle3 text-ink-2 q-ml-sm">{{
							$t(`APP_STATUS.${item.state}`)
						}}</span>
					</div>
				</div>
			</div>
			<div>
				<Metrics
					:namespace="item.namespace"
					:name="item.deployment"
					:cpu_usage="dataFilter('cpu_usage', item.name)"
					:memory_usage="dataFilter('memory_usage', item.name)"
					:net_transmitted="dataFilter('net_transmitted', item.name)"
					:net_received="dataFilter('net_received', item.name)"
					:ownerKind="item.ownerKind"
					:loading="loading"
				></Metrics>
			</div>
		</MyCard>
	</MyGridLayout>
	<Empty
		v-if="sortApps.length === 0"
		size="large"
		style="margin-top: 240px"
	></Empty>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import MyAvatarImgVue from '@apps/control-panel-common/src/components/MyAvatarImg.vue';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import Metrics from './MetricsPage.vue';
import MyBadge from '@apps/control-panel-common/src/components/MyBadge.vue';
import { useRouter } from 'vue-router';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import { get, toLower, capitalize, isEmpty } from 'lodash';
import SortButtom from '@apps/control-panel-common/src/components/SortButton.vue';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import { t } from '@apps/dashboard/src/boot/i18n';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import {
	fetchWorkloadsMetrics,
	loadingApps,
	loadingData,
	resetAppMetricsData
} from './config';
import { useAppList } from '@apps/dashboard/src/stores/AppList';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
import { OLARES_APP } from '@apps/control-panel-common/src/constant/user';
import { Locker } from '../../types/main';
import QBtnToggleStyle from '@apps/control-panel-common/components/QBtnToggleStyle.vue';
import QButtonStyle from '@apps/control-panel-common/components/QButtonStyle.vue';
const appList = useAppList();
enum EntranceState {
	Public = 'public',
	Private = 'private'
}

const options = [
	{
		label: t('SORT_BY_NODE_CPU_UTILISATION'),
		value: 'namespace_cpu_usage',
		sortBy: 'cpu_usage'
	},
	{
		label: t('SORT_BY_NODE_MEMORY_UTILISATION'),
		value: t('SORT_BY_NODE_MEMORY_UTILISATION'),
		sortBy: 'memory_usage'
	},
	{
		label: t('SORT_BY_INBOUND_TRAFFIC'),
		value: 'namespace_net_bytes_received',
		sortBy: 'net_received'
	},
	{
		label: t('SORT_BY_OUTBOUND_TRAFFIC'),
		value: 'namespace_net_bytes_transmitted',
		sortBy: 'net_transmitted'
	}
];

const selected = ref(options[0]);
const router = useRouter();
const sort = ref<any>('desc');
const name = ref();
const isRunning = ref(true);

const appDetail = useAppDetailStore();
const userNamespace = `user-space-${appDetail.user.username}`;

const apps = ref(loadingApps);
const loading = ref(true);
const monitoringData = ref(loadingData);

const fetchData = (showLoading = true, autofresh = false) => {
	const data = appList.appsWithNamespace;
	if (showLoading) {
		loading.value = true;
	}
	if (!autofresh) {
		clearLocker();
	}
	apps.value = data.map((item) => ({
		...item,
		isSystem: item.namespace === userNamespace
	}));
	fetchWorkloadsMetrics(apps.value, userNamespace, sort.value, autofresh)
		.then((data) => {
			monitoringData.value = data;
		})
		.finally(() => {
			loading.value = false;
			isRunning.value && refresh();
		});
};
const sortApps = computed(() => {
	const target = get(monitoringData.value, `${selected.value.sortBy}`, []);
	const data = target.map((item) => {
		const isOlaresApp = item.name === OLARES_APP;
		const firstEntrance = get(item, 'entrances[0]', {});
		const firstEntranceFormat = isOlaresApp
			? { ...firstEntrance, title: item.title, icon: item.icon }
			: firstEntrance;
		return {
			...item,
			entrances: isOlaresApp
				? [{ ...firstEntrance, icon: item.icon }]
				: item.entrances,
			currentEntrance: item.currentEntrance
				? item.currentEntrance
				: firstEntranceFormat
		};
	});
	if (!name.value) {
		return data;
	} else {
		return data.filter((item: any) =>
			toLower(item.title).includes(toLower(name.value))
		);
	}
});

const dataFilter = (type: any, name: string) => {
	return monitoringData.value[type].find((item: any) => item.name === name);
};

const sortHandler = (value: string) => {
	sort.value = value;
	monitoringData.value = { ...loadingData };
	initData();
};

const initData = (loading = true) => {
	clearLocker();
	resetAppMetricsData();
	fetchData(loading);
};

const EntranceClickHandler = (data: any, itemIndex: number) => {
	// monitoringData.value[selected.value.sortBy][itemIndex].currentEntrance = data;
};

const authLevelFilter = (state: EntranceState) => {
	return state === EntranceState.Public ? capitalize(state) : '';
};
let locker: Locker = undefined;
const clearLocker = () => {
	locker && clearTimeout(locker);
};

const refresh = () => {
	clearLocker();
	locker = setTimeout(() => {
		fetchData(false, true);
	}, 5 * 1000);
};

const toggleRunning = () => {
	isRunning.value = !isRunning.value;
	if (isRunning.value) {
		initData(false);
	} else {
		clearLocker();
	}
};

resetAppMetricsData();

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

onBeforeUnmount(() => {
	clearLocker();
});
</script>
<style lang="scss" scoped>
.action-container {
	position: absolute;
	right: 0;
	top: 60px;
}
.avatar-image {
	margin-left: -20px;
}
</style>
