<template>
	<QTableStyle>
		<q-table
			:rows="rows"
			flat
			:columns="columns"
			row-key="name"
			:loading="loading"
			hide-pagination
			v-model:pagination="pagination"
		>
			<template v-slot:body-cell-status="props">
				<q-td :props="props">
					<div class="row items-center no-wrap">
						<MyBadge class="q-mr-xs" :type="props.row.type" />
						<span class="event-type">{{ props.row.type }}</span>
					</div>
				</q-td>
			</template>
			<template v-slot:no-data>
				<div class="row justify-center full-width q-mt-lg">
					<Empty></Empty>
				</div>
			</template>
		</q-table>
	</QTableStyle>
</template>

<script setup lang="ts">
import { orderBy } from 'lodash';
import { getEvent } from '@apps/control-hub/src/network';
import { joinSelector } from '@apps/control-hub/src/utils';
import { MODULE_KIND_MAP } from '@apps/control-hub/src/utils/constants';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { onMounted, ref, watch, computed } from 'vue';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import MyBadge from '@apps/control-panel-common/src/components/MyBadge.vue';
import QTableStyle from '@apps/control-panel-common/src/components/QTableStyle.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const loading = ref(false);
const rows = ref([]);
const pagination = ref({
	rowsNumber: 0
});
const columns: any = computed(() => [
	{
		label: t('TYPE'),
		align: 'center',
		name: 'type',
		field: 'type',
		width: '10%'
	},
	{
		label: t('REASON'),
		align: 'center',
		name: 'reason',
		field: 'reason',
		width: '16%'
	},
	{
		label: t('EVENT_AGE'),
		align: 'center',
		name: 'age',
		field: 'age',
		width: '16%'
	},
	{
		label: t('SOURCE'),
		align: 'center',
		name: 'from',
		field: 'from',
		width: '18%'
	},
	{
		label: t('MESSAGE'),
		align: 'center',
		name: 'message',
		field: 'message'
	}
]);

interface Props {
	detail: any;
}
const props = withDefaults(defineProps<Props>(), {});

const fetchData = () => {
	const { uid, name, namespace, _originData = {} } = props.detail;

	const fields = {
		'involvedObject.name': name,
		'involvedObject.namespace': namespace,
		'involvedObject.kind': _originData.kind,
		// TODO: src/core/containers/Base/Detail/Events/index.jsx
		// 'involvedObject.kind': _originData.kind || MODULE_KIND_MAP[this.module],
		'involvedObject.uid': uid
	};
	const params = {
		namespace,
		fieldSelector: joinSelector(fields)
	};
	loading.value = true;
	getEvent(params)
		.then((res) => {
			rows.value = orderBy(
				res.data.items.map(ObjectMapper.events),
				'lastTimestamp'
			);
		})
		.finally(() => {
			loading.value = false;
		});
};
watch(
	() => props.detail,
	() => {
		rows.value = [];
		fetchData();
	}
);
onMounted(() => {
	fetchData();
});
</script>

<style></style>
