<template>
	<q-card flat class="items-center q-px-lg relative-position">
		<div>
			<div class="q-py-lg" v-if="!isLoading && workloads.length === 0">
				<Empty> </Empty>
			</div>
			<Item
				v-for="workload in workloads"
				:key="`${workload.module}-${workload.name}`"
				prefix="prefix"
				:detail="workload"
			>
			</Item>
			<q-inner-loading :showing="isLoading"> </q-inner-loading>
		</div>
	</q-card>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { getWorkloads } from '@apps/control-panel-common/src/network';
import Item from './WorkloadItem.vue';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import Empty from '../components/Empty.vue';

interface Props {
	selector: string;
}

const props = withDefaults(defineProps<Props>(), {});
const emits = defineEmits<{
	(e: 'update', data: any[]): void;
}>();

const route = useRoute();
const workloads = ref();
const isLoading = ref(true);

const mapper = (module: string) => {
	// eslint-disable-next-line @typescript-eslint/ban-ts-comment
	// @ts-ignore
	return ObjectMapper[module] || ((data: any) => data);
};
async function fetchList(
	namespace: string,
	module: string,
	labelSelector: string
) {
	const result = await getWorkloads({
		namespace,
		type: module,
		labelSelector
	});

	const data = Array.isArray(result.data.items)
		? result.data.items.map((item) => ({
				// cluster: 'default',
				module: module,
				...mapper(module)(item)
		  }))
		: [];

	return data;
}

const fetchData = async (selector: string) => {
	isLoading.value = true;
	workloads.value = [];
	const { namespace, kind }: Record<string, any> = route.params;
	const [deployments, statefulsets, daemonsets] = await Promise.all([
		fetchList(namespace, 'deployments', selector),
		fetchList(namespace, 'statefulsets', selector),
		fetchList(namespace, 'daemonsets', selector)
	]);
	workloads.value = [...deployments, ...statefulsets, ...daemonsets];
	emits('update', workloads.value);
	isLoading.value = false;
};
watch(
	() => props.selector,
	(newData, oldData) => {
		if (newData && newData !== oldData) {
			fetchData(newData);
		}
	}
);

watch(
	() => route.params.pods_uid,
	() => {
		workloads.value = [];
	}
);
</script>

<style lang="scss" scoped>
.items-center {
	background: $background-1;
}
.empty {
	text-align: center;
}
</style>
