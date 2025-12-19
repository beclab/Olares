<template>
	<QTableStyle>
		<q-table
			:rows="ports"
			flat
			:columns="columns"
			row-key="name"
			hide-pagination
			v-ripple="false"
			v-model:pagination="pagination"
			table-class="port-table-wrapper"
		>
			<template v-slot:no-data>
				<div class="row justify-center full-width q-mt-lg">
					<Empty></Empty>
				</div>
			</template>
		</q-table>
	</QTableStyle>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import Empty from '../components/Empty.vue';
import QTableStyle from '../components/QTableStyle2.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	ports: any[];
}

withDefaults(defineProps<Props>(), {});
const pagination = ref({
	rowsNumber: 0
});
const columns: any = computed(() => [
	{
		label: t('NAME'),
		name: 'name',
		field: 'name',
		align: 'left'
	},
	{
		label: t('PROTOCOL'),
		name: 'protocol',
		field: 'protocol',
		align: 'left'
	},
	{
		label: t('PORT'),
		name: 'containerPort',
		field: 'containerPort',
		align: 'left'
	}
]);
</script>

<style scoped lang="scss">
::v-deep(.port-table-wrapper table) {
	table-layout: fixed;
}
</style>
