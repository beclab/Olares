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
			style="width: 100%"
		>
			<template v-slot:body-cell-type="props">
				<q-td :props="props">
					<div class="row items-center no-wrap text-ink-1">
						<MyBadge class="q-mr-xs" :type="props.row.type" />
						<span class="event-type">{{ props.row.type }}</span>
					</div>
				</q-td>
			</template>
			<template v-slot:body-cell-age="props">
				<q-td :props="props">
					<div v-html="props.row.age"></div>
				</q-td>
			</template>
			<template v-slot:body-cell-message="props">
				<q-td :props="props">
					<div class="my-table-td-item">
						{{ props.row.message }}
					</div>
				</q-td>
			</template>
			<template v-slot:no-data>
				<div style="flex: 1" class="row justify-center q-mt-lg">
					<Empty v-show="!loading"></Empty>
				</div>
			</template>
		</q-table>
	</QTableStyle>
</template>

<script setup lang="ts">
import { orderBy } from 'lodash';
import { getEvent } from '@apps/control-panel-common/src/network';
import { joinSelector } from '@apps/control-panel-common/src/utils';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import { ref, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import Empty from '../components/Empty.vue';
import QTableStyle from '../components/QTableStyle2.vue';
import MyBadge from '../components/MyBadge.vue';
import { MODULE_KIND_MAP } from '@apps/control-panel-common/src/utils/constants';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const route = useRoute();
const loading = ref(false);
const rows = ref([]);
const pagination = ref({
	page: 1,
	rowsPerPage: 20,
	rowsNumber: 0
});

const columns: any = computed(() => [
	{
		label: t('TYPE'),
		align: 'left',
		name: 'type',
		field: 'type',
		width: '10%'
	},
	{
		label: t('REASON'),
		align: 'left',
		name: 'reason',
		field: 'reason',
		width: '16%'
	},
	{
		label: t('EVENT_AGE'),
		align: 'left',
		name: 'age',
		field: 'age',
		width: '16%'
	},
	{
		label: t('SOURCE'),
		align: 'left',
		name: 'from',
		field: 'from',
		width: '18%'
	},
	{
		label: t('MESSAGE'),
		align: 'left',
		name: 'message',
		field: 'message'
	}
]);

interface Props {
	detail: any;
}

const props = withDefaults(defineProps<Props>(), {});

const fetchData = () => {
	if (!props.detail) {
		return;
	}
	const { uid, name, namespace, _originData = {} } = props.detail;

	const fields = {
		'involvedObject.name': name,
		'involvedObject.namespace': namespace,
		'involvedObject.kind': _originData.kind,
		'involvedObject.uid': uid
	};
	const params = {
		namespace: namespace || _originData.namespace,
		fieldSelector: joinSelector(fields)
	};
	loading.value = true;
	getEvent(params)
		.then((res) => {
			rows.value = orderBy(
				res.data.items.map(ObjectMapper.events),
				'lastTimestamp'
			);
			pagination.value.rowsPerPage = rows.value.length;
		})
		.finally(() => {
			loading.value = false;
		});
};

watch([() => route.params, () => props.detail], () => {
	fetchData();
});
fetchData();

defineExpose({ fetchData });
</script>

<style lang="scss" scoped>
.my-table-wrapper {
	position: absolute;
	left: 0;
	right: 0;
}
.my-table-td-item {
	word-break: break-all;
	white-space: pre-wrap;
	text-align: left;
	width: max-content;
}
.event-type {
	font-weight: 500;
	line-height: 16px;
}
</style>
