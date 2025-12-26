<template>
	<div class="q-pa-md column" v-if="template">
		<div class="row q-py-sm justify-start">
			<q-btn
				:label="t('back')"
				class="col-1"
				flat
				color="primary"
				@click="
					() => {
						this.$router.push('/template');
					}
				"
			/>
		</div>
		<div class="column">
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('body') }}</div>
				<div class="q-pa-sm">{{ template }}</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('name') }}</div>
				<div class="q-pa-sm">{{ template.name }}</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('base.app') }}</div>
				<div class="q-pa-sm">
					{{ template.appId }} / {{ template.appName }} /
					{{ template.appTemplateName }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('variables') }}</div>
				<div class="q-pa-sm">
					{{ template.variables.join() }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('created_at') }}</div>
				<div class="q-pa-sm">
					{{ template.createdAt }}
				</div>
			</div>
			<div class="row justify-start">
				<div class="q-pa-sm text-bold">{{ t('updated_at') }}</div>
				<div class="q-pa-sm">
					{{ template.updatedAt }}
				</div>
			</div>
		</div>
		<q-table
			flat
			bordered
			title="Template Content"
			:rows="templateContent"
			:columns="columns"
			row-key="id"
			:pagination="initialPagination"
		>
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
import { ref, onMounted } from 'vue';
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
		name: 'language',
		align: 'left',
		label: i18n.global.t('language'),
		field: 'language',
		sortable: false
	},
	{
		name: 'title',
		align: 'left',
		label: i18n.global.t('title'),
		field: 'title',
		sortable: false
	},
	{
		name: 'body',
		align: 'left',
		label: i18n.global.t('body'),
		field: 'body',
		sortable: false
	}
];

export default {
	setup() {
		const applicationStore = useNotificationStore();
		const Route = useRoute();
		let id = ref();

		let template = ref();
		let templateContent = ref([]);
		const { t } = useI18n();

		onMounted(() => {
			id.value = Route.params.id;
			template.value = applicationStore.templates.find((t) => t.id == id.value);
			templateContent.value = applicationStore.templateContent.filter(
				(t) => t.templateId == id.value
			);
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
			templateContent,
			template,
			t
		};
	}
};
</script>
