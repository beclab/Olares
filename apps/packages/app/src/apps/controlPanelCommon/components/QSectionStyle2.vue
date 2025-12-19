<template>
	<div class="q-section-style-contianer">
		<slot></slot>
	</div>
</template>

<style lang="scss" scoped>
.q-section-style-contianer {
	::v-deep(.q-field--auto-height.q-field .q-field__native) {
		min-height: 32px;
		color: $ink-2;
	}

	::v-deep(.q-field--auto-height.q-field .q-field__control) {
		min-height: 32px;
	}

	::v-deep(.q-field--auto-height.q-field .q-field__control::before) {
		border-color: $btn-stroke;
	}

	::v-deep(.q-field .q-field__marginal) {
		min-height: 32px;
		color: $ink-2;
	}
}
</style>
