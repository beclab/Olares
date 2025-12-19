<template>
	<BtSelect
		style="min-width: 200px"
		v-model="selectValue"
		@update:model-value="selectHandler"
		:options="options"
		dense
		outlined
		:disable="!appDetailStore.isAdmin"
	>
		<template v-slot:option-label="scope">
			{{ scope.opt.label }}
		</template>
	</BtSelect>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
	getPodsList,
	getUserList
} from '@apps/control-panel-common/src/network';
import QSectionStyle from '@apps/control-panel-common/src/components/QSectionStyle.vue';
import { initOptions, options, selectValue } from './config';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';

const appDetailStore = useAppDetailStore();

const emits = defineEmits(['update:model-value']);

const users = ref<any[]>([]);

const fetchUsers = () => {
	getUserList().then((res) => {
		users.value = res.data.data;

		if (users.value.length > 0) {
			const userOptions = users.value.map((item) => ({
				label: item.name,
				value: item.name
			}));
			options.value = initOptions.concat(userOptions);
			if (!appDetailStore.isAdmin) {
				const target = options.value.find(
					(item) => appDetailStore.user.username === item.value
				);
				if (target) {
					selectValue.value = target;
				}
			}
		}
	});
};
const selectHandler = () => {
	emits('update:model-value', selectValue.value);
};
onMounted(() => {
	fetchUsers();
});
</script>

<style></style>
