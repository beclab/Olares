<template>
	<div class="q-section-style-contianer">
		<slot></slot>
	</div>
</template>

<style lang="scss" scoped>
.q-section-style-contianer {
	max-width: 100%;
	::v-deep(.q-field--auto-height.q-field--dense .q-field__native) {
		height: 20px;
		min-height: 20px;
		color: $ink-2;
		padding: 0px;
		overflow: hidden;
		width: 100%;
		padding: 4px 4px 4px 8px;
	}

	::v-deep(.q-field--auto-height.q-field--dense .q-field__control) {
		height: 20px;
		min-height: 20px;
		border-radius: 999px;
	}

	::v-deep(.q-field--auto-height.q-field--dense .q-field__control::before) {
		border-color: $btn-stroke;
	}

	::v-deep(.q-field--dense .q-field__marginal) {
		height: 20px;
		min-height: 20px;
		color: $ink-2;
		padding-left: 0px;
	}
	::v-deep(.q-field--dense .q-field__marginal .q-select__dropdown-icon) {
		font-size: 18px;
	}
}
</style>
