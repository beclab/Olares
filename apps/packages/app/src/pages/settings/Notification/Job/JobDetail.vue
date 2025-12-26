<template>
	<div class="q-pa-md column" v-if="job">
		<div class="row q-py-sm justify-start">
			<q-btn
				:label="t('back')"
				class="col-1"
				flat
				color="primary"
				@click="
					() => {
						this.$router.push('/job');
					}
				"
			/>
		</div>
		<div class="column">
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('id') }}</div>
				<div class="q-pa-sm">{{ job.id }}</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('body') }}</div>
				<div class="q-pa-sm">{{ job }}</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('language') }}</div>
				<div class="q-pa-sm">
					{{ job.language }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('message_num') }}</div>
				<div class="q-pa-sm">
					{{ job.messageNum }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('sent_num') }}</div>
				<div class="q-pa-sm">
					{{ job.sentNum }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('success_num') }}</div>
				<div class="q-pa-sm">
					{{ job.successNum }}
				</div>
			</div>

			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('created_at') }}</div>
				<div class="q-pa-sm">
					{{ job.createdAt }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('updated_at') }}</div>
				<div class="q-pa-sm">
					{{ job.updatedAt }}
				</div>
			</div>
		</div>
		<q-table
			flat
			bordered
			:title="t('message')"
			:rows="messages"
			:columns="columns"
			row-key="id"
			:pagination="initialPagination"
		>
			<template v-slot:body-cell-edit="props">
				<q-td :props="props"></q-td>
			</template>
		</q-table>
	</div>
</template>

<script lang="ts">
import { ref, onMounted } from 'vue';
import { useNotificationStore } from 'src/stores/settings/notification';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { i18n } from 'src/boot/i18n';

const columns = [
	{
		name: 'id',
		align: 'left',
		label: i18n.global.t('id'),
		field: 'id',
		sortable: false
	},
	{
		name: 'message',
		align: 'left',
		label: i18n.global.t('message'),
		field: 'message',
		sortable: false
	},
	{
		name: 'recipientType',
		align: 'left',
		label: i18n.global.t('recipient_type'),
		field: 'recipientType',
		sortable: false
	},
	{
		name: 'recipient',
		align: 'left',
		label: i18n.global.t('recipient'),
		field: 'recipient',
		sortable: false,
		style: 'width: 150px'
	},
	{
		name: 'senderType',
		align: 'left',
		label: i18n.global.t('sender_type'),
		field: 'senderType',
		sortable: false
	},
	{
		name: 'sender',
		align: 'left',
		label: i18n.global.t('sender'),
		field: 'sender',
		sortable: false,
		style: 'width: 200px'
	},
	{
		name: 'status',
		align: 'left',
		label: i18n.global.t('status'),
		field: 'status',
		sortable: false
	},
	{
		name: 'createdAt',
		align: 'left',
		label: i18n.global.t('created_at'),
		field: 'createdAt',
		sortable: false
	},
	{
		name: 'updatedAt',
		align: 'left',
		label: i18n.global.t('updated_at'),
		field: 'updatedAt',
		sortable: false
	},
	{
		name: 'action',
		align: 'right',
		label: i18n.global.t('action'),
		sortable: false
	}
];

export default {
	setup() {
		const applicationStore = useNotificationStore();
		const Route = useRoute();
		const { t } = useI18n();
		let id = ref();

		let job = ref();
		let messages = ref([]);

		onMounted(async () => {
			id.value = Route.params.id;
			job.value = applicationStore.jobs.find((t) => t.id == id.value);
			messages.value = await applicationStore.getMessageByJob(id.value);
		});

		return {
			initialPagination: {
				sortBy: 'desc',
				descending: false,
				page: 1,
				rowsPerPage: 1000
			},
			columns,
			applicationStore,
			messages,
			job,
			t
		};
	}
};
</script>
