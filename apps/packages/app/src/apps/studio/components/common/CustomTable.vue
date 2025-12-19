<template>
	<div class="header-row text-ink-3 text-body3">
		<div
			v-for="column in columns"
			:key="column.name"
			:style="{ textAlign: column.align }"
		>
			{{ column.label }}
		</div>
	</div>

	<div
		class="body-row text-ink-1 text-subtitle2"
		v-for="item in data"
		:key="item.key"
	>
		<div :style="{ textAlign: columns[0].align }">{{ item.key }}</div>
		<div :style="{ textAlign: columns[1].align }">{{ item.value }}</div>
		<div :style="{ textAlign: columns[2].align }">
			<slot name="body-cell-actions" :props="item" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, PropType } from 'vue';

interface ColumnsType {
	name: string;
	label: string;
	align: 'left' | 'right' | 'center';
}

defineProps({
	columns: {
		type: Array as PropType<ColumnsType[]>,
		required: true
	},
	data: {
		type: Array as PropType<any[]>,
		required: false,
		default: () => []
	}
});
</script>

<style lang="scss" scoped>
.header-row {
	height: 36px;
	padding: 0 12px;
	border-bottom: 1px solid $input-stroke;
	display: grid;
	grid-template-columns: 2fr 2fr 1fr;
	gap: 10px;
	padding: 10px;
}

.body-row {
	height: 64px;
	line-height: 64px;
	padding: 0 12px;
	border-bottom: 1px solid $input-stroke;
	display: grid;
	grid-template-columns: 2fr 2fr 1fr;
	gap: 10px;

	&:hover {
		background-color: rgba($color: #000000, $alpha: 0.05);
	}
}
</style>
