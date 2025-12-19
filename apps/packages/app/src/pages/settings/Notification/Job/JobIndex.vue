<template>
	<page-title-component :show-back="true" :title="t('jobs')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="q-pa-md column q-list-class">
			<q-table
				flat
				:rows="applicationStore.jobs"
				:columns="columns"
				row-key="id"
				:pagination="initialPagination"
			>
				<template v-slot:top-right>
					<q-btn
						:label="t('create')"
						class="col-1"
						flat
						color="primary"
						@click="addJob()"
					/>
				</template>

				<template v-slot:body-cell-notifyPolicyId="props">
					<q-td :props="props">
						{{
							applicationStore.notifyPolicy.find(
								(notifyPolicy) => notifyPolicy.id == props.row.notifyPolicyId
							).name
						}}
					</q-td>
				</template>

				<template v-slot:body-cell-templateId="props">
					<q-td :props="props">
						{{
							applicationStore.templates.find(
								(template) => template.id == props.row.templateId
							).name +
							'(' +
							applicationStore.templates.find(
								(template) => template.id == props.row.templateId
							).appName +
							')'
						}}
					</q-td>
				</template>

				<template v-slot:body-cell-message="props">
					<q-td :props="props">
						{{ props.row.messageNum }} / {{ props.row.sentNum }}/
						{{ props.row.successNum }}
					</q-td>
				</template>

				<template v-slot:body-cell-action="props">
					<q-td :props="props">
						<q-btn
							:label="t('view')"
							class="col-1"
							flat
							color="primary"
							@click="
								() => {
									this.$router.push('/job/' + props.row.id);
								}
							"
						/>
					</q-td>
				</template>
			</q-table>
		</div>
	</bt-scroll-area>
</template>

<script lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useNotificationStore } from 'src/stores/settings/notification';
import CreateJob from './CreateJob.vue';
import { useI18n } from 'vue-i18n';
import { defineComponent } from 'vue';
import { i18n } from 'src/boot/i18n';
import { useQuasar } from 'quasar';
import { ref } from 'vue';

const columns = [
	{
		name: 'id',
		align: 'left',
		label: i18n.global.t('id'),
		field: 'id',
		sortable: false
	},
	{
		name: 'templateId',
		align: 'left',
		label: i18n.global.t('template_id'),
		field: 'templateId',
		sortable: false
	},
	{
		name: 'notifyPolicyId',
		align: 'left',
		label: i18n.global.t('notify_policy_id'),
		field: 'notifyPolicyId',
		sortable: false
	},

	{
		name: 'message',
		align: 'left',
		label: i18n.global.t('message'),

		sortable: false
	},
	{
		name: 'action',
		align: 'right',
		label: i18n.global.t('action'),
		sortable: false
	}
];

export default defineComponent({
	name: 'JobIndex',
	components: {
		PageTitleComponent
	},
	setup() {
		const $q = useQuasar();
		const { t } = useI18n();
		const applicationStore = useNotificationStore();

		const category = ref('All');
		const list = ref([]);

		function addJob() {
			$q.dialog({
				component: CreateJob,
				persistent: true,
				componentProps: {}
			}).onOk(async () => {
				await applicationStore.refreshJobs();
			});
		}

		return {
			initialPagination: {
				sortBy: 'desc',
				descending: false,
				page: 1,
				rowsPerPage: 1000
			},
			columns,
			category,
			list,
			applicationStore,
			addJob,
			t
		};
	}
});
</script>
