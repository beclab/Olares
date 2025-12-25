<template>
	<div class="my-expansion-container">
		<slot></slot>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.my-expansion-container {
	::v-deep(.q-expansion-item > .q-expansion-item__container > .q-item) {
		padding: 0 8px;
		color: $grey-10;
		font-size: 14px;
		font-style: normal;
		font-weight: 500;
		line-height: 20px;
		min-height: 40px;
	}

	::v-deep(
			.q-expansion-item
				> .q-expansion-item__container
				> .q-item
				> .q-item__section
		) {
		min-width: 24px;
		padding-right: 8px;
	}

	::v-deep(
			.q-expansion-item
				> .q-expansion-item__container
				> .q-item
				> .q-item__section--side
				> .q-icon
		) {
		font-size: 20px;
	}
}
</style>
