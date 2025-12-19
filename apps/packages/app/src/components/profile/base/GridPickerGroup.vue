<template>
	<div
		:class="grid ? 'grid-picker-group' : 'row'"
		:style="{
			'--columns': columns,
			'--columnGap': columnGap,
			'--rowGap': rowGap
		}"
	>
		<slot></slot>
	</div>
</template>

<script lang="ts" setup>
import { ref, provide, watch, onMounted } from 'vue';

const props = defineProps({
	modelValue: {
		type: [String, Number],
		default: '',
		required: false
	},
	grid: {
		type: Boolean,
		default: true
	},
	columns: {
		type: Number,
		default: 3
	},
	columnGap: {
		type: String,
		default: '20px'
	},
	rowGap: {
		type: String,
		default: '20px'
	}
});

const childComponents = ref<any[]>([]);
const childValues = ref<any[]>([]);
const emit = defineEmits(['update:modelValue']);

function handleItemClick(button: any, value: string | number) {
	readyRefresh(value);
	if (value !== '') {
		console.log(value);
		emit('update:modelValue', value);
	}
}

const addChildComponent = (componentRef: any, value: string | number) => {
	childComponents.value.push(componentRef);
	childValues.value.push(value);
};

provide('addChildComponent', addChildComponent);
provide('handleItemClick', handleItemClick);

const executeRefresh = (selected: number) => {
	childComponents.value.forEach((item, index) => {
		if (index === selected) {
			item.value.exposed.setSelected(true);
		} else {
			item.value.exposed.setSelected(false);
		}
	});
};

const readyRefresh = (value: any) => {
	if (value) {
		const index = childValues.value.findIndex((item) => item === value);
		console.log(childValues.value);
		console.log(props.modelValue);
		if (index >= 0) {
			executeRefresh(index);
		}
	} else {
		executeRefresh(-1);
	}
};

watch(
	() => props.modelValue,
	() => {
		console.log('on watch');
		readyRefresh(props.modelValue);
	}
);

onMounted(() => {
	readyRefresh(props.modelValue);
});
</script>

<style scoped lang="scss">
.grid-picker-group {
	display: grid;
	align-items: center;
	justify-items: center;
	justify-content: center;
	grid-row-gap: var(--rowGap);
	grid-column-gap: var(--columnGap);
	grid-template-columns: repeat(var(--columns), minmax(0, 1fr));
}
</style>
