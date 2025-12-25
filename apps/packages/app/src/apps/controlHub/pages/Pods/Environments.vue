<template>
	<MyCard no-content-gap square flat :title="t('ENVIRONMENT_VARIABLE_PL')">
		<div class="q-gutter-y-md">
			<div
				class="environments-wrapper q-px-lg q-py-md"
				v-for="item in envlist"
				:key="item.name"
			>
				<MyExpansion
					:label="labelFormat(item)"
					:default-opened="!loading && !item.variables.length"
				>
					<MyChipList :data="variablesFilter(item.variables)"> </MyChipList>
					<div v-if="!item.variables.length">
						<Empty></Empty>
					</div>
				</MyExpansion>
			</div>
		</div>
		<Empty v-if="!loading && !envlist?.length"></Empty>
		<q-inner-loading :showing="loading"></q-inner-loading>
	</MyCard>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { onMounted, ref, watch, computed } from 'vue';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import { get, isEmpty, isNil } from 'lodash';
import { t } from '@apps/control-hub/src/boot/i18n';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import MyChipList from '@apps/control-panel-common/src/containers/MyChipList.vue';
import {
	getDeploymentsDetail,
	getPodDetail
} from '@apps/control-panel-common/src/network';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { fetcEnvList } from '@apps/control-panel-common/src/containers/env';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import MyExpansion from '@apps/control-panel-common/src/components/MyExpansion.vue';

interface Props {
	module?: string;
}

let loading = ref(false);
const route = useRoute();
const envDetail = ref();
const envlist = ref();
const variables = ref();

const props = withDefaults(defineProps<Props>(), {});

const labelFormat = (item: any) => {
	let label = '';
	label =
		item.type === 'init'
			? t('INIT_CONTAINER_VALUE', { value: item.name })
			: t('CONTAINER_VALUE', { value: item.name });
	return label;
};

const containers = computed(() => {
	const data = envDetail.value || {};
	const { spec, containers = [] } = data;

	if (!isEmpty(containers)) return containers;
	if (!isEmpty(spec)) return get(spec, 'template.spec.containers', []);

	return [];
});

const initContainers = computed(() => {
	const data = envDetail.value || {};
	const { spec, initContainers: initContainersData = [] } = data;

	if (!isEmpty(initContainersData)) return initContainersData;
	if (!isEmpty(spec)) return get(spec, 'template.spec.initContainers', []);

	return [];
});

const fetchData = async () => {
	const { namespace, cluster }: { [key: string]: any } = route.params;
	try {
		envlist.value = await fetcEnvList({
			namespace: namespace,
			cluster: cluster,
			containers: containers.value,
			initContainers: initContainers.value
		});
	} catch (error) {
		envlist.value = [];
	}
};

const variablesFilter = (data) => data.filter((item) => !!item.value);

const fetchEnv = () => {
	const { namespace, name, container }: { [key: string]: any } = route.params;
	loading.value = true;
	getPodDetail({ namespace, podName: name })
		.then((res) => {
			envDetail.value = ObjectMapper.pods(res.data);

			fetchData();
		})
		.finally(() => {
			loading.value = false;
		})
		.catch(() => {
			envDetail.value = {};
			fetchData();
		});
};

watch(
	() => route.params,
	async (newId) => {
		fetchEnv();
	}
);
onMounted(() => {
	fetchEnv();
});
</script>
<style lang="scss" scoped>
.environments-wrapper {
	border-radius: 8px;
	border: 1px solid $separator;
}
</style>
