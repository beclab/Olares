<template>
	<div
		class="row text-subtitle3-m text-ink-3 q-gutter-x-sm q-gutter-y-sm select-bg"
	>
		<div
			v-for="item in options"
			:key="item.value"
			class="item-base row item-center justify-center"
			:class="{
				normal: !item.selected,
				select: item.selected
			}"
			:style="
				rowCount > 0
					? `min-width: calc((100% - ${rowCount - 1}*12px) / ${rowCount})`
					: ''
			"
			@click="itemClick(item)"
		>
			{{ item.label }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, PropType } from 'vue';

interface SelectItem {
	value: string | number;
	label: string;
	selected: boolean;
	isAll: boolean;
	isDefault: boolean;
}

const props = defineProps({
	rowCount: {
		type: Number,
		required: false,
		default: 0
	},
	options: {
		type: Object as PropType<SelectItem[]>,
		require: true,
		default: [] as SelectItem[]
	},
	singleSelect: {
		type: Boolean,
		required: false,
		default: true
	}
});

const itemClick = (item: SelectItem) => {
	if (!props.singleSelect) {
		if (item.isAll) {
			props.options.forEach((e) => {
				if (item.selected) {
					if (e.isAll || !e.selected) {
						return;
					}
					e.selected = false;
				} else {
					if (e.isAll || e.selected) {
						return;
					}
					e.selected = true;
				}
			});
			item.selected = !item.selected;
			return;
		} else {
			item.selected = !item.selected;
			if (isAllItem.value) {
				isAllItem.value.selected = !hasNoSelected.value;
			}
		}
	} else {
		if (item.selected) {
			return;
		}
		item.selected = true;
		props.options.forEach((e) => {
			if (e.selected && e.value != item.value) {
				e.selected = false;
			}
		});
	}
};

const isAllItem = computed(() => {
	return props.options.find((e) => e.isAll);
});

const hasNoSelected = computed(() => {
	return props.options.find((e) => !e.isAll && !e.selected) != undefined;
});
</script>

<style scoped lang="scss">
.select-bg {
	padding-bottom: 15px;
	padding-top: 5px;
}
.item-base {
	border-radius: 4px;
	height: 28px;
	line-height: 28px;
	padding-left: 10px;
	padding-right: 10px;
	min-width: 78px;
}

.normal {
	background: $background-3;
}

.select {
	background: $yellow-soft;
}
</style>
