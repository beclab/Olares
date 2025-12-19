<template>
	<div class="q-pa-md column" v-if="recipients">
		<div class="row q-py-sm justify-start">
			<q-btn
				:label="t('back')"
				class="col-1"
				flat
				color="primary"
				@click="
					() => {
						this.$router.push('/recipients');
					}
				"
			/>
		</div>
		<div class="column">
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('name') }}</div>
				<div class="q-pa-sm">{{ recipients.name }}</div>
			</div>

			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('type') }}</div>
				<div class="q-pa-sm">{{ recipients.type }}</div>
			</div>

			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('created_at') }}</div>
				<div class="q-pa-sm">
					{{ recipients.createdAt }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('updated_at') }}</div>
				<div class="q-pa-sm">
					{{ recipients.updatedAt }}
				</div>
			</div>
		</div>
		<q-table
			flat
			bordered
			:title="t('recipients_address')"
			:rows="recipientAddress"
			:columns="columns"
			row-key="id"
			:pagination="initialPagination"
		>
			<template v-slot:top-right>
				<q-btn
					:label="t('add')"
					class="col-1"
					flat
					color="primary"
					@click="addRecipientsAddress()"
				/>
			</template>

			<template v-slot:body-cell-data="props">
				<q-td :props="props">
					{{ JSON.stringify(props.row.data) }}
				</q-td>
			</template>

			<template v-slot:body-cell-edit="props">
				<q-td :props="props">
					<q-btn
						:label="t('edit')"
						class="col-1"
						flat
						color="primary"
						@click="
							() => {
								this.$router.push('/topic/' + props.row.id);
							}
						"
					/>
				</q-td>
			</template>
		</q-table>
	</div>
</template>

<script lang="ts">
import { useNotificationStore } from 'src/stores/settings/notification';
import CreateRecipients from './CreateRecipient';
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
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
		name: 'name',
		align: 'left',
		label: i18n.global.t('name'),
		field: 'name',
		sortable: false
	},
	{
		name: 'data',
		align: 'left',
		label: i18n.global.t('data'),
		field: 'data',
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
		const applicationStore = useNotificationStore();
		const Route = useRoute();
		let id = ref();
		const { t } = useI18n();

		let recipients = ref();
		let recipientAddress = ref([]);

		onMounted(() => {
			id.value = Route.params.id;
			recipients.value = applicationStore.recipients.find(
				(t) => t.id == id.value
			);
			recipientAddress.value = applicationStore.recipientAddress.filter(
				(t) => t.recipientsId == id.value
			);
		});

		function addRecipientsAddress() {
			$q.dialog({
				component: CreateRecipients,
				persistent: true,
				componentProps: {
					id: id.value
				}
			}).onOk(async () => {
				await applicationStore.refreshRecipientAddress();
				recipientAddress.value = applicationStore.recipientAddress.filter(
					(t) => t.recipientsId == id.value
				);
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
			addRecipientsAddress,
			applicationStore,
			recipientAddress,
			recipients,
			t
		};
	}
};
</script>
