<template>
	<span class="btn-style-container">
		<slot></slot>
	</span>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.btn-style-container {
	::v-deep(.q-btn) {
		padding: 0;
		min-height: 26px;
		height: 26px;
		padding: 0 8px;
	}
	::v-deep(.q-btn--outline:before) {
		border: 1px solid $btn-stroke;
	}
	::v-deep(.q-btn__content) {
		padding: 0;
		line-height: 14px;
	}
}
</style>
