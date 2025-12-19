<template>
	<div class="q-py-xs q-px-md">
		<span>{{ text }}</span>
	</div>
</template>

<script setup lang="ts">
interface Props {
	text: string;
}

defineProps<Props>();
</script>

<style scoped lang="scss"></style>
