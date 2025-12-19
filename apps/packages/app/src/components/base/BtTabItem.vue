<template>
	<div
		class="row justify-center items-center"
		:class="selected ? 'tab-item-root-selected' : 'tab-item-root'"
	>
		<q-icon v-if="icon" :name="icon" class="font-icon q-mr-xs" size="16px" />
		<div class="tab-item-title text-body2">{{ label }}</div>
	</div>
</template>

<script setup lang="ts">
defineProps({
	selected: {
		type: Boolean,
		default: false
	},
	label: {
		type: String,
		default: '',
		require: false
	},
	icon: {
		type: String,
		default: '',
		require: false
	}
});
</script>

<style scoped lang="scss">
.tab-item-root {
	height: 32px;
	width: 100px;
	padding-left: 12px;
	padding-right: 12px;
	color: $ink-2;
	text-align: center;
	border-radius: 8px;
}

.tab-item-root-selected {
	@extend .tab-item-root;
	border-radius: 8px;
	color: $blue-default;
	background-color: $blue-soft;
}

.tab-item-title {
	text-align: center;
	text-transform: capitalize;
}
</style>
