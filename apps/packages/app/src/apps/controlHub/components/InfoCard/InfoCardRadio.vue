<template>
	<div class="row justify-between my-info-card-radio-container">
		<InfoCardItem
			v-for="(item, index) in list"
			:key="index"
			:active="index === active"
			@click="itemClick(index, item)"
			:used="item.used"
			:total="item.total"
			:name="item.name"
			:unit-type="item.unitType"
		></InfoCardItem>
	</div>
</template>

<script setup lang="ts">
import { ref, toRef } from 'vue';
import InfoCardItem from './InfoCardItem.vue';
import { InfoCardItemProps } from './InfoCardItem.vue';

export interface InfoCardRadioProps {
	list?: Array<InfoCardItemProps>;
	defaultActive?: number;
}
const emit = defineEmits<{
	(e: 'change', data: InfoCardItemProps, index: number): void;
}>();

const props = withDefaults(defineProps<InfoCardRadioProps>(), {
	defaultActive: 0
});

const active = ref(props.defaultActive);
const itemClick = (index: number, data: InfoCardItemProps) => {
	active.value = index;
	emit('change', data, index);
};
</script>

<style>
.my-info-card-radio-container {
}
</style>
