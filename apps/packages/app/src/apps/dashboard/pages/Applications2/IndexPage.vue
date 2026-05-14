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
					:disable="viewLoading"
					style="width: 240px"
				>
					<template v-slot:prepend>
						<q-icon name="search" color="ink-2" size="xs" />
					</template>
				</q-input>
			</QInputStyle>
			<BtSelect
				v-model="selected"
				:options="sortMetricOptions"
				dense
				outlined
				:disable="viewLoading"
				@update:model-value="onSortMetricSelect"
			/>
			<SortButtom
				@change="sortHandler"
				:default-value="sort"
				:disable="viewLoading"
			></SortButtom>
			<QButtonStyle>
				<q-btn
					:icon="isRunning ? 'stop' : 'play_arrow'"
					dense
					outline
					color="ink-2"
					@click="toggleRunning"
					:disable="viewLoading"
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
	<q-infinite-scroll
		:key="infiniteResetKey"
		@load="onInfiniteLoad"
		:offset="infiniteScrollOffset"
		:disable="infiniteScrollDisabled"
		:debounce="40"
	>
		<div ref="gridWrapperRef" class="applications2-grid relative-position">
			<MyGridLayout col-width="510px" gap="xl" :contentLength="sortApps.length">
				<MyCard v-for="item in sortApps" :key="`${item.id}-${item.namespace}`">
					<div class="row items-center q-pb-xl">
						<q-skeleton
							v-if="viewLoading"
							type="rect"
							width="32px"
							height="32px"
						/>
						<div v-else class="row no-wrap relative-position flex-gap-x-md">
							<MyAvatarImgVue
								:src="item.icon"
								:loading="viewLoading"
							></MyAvatarImgVue>
						</div>
						<div
							class="row items-center justify-between text-h4 text-ink-1 q-ml-md"
							style="flex: 1"
						>
							<div class="row items-center">
								<q-skeleton v-if="viewLoading" type="text" width="80px" />
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
							:pod_acount="item.pod_acount"
							:loading="viewLoading"
						></Metrics>
					</div>
				</MyCard>
			</MyGridLayout>
		</div>
	</q-infinite-scroll>
	<Empty
		v-if="sortApps.length === 0 && !viewLoading"
		size="large"
		style="margin-top: 240px"
	></Empty>
</template>

<script setup lang="ts">
import {
	computed,
	nextTick,
	onBeforeUnmount,
	onMounted,
	ref,
	watch
} from 'vue';
import { useI18n } from 'vue-i18n';
import MyAvatarImgVue from '@apps/control-panel-common/src/components/MyAvatarImg.vue';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import Metrics from './MetricsPage.vue';
import MyBadge from '@apps/control-panel-common/src/components/MyBadge.vue';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import { get, capitalize, isEmpty, toLower, trim } from 'lodash';
import SortButtom from '@apps/control-panel-common/src/components/SortButton.vue';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import QInputStyle from '@apps/control-panel-common/src/components/QInputStyle.vue';
import { t } from 'src/boot/dashboard-i18n';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import {
	fetchWorkloadsMetrics,
	fetchWorkloadsMetricsForApps,
	cancelMainMetricsFetch,
	loadingData,
	resetAppMetricsData,
	buildSkeletonMonitoringData,
	mergeMonitoringData,
	removeMonitoringDataApps,
	removeAppsFromMetricsCache
} from './config';
import { useAppList } from '@apps/dashboard/src/stores/AppList';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
import { OLARES_APP } from '@apps/control-panel-common/src/constant/user';
import { Locker } from '../../types/main';
import QButtonStyle from '@apps/control-panel-common/components/QButtonStyle.vue';
import {
	APPLICATIONS2_GRID_INFINITE_DEFAULTS,
	clampVisibleLimit,
	computeAppendStep,
	computeColumnCount,
	computeInfiniteScrollOffsetPx,
	getScrollViewportHeight,
	itemsPerPageFromGrid,
	skeletonCountCap,
	targetVisibleAfterReset as targetVisibleAfterResetFn,
	type Applications2GridInfiniteConfig
} from './infiniteLoad';

const gridCfg: Applications2GridInfiniteConfig =
	APPLICATIONS2_GRID_INFINITE_DEFAULTS;

const appList = useAppList();
const { locale } = useI18n();

enum EntranceState {
	Public = 'public',
	Private = 'private'
}

const sortMetricOptions = computed(() => [
	{
		label: t('SORT_BY_NODE_CPU_UTILISATION'),
		value: 'namespace_cpu_usage',
		sortBy: 'cpu_usage'
	},
	{
		label: t('SORT_BY_NODE_MEMORY_UTILISATION'),
		value: 'namespace_memory_usage_wo_cache',
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
]);

const selected = ref(sortMetricOptions.value[0]);
watch(sortMetricOptions, (opts) => {
	const cur = selected.value?.value;
	const next = opts.find((o) => o.value === cur) || opts[0];
	selected.value = next;
});

const sort = ref<any>('desc');
const name = ref<string | undefined>();
const isRunning = ref(true);

const monitoringData = ref(loadingData);
const fullSortedNamesRef = ref<string[]>([]);

const mapMetricRows = (target: any[]) => {
	return target.map((item) => {
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
};

const fullSortedRows = computed(() => {
	const target = get(
		monitoringData.value,
		`${selected.value.sortBy}`,
		[]
	) as any[];
	return mapMetricRows(Array.isArray(target) ? target : []);
});

const displaySortedRows = computed(() =>
	fullSortedRows.value.filter((row) => matchesSearchQuery(row))
);

const totalRowCount = computed(() => displaySortedRows.value.length);

const colCount = ref(2);
const visibleLimit = ref(
	targetVisibleAfterResetFn(
		itemsPerPageFromGrid(gridCfg.fixedGridRows, colCount.value),
		colCount.value,
		gridCfg
	)
);
const infiniteResetKey = ref(0);

const itemsPerPage = computed(() =>
	itemsPerPageFromGrid(gridCfg.fixedGridRows, colCount.value)
);

const targetVisibleAfterReset = computed(() =>
	targetVisibleAfterResetFn(itemsPerPage.value, colCount.value, gridCfg)
);

const gridWrapperRef = ref<HTMLElement | null>(null);
let gridResizeObserver: ResizeObserver | null = null;

const rebindGridResizeObserver = () => {
	gridResizeObserver?.disconnect();
	gridResizeObserver = null;
	const el = gridWrapperRef.value;
	if (!el || typeof ResizeObserver === 'undefined') return;
	gridResizeObserver = new ResizeObserver(onGridBoxResize);
	gridResizeObserver.observe(el);
};

const infiniteScrollOffset = ref(
	typeof window !== 'undefined'
		? computeInfiniteScrollOffsetPx(window.innerHeight, colCount.value, gridCfg)
		: gridCfg.offsetMinPx
);

const updateInfiniteScrollOffset = () => {
	if (typeof window === 'undefined') return;
	const viewport = getScrollViewportHeight(gridWrapperRef.value);
	infiniteScrollOffset.value = computeInfiniteScrollOffsetPx(
		viewport,
		colCount.value,
		gridCfg
	);
};

const onGridBoxResize = () => {
	updateColCount(totalRowCount.value);
};

const updateColCount = (listTotal: number) => {
	if (!gridWrapperRef.value) return;
	const width = gridWrapperRef.value.clientWidth;
	const cols = computeColumnCount(width, gridCfg);
	if (cols !== colCount.value) {
		colCount.value = cols;
		updateInfiniteScrollOffset();
		visibleLimit.value = clampVisibleLimit(visibleLimit.value, listTotal);
	}
};

const syncGridScrollMetrics = () => {
	updateColCount(totalRowCount.value);
	updateInfiniteScrollOffset();
};

watch(
	gridWrapperRef,
	() => {
		void nextTick(() => {
			rebindGridResizeObserver();
			syncGridScrollMetrics();
		});
	},
	{ flush: 'post' }
);

onMounted(() => {
	window.addEventListener('resize', updateInfiniteScrollOffset);
});

const appDetail = useAppDetailStore();
const userNamespace = `user-space-${appDetail.user.username}`;

const allListedApps = computed(() => appList.appsWithNamespace);

const currentAppNames = computed(() =>
	allListedApps.value.map((item: any) => item.name)
);

const prevAppsNamesRef = ref<string[]>([]);

const matchesSearchQuery = (row: {
	name?: string;
	title?: string;
	currentEntrance?: { title?: string };
}) => {
	const q = trim(toLower(name.value || ''));
	if (!q) return true;
	const byName = toLower(row.name || '').includes(q);
	const byTitle = toLower(
		row.title || row.currentEntrance?.title || ''
	).includes(q);
	return byName || byTitle;
};

const loading = ref(false);
const interactionLoading = ref(false);
const viewLoading = computed(() => loading.value || interactionLoading.value);

const initialLoadComplete = ref(false);

let fetchRequestId = 0;

const fetchData = async (
	showLoading = true,
	autofresh = false,
	interaction = false
) => {
	const reqId = ++fetchRequestId;
	const shouldShowInteractionLoading = interaction && showLoading;
	const shouldShowInitialLoading =
		!initialLoadComplete.value &&
		!autofresh &&
		!shouldShowInteractionLoading &&
		allListedApps.value.length > 0;

	if (!autofresh) {
		clearLocker();
	}

	if (!allListedApps.value.length) {
		resetAppMetricsData();
		monitoringData.value = {
			cpu_usage: [],
			memory_usage: [],
			net_transmitted: [],
			net_received: []
		};
		fullSortedNamesRef.value = [];
		loading.value = false;
		interactionLoading.value = false;
		return;
	}

	if (shouldShowInteractionLoading || shouldShowInitialLoading) {
		resetAppMetricsData();
		interactionLoading.value = shouldShowInteractionLoading;
		loading.value = true;
		fullSortedNamesRef.value = [];
		const skelCap = skeletonCountCap(colCount.value, gridCfg);
		const skelN = Math.min(
			skelCap,
			itemsPerPage.value,
			Math.max(allListedApps.value.length, 1)
		);
		monitoringData.value = buildSkeletonMonitoringData(skelN) as any;
	}

	const preserveOrder =
		autofresh && fullSortedNamesRef.value.length > 0
			? [...fullSortedNamesRef.value]
			: undefined;

	try {
		const appsPayload = allListedApps.value.map((item) => ({
			...item,
			isSystem: item.namespace === userNamespace
		}));
		const metrics = await fetchWorkloadsMetrics(
			appsPayload,
			userNamespace,
			sort.value,
			autofresh,
			preserveOrder ? { preserveOrder } : undefined
		);
		if (reqId !== fetchRequestId) return;

		monitoringData.value = metrics as any;
		if (!autofresh) {
			const sortKey = selected.value.sortBy;
			const arr = (metrics as any)[sortKey] as any[];
			fullSortedNamesRef.value = (arr || []).map((r) => r.name);
		}
	} catch {
		//
	} finally {
		if (reqId === fetchRequestId) {
			loading.value = false;
			interactionLoading.value = false;
			if (!initialLoadComplete.value && allListedApps.value.length > 0) {
				initialLoadComplete.value = true;
			}
			if (isRunning.value) {
				refresh();
			}
		}
	}
};

const infiniteScrollDisabled = computed(
	() =>
		viewLoading.value ||
		totalRowCount.value === 0 ||
		visibleLimit.value >= totalRowCount.value
);

const sortApps = computed(() => {
	const rows = displaySortedRows.value;
	const limit = Math.min(visibleLimit.value, rows.length);
	return rows.slice(0, limit);
});

const resetVisibleWindow = () => {
	visibleLimit.value = targetVisibleAfterReset.value;
	infiniteResetKey.value += 1;
};

const syncFullSortedNamesFromMonitoring = () => {
	const sortKey = selected.value.sortBy;
	const arr = (monitoringData.value as any)[sortKey] as any[];
	fullSortedNamesRef.value = (arr || []).map((r) => r.name);
};

let infiniteAppendInFlight = false;

const onInfiniteLoad = async (
	_index: number,
	done: (stop?: boolean) => void
) => {
	if (viewLoading.value) {
		done();
		return;
	}
	const total = totalRowCount.value;
	if (visibleLimit.value >= total) {
		done(true);
		return;
	}
	if (infiniteAppendInFlight) {
		done();
		return;
	}
	infiniteAppendInFlight = true;
	const step = computeAppendStep({
		visible: visibleLimit.value,
		total,
		itemsPerPage: itemsPerPage.value,
		cols: colCount.value,
		cfg: gridCfg
	});
	try {
		visibleLimit.value = Math.min(visibleLimit.value + step, total);
		await nextTick();
	} finally {
		infiniteAppendInFlight = false;
	}
	done();
};

const dataFilter = (type: any, appName: string) => {
	const list = monitoringData.value[type] as any[];
	return list?.find((item: any) => item.name === appName);
};

const resortMonitoringDataLocally = () => {
	if (viewLoading.value) return;
	const hasMetricRows = MONITORING_KEYS.some((key) =>
		((monitoringData.value as any)[key] || []).some((row: any) => row?.name)
	);
	if (!hasMetricRows) return;
	monitoringData.value = mergeMonitoringData(
		monitoringData.value,
		{},
		sort.value
	) as any;
	syncFullSortedNamesFromMonitoring();
};

const invalidateInFlightMainFetch = () => {
	fetchRequestId += 1;
	cancelMainMetricsFetch();
	clearLocker();
	if (isRunning.value) {
		refresh();
	}
};

const sortHandler = (value: string) => {
	invalidateInFlightMainFetch();
	sort.value = value;
	resetVisibleWindow();
	resortMonitoringDataLocally();
};

const authLevelFilter = (state: EntranceState) => {
	return state === EntranceState.Public ? capitalize(state) : '';
};
let locker: Locker = undefined;
const clearLocker = () => {
	locker && clearTimeout(locker);
};

let pollingPaused = false;

const refresh = () => {
	clearLocker();
	if (pollingPaused) return;
	locker = setTimeout(() => {
		fetchData(false, true, false);
	}, 5 * 1000);
};

const toggleRunning = () => {
	isRunning.value = !isRunning.value;
	if (isRunning.value) {
		void fetchData(false, true, false);
	} else {
		clearLocker();
	}
};

const onSortMetricSelect = () => {
	invalidateInFlightMainFetch();
	resetVisibleWindow();
	resortMonitoringDataLocally();
};

resetAppMetricsData();

const MONITORING_KEYS = [
	'cpu_usage',
	'memory_usage',
	'net_transmitted',
	'net_received'
] as const;

let partialQueue: Promise<void> = Promise.resolve();
const enqueuePartial = (task: () => Promise<void>) => {
	const next = partialQueue.then(task, task);
	partialQueue = next.catch(() => undefined);
	return next;
};

const buildAppPayload = (app: any) => ({
	...app,
	isSystem: app.namespace === userNamespace
});

const applyAppRemoval = (removedNames: string[]) => {
	if (!removedNames.length) return;
	const removedSet = new Set(removedNames);

	const removedMeta: any[] = [];
	MONITORING_KEYS.forEach((key) => {
		const rows: any[] = (monitoringData.value as any)[key] || [];
		rows.forEach((row: any) => {
			if (
				row?.name &&
				removedSet.has(row.name) &&
				!removedMeta.find((m) => m.name === row.name)
			) {
				removedMeta.push({
					name: row.name,
					namespace: row.namespace,
					deployment: row.deployment,
					isSystem: !!row.isSystem
				});
			}
		});
	});

	monitoringData.value = removeMonitoringDataApps(
		monitoringData.value,
		removedNames
	) as any;

	fullSortedNamesRef.value = fullSortedNamesRef.value.filter(
		(n) => !removedSet.has(n)
	);

	const remainingApps = allListedApps.value.map(buildAppPayload);
	removeAppsFromMetricsCache(removedMeta, remainingApps);
};

const applyAppAddition = async (addedNames: string[]) => {
	if (!addedNames.length) return;
	const addedSet = new Set(addedNames);
	const addedApps = allListedApps.value
		.filter((item: any) => addedSet.has(item.name))
		.map(buildAppPayload);
	if (!addedApps.length) return;

	clearLocker();
	cancelMainMetricsFetch();
	pollingPaused = true;

	try {
		const partial = await fetchWorkloadsMetricsForApps(
			addedApps,
			userNamespace,
			sort.value
		);
		if (!partial) return;

		monitoringData.value = mergeMonitoringData(
			monitoringData.value,
			partial,
			sort.value
		) as any;

		syncFullSortedNamesFromMonitoring();
	} catch {
		//
	} finally {
		pollingPaused = false;
		if (isRunning.value) {
			refresh();
		}
	}
};

watch(
	currentAppNames,
	(newNames) => {
		const prev = prevAppsNamesRef.value;
		const newList = Array.isArray(newNames) ? newNames : [];

		if (!initialLoadComplete.value) {
			prevAppsNamesRef.value = [...newList];
			if (isEmpty(newList)) {
				void fetchData(false, false, false);
				return;
			}
			resetVisibleWindow();
			void fetchData(false, false, false);
			return;
		}

		const newSet = new Set(newList);
		const prevSet = new Set(prev);
		const added = newList.filter((n) => !prevSet.has(n));
		const removed = prev.filter((n) => !newSet.has(n));

		prevAppsNamesRef.value = [...newList];

		if (!added.length && !removed.length) {
			return;
		}

		void enqueuePartial(async () => {
			if (removed.length) {
				applyAppRemoval(removed);
			}
			if (added.length) {
				await applyAppAddition(added);
			}
		});
	},
	{ immediate: true }
);

watch(name, () => {
	resetVisibleWindow();
});

watch(locale, () => {
	resetVisibleWindow();
	fetchData(true, false, true);
});

watch(totalRowCount, (total) => {
	if (total === 0) return;
	visibleLimit.value = clampVisibleLimit(visibleLimit.value, total);
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', updateInfiniteScrollOffset);
	clearLocker();
	gridResizeObserver?.disconnect();
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
.applications2-grid {
	min-height: 120px;
}
</style>
