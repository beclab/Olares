<template>
	<div
		class="row justify-center items-center"
		:class="selected ? 'avatar-selector-selected' : 'avatar-selector-normal'"
	>
		<q-img v-if="src" :src="src" class="selector-avatar" />
		<q-img v-else src="upload_default.svg" class="selector-avatar" />
	</div>
</template>

<script lang="ts" setup>
defineProps({
	src: String,
	selected: {
		type: Boolean,
		default: false
	}
});
</script>

<style scoped lang="scss">
.avatar-selector-selected {
	width: 48px;
	height: 48px;
	padding: 2px;
	border-radius: 4px;
	border: 1px solid $green-default;
}

.avatar-selector-normal {
	width: 48px;
	height: 48px;
	padding: 2px;
	border-radius: 4px;
	border: none;
}

.selector-avatar {
	width: 42px;
	height: 42px;
	border-radius: 4px;
}
</style>
