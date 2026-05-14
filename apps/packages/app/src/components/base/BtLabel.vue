<template>
	<div class="row justify-start items-center cursor-pointer">
		<q-icon
			v-if="name"
			size="24px"
			color="primary"
			:name="name"
			style="margin-right: 4px"
		/>
		<div class="text-body-1 text-primary text-right">{{ label }}</div>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	name: String,
	label: String
});
</script>

<style scoped lang="scss"></style>
