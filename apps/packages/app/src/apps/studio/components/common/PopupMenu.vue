<template>
	<q-menu class="popup-menu">
		<q-list class="menu-list" dense padding>
			<q-item
				class="row items-center justify-start q-mx-xs menu-item"
				clickable
				v-close-popup
				v-for="item in items"
				:key="item.name"
				@click="handleEvent(item.name)"
			>
				<q-icon :name="item.icon" size="16px" class="q-mr-xs" color="ink-2" />
				<q-item-section class="text-ink-2 text-body3">
					{{ item.label }}</q-item-section
				>
			</q-item>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { OPERATE_ACTION } from '@apps/studio/src/types/core';

const props = defineProps({
	items: {
		type: Object,
		required: false
	},
	path: {
		type: String,
		required: false
	},
	label: {
		type: String,
		required: false
	}
});

const emits = defineEmits(['handleEvent']);

const handleEvent = async (action: OPERATE_ACTION) => {
	emits('handleEvent', action, props.path, props.label);
};
</script>

<style lang="scss" scoped>
.popup-menu {
	box-shadow: 0px 4px 10px 0px #00000033;
	padding: 0 4px;
	.menu-list {
		padding: 4px 0;
		.menu-item {
			border-radius: 4px;
			padding: 0 4px;
			margin: 0;
		}
	}
}
</style>
