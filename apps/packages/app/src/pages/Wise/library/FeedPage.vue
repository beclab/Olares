<template>
	<common-page-container :menu-type="currentMenu" margin="80px">
		<template v-slot:title-right>
			<q-btn
				class="btn-size-sm btn-no-text btn-no-border"
				color="ink-2"
				outline
				no-caps
				:loading="loading"
				:icon="setUnRead ? 'sym_r_format_list_bulleted' : 'sym_r_checklist_rtl'"
				@click="readAll()"
			>
				<bt-tooltip
					:label="
						setUnRead ? t('main.mask_all_unseen') : t('main.mask_all_seen')
					"
				/>
				<template v-slot:loading>
					<bt-loading :loading="loading" />
				</template>
			</q-btn>
		</template>
		<template v-slot:tab-content-1>
			<source-entry-list :array="unseen" :time-type="SORT_TYPE.PUBLISHED" />
		</template>
		<template v-slot:tab-content-2>
			<source-entry-list :array="seen" :time-type="SORT_TYPE.PUBLISHED" />
		</template>
	</common-page-container>
</template>

<script lang="ts" setup>
import BtTooltip from '../../../components/base/BtTooltip.vue';
import SourceEntryList from './content/SourceEntryList.vue';
import CommonPageContainer from './content/CommonPageContainer.vue';
import { MenuType, TabType } from '../../../utils/rss-menu';
import BtLoading from '../../../components/base/BtLoading.vue';
import { useConfigStore } from '../../../stores/rss-config';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { liveQuery } from '../database/sqliteService';
import { useRssStore } from '../../../stores/rss';
import { onActivated, onDeactivated } from 'vue-demi';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Entry, SORT_TYPE } from '../../../utils/rss-types';
import { useReaderStore } from '../../../stores/rss-reader';
import HotkeyManager from '../../../directives/hotkeyManager';
// import { paramsToEntry } from '../../Wise/database/tables/entry';

const { t } = useI18n();
const rssStore = useRssStore();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const currentMenu = MenuType.History;
const loading = ref(false);
const setUnRead = computed(() => {
	return configStore.menuChoice.tab === TabType.Seen;
});
const unseen = ref([]);
const seen = ref([]);
let subscriptionSeen: any;
let subscriptionUnseen: any;

const updateNavigationList = (tab: TabType, list: Entry[]) => {
	if (configStore.menuChoice.tab === tab) {
		readerStore.setNavigationList(list);
	}
};

onActivated(() => {
	// HotkeyManager.setScope(MenuType.Feed);
	HotkeyManager.logAllKeyCodes();
	subscriptionUnseen = liveQuery(
		'unreadFeed',
		"SELECT entries.* FROM entries WHERE unread = true AND EXISTS (SELECT 1 FROM json_each(sources) WHERE value = 'wise')"
	).subscribe((data) => {
		if (data && data.length > 0) {
			console.log('get data 1 !!', data);
			unseen.value = data;
		} else {
			unseen.value = [];
		}
		updateNavigationList(TabType.UnSeen, unseen.value);
	});

	subscriptionSeen = liveQuery(
		'readFeed',
		"SELECT entries.* FROM entries WHERE unread = false AND EXISTS (SELECT 1 FROM json_each(sources) WHERE value = 'wise')"
	).subscribe((data) => {
		if (data && data.length > 0) {
			console.log('get data 2 !!', data);
			seen.value = data;
		} else {
			seen.value = [];
		}
		updateNavigationList(TabType.Seen, seen.value);
	});
});

watch(
	() => configStore.menuChoice,
	() => {
		// if (configStore.menuChoice.type === MenuType.Feed) {
		// 	updateNavigationList(TabType.UnSeen, unseen.value);
		// 	updateNavigationList(TabType.Seen, seen.value);
		// }
	},
	{
		deep: true,
		immediate: true
	}
);

onDeactivated(() => {
	subscriptionUnseen.unsubscribe();
	subscriptionSeen.unsubscribe();
});

async function readAll() {
	if (setUnRead.value) {
		const ids = seen.value.map((item) => {
			return item.id;
		});
		if (ids.length > 0) {
			console.log(ids);
			loading.value = true;
			rssStore.markEntryUnread(ids, true).finally(() => {
				loading.value = false;
			});
		} else {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('base.no_matching_content')
			});
		}
	} else {
		const ids = unseen.value.map((item) => {
			return item.id;
		});
		if (ids.length > 0) {
			console.log(ids);
			loading.value = true;
			rssStore.markEntryUnread(ids, false).finally(() => {
				loading.value = false;
			});
		} else {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('base.no_matching_content')
			});
		}
	}
}
</script>
<style lang="scss" scoped></style>
