<template>
	<div class="olares-info-root">
		<div class="row items-center justify-between">
			<div class="text-body2 text-ink-1">
				{{ title }}
			</div>
			<q-icon :name="iconName" size="20px" color="ink-3" v-if="!hiddenArrow" />
		</div>
		<slot name="detail"></slot>
	</div>
</template>

<script setup lang="ts">
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	hiddenArrow: {
		type: Boolean,
		required: false,
		default: false
	},
	iconName: {
		type: String,
		required: false,
		default: 'sym_r_keyboard_arrow_right'
	}
});
</script>

<style scoped lang="scss"></style>
