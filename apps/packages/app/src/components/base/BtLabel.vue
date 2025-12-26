<template>
	<div class="row justify-start items-center cursor-pointer">
		<q-icon
			v-if="name"
			size="24px"
			color="blue-default"
			:name="name"
			style="margin-right: 4px"
		/>
		<div class="text-body-1 text-blue-default text-right">{{ label }}</div>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	name: String,
	label: String
});
</script>

<style scoped lang="scss"></style>
