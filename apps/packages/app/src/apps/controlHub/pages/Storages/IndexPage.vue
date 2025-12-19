<template>
	<div class="relative-position">
		<MenuHeader size="xxl"></MenuHeader>
		<bt-scroll-area style="height: calc(100vh - 56px)">
			<div class="relative-position q-px-xxl q-py-lg">
				<PersistentVolumeClaim></PersistentVolumeClaim>
			</div>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import PersistentVolumeClaim from './PersistentVolumeClaim.vue';
</script>

<style></style>
