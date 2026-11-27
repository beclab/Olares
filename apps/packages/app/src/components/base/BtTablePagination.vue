<template>
	<div class="bt-table-pagination row items-center justify-end no-wrap">
		<q-select
			dense
			borderless
			options-dense
			emit-value
			map-options
			behavior="menu"
			:model-value="scope.pagination.rowsPerPage"
			:options="computedOptions"
			class="bt-table-pagination__select"
			popup-content-class="bt-table-pagination__popup"
			:popup-content-style="{ width: '72px', minWidth: '72px' }"
			@update:model-value="onRowsPerPageChange"
		/>

		<span class="text-body3 text-ink-3 bt-table-pagination__range q-mx-sm">
			{{ rangeText }}
		</span>

		<q-btn
			dense
			flat
			round
			size="sm"
			icon="chevron_left"
			color="ink-2"
			:disable="scope.isFirstPage"
			@click="scope.prevPage"
		/>
		<q-btn
			dense
			flat
			round
			size="sm"
			icon="chevron_right"
			color="ink-2"
			:disable="scope.isLastPage"
			@click="scope.nextPage"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useQuasar } from 'quasar';

interface PaginationScope {
	pagination: {
		page: number;
		rowsPerPage: number;
		rowsNumber?: number | undefined;
	};
	isFirstPage: boolean;
	isLastPage: boolean;
	prevPage: () => void;
	nextPage: () => void;
}

const props = withDefaults(
	defineProps<{
		/** The scope object provided by q-table's `pagination` slot. */
		scope: PaginationScope;
		rowsPerPageOptions?: number[];
	}>(),
	{
		// Mirror Quasar QTable's native default rows-per-page options.
		rowsPerPageOptions: () => [5, 7, 10, 15, 20, 25, 50, 0]
	}
);

const emit = defineEmits<{
	(e: 'update:rows-per-page', value: number): void;
}>();

const $q = useQuasar();

// Reproduce Quasar's `computedRowsPerPageOptions`: keep the current value in the
// list when it isn't among the configured options, and render 0 as "All".
const computedOptions = computed(() => {
	const current = props.scope.pagination.rowsPerPage;
	const base = props.rowsPerPageOptions.includes(current)
		? props.rowsPerPageOptions
		: [current, ...props.rowsPerPageOptions];
	return base.map((count) => ({
		label: count === 0 ? $q.lang.table.allRows : String(count),
		value: count
	}));
});

const rangeText = computed(() => {
	const { page, rowsPerPage } = props.scope.pagination;
	const rowsNumber = props.scope.pagination.rowsNumber ?? 0;
	if (!rowsPerPage || rowsPerPage <= 0) {
		return `${rowsNumber} / ${rowsNumber}`;
	}
	const start = rowsNumber === 0 ? 0 : (page - 1) * rowsPerPage + 1;
	const end = Math.min(page * rowsPerPage, rowsNumber);
	return `${start}-${end} / ${rowsNumber}`;
});

function onRowsPerPageChange(value: number) {
	emit('update:rows-per-page', value);
}
</script>

<style lang="scss" scoped>
.bt-table-pagination {
	&__select {
		width: 60px;
		min-width: 60px;

		:deep(.q-field__control),
		:deep(.q-field__control-container) {
			min-height: 28px;
			height: 28px;
		}

		:deep(.q-field__native) {
			min-height: 28px;
			padding: 0;
			align-items: center;
			line-height: 28px;
		}

		:deep(.q-field__append) {
			height: 28px;
			align-items: center;
			padding-left: 2px;
		}
	}

	&__range {
		white-space: nowrap;
	}
}
</style>

<!--
	The q-select dropdown is teleported to <body>, so scoped styles can't reach
	it. Use a global rule to keep the popup narrow on mobile instead of letting
	it stretch full-width.
-->
<style lang="scss">
.bt-table-pagination__popup {
	min-width: 72px !important;
	width: 72px !important;
	border-radius: 8px;

	.q-item {
		min-height: 36px;
		padding: 0 12px;
		justify-content: center;
	}
}
</style>
