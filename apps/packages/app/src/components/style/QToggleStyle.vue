<template>
	<span>
		<slot></slot>
	</span>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
::v-deep(.q-toggle__inner) {
	height: 16px;
	line-height: 16px;
	padding-top: 0;
	padding-bottom: 0;
}
::v-deep(.q-toggle__track) {
	height: 16px;
	border-radius: 8px;
	background: $background-5;
}
::v-deep(.q-toggle__thumb) {
	height: 12px !important;
	width: 12px;
	top: 2px;
}
::v-deep(.q-toggle__thumb:after) {
	box-shadow: none;
	background: $ink-on-brand;
	filter: drop-shadow(0px 2px 4px rgba(0, 0, 0, 0.1));
	left: 4px;
	right: -4px;
}

::v-deep(.q-toggle__inner--truthy .q-toggle__thumb:after) {
	right: 0px;
	left: 0px;
}
::v-deep(.q-toggle__thumb:before) {
	display: none;
}
</style>
