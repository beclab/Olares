<template>
	<div class="q-pa-md column">
		<div class="row q-py-sm justify-start">
			<q-btn
				:label="t('back')"
				class="col-1"
				flat
				color="primary"
				@click="
					() => {
						this.$router.push('/sender');
					}
				"
			/>
		</div>
		<q-table
			flat
			bordered
			title="Sender"
			:rows="senderTemplates"
			:columns="columns"
			row-key="id"
			:pagination="initialPagination"
		>
			<template v-slot:body-cell-action="props">
				<q-td :props="props">
					<q-btn
						:label="t('add')"
						class="col-1"
						flat
						color="primary"
						@click="addSender(props.row)"
					/>
				</q-td>
			</template>
		</q-table>
	</div>
</template>

<script lang="ts">
import { useNotificationStore } from 'src/stores/settings/notification';
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import {
	senderTemplates,
	TerminusNotificationSenderTemplate
} from 'src/utils/settings/notification';
import CreateSender from './CreateSender.vue';
import { i18n } from 'src/boot/i18n';
import { useI18n } from 'vue-i18n';

const columns = [
	{
		name: 'name',
		align: 'left',
		label: i18n.global.t('name'),
		field: 'name',
		sortable: false
	},
	{
		name: 'type',
		align: 'left',
		label: i18n.global.t('type'),
		field: 'type',
		sortable: false
	},
	{
		name: 'description',
		align: 'left',
		label: i18n.global.t('description'),
		field: 'description',
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
		const $q = useQuasar();
		const router = useRouter();
		const applicationStore = useNotificationStore();

		const category = ref('All');
		const list = ref([]);
		const { t } = useI18n();

		function addSender(template: TerminusNotificationSenderTemplate) {
			$q.dialog({
				component: CreateSender,
				persistent: true,
				componentProps: {
					sender: template.name
				}
			}).onOk(async () => {
				await applicationStore.refreshSenders();
				router.push('/sender');
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
			senderTemplates,
			addSender,
			t
		};
	}
};
</script>
