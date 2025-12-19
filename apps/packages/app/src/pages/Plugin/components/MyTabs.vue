<template>
	<div class="row my-tabs-container">
		<div
			class="col tab-item-wrapper"
			v-for="item in options"
			:key="item.value"
			@click="clickHandler(item)"
		>
			<div
				class="row items-center justify-center tab-item"
				:class="[
					modelValue === item.value ? 'active text-body3 text-grey-10' : ''
				]"
			>
				<q-img
					:src="item.img"
					:ratio="1"
					width="20px"
					spinner-size="xs"
					class="img"
				/>
				<div class="label text-body3 text-ink-1">{{ item.label }}</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
interface ListItem {
	label: string;
	value: string;
	img?: string;
}
interface Props {
	options: ListItem[];
	modelValue: string;
}
withDefaults(defineProps<Props>(), {});

const emits = defineEmits(['update:modelValue']);
const clickHandler = (data) => {
	emits('update:modelValue', data.value);
};
</script>

<style lang="scss" scoped>
.my-tabs-container {
	background: $grey-1;
	border-radius: 4px !important;
	.tab-item-wrapper {
		padding: 4px;
		cursor: pointer;
	}
	.tab-item {
		padding: 2px;
	}
	.active {
		background: white;
		border-radius: 4px;
	}
	.img {
		border-radius: 50%;
	}
	.label {
		margin-left: 6px;
	}
}
</style>
