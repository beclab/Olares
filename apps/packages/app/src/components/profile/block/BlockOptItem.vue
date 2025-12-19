<template>
	<base-popup-item class="view-item-root row justify-start items-center">
		<q-icon v-if="icon" class="text-ink-2" size="16px" :name="icon" />
		<div class="text-body3 text-ink-2 q-ml-sm">
			{{ label }}
		</div>
	</base-popup-item>
</template>

<script lang="ts" setup>
import BasePopupItem from '@apps/profile/src/components/profile/base/BasePopupItem.vue';

defineProps({
	label: { type: String, require: true },
	icon: { type: String, require: true }
});
</script>

<style scoped lang="scss">
.view-item-root {
	width: 120px;
}
</style>
