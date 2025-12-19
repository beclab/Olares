<template>
	<div class="q-input-style-contianer">
		<slot></slot>
	</div>
</template>

<style lang="scss" scoped>
.q-input-style-contianer {
	::v-deep(.q-field__control) {
		height: 32px;
		line-height: 16px;
	}
	::v-deep(.q-field--dark .q-field__native) {
		color: $ink-2;
	}
	::v-deep(.q-field__marginal) {
		height: 32px;
		font-size: 16px;
	}
	::v-deep(.q-field__control::before) {
		border-color: $btn-stroke;
	}
}
</style>
