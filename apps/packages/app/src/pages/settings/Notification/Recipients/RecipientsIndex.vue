<template>
	<page-title-component :show-back="true" :title="t('recipients')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="q-pa-md column q-list-class">
			<q-table
				flat
				:rows="applicationStore.recipients"
				:columns="columns"
				row-key="appid"
				:pagination="initialPagination"
			>
				<template v-slot:top-right>
					<q-btn
						:label="t('create')"
						class="col-1"
						flat
						color="primary"
						@click="addRecipientCollection()"
					/>
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
									this.$router.push('/recipients/' + props.row.id);
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
import { useNotificationStore } from 'src/stores/settings/notification';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AddRecipientsCollection from './AddRecipientsCollection';
import { useQuasar } from 'quasar';
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
		name: 'type',
		align: 'left',
		label: i18n.global.t('type'),
		field: 'type',
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
		name: 'action',
		align: 'right',
		label: i18n.global.t('action'),
		sortable: false
	}
];

export default {
	components: { PageTitleComponent },
	setup() {
		const $q = useQuasar();
		const applicationStore = useNotificationStore();
		const { t } = useI18n();
		function addRecipientCollection() {
			$q.dialog({
				component: AddRecipientsCollection,
				persistent: true,
				componentProps: {}
			}).onOk(async () => {
				await applicationStore.refreshRecipients();
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
			applicationStore,
			addRecipientCollection,
			t
		};
	}
};
</script>
