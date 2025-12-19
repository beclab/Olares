<template>
	<div class="q-section-style-contianer">
		<slot></slot>
	</div>
</template>

<style lang="scss" scoped>
.q-section-style-contianer {
	::v-deep(.q-field--auto-height.q-field--dense .q-field__native) {
		height: 32px;
		min-height: 32px;
		color: $ink-2;
	}

	::v-deep(.q-field--auto-height.q-field--dense .q-field__control) {
		height: 32px;
		min-height: 32px;
	}

	::v-deep(.q-field--auto-height.q-field--dense .q-field__control::before) {
		border-color: $btn-stroke;
	}

	::v-deep(.q-field--dense .q-field__marginal) {
		height: 32px;
		min-height: 32px;
		color: $ink-2;
	}
}
</style>
