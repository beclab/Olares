<template>
	<div class="workflow-list-root">
		<q-table
			:pagination="initialPagination"
			:rows="rows"
			flat
			class="bg-background-1"
			:loading="argoStore.workflowsLoading"
			:columns="columns"
			row-key="name"
			@row-click="onClick"
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

			<template v-slot:body-cell-phase="props">
				<q-td :props="props" class="left-align text-body2 text-ink-2">
					{{ props.row.phase }}
				</q-td>
			</template>

			<template v-slot:body-cell-startedAt="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{
						props.row.startedAt
							? getPastTime(new Date(), new Date(props.row.startedAt))
							: '-'
					}}
				</q-td>
			</template>

			<template v-slot:body-cell-finishedAt="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{
						props.row.finishedAt
							? getPastTime(new Date(), new Date(props.row.finishedAt))
							: '-'
					}}
				</q-td>
			</template>

			<template v-slot:body-cell-progress="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{ props.row.progress }}
				</q-td>
			</template>

			<template v-slot:body-cell-message="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{ props.row.message }}
				</q-td>
			</template>
			<template v-slot:no-data>
				<empty-view :is-table="true" />
			</template>
		</q-table>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { getPastTime } from '../../../../utils/rss-utils';
import { useArgoStore } from '../../../../stores/argo';
import { useI18n } from 'vue-i18n';
import EmptyView from '../../../../components/rss/EmptyView.vue';

const i18n = useI18n();

const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: i18n.t('base.name'),
		field: 'name'
	},
	{
		name: 'phase',
		align: 'left',
		label: i18n.t('base.phase'),
		field: 'phase'
	},
	{
		name: 'startedAt',
		align: 'right',
		label: i18n.t('base.started_at'),
		field: 'startedAt'
	},
	{
		name: 'finishedAt',
		align: 'right',
		label: i18n.t('base.finished_at'),
		field: 'finishedAt'
	},
	{
		name: 'progress',
		align: 'right',
		label: i18n.t('base.progress'),
		field: 'progress'
	},
	{
		name: 'message',
		align: 'right',
		label: i18n.t('base.message'),
		field: 'message'
	}
];

const rows = ref<any[]>([]);
const argoStore = useArgoStore();
const initialPagination = ref({
	page: 0,
	rowsPerPage: 100
});

const onClick = (eve: any, row: any, index: number) => {
	if (argoStore.workflows[index]) {
		argoStore.workflow_id = argoStore.workflows[index].metadata.name;
	}
};

watch(
	() => argoStore.workflows,
	(newValue) => {
		if (newValue) {
			rows.value = argoStore.workflows.map((workflow) => ({
				phase: workflow.status.phase,
				name: workflow.metadata.name,
				startedAt: workflow.status.startedAt,
				finishedAt: workflow.status.finishedAt,
				progress: workflow.status.progress,
				message: workflow.status.message
			}));

			console.log('list=>', rows);
		} else {
			rows.value = [];
		}
	},
	{
		immediate: true
	}
);
</script>

<style scoped lang="scss">
.workflow-list-root {
	width: 100%;
	height: 100%;
	padding-left: 44px;
	padding-right: 44px;
}
::v-deep(.q-btn.text-grey-8:before) {
	border: unset !important;
}
</style>
