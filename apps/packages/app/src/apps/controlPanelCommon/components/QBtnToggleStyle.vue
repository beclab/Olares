<template>
	<div class="my-btn-toggle-wrapper">
		<slot></slot>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.my-btn-toggle-wrapper {
	border: 1px solid $btn-stroke;
	border-radius: 8px;
	overflow: hidden;
	::v-deep(> .q-btn-group > .q-btn-item:first-child) {
		border-radius: 0px;
		border-top-right-radius: 8px;
		border-bottom-right-radius: 8px;
	}

	::v-deep(> .q-btn-group > .q-btn-item:last-child) {
		border-radius: 0px;
		border-top-left-radius: 8px;
		border-bottom-left-radius: 8px;
	}
}
</style>
