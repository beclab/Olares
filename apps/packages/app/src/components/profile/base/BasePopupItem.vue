<template>
	<q-item
		dense
		clickable
		@mouseenter="hover = true"
		@mouseleave="hover = false"
		:active="hover"
		active-class="bg-grey-1"
		@click="emit('onItemClick')"
	>
		<slot />
	</q-item>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const hover = ref();
const emit = defineEmits(['onItemClick']);
</script>

<style lang="scss" scoped>
.q-list--dense > .q-item,
.q-item--dense {
	min-height: 40px;
	padding: 8px !important;
}

.q-item--active {
	color: $grey-1 !important;
	background: $grey-1 !important;
}

.q-item {
	padding: 8px !important;
}
</style>
