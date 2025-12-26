<template>
	<div>
		<div
			class="my-table-container"
			:class="[
				stickyFirst ? 'first-sticky' : '',
				stickyLast ? 'last-sticky' : ''
			]"
		>
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { toRefs } from 'vue';

interface Props {
	stickyFirst: boolean;
	stickyLast: boolean;
}

const props = withDefaults(defineProps<Props>(), {});

const { stickyFirst, stickyLast } = toRefs(props);
</script>

<style lang="scss" scoped>
.my-table-container {
	border: 2px solid red;

	::v-deep(.q-table__container thead) {
		color: $ink-3;
		font-size: 12px;
		font-style: normal;
		font-weight: 400;
		line-height: 16px;
		background: $background-1;
	}
	::v-deep(.q-table__container thead tr) {
		height: auto;
	}
	::v-deep(td) {
		padding: 20px 20px 20px 0;
		font-size: 14px;
		font-style: normal;
		font-weight: 400;
		line-height: 20px;
		color: $ink-1;
	}
	::v-deep(th) {
		padding: 8px 0 8px 0;
	}
	::v-deep(.q-table th.sorted .q-table__sort-icon) {
		color: $ink-3;
	}
	::v-deep(.q-table__card .q-table__bottom) {
		font-size: 12px;
		color: $ink-3;
		border-top: 0 solid;
		padding-right: 0;
	}
	::v-deep(
			.q-table__container.q-table--horizontal-separator tbody tr:last-child > td
		) {
		border-bottom-width: 1px;
	}
	::v-deep(.q-table th) {
		border-color: $separator;
	}
	::v-deep(.q-table--horizontal-separator tbody tr:not(:last-child) > td) {
		border-color: $separator;
	}
	::v-deep(.q-table td) {
		border-color: $separator;
	}
	::v-deep(.q-field__native, .q-field__marginal) {
		color: $ink-2;
	}
}
</style>

<style lang="scss" scoped>
.first-sticky {
	::v-deep(thead tr:first-child th:first-child) {
		background-color: $background-1;
	}

	::v-deep(td) {
		background-color: $background-1;
	}

	::v-deep(th) {
		background-color: $background-1;
	}

	::v-deep(th:first-child) {
		position: sticky;
		left: 0;
		z-index: 1;
	}
	::v-deep(td:first-child) {
		position: sticky;
		left: 0;
		z-index: 1;
	}
}

.last-sticky {
	::v-deep(thead tr:first-child th:first-child) {
		background-color: $background-1;
	}

	::v-deep(td) {
		background-color: $background-1;
	}

	::v-deep(th:last-child) {
		position: sticky;
		right: 0;
		z-index: 1;
	}
	::v-deep(td:last-child) {
		position: sticky;
		right: 0;
		z-index: 1;
	}
}
</style>
