<template>
	<div class="q-pa-md column" v-if="policy">
		<div class="row q-py-sm justify-start">
			<q-btn
				:label="t('back')"
				class="col-1"
				flat
				color="primary"
				@click="
					() => {
						this.$router.push('/notify');
					}
				"
			/>
		</div>
		<div class="column">
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('name') }}</div>
				<div class="q-pa-sm">{{ policy.name }}</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('is_default') }}</div>
				<div class="q-pa-sm">
					{{ policy.isDefault }}
				</div>
			</div>
		</div>
		<q-table
			flat
			bordered
			:title="t('rules')"
			:rows="rules"
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
					@click="addNotifyRule()"
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

			<template v-slot:body-cell-sender="props">
				<q-td :props="props">
					{{
						applicationStore.senders.find(
							(sender) => sender.id == props.row.sender
						).name
					}}
				</q-td>
			</template>
			<template v-slot:body-cell-recipients="props">
				<q-td :props="props">
					{{
						applicationStore.recipients.find(
							(recipients) => recipients.id == props.row.recipients
						).name
					}}
				</q-td>
			</template>
			<template v-slot:body-cell-action="props">
				<q-td :props="props">
					<q-btn
						:label="t('delete')"
						class="col-1"
						flat
						color="primary"
						@click="onDelete(props.row.id)"
					/>
				</q-td>
			</template>
		</q-table>
	</div>
</template>

<script lang="ts">
import { ref, onMounted } from 'vue';
import { useNotificationStore } from 'src/stores/settings/notification';
import { Loading, useQuasar } from 'quasar';
import { useRoute } from 'vue-router';
import AddRule from './AddRule';
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
		name: 'sender',
		align: 'left',
		label: i18n.global.t('sender'),
		field: 'sender',
		sortable: false
	},
	{
		name: 'recipients',
		align: 'left',
		label: i18n.global.t('recipients'),
		field: 'recipients',
		sortable: false
	},
	{
		name: 'status',
		align: 'left',
		label: i18n.global.t('status'),
		field: 'status',
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
		const $q = useQuasar();
		const applicationStore = useNotificationStore();
		const Route = useRoute();
		let id = ref();

		let policy = ref();
		let rules = ref([]);
		const { t } = useI18n();

		onMounted(async () => {
			id.value = Route.params.id;
			policy.value = applicationStore.notifyPolicy.find(
				(t) => t.id == id.value
			);
			rules.value = await applicationStore.getNotifyRuleByPolicyId(id.value);
		});

		function addNotifyRule() {
			$q.dialog({
				component: AddRule,
				persistent: true,
				componentProps: {
					id: id.value
				}
			}).onOk(async () => {
				rules.value = await applicationStore.getNotifyRuleByPolicyId(id.value);
			});
		}

		async function onDelete(ruleId) {
			Loading.show();
			try {
				await applicationStore.deleteNotifyRule(ruleId);
				rules.value = await applicationStore.getNotifyRuleByPolicyId(id.value);
			} finally {
				Loading.hide();
			}
		}

		return {
			initialPagination: {
				sortBy: 'desc',
				descending: false,
				page: 1,
				rowsPerPage: 1000
			},
			columns,
			applicationStore,
			policy,
			rules,
			addNotifyRule,
			onDelete,
			t
		};
	}
};
</script>
