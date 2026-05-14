<template>
	<q-icon
		class="bt-menu-icon item-icon"
		:size="btIconSize"
		v-if="data.icon"
		:name="data.icon"
	/>
	<img
		v-else-if="data.img"
		:src="active && data.activeImg ? data.activeImg : data.img"
		:style="{ height: btIconSize, maxWidth: btIconSize }"
	/>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { defaultSize, ItemCell, Size } from './Menu';

interface Props {
	data: ItemCell;
	active: boolean;
	size: Size;
}
const props = withDefaults(defineProps<Props>(), {
	size: defaultSize
});

const btIconSize = computed(() => {
	switch (props.size) {
		case 'md':
			return '20px';
		case 'sm':
			return '18px';
		default:
			return '20px';
	}
});
</script>

<style lang="scss" scoped>
.bt-menu-icon {
	&.item-icon {
		font-size: 24px;
	}
	&.active {
	}
}
</style>
