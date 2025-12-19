<template>
	<page-title-component :show-back="true" :title="t('sender')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="q-pa-md column q-list-class">
			<q-table
				flat
				:rows="applicationStore.senders"
				:columns="columns"
				row-key="id"
				:pagination="initialPagination"
			>
				<template v-slot:top-right>
					<q-btn
						:label="t('new')"
						class="col-1"
						flat
						color="primary"
						@click="
							() => {
								this.$router.push('/sender/create');
							}
						"
					/>
				</template>

				<template v-slot:body-cell-action="props">
					<q-td :props="props"> </q-td>
				</template>
			</q-table>
		</div>
	</bt-scroll-area>
</template>

<script lang="ts">
import { useNotificationStore } from 'src/stores/settings/notification';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { i18n } from 'src/boot/i18n';
import { useI18n } from 'vue-i18n';
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
		name: 'app',
		align: 'left',
		label: i18n.global.t('base.app'),
		field: 'app',
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
	components: {
		PageTitleComponent
	},
	setup() {
		const $q = useQuasar();

		const applicationStore = useNotificationStore();
		const { t } = useI18n();
		const category = ref('All');
		const list = ref([]);
		let isNew = false;

		function updateList() {
			let result = applicationStore.topicList.find(
				(t) => t.category == category.value
			);
			if (!result) {
				isNew = true;
				list.value = [];
				return;
			}
			isNew = false;

			let l = JSON.parse(result.content);
			let r = [];
			for (const t of l) {
				let tt = applicationStore.topics.find((to) => to.id == t);
				r.push(tt);
			}

			list.value = r;
		}

		async function onSave() {
			$q.loading.show();
			try {
				let obj = {
					category: category.value,
					content: JSON.stringify(list.value.map((t) => t.id))
				};

				if (isNew) {
					await applicationStore.createTopicList(obj);
				} else {
					await applicationStore.updateTopicList(obj);
				}

				updateList();
			} catch (e) {
				notifyFailed(e.message || t('errors.something_wrong_please_try_again'));
			} finally {
				$q.loading.hide();
			}
		}

		async function onDelete(row) {
			let index = list.value.findIndex((t) => t.id == row.id);

			if (index < 0 || index > list.value.length - 1) {
				return;
			}
			list.value.splice(index, 1);
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
			onSave,
			onDelete,
			t
		};
	}
};
</script>
