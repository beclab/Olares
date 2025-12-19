<template>
	<div class="adaptive-grid">
		<table>
			<tbody>
				<tr v-for="(row, index) in formattedRows" :key="index">
					<td
						v-for="(item, cellIndex) in row"
						:key="cellIndex"
						class="grid-cell"
					>
						<div class="cell-content">
							<div class="text-body3 text-ink-3 label">
								{{ item.label }}&nbsp;<span>:</span>&nbsp;
							</div>
							<div class="value">
								<div v-if="$slots[`field-${item.field}`]">
									<slot
										:name="`field-${item.field}`"
										:value="data[item.field]"
										:data="data"
									></slot>
								</div>
								<div v-else>{{ data[item.field] || '-' }}</div>
							</div>
						</div>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';

interface Column {
	label: string;
	field: string;
}

interface Props {
	columns: Column[];
	data: Record<string, any>;
	minColumnWidth?: number;
	maxColumn?: number;
}

const props = withDefaults(defineProps<Props>(), {
	columns: () => [],
	data: () => ({}),
	minColumnWidth: 100,
	maxColumn: 3
});

const containerWidth = ref(0);
let observer: ResizeObserver | null = null;

const columnsCount = computed(() => {
	if (!containerWidth.value) return 1;
	const maxColumns = Math.floor(containerWidth.value / props.minColumnWidth);
	return Math.max(1, Math.min(maxColumns, props.maxColumn));
});

const formattedRows = computed<Column[][]>(() => {
	const result: Column[][] = [];
	for (let i = 0; i < props.columns.length; i += columnsCount.value) {
		result.push(props.columns.slice(i, i + columnsCount.value));
	}
	return result;
});

onMounted(() => {
	const el = document.querySelector('.adaptive-grid');
	if (el) {
		observer = new ResizeObserver((entries) => {
			containerWidth.value = entries[0].contentRect.width;
		});
		observer.observe(el);
	}
});

onBeforeUnmount(() => {
	if (observer) {
		observer.disconnect();
	}
});
</script>

<style scoped>
.adaptive-grid {
	width: 100%;
}

table {
	width: 100%;
	border-collapse: collapse;
}

.grid-cell {
	padding: 16px;
}

tr:first-child .grid-cell {
	padding-top: 0;
}

tr td:first-child.grid-cell {
	padding-left: 0;
}

tr:last-child .grid-cell {
	padding-bottom: 0;
}

tr td:last-child.grid-cell {
	padding-right: 0;
}

.cell-content {
	display: flex;
	align-items: center;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.label {
	flex-shrink: 0;
}

.value {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
}
</style>
