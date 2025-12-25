<template>
	<div class="wise-page-root bg-color-white column justify-start">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="
						filterInfo
							? filterInfo.system
								? t(`main.${filterInfo.name}`)
								: filterInfo.name
							: ''
					"
					:icon="
						filterInfo && filterInfo.icon ? filterInfo.icon : 'sym_r_filter'
					"
					margin="180px"
				>
					<template v-slot:more>
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_keyboard_arrow_down"
							color="ink-2"
							outline
							no-caps
						>
							<bt-popup v-if="filterInfo" style="width: 200px">
								<bt-popup-item
									v-close-popup
									class="q-mt-xs"
									:title="t('main.split_view')"
									@on-item-click="onSplitViewClick"
								/>
								<bt-popup-item
									v-close-popup
									:title="
										filterInfo.showbadge
											? t('main.hide_unread_badge')
											: t('main.show_unread_badge')
									"
									class="q-mt-xs"
									@on-item-click="
										updateFilter({ showbadge: !filterInfo.showbadge }, false)
									"
									:hotkey="WISE_HOTKEY.TAB.UNREAD_NUM"
								/>
								<bt-popup-item
									v-close-popup
									class="q-mt-xs"
									:title="
										filterInfo.pin
											? t('main.unpin_from_menu')
											: t('main.pin_from_menu')
									"
									@on-item-click="updateFilter({ pin: !filterInfo.pin }, false)"
									:hotkey="WISE_HOTKEY.TAB.PIN"
								/>
							</bt-popup>
						</q-btn>
					</template>
					<template v-slot:end>
						<q-tabs
							v-if="menuTabs.length > 0"
							v-model="configStore.menuChoice.tab"
							dense
							class="wise-page-tabs"
							active-color="orange-default"
							indicator-color="orange-default"
							align="left"
							:breakpoint="0"
						>
							<template v-for="item in menuTabs" :key="item">
								<q-tab class="wise-page-tab" :name="item">
									<template v-slot:default>
										<tab-item
											:hide-border="true"
											:selected="configStore.menuChoice.tab === item"
											:title="TabInfoMap[item]?.title"
											@click="configStore.setMenuTab(item)"
										/>
									</template>
								</q-tab>
							</template>
						</q-tabs>
					</template>
				</bt-breadcrumbs>
			</template>

			<template v-slot:after>
				<title-right-layout>
					<q-btn
						v-if="
							filterInfo &&
							!(
								filterInfo.splitview == SPLIT_TYPE.SEEN &&
								configStore.menuChoice.tab === TabType.Seen
							)
						"
						class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
						color="ink-2"
						outline
						no-caps
						:disable="unreadAll"
						:loading="readAll && readSetLoading"
						icon="sym_r_checklist_rtl"
						@click="setReadAll(true)"
					>
						<template v-slot:loading>
							<bt-loading :loading="readSetLoading" />
						</template>
						<bt-tooltip
							:label="t('main.mask_all_seen')"
							:hotkey="WISE_HOTKEY.TAB.READ_ALL"
						/>
					</q-btn>
					<q-btn
						v-if="
							filterInfo &&
							!(
								filterInfo.splitview == SPLIT_TYPE.SEEN &&
								configStore.menuChoice.tab === TabType.UnSeen
							)
						"
						class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
						color="ink-2"
						outline
						no-caps
						:disable="readAll"
						:loading="unreadAll && readSetLoading"
						icon="sym_r_format_list_bulleted"
						@click="setReadAll(false)"
					>
						<template v-slot:loading>
							<bt-loading :loading="readSetLoading" />
						</template>
						<bt-tooltip
							:label="t('main.mask_all_unseen')"
							:hotkey="WISE_HOTKEY.TAB.UNREAD_ALL"
						/>
					</q-btn>
					<q-btn
						class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_sort"
						color="ink-2"
						outline
						no-caps
					>
						<bt-tooltip :label="t('main.sort')" />
						<bt-popup style="width: 176px">
							<div class="q-mt-sm q-ml-sm text-ink-3 text-overline">
								{{ t('main.sort_by') }}
							</div>
							<!--							<bt-popup-item-->
							<!--								v-close-popup-->
							<!--								:title="t('main.date_updated')"-->
							<!--								class="q-mt-xs"-->
							<!--								:selected="-->
							<!--									filterInfo && filterInfo.sortby && filterInfo.sortby === SORT_TYPE.UPDATED-->
							<!--								"-->
							<!--								@on-item-click="updateFilter({ sortby: SORT_TYPE.UPDATED })"-->
							<!--							/>-->
							<bt-popup-item
								v-close-popup
								class="q-mt-xs"
								:selected="
									filterInfo &&
									filterInfo.sortby &&
									filterInfo.sortby === SORT_TYPE.PUBLISHED
								"
								:title="t('main.date_published')"
								@on-item-click="updateFilter({ sortby: SORT_TYPE.PUBLISHED })"
							/>
							<bt-popup-item
								v-close-popup
								:title="t('main.date_created')"
								class="q-mt-xs"
								:selected="
									filterInfo &&
									filterInfo.sortby &&
									filterInfo.sortby === SORT_TYPE.CREATED
								"
								@on-item-click="updateFilter({ sortby: SORT_TYPE.CREATED })"
							/>
							<div
								class="full-width bg-separator q-mt-sm"
								style="height: 1px"
							/>
							<div class="q-mt-sm q-ml-sm text-ink-3 text-overline">
								{{ t('main.order_by') }}
							</div>
							<bt-popup-item
								v-close-popup
								class="q-mt-xs"
								:title="t('main.recent_to_old')"
								:selected="
									filterInfo &&
									filterInfo.orderby &&
									filterInfo.orderby === ORDER_TYPE.DESC
								"
								@on-item-click="updateFilter({ orderby: ORDER_TYPE.DESC })"
							/>
							<bt-popup-item
								v-close-popup
								class="q-mt-xs"
								:title="t('main.old_to_recent')"
								:selected="
									filterInfo &&
									filterInfo.orderby &&
									filterInfo.orderby === ORDER_TYPE.ASC
								"
								@on-item-click="updateFilter({ orderby: ORDER_TYPE.ASC })"
							/>
						</bt-popup>
					</q-btn>
				</title-right-layout>
			</template>
		</title-bar>
		<div class="wise-page-content column justify-start">
			<q-tab-panels
				v-if="menuTabs.length > 0"
				v-model="configStore.menuChoice.tab"
				animated
				class="wise-page-tab-panels"
				keep-alive
			>
				<template v-for="item in menuTabs" :key="item">
					<q-tab-panel :name="item" class="wise-page-tab-panel">
						<source-entry-list
							:array="listMap[item] || []"
							:time-type="filterInfo ? filterInfo.sortby : SORT_TYPE.PUBLISHED"
						/>
					</q-tab-panel>
				</template>
			</q-tab-panels>
			<div v-else class="full-width full-height">
				<source-entry-list
					:array="listMap[TabType.Empty] || []"
					:time-type="filterInfo ? filterInfo.sortby : SORT_TYPE.PUBLISHED"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import SplitViewDialog from '../../../components/rss/dialog/SplitViewDialog.vue';
import TitleRightLayout from '../../../components/base/TitleRightLayout.vue';
import { MenuType, TabInfoMap, TabType } from '../../../utils/rss-menu';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import BtPopupItem from '../../../components/base/BtPopupItem.vue';
import BtTooltip from '../../../components/base/BtTooltip.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import BtPopup from '../../../components/base/BtPopup.vue';
import SourceEntryList from './content/SourceEntryList.vue';
import TabItem from '../../../components/rss/TabItem.vue';
import { useConfigStore } from '../../../stores/rss-config';
import { useFilterStore } from '../../../stores/rss-filter';
import BtLoading from '../../../components/base/BtLoading.vue';
import { WISE_HOTKEY } from '../../../directives/wiseHotkey';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import HotkeyManager from '../../../directives/hotkeyManager';
import { useReaderStore } from '../../../stores/rss-reader';
import { FilterFormat } from '../database/filterFormat';
import { onActivated, onDeactivated } from 'vue-demi';
import { liveQuery } from '../database/sqliteService';
import { useRssStore } from '../../../stores/rss';
import { ref, reactive, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useRoute } from 'vue-router';
import {
	Entry,
	FilterInfo,
	ORDER_TYPE,
	SORT_TYPE,
	SPLIT_TYPE
} from '../../../utils/rss-types';
import { useAbilityStore } from '../../../stores/rss-ability';
import { notifyFailed } from '../../../utils/settings/btNotify';

const { t } = useI18n();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const filterStore = useFilterStore();
const $q = useQuasar();
const readSetLoading = ref(false);
const readAll = ref(false);
const unreadAll = ref(false);
const rssStore = useRssStore();
const route = useRoute();
let listMap = reactive<Record<string, Entry[]>>({});
let subscriptionMap: Map<string, any> = new Map();
const menuTabs = ref<any[]>([]);
const refreshTabs = ref<any[]>([]);
let firstLoadTimer: NodeJS.Timer | null = null;
let filterInfo = ref<FilterInfo | undefined>(undefined);
let hotKeyMap;
const abilityStore = useAbilityStore();

onActivated(async () => {
	console.log('[RssCustomPage] Page activated');
	hotKeyMap = {
		[WISE_HOTKEY.TAB.SPLIT_NEXT]: () => {
			let nextSplit;
			if (!filterInfo.value) {
				return;
			}
			switch (filterInfo.value.splitview) {
				case SPLIT_TYPE.LOCATION:
					nextSplit = SPLIT_TYPE.SEEN;
					break;
				case SPLIT_TYPE.SEEN:
					nextSplit = SPLIT_TYPE.NONE;
					break;
				case SPLIT_TYPE.NONE:
					nextSplit = SPLIT_TYPE.LOCATION;
					break;
			}
			if (!nextSplit) {
				return;
			}
			updateFilterBySpite(nextSplit);
		},
		[WISE_HOTKEY.TAB.SPLIT_PRE]: () => {
			let preSplit;
			if (!filterInfo.value) {
				return;
			}
			switch (filterInfo.value.splitview) {
				case SPLIT_TYPE.LOCATION:
					preSplit = SPLIT_TYPE.NONE;
					break;
				case SPLIT_TYPE.SEEN:
					preSplit = SPLIT_TYPE.LOCATION;
					break;
				case SPLIT_TYPE.NONE:
					preSplit = SPLIT_TYPE.SEEN;
					break;
			}
			if (!preSplit) {
				return;
			}
			updateFilterBySpite(preSplit);
		},
		[WISE_HOTKEY.TAB.READ_ALL]: () => {
			if (readAll.value && readSetLoading.value) {
				return;
			}
			if (!filterInfo.value) {
				return;
			}
			if (filterInfo.value.splitview !== SPLIT_TYPE.SEEN) {
				return;
			}
			setReadAll(true);
		},
		[WISE_HOTKEY.TAB.UNREAD_ALL]: () => {
			if (unreadAll.value && readSetLoading.value) {
				return;
			}
			if (!filterInfo.value) {
				return;
			}
			if (filterInfo.value.splitview !== SPLIT_TYPE.SEEN) {
				return;
			}
			setReadAll(false);
		},
		[WISE_HOTKEY.TAB.UNREAD_NUM]: () => {
			if (!filterInfo.value) {
				return;
			}
			updateFilter({ showbadge: !filterInfo.value.showbadge }, false);
		},
		[WISE_HOTKEY.TAB.PIN]: () => {
			if (!filterInfo.value) {
				return;
			}
			updateFilter({ pin: !filterInfo.value.pin }, false);
		}
	};
	HotkeyManager.setScope(MenuType.Custom);
	HotkeyManager.registerHotkeys(hotKeyMap, [MenuType.Custom]);
	firstLoadTimer = setInterval(() => {
		firstLoad();
	}, 500);

	if (
		filterInfo.value &&
		filterInfo.value.name == 'feeds' &&
		filterInfo.value.system
	) {
		await abilityStore.getAbiAbility();
		if (!abilityStore.rssubscribe) {
			notifyFailed(t('Rss Subscribe not installed'));
		}
	}
});

onDeactivated(() => {
	console.log('[RssCustomPage] Page deactivated');
	if (hotKeyMap) {
		HotkeyManager.unregisterHotkeys(hotKeyMap, [MenuType.Custom]);
	}
	if (firstLoadTimer) {
		clearInterval(firstLoadTimer);
	}
	unsubscribeAll();
});

const firstLoad = () => {
	console.log('[RssCustomPage] query...');
	if (
		configStore.menuInited &&
		filterStore.inited &&
		refreshTabs.value.length > 0 &&
		filterInfo.value &&
		subscriptionMap.size === 0
	) {
		clearInterval(firstLoadTimer);
		refreshTabs.value.forEach(tabSubscribe);
	}
};

const setReadAll = async (read: boolean) => {
	const currentTab =
		menuTabs.value.length > 0 ? configStore.menuChoice.tab : TabType.Empty;
	const list = listMap[currentTab] || [];
	const ids = list
		.filter((item) => item.unread !== read)
		.map((item) => item.id);

	if (ids.length === 0) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('base.no_matching_content')
		});
		console.warn(
			'[RssCustomPage] No matching content, no need to update read status'
		);
		return;
	}

	readSetLoading.value = true;
	read ? (readAll.value = true) : (unreadAll.value = true);

	try {
		await rssStore.markEntryUnread(ids, !read);
		console.log(
			`[RssCustomPage] Successfully batch set read status, read: ${read}`
		);
	} catch (error) {
		console.error('[RssCustomPage] Failed to batch set read status', error);
	} finally {
		readAll.value = unreadAll.value = readSetLoading.value = false;
	}
};

const updateNavigationList = (tab: TabType | string, list: Entry[]) => {
	if (configStore.menuChoice.tab === tab) {
		console.log(
			`[RssCustomPage] Update navigation list, current tab=${tab}, data length=${
				list ? list.length : 0
			}`
		);
		readerStore.setNavigationList(list);
	}
};

const onSplitViewClick = () => {
	$q.dialog({
		component: SplitViewDialog,
		componentProps: { filter: filterInfo.value }
	}).onOk((splitview) => {
		updateFilterBySpite(splitview);
	});
};

const updateFilterBySpite = (splitview: SPLIT_TYPE) => {
	console.log('[RssCustomPage] SplitViewDialog confirmed ' + splitview);
	const newFilter = { ...filterInfo.value, splitview };
	filterInfo.value = newFilter;
	filterStore.modifyFilter(newFilter);
	filterStore.updateMenuBySplit(filterInfo.value.id, splitview, true);
	//refresh immediate
	menuTabs.value = configStore.userTabs.get(configStore.menuChoice.type) ?? [];
	refreshTabs.value =
		menuTabs.value.length > 0 ? menuTabs.value : [TabType.Empty];
	queryAllTabs();
};

const updateFilter = (params: any, refresh = true) => {
	if (filterInfo.value) {
		const newFilter = { ...filterInfo.value, ...params };
		filterInfo.value = newFilter;
		filterStore.modifyFilter(newFilter);
		if (refresh) {
			queryAllTabs();
		}
	}
};

const queryAllTabs = () => {
	unsubscribeAll();
	listMap = reactive<Record<string, Entry[]>>({});
	refreshTabs.value.forEach(tabSubscribe);
};

function tabSubscribe(tab: string) {
	if (!filterInfo.value) {
		console.warn(
			`[RssCustomPage] Filter is empty, cannot subscribe to tab: ${tab}`
		);
		return;
	}

	console.log(`[RssCustomPage] Start subscription for tab: ${tab}`);

	subscriptionMap.set(
		tab,
		liveQuery(
			`${filterInfo.value.id}_${tab}`,
			FilterFormat.fromFilterInfo(filterInfo.value, tab).buildQuery()
		)
			.subscribe((res: any, id: string) => {
				const updateTab = id.split('_')[1];
				listMap[updateTab] = res;
				updateNavigationList(updateTab, res);
				console.log(
					`[RssCustomPage] Data loaded successfully, tab: ${updateTab}`
				);
			})
			.catch((error) => {
				console.error(`[RssCustomPage] Query failed, tab: ${tab}`, error);
				listMap[tab] = [];
			})
	);
}

const unsubscribeAll = () => {
	subscriptionMap.forEach((sub, key) => {
		sub?.unsubscribe();
		console.log(`[RssCustomPage] Unsubscribe: ${key}`);
	});
	subscriptionMap.clear();
};

watch(
	() => configStore.menuChoice,
	() => {
		console.log('[RssCustomPage] Get tabs:', configStore.menuChoice);
		menuTabs.value =
			configStore.userTabs.get(configStore.menuChoice.type) ?? [];
		refreshTabs.value =
			menuTabs.value.length > 0 ? menuTabs.value : [TabType.Empty];
	},
	{
		immediate: true,
		deep: true
	}
);

watch(
	() => filterStore.inited,
	() => {
		if (filterStore.inited) {
			filterInfo.value = filterStore.getFilterListById(
				route.params.filterId as string
			);
			console.log('sdasdadsasdadd', filterInfo.value);
		}
	},
	{
		immediate: true
	}
);

watch(
	() => configStore.menuChoice.tab,
	() => {
		updateNavigationList(
			configStore.menuChoice.tab,
			listMap[configStore.menuChoice.tab]
		);
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped></style>
