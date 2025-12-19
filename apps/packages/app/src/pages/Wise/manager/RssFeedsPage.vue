<template>
	<div class="feed-page column">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('main.rss_feeds')"
					icon="sym_r_rss_feed"
					margin="400px"
				/>
			</template>

			<template v-slot:after>
				<div class="row justify-end items-center" style="margin-right: 44px">
					<rss-search
						v-model="search"
						class="q-mr-lg q-mt-xs"
						style="width: 200px"
					/>
					<!--					<q-btn-->
					<!--						class="q-mr-sm btn-size-sm"-->
					<!--						:label="t('base.add_feed')"-->
					<!--						color="orange-6"-->
					<!--						@click="addFeed"-->
					<!--						no-caps-->
					<!--					/>-->
				</div>
			</template>
		</title-bar>

		<bt-scroll-area class="feed-tab">
			<q-table
				:pagination="initialPagination"
				:rows="rows"
				class="bg-background-1"
				flat
				:columns="columns"
				row-key="name"
			>
				<template v-slot:header="props">
					<q-tr :props="props" style="height: 32px">
						<q-th
							v-for="col in props.cols"
							:key="col.name"
							:props="props"
							class="recommend-header-field text-body3 text-ink-3"
						>
							{{ col.label }}
						</q-th>
					</q-tr>
				</template>
				<template v-slot:body-cell-name="props">
					<q-td :props="props" class="left-align text-subtitle3 text-ink-1">
						<div class="row justify-start items-center">
							<feed-icon
								:feed="props.row.feed"
								size="24px"
								style="margin-right: 8px"
							/>
							<div>{{ props.row.name }}</div>
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-description="props">
					<q-td :props="props" class="left-align text-body2 text-ink-2">
						{{ props.row.description }}
					</q-td>
				</template>

				<template v-slot:body-cell-documents="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2"
						style="text-align: center"
					>
						{{ props.row.documents }}
					</q-td>
				</template>

				<template v-slot:body-cell-views="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 cursor-pointer"
					>
						<div
							v-if="
								filterStore.feedMap.get(props.row.id) &&
								filterStore.feedMap.get(props.row.id)?.size > 0
							"
							class="row"
						>
							<template
								v-for="item in filterStore.feedMap.get(props.row.id)"
								:key="item.id"
							>
								<create-view class="q-mr-xs q-my-xs" :name="item.name" />
							</template>
						</div>
						<div class="text-ink-3 text-body3" v-else>
							{{ t('main.manager_views') }}
						</div>

						<view-edit-popup :data="props.row" type="feed_id" />
					</q-td>
				</template>

				<template v-slot:body-cell-lastUpdated="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2"
						style="text-align: right"
					>
						{{ getPastTime(new Date(), new Date(props.row.lastUpdated)) }}
					</q-td>
				</template>

				<template v-slot:body-cell-operations="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2"
						style="text-align: right"
					>
						<div class="row justify-end items-center">
							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_content_copy"
								color="ink-2"
								outline
								@click.stop
								no-caps
								@click="copyUrl(props.row.feed)"
							>
								<bt-tooltip :label="t('base.copy')" />
							</q-btn>

							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_edit_square"
								color="ink-2"
								outline
								@click.stop="editFeed(props.row.feed)"
								no-caps
							>
								<bt-tooltip :label="t('base.edit')" />
							</q-btn>
							<q-btn
								class="btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_delete"
								color="ink-2"
								outline
								:loading="props.row.loading"
								@click.stop
								no-caps
								@click="deleteFeed(props.row)"
							>
								<bt-tooltip :label="t('base.remove')" />
								<template v-slot:loading>
									<bt-loading :loading="props.row.loading" />
								</template>
							</q-btn>
						</div>
					</q-td>
				</template>
				<template v-slot:no-data>
					<empty-view :is-table="true" />
				</template>
			</q-table>
		</bt-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import BaseCheckBoxDialog from '../../../components/base/BaseCheckBoxDialog.vue';
import FeedAddDialog from '../../../components/rss/dialog/FeedAddDialog.vue';
import FeedEditDialog from '../../../components/rss/dialog/FeedEditDialog.vue';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import ViewEditPopup from '../../../components/rss/ViewEditPopup.vue';
import CreateView from '../../../components/rss/CreateView.vue';
import BtLoading from '../../../components/base/BtLoading.vue';
import FeedIcon from '../../../components/rss/FeedIcon.vue';
import EmptyView from '../../../components/rss/EmptyView.vue';
import BtTooltip from '../../../components/base/BtTooltip.vue';
import RssSearch from '../../../components/rss/RssSearch.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import { getEntriesByFeedId } from '../database/tables/entry';
import { useConfigStore } from '../../../stores/rss-config';
import { useFilterStore } from '../../../stores/rss-filter';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { notifyFailed } from '../../../utils/settings/btNotify';
import { useAbilityStore } from '../../../stores/rss-ability';
import { Feed, SOURCE_TYPE } from '../../../utils/rss-types';
import { getPastTime } from '../../../utils/rss-utils';
import { useQuasar } from 'quasar';
import { useRssStore } from '../../../stores/rss';
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { onActivated } from 'vue-demi';
import { getApplication } from '../../../application/base';

const { t } = useI18n();
const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: t('base.feed_name'),
		field: 'name'
	},
	{
		name: 'description',
		align: 'left',
		label: t('base.description'),
		field: 'description'
	},
	{
		name: 'documents',
		align: 'center',
		label: t('base.documents'),
		field: 'documents'
	},
	{
		name: 'views',
		align: 'left',
		label: t('base.add_view'),
		field: 'views'
	},
	{
		name: 'lastUpdated',
		align: 'right',
		label: t('base.last_updated'),
		field: 'lastUpdated'
	},
	{
		name: 'operations',
		align: 'right',
		label: t('base.operations'),
		field: 'operations'
	}
];

const $q = useQuasar();
const rows = ref<any[]>([]);
const rssStore = useRssStore();
const abilityStore = useAbilityStore();
const search = ref('');
const configStore = useConfigStore();
const filterStore = useFilterStore();

onActivated(async () => {
	await abilityStore.getAbiAbility();
	if (!abilityStore.rssubscribe) {
		notifyFailed(t('Rss Subscribe not installed'));
	}
});

const initialPagination = ref({
	page: 0,
	rowsPerPage: 100
});

watch(
	() => [rssStore.feeds, search.value],
	async () => {
		const newRows = [];
		const list = rssStore.feeds.filter((feed) =>
			feed.sources.includes(SOURCE_TYPE.WISE)
		);
		console.log('==!!!! list', list);
		for (const feed of list) {
			if (!search.value || feed.title.includes(search.value)) {
				const list = await getEntriesByFeedId(feed.id);
				console.log('==!!!! push', feed);
				newRows.push({
					id: feed.id,
					feed: feed,
					name: feed.title ? feed.title : feed.feed_url,
					description: feed.description,
					documents: list.length,
					lastUpdated: new Date(feed.updated_at).getTime(),
					loading: false
				});
			}
		}
		rows.value = newRows;
		console.log('==!!!! rows', rows.value);
	},
	{
		immediate: true,
		deep: true
	}
);

const copyUrl = (feed: Feed) => {
	getApplication()
		.copyToClipboard(feed.feed_url)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('copy_success')
			});
		})
		.catch((e) => {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('copy_failure_message', e.message)
			});
		});
};

const addFeed = () => {
	$q.dialog({
		component: FeedAddDialog
	});
};

const deleteFeed = (row: any) => {
	const remove = configStore.feedRemoveWithFile;
	$q.dialog({
		component: BaseCheckBoxDialog,
		componentProps: {
			label: t('dialog.remove_subscription'),
			content: t('dialog.remove_subscription_desc'),
			modelValue: remove,
			showCheckbox: true,
			boxLabel: t('dialog.delete_the_files_if_present')
		}
	})
		.onOk(async (selected) => {
			console.log('remove task ok');
			row.loading = true;
			configStore.setFeedRemoveWithFile(selected);
			rssStore.removeFeed(row.feed.feed_url, selected).finally(() => {
				row.loading = false;
			});
		})
		.onCancel(() => {
			console.log('remove task cancel');
		});
};

const editFeed = (feed: Feed) => {
	$q.dialog({
		component: FeedEditDialog,
		componentProps: {
			feed
		}
	});
};
</script>

<style scoped lang="scss">
.feed-page {
	height: 100%;
	width: 100%;

	.feed-tab {
		width: 100%;
		height: calc(100% - 56px);
		padding-left: 44px;
		padding-right: 44px;
	}
}

::v-deep(.q-btn.text-grey-8:before) {
	border: unset !important;
}
</style>
