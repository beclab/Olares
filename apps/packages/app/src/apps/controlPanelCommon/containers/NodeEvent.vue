<template>
	<div class="my-table-wrapper">
		<q-table
			:rows="rows"
			flat
			:columns="columns"
			row-key="name"
			:loading="loading"
			hide-pagination
			style="width: 100%"
		>
			<template v-slot:body-cell-status="props">
				<q-td :props="props">
					<q-badge
						rounded
						:color="props.row.status ? 'primary' : 'red'"
						class="q-mr-sm"
					/>
					<span>{{ props.row.status ? t('SUCCESSFUL') : t('FAILED') }}</span>
				</q-td>
			</template>
			<template v-slot:body-cell-age="props">
				<q-td :props="props">
					<div v-html="props.row.age"></div>
				</q-td>
			</template>
			<template v-slot:no-data>
				<Empty></Empty>
			</template>
		</q-table>
	</div>
</template>

<script setup lang="ts">
import { orderBy } from 'lodash';
import { getEvent, getNodeEvent } from '@apps/control-panel-common/src/network';
import { joinSelector } from '@apps/control-panel-common/src/utils';
import { MODULE_KIND_MAP } from '@apps/control-panel-common/src/utils/constants';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import { onBeforeUnmount, ref, watch } from 'vue';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
import { useRoute } from 'vue-router';
import Empty from '../components/Empty2.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
const route = useRoute();
const loading = ref(false);
const rows = ref([]);

const columns: any = [
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
];

interface Props {
	detail: any;
}
const props = withDefaults(defineProps<Props>(), {});

const fetchData = () => {
	const { uid, name, namespace, _originData = {} } = props.detail;

	const fields = {
		'involvedObject.name': name,
		'involvedObject.kind': _originData.kind
		// TODO: src/core/containers/Base/Detail/Events/index.jsx
		// 'involvedObject.kind': _originData.kind || MODULE_KIND_MAP[this.module],
	};
	const params = {
		fieldSelector: joinSelector(fields)
	};
	loading.value = true;
	getNodeEvent(params)
		.then((res) => {
			rows.value = orderBy(
				res.data.items.map(ObjectMapper.events),
				'lastTimestamp'
			);
			console.log('rows', rows.value);
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(
	() => route.params,
	() => {
		fetchData();
	}
);
fetchData();
</script>

<style lang="scss" scoped>
.my-table-wrapper {
	position: absolute;
	left: 0;
	right: 0;
}
</style>
