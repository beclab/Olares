<template>
	<q-select
		v-bind="$attrs"
		v-model="modelValue"
		use-input
		hide-selected
		fill-input
		@input-value="setModel"
	>
		<!-- <template v-slot:no-option>
      <q-item>
        <q-item-section class="text-grey"> No results </q-item-section>
      </q-item>
    </template> -->
	</q-select>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const modelValue = ref();
const emits = defineEmits(['update:modelValue']);
const setModel = (val: any) => {
	modelValue.value = val;
	emits('update:modelValue', val);
};
</script>
