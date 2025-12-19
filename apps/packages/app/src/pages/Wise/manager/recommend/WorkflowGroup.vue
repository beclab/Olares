<template>
	<div class="workflow-group-root">
		<q-table
			:pagination="initialPagination"
			:rows="rows"
			flat
			class="bg-background-1"
			:loading="argoStore.cronWorkflowsLoading"
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

			<template v-slot:body-cell-schedule="props">
				<q-td :props="props" class="left-align text-body2 text-ink-2">
					{{ props.row.schedule }}
				</q-td>
			</template>

			<template v-slot:body-cell-created="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{ props.row.created }}
				</q-td>
			</template>

			<template v-slot:body-cell-nextRun="props">
				<q-td
					:props="props"
					class="left-align text-body2 text-ink-2"
					style="text-align: right"
				>
					{{ props.row.nextRun }}
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
import { useArgoStore } from '../../../../stores/argo';
import cronstrue from 'cronstrue';
import 'cronstrue/locales/zh_CN';
import 'cronstrue/locales/es';
import cronParser from 'cron-parser';
import { useI18n } from 'vue-i18n';
import EmptyView from 'src/components/rss/EmptyView.vue';
import { i18n } from '../../../../boot/i18n';
const { t } = useI18n();
const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: t('base.name'),
		field: 'name'
	},
	{
		name: 'nameSpace',
		align: 'left',
		label: t('base.name_space'),
		field: 'nameSpace'
	},
	{
		name: 'schedule',
		align: 'left',
		label: t('base.schedule'),
		field: 'schedule'
	},
	{
		name: 'created',
		align: 'right',
		label: t('base.created'),
		field: 'created'
	},
	{
		name: 'nextRun',
		align: 'right',
		label: t('base.next_run'),
		field: 'nextRun'
	}
];

const rows = ref<any[]>([]);
const argoStore = useArgoStore();

const initialPagination = ref({
	page: 0,
	rowsPerPage: 100
});

const onClick = (eve: any, row: any, index: number) => {
	if (argoStore.cron_workflows[index]) {
		argoStore.cronLabel = argoStore.cron_workflows[index].metadata.name;
		argoStore.namespace = argoStore.cron_workflows[index].metadata.namespace;
		argoStore.get_workflows(argoStore.namespace, argoStore.cronLabel);
	}
};

watch(
	() => argoStore.cron_workflows,
	(newValue) => {
		if (newValue) {
			const getSchedule = (schedule) => {
				try {
					let language: string;
					if ((i18n.global.locale.value as string).startsWith('zh')) {
						language = 'zh_CN';
					} else {
						language = (i18n.global.locale.value as string).substring(0, 2);
					}
					return cronstrue.toString(schedule, {
						locale: language
					});
				} catch (e) {
					return schedule;
				}
			};

			const getNextRun = (schedule, lastSchedule) => {
				try {
					const now = new Date();
					const options = {
						currentDate: lastSchedule ? new Date(lastSchedule) : now,
						iterator: true
					};
					const interval = cronParser.parseExpression(schedule, options);
					const nextExecutionTime = interval.next().value.toDate();
					let diffInMs = nextExecutionTime - now;
					if (diffInMs < 0) {
						const nextNextExecutionTime = interval.next().value.toDate();
						diffInMs = nextNextExecutionTime - now;
					}
					const diffInSeconds = Math.floor(diffInMs / 1000);
					const hours = Math.floor(diffInSeconds / 3600);
					const remainingSeconds = diffInSeconds % 3600;
					const minutes = Math.floor(remainingSeconds / 60);
					const seconds = remainingSeconds % 60;

					let timeDifferenceString = '';
					if (hours > 0) {
						timeDifferenceString += `${hours}h`;
					}
					if (minutes > 0) {
						timeDifferenceString += `${minutes}m`;
					}
					if (seconds > 0) {
						timeDifferenceString += `${seconds}s`;
					}
					if (hours === 0 && minutes === 0 && seconds === 0) {
						timeDifferenceString = 'in less than 1s';
					}
					// console.log('time===>>>');
					// console.log(diffInMs);
					// console.log(diffInSeconds);
					// console.log(timeDifferenceString);
					return timeDifferenceString;
				} catch (e) {
					console.log(e);
					return '-';
				}
			};

			rows.value = newValue.map((cron_workflow) => ({
				name: cron_workflow.metadata.name,
				nameSpace: cron_workflow.metadata.namespace,
				schedule: getSchedule(cron_workflow.spec.schedule),
				created: calculateTimeDifference(
					cron_workflow.metadata.creationTimestamp
				),
				nextRun: getNextRun(
					cron_workflow.spec.schedule,
					cron_workflow.status.lastScheduledTime
				)
			}));
		} else {
			rows.value = [];
		}
	}
);

function calculateTimeDifference(timestamp: string): string {
	const inputTime = new Date(timestamp);
	const now = new Date();

	const isSameDay =
		inputTime.getFullYear() === now.getFullYear() &&
		inputTime.getMonth() === now.getMonth() &&
		inputTime.getDate() === now.getDate();

	if (isSameDay) {
		const hours = inputTime.getHours().toString().padStart(2, '0');
		const minutes = inputTime.getMinutes().toString().padStart(2, '0');
		return `${hours}:${minutes}`;
	} else {
		const year = inputTime.getFullYear();
		const month = (inputTime.getMonth() + 1).toString().padStart(2, '0');
		const day = inputTime.getDate().toString().padStart(2, '0');
		return `${year}-${month}-${day}`;
	}
}
</script>

<style scoped lang="scss">
.workflow-group-root {
	width: 100%;
	height: 100%;
	padding-left: 44px;
	padding-right: 44px;
}
::v-deep(.q-btn.text-grey-8:before) {
	border: unset !important;
}
</style>
