<template>
	<div style="width: 100%">
		<q-resize-observer @resize="onResize" />
		<slot v-bind:width="pageSize.width" v-bind:height="pageSize.height"></slot>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const pageSize = ref({ width: 0, height: 0 });
const onResize = (size: { width: number; height: number }) => {
	console.log('onResize', size);
	pageSize.value = size;
};
</script>

<style></style>
