<template>
	<div class="around-borders-lg">
		<div class="my-table-container">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.my-table-container {
	& > ::v-deep(.q-table__container .q-table__card) {
		color: $grey-10;
	}

	& > ::v-deep(.q-table__container .q-table thead tr) {
		height: 40px;
		color: $ink-3;
	}

	& > ::v-deep(.q-table__container .q-table tbody td) {
		height: 56px;
		color: $ink-1;
	}

	& > ::v-deep(.q-table__container .q-table__control .q-table__bottom-item) {
		color: $grey-5;
	}

	& > ::v-deep(.q-table__container .q-table__bottom) {
		border-top: none;
	}
	::v-deep(.q-table th) {
		border-color: $separator;
	}
	::v-deep(.q-table--horizontal-separator tbody tr:not(:last-child) > td) {
		border-color: $separator;
	}
	::v-deep(.q-field__native, .q-field__marginal) {
		color: $ink-2;
	}
}
</style>
