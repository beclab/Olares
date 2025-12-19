<template>
	<div class="filtered-page column">
		<title-bar>
			<template v-slot:before>
				<q-breadcrumbs dense class="left-align text-h6 text-ink-1">
					<div
						class="row justify-center items-center"
						style="
							width: 24px;
							height: 24px;
							margin-right: 8px;
							margin-left: 44px;
						"
					>
						<q-icon size="22px" name="sym_r_stacks" />
					</div>

					<q-breadcrumbs-el
						class="left-align text-h6 text-ink-1 cursor-pointer"
						:label="t('main.filtered_views')"
					/>
				</q-breadcrumbs>
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
						:label="t('base.add_view')"
						color="orange-6"
						@click="addView"
						no-caps
					/>
				</div>
			</template>
		</title-bar>

		<bt-scroll-area class="filtered-tab">
			<q-table
				:pagination="initialPagination"
				:rows="rows"
				class="bg-background-1"
				flat
				wrap-cells
				:columns="columns"
				row-key="name"
				@row-click="onItemClick"
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
					<q-td
						:props="props"
						class="left-align text-subtitle3 text-ink-1 filter-name"
					>
						<div class="row justify-start items-center filter-text">
							{{ getRowName(props.row) }}
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-description="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 filter-description"
					>
						<div class="filter-text">
							{{ getRowDescription(props.row) }}
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-documents="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 filter-documents"
						style="text-align: center"
					>
						<div class="filter-text">
							{{ getRowDocuments(props.row) }}
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-query="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 filter-query"
						style="text-align: left"
					>
						<div class="filter-text">
							{{ props.row.query }}
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-lastUpdated="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 filter-lastUpdated"
						style="text-align: right"
					>
						<div class="filter-text">
							{{ getPastTime(new Date(), new Date(props.row.updated_at)) }}
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-operations="props">
					<q-td
						:props="props"
						class="left-align text-body2 text-ink-2 filter-operations"
						style="text-align: right"
					>
						<div class="row justify-end items-center">
							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text btn-no-border text-ink-2"
								:icon="props.row.pin ? 'sym_r_keep_off' : 'sym_r_keep'"
								color="ink-2"
								outline
								@click.stop="
									filterStore.modifyFilter({
										...props.row,
										pin: !props.row.pin
									})
								"
								no-caps
							>
								<bt-tooltip
									:label="
										props.row.pin
											? t('main.unpin_from_menu')
											: t('main.pin_from_menu')
									"
								/>
							</q-btn>

							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_edit_square"
								color="ink-2"
								outline
								@click.stop="editView(props.row)"
								:disable="props.row.system"
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
								:disable="props.row.system"
								no-caps
								@click="deleteView(props.row)"
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
import FilterEditDialog from '../../../components/rss/dialog/FilterEditDialog.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import RssSearch from '../../../components/rss/RssSearch.vue';
import BtTooltip from '../../../components/base/BtTooltip.vue';
import EmptyView from '../../../components/rss/EmptyView.vue';
import BtLoading from '../../../components/base/BtLoading.vue';
import { useFilterStore } from '../../../stores/rss-filter';
import { getPastTime } from '../../../utils/rss-utils';
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { FilterInfo } from '../../../utils/rss-types';
import { useQuasar } from 'quasar';
import { useConfigStore } from '../../../stores/rss-config';
import { sendMessageToWorker } from '../database/sqliteService';
import { FilterFormat } from '../database/filterFormat';
import { BtDialog, useColor } from '@bytetrade/ui';

const search = ref();
const { t } = useI18n();
const $q = useQuasar();
const filterStore = useFilterStore();
const configStore = useConfigStore();
let listMap = reactive<Record<string, number>>({});

const rows = computed(() => {
	return filterStore.filterList
		.filter((item) => {
			if (search.value) {
				if (
					getRowName(item).toLowerCase().indexOf(search.value.toLowerCase()) >
						-1 ||
					getRowDescription(item)
						.toLowerCase()
						.indexOf(search.value.toLowerCase()) > -1
				) {
					return true;
				} else {
					return false;
				}
			}
			return true;
		})
		.map((item) => {
			return { ...item, loading: false };
		});
});

const getRowName = (item) => {
	if (item.system) {
		return t(`main.${item.name}`);
	}
	return item.name;
};

const getRowDescription = (item) => {
	if (item.system) {
		return t(`main.${item.name}_description`);
	}
	return item.description;
};

const getRowDocuments = (item) => {
	return listMap[item.id] || 0;
};

watch(
	() => filterStore.filterList,
	async () => {
		const tempListMap = {};

		await Promise.all(
			filterStore.filterList.map((item, index) =>
				sendMessageToWorker(
					'query',
					{
						sql: FilterFormat.fromFilterInfo(item).buildQuery()
					},
					item.id + '_filter_' + index
				).then((filter: any) => {
					tempListMap[item.id] = filter.length;
				})
			)
		);

		Object.assign(listMap, tempListMap);
	},
	{
		deep: true,
		immediate: true
	}
);

const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: t('base.name'),
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
		name: 'query',
		align: 'left',
		label: t('base.query'),
		field: 'query'
	},
	{
		name: 'lastUpdated',
		align: 'right',
		label: t('base.last_updated'),
		field: 'updated_at'
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

const onItemClick = (evt: any, row: FilterInfo) => {
	configStore.setMenuType(row.id, { filterId: row.id });
};

const addView = () => {
	$q.dialog({
		component: FilterEditDialog,
		componentProps: {
			createWithQuery: true
		}
	});
};

const editView = (info: FilterInfo) => {
	$q.dialog({
		component: FilterEditDialog,
		componentProps: {
			data: info
		}
	});
};

const { color: orange } = useColor('orange-default');
const { color: textInk } = useColor('ink-on-brand');

const deleteView = (row: any) => {
	BtDialog.show({
		title: t('dialog.remove_view'),
		message: t('dialog.remove_view_desc'),
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
				row.loading = true;
				filterStore.deleteFilter(row.id).finally(() => {
					row.loading = false;
				});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};
</script>

<style scoped lang="scss">
.filtered-page {
	height: 100%;
	width: 100%;

	.filtered-tab {
		width: 100%;
		height: calc(100% - 56px);
		padding-left: 44px;
		padding-right: 44px;

		.filter-text {
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 1;
			-webkit-box-orient: vertical;
		}

		.filter-name {
			min-width: 60px;
			max-width: 80px;
			padding: 0;
		}

		.filter-description {
			min-width: 130px;
			max-width: 330px;
			padding: 0;
		}

		.filter-documents {
			min-width: 100px;
			max-width: 120px;
		}

		.filter-query {
			min-width: 140px;
			max-width: 400px;
		}

		.filter-lastUpdated {
			min-width: 100px;
			max-width: 120px;
		}

		.filter-operations {
			min-width: 160px;
			max-width: 160px;
		}
	}
}
::v-deep(.q-btn.text-grey-8:before) {
	border: unset !important;
}
</style>
