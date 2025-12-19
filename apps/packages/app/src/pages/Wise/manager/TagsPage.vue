<template>
	<div class="tag-page column">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('base.tags')"
					icon="sym_r_sell"
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

					<q-btn
						class="q-mr-sm btn-size-sm"
						:label="t('base.add_tag')"
						color="orange-6"
						@click="addTag"
						no-caps
					/>
				</div>
			</template>
		</title-bar>

		<bt-scroll-area class="tag-tab">
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
						{{ props.row.name }}
					</q-td>
				</template>

				<template v-slot:body-cell-highlights="props">
					<q-td :props="props" class="left-align text-body2 text-ink-2">
						{{ props.row.highlights }}
					</q-td>
				</template>

				<template v-slot:body-cell-views="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 cursor-pointer"
					>
						<div
							v-if="
								filterStore.labelMap.get(props.row.id) &&
								filterStore.labelMap.get(props.row.id)?.size > 0
							"
							class="row"
						>
							<template
								v-for="item in filterStore.labelMap.get(props.row.id)"
								:key="item.id"
							>
								<create-view class="q-mr-xs q-my-xs" :name="item.name" />
							</template>
						</div>
						<div class="text-ink-3 text-body3" v-else>
							{{ t('main.manager_views') }}
						</div>

						<view-edit-popup :data="props.row" type="tag_id" />
					</q-td>
				</template>

				<template v-slot:body-cell-lastUpdated="props">
					<q-td :props="props" class="left-align text-body2 text-ink-2">
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
								icon="sym_r_edit_square"
								color="ink-2"
								outline
								@click.stop
								no-caps
								@click="editTagName(props.row.label)"
							>
								<bt-tooltip :label="t('base.edit')" />
							</q-btn>

							<q-btn
								class="btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_delete"
								color="ink-2"
								outline
								@click.stop
								no-caps
								@click="deleteTag(props.row.label)"
							>
								<bt-tooltip :label="t('base.remove')" />
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
import TitleBar from '../../../components/rss/TitleBar.vue';
import BtTooltip from '../../../components/base/BtTooltip.vue';
import EmptyView from '../../../components/rss/EmptyView.vue';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import ViewEditPopup from '../../../components/rss/ViewEditPopup.vue';
import TagEditDialog from '../../../components/rss/dialog/TagEditDialog.vue';
import RssSearch from '../../../components/rss/RssSearch.vue';
import CreateView from '../../../components/rss/CreateView.vue';
import { useFilterStore } from '../../../stores/rss-filter';
import { getPastTime } from '../../../utils/rss-utils';
import { useRssStore } from '../../../stores/rss';
import { Label } from '../../../utils/rss-types';
import { BtDialog, useColor } from '@bytetrade/ui';
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const filterStore = useFilterStore();
const rows = ref<any[]>([]);
const rssStore = useRssStore();
const search = ref('');
const { t } = useI18n();
const $q = useQuasar();

const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: t('tag'),
		field: 'name'
	},
	{
		name: 'document',
		align: 'left',
		label: t('base.documents'),
		field: 'document'
	},
	{
		name: 'highlights',
		align: 'left',
		label: t('base.highlights'),
		field: 'highlights'
	},
	{
		name: 'views',
		align: 'left',
		label: t('base.add_view'),
		field: 'views'
	},
	{
		name: 'lastUpdated',
		align: 'left',
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

const initialPagination = ref({
	page: 0,
	rowsPerPage: 100
});

watch(
	() => [rssStore.labels, search],
	() => {
		const newRows = [];
		rssStore.labels.forEach((label) => {
			if (!search.value || label.name.includes(search.value)) {
				newRows.push({
					id: label.id,
					label: label,
					name: label.name,
					document: label.entries ? label.entries.length : 0,
					highlights: label.notes ? label.notes.length : 0,
					lastUpdated: label.updated_at
				});
			}
		});
		rows.value = newRows;
	},
	{
		immediate: true,
		deep: true
	}
);

const addTag = () => {
	$q.dialog({
		component: TagEditDialog
	});
};

const deleteTag = (label: Label) => {
	const { color: orange } = useColor('orange-default');
	const { color: textInk } = useColor('ink-on-brand');

	BtDialog.show({
		title: t('remove_tag'),
		message: t('remove_tag_desc'),
		okStyle: {
			background: orange.value,
			color: textInk.value
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				console.log('click ok');
				rssStore.removeLabel(label.id);
			} else {
				console.log('click cancel');
			}
		})
		.catch((err: Error) => {
			console.log('click ok', err);
		});
};

const editTagName = (label: Label) => {
	$q.dialog({
		component: TagEditDialog,
		componentProps: {
			data: label
		}
	});
};
</script>

<style scoped lang="scss">
.tag-page {
	height: 100%;
	width: 100%;

	.tag-tab {
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
