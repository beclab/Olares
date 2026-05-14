<template>
	<q-item-section class="bt-itemlabel-container">
		<q-item-label lines="1" style="display: flex; align-items: center">
			<div :class="labelTextClass" class="label-content">
				{{ data.label }}
			</div>
			<BadgeCount
				v-if="data.count != undefined"
				class="q-ml-sm badge"
				:active="active"
				:activeClass="activeClass"
			>
				{{ data.count }}
			</BadgeCount>
		</q-item-label>
	</q-item-section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import BadgeCount from './BadgeCount.vue';
import { defaultSize, ItemCell, Size } from './Menu';

interface Props {
	data: ItemCell;
	active: boolean;
	activeClass: string;
	size: Size;
}
const props = withDefaults(defineProps<Props>(), {
	size: defaultSize
});

const labelTextClass = computed(() => {
	switch (props.size) {
		case 'md':
			return 'text-body1';
		case 'sm':
			return 'text-body2';
		default:
			return 'text-body1';
	}
});
</script>

<style lang="scss" scoped>
.bt-itemlabel-container {
	.label-content {
		flex-shrink: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.badge {
		flex-shrink: 0;
	}
}
</style>
