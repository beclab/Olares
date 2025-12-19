<template>
	<q-menu
		@update:model-value="showPopupProxy"
		class="popup-menu"
		:offset="offset"
	>
		<q-list dense padding style="width: 120px">
			<q-item
				class="row items-center justify-start popup-item text-ink-2"
				clickable
				v-close-popup
			>
				<q-item-section class="menuName">ssss</q-item-section>
			</q-item>
			<q-item
				class="row items-center justify-start popup-item text-ink-2"
				clickable
				v-close-popup
			>
				<q-item-section class="menuName">sss</q-item-section>
			</q-item>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { PropType } from 'vue';

defineProps({
	item: {
		type: Object,
		required: false
	},
	offset: {
		type: Array as PropType<number[]>,
		default: () => [100, 5]
	}
});

const emits = defineEmits(['showPopupProxy']);

const showPopupProxy = (value: boolean) => {
	if (value) {
		onBeforeShow();
	}
	emits('showPopupProxy', value);
};

const onBeforeShow = async () => {};
</script>
<style lang="scss">
.menuName {
	white-space: nowrap;
}
</style>
