<template>
	<QTableStyle2>
		<q-table
			:rows="rows"
			:columns="columns"
			row-key="__key"
			flat
			:pagination="pagination"
			hide-pagination
		>
			<template #header-cell-size="props">
				<q-th :props="props">
					<div class="row items-center no-wrap q-gutter-xs">
						<span>{{ props.col.label }}</span>
						<q-select
							v-model="sizeUnitMode"
							:options="unitSelectOptions"
							dense
							borderless
							hide-selected
							dropdown-icon="expand_more"
							emit-value
							map-options
							class="lsblk-unit-select"
						/>
					</div>
				</q-th>
			</template>
			<template #header-cell-fsused="props">
				<q-th :props="props">
					<div class="row items-center no-wrap q-gutter-xs">
						<span>{{ props.col.label }}</span>
						<q-select
							v-model="fsusedUnitMode"
							:options="unitSelectOptions"
							dense
							borderless
							hide-selected
							dropdown-icon="expand_more"
							emit-value
							map-options
							class="lsblk-unit-select"
						/>
					</div>
				</q-th>
			</template>
			<template #body-cell-name="slotProps">
				<q-td :props="slotProps">
					<span class="lsblk-name-cell text-ink-1">
						<span class="lsblk-name-text">{{ slotProps.row.name }}</span>
					</span>
				</q-td>
			</template>
			<template #body-cell-size="slotProps">
				<q-td :props="slotProps">
					{{ formatLsblkDiskValue(slotProps.row.size, sizeUnitMode) }}
				</q-td>
			</template>
			<template #body-cell-fsused="slotProps">
				<q-td :props="slotProps">
					{{ formatLsblkDiskValue(slotProps.row.fsused, fsusedUnitMode) }}
				</q-td>
			</template>
		</q-table>
	</QTableStyle2>
</template>

<script setup lang="ts">
import QTableStyle2 from '@apps/control-panel-common/components/QTableStyle2.vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	formatLsblkDiskValue,
	LSBLK_DISK_FIXED_UNITS,
	type LsblkFlatRow
} from './config';
type LsblkDiskUnitMode = 'auto' | (typeof LSBLK_DISK_FIXED_UNITS)[number];

const props = defineProps<{
	rows: LsblkFlatRow[];
	columns: any[];
	pagination: { rowsNumber: number };
}>();

const { t, locale } = useI18n();

const sizeUnitMode = ref<LsblkDiskUnitMode>('auto');
const fsusedUnitMode = ref<LsblkDiskUnitMode>('auto');

const unitSelectOptions = computed(() => {
	locale.value;
	return [
		{ label: t('DISK_OP.LSBLK_UNIT_AUTO'), value: 'auto' as const },
		...LSBLK_DISK_FIXED_UNITS.map((u) => ({ label: u, value: u }))
	];
});

watch(
	() => props.rows,
	() => {
		sizeUnitMode.value = 'auto';
		fsusedUnitMode.value = 'auto';
	}
);
</script>

<style lang="scss" scoped>
.lsblk-unit-select {
	width: 16px;
	::v-deep(.q-field__native) {
		color: $ink-3 !important;
	}
	::v-deep(.q-field__append) {
		font-size: 16px !important;
		padding-left: 0px;
		color: $ink-3 !important;
	}
}

.lsblk-name-cell {
	display: inline-block;
	min-width: 0;
	line-height: 20px;
	word-break: break-all;
}

.lsblk-tree-prefix {
	display: inline;
	white-space: pre;
	font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
		'Liberation Mono', 'Courier New', monospace;
	font-size: 13px;
	letter-spacing: 0;
	user-select: none;
}
</style>
