<template>
	<BtSelect
		style="min-width: 200px"
		v-model="selectValue"
		@update:model-value="selectHandler"
		:options="options"
		dense
		outlined
	>
		<template v-slot:option-label="scope">
			{{ scope.opt.label }}
		</template>
	</BtSelect>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
	getNodesList,
	getPodsList
} from '@apps/control-panel-common/src/network';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import { GLOBAL_ROLE } from '@apps/control-panel-common/src/constant/user';
import { initOptions, options, selectValue } from './config';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';

const emits = defineEmits(['update:model-value']);

const users = ref<any[]>([]);

const getNodes = () => {
	const params = {
		sortBy: 'createTime'
	};
	getNodesList(params).then((res) => {
		users.value = res.data.items;

		if (users.value.length > 0) {
			const userOptions = users.value.map((item) => ({
				label: item.metadata.name,
				value: item.metadata.name,
				type: 'node'
			}));
			options.value = initOptions.concat(userOptions);
		}
	});
};
const selectHandler = () => {
	emits('update:model-value', selectValue.value);
};
onMounted(() => {
	getNodes();
});
</script>

<style></style>
