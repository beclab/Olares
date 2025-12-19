<template>
	<div class="q-gutter-y-md">
		<div
			class="environments-wrapper q-px-lg q-py-md"
			v-for="item in envlist"
			:key="item.name"
		>
			<MyExpansion :label="item.name" :default-opened="!item.env">
				<MyChipList
					v-if="envFilter(item.env).length > 0"
					:data="envFilter(item.env)"
					dense-toggle
				>
				</MyChipList>
				<div v-else>
					<Empty></Empty>
				</div>
			</MyExpansion>
		</div>
		<Empty v-if="noData"></Empty>
	</div>
</template>

<script setup lang="ts">
import { ref, watchEffect, computed } from 'vue';
import { get, isEmpty } from 'lodash';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import MyChipList from '@apps/control-panel-common/src/containers/MyChipList.vue';
import { fetcEnvList } from '@apps/control-panel-common/src/containers/env';
import MyExpansion from '@apps/control-panel-common/src/components/MyExpansion.vue';

interface Props {
	detail: Record<string, any>;
}

const envlist = ref();

const props = withDefaults(defineProps<Props>(), {});

const envFilter = (data: any[]) => {
	const newData = data || [];
	const temp = newData.filter((item) => item.value);
	return temp;
};

const noData = computed(() => {
	return isEmpty(envlist.value);
});

watchEffect(async () => {
	if (props.detail) {
		const cluster = get(props.detail, 'cluster');
		const namespace = get(props.detail, 'namespace');
		const containers = get(props.detail, 'containers');
		const initContainers = get(props.detail, 'initContainers');

		envlist.value = await fetcEnvList({
			namespace: namespace,
			cluster: cluster,
			containers: containers,
			initContainers: initContainers
		});
	}
});
</script>
<style lang="scss" scoped>
.environments-wrapper {
	border-radius: 8px;
	border: 1px solid $separator;
}
</style>
