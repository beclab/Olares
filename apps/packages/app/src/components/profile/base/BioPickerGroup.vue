<template>
	<div class="picker-group-layout row">
		<slot></slot>
	</div>
</template>

<script lang="ts" setup>
import { ref, provide, nextTick, onMounted } from 'vue';

const props = defineProps({
	activeIndex: {
		type: Number,
		default: 0,
		required: false
	}
});

const childComponents = ref([]);
let selectedButton = ref(null);
const emit = defineEmits(['update:activeIndex']);

function handleItemClick(button) {
	selectedButton.value = button;
	const index = childComponents.value.findIndex((item) => item == button);
	console.log(index);
	executeRefresh(index);
	emit('update:activeIndex', index);
}

const addChildComponent = (componentRef: any) => {
	childComponents.value.push(componentRef);
	// console.log(childComponents)
};

provide('addChildComponent', addChildComponent);
provide('handleItemClick', handleItemClick);

const executeRefresh = (selected: number) => {
	nextTick(() => {
		childComponents.value.forEach((item, index) => {
			if (index === selected) {
				item.value.exposed.setSelected(true);
			} else {
				item.value.exposed.setSelected(false);
			}
		});
	});
};

onMounted(() => {
	executeRefresh(props.activeIndex);
});
</script>

<style scoped lang="scss">
.picker-group-layout {
	width: 100%;
	display: grid;
	align-items: center;
	justify-items: center;
	justify-content: center;
	padding-bottom: 20px;
	grid-column-gap: 20px;
	grid-row-gap: 16px;
	grid-template-columns: repeat(3, minmax(0, 1fr));
}
</style>
