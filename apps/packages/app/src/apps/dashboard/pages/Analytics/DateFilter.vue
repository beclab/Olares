<template>
	<BtSelect
		v-model="dateValue[websiteId]"
		:options="dateFilterOption"
		outlined
		dense
		option-label="label"
		option-value="value"
		@update:model-value="change"
	>
	</BtSelect>
</template>

<script setup lang="ts">
import { dateValue, dateFilterOption } from './DatefilterState';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';

interface Props {
	websiteId: string;
}

withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['change']);
const change = () => {
	emits('change');
};
</script>

<style></style>
