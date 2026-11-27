<template>
	<div class="entry-diff-viewer">
		<header class="toolbar">
			<div class="title">🧪 {{ t('debug.entryDiff.title') }}</div>
			<div class="actions">
				<button class="btn" @click="refreshData">
					🔄 {{ t('debug.common.refresh') }}
				</button>
				<button class="btn" @click="exportReportJson">
					📤 {{ t('debug.entryDiff.exportJson') }}
				</button>
				<button class="btn primary" @click="exportReportMarkdown">
					📝 {{ t('debug.entryDiff.exportDoc') }}
				</button>
			</div>
		</header>

		<div class="summary">
			<div class="summary-item">
				<span class="label">{{ t('debug.entryDiff.total') }}</span>
				<span class="value">{{ stats.total }}</span>
			</div>
			<div class="summary-item">
				<span class="label">{{ t('debug.entryDiff.comparable') }}</span>
				<span class="value">{{ stats.comparable }}</span>
			</div>
			<div class="summary-item warn">
				<span class="label">{{ t('debug.entryDiff.mismatch') }}</span>
				<span class="value">{{ mismatchItems.length }}</span>
			</div>
			<div class="summary-item danger">
				<span class="label">{{ t('debug.entryDiff.missingRaw') }}</span>
				<span class="value">{{ stats.missingRaw }}</span>
			</div>
		</div>

		<div class="filter-row">
			<input
				v-model="searchQuery"
				class="search-input"
				type="text"
				:placeholder="`🔍 ${t('debug.entryDiff.searchPlaceholder')}`"
			/>
			<select v-model="selectedSource" class="source-select">
				<option value="">
					{{ t('debug.entryDiff.allSources') }} ({{ stats.mismatchTotal }}/{{
						stats.total
					}})
				</option>
				<option
					v-for="source in sourceOptions"
					:key="source.sourceId"
					:value="source.sourceId"
				>
					{{ source.sourceId }} ({{ source.mismatch }}/{{ source.total }})
				</option>
			</select>
			<button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
				{{ t('debug.common.clear') }}
			</button>
		</div>
		<div class="source-stats" v-if="selectedSource && selectedSourceStats">
			{{ t('debug.entryDiff.currentSource') }}
			<strong>{{ selectedSource }}</strong>
			: {{ t('debug.entryDiff.mismatch') }}
			<strong>{{ selectedSourceStats.mismatch }}</strong>
			/ {{ t('debug.entryDiff.total') }}
			<strong>{{ selectedSourceStats.total }}</strong>
		</div>
		<div class="category-row">
			<button
				:class="['category-chip', { active: selectedField === '' }]"
				@click="selectedField = ''"
			>
				{{ t('debug.entryDiff.allFields') }} ({{
					sourceScopedMismatchItems.length
				}})
			</button>
			<button
				v-for="item in fieldOptions"
				:key="item.field"
				:class="['category-chip', { active: selectedField === item.field }]"
				@click="selectedField = item.field"
			>
				{{ item.field }} ({{ item.count }})
			</button>
		</div>

		<div class="main">
			<div class="left">
				<div class="panel-title">
					{{ t('debug.entryDiff.mismatchItems') }}
					<span>{{ filteredMismatchItems.length }}</span>
				</div>
				<div class="list">
					<div
						v-for="item in filteredMismatchItems"
						:key="item.key"
						:class="['list-item', { active: selectedKey === item.key }]"
						@click="selectedKey = item.key"
					>
						<div class="line key">{{ item.key }}</div>
						<div class="line meta">
							<span>{{ item.appName || '-' }}</span>
							<span class="source-badge">{{ item.sourceId || 'unknown' }}</span>
							<span class="category-badge">
								{{ item.fieldGroups.slice(0, 2).join(', ') }}
								<span v-if="item.fieldGroups.length > 2">
									+{{ item.fieldGroups.length - 2 }}
								</span>
							</span>
							<span class="badge" :class="item.status">{{
								item.statusLabel
							}}</span>
						</div>
						<div class="line diff" v-if="item.diffPaths.length">
							{{ item.diffPaths.slice(0, 4).join(', ') }}
							<span v-if="item.diffPaths.length > 4">
								... +{{ item.diffPaths.length - 4 }}
							</span>
						</div>
					</div>
					<div v-if="filteredMismatchItems.length === 0" class="empty">
						{{ t('debug.entryDiff.noMismatchDetected') }}
					</div>
				</div>
			</div>

			<div class="right">
				<div class="panel-title">{{ t('debug.entryDiff.detail') }}</div>
				<div v-if="selectedItem" class="detail-wrap">
					<div class="detail-head">
						<div><strong>Key:</strong> {{ selectedItem.key }}</div>
						<div><strong>App:</strong> {{ selectedItem.appName || '-' }}</div>
						<div>
							<strong>Source:</strong> {{ selectedItem.sourceId || 'unknown' }}
						</div>
						<div>
							<strong>Fields:</strong> {{ selectedItem.fieldGroups.join(', ') }}
						</div>
						<div class="paths" v-if="selectedItem.diffPaths.length">
							<strong>Diff:</strong> {{ selectedItem.diffPaths.join(', ') }}
						</div>
					</div>
					<div class="diff-wrap">
						<div class="diff-header">
							<div class="diff-col-title">app_info.app_entry</div>
							<div class="diff-col-title">raw_data</div>
						</div>
						<div class="diff-body">
							<div
								v-for="(row, idx) in selectedDiffRows"
								:key="`${idx}-${row.leftNo}-${row.rightNo}-${row.type}`"
								:class="['diff-row', row.type]"
							>
								<div class="line-no left">{{ row.leftNo || '' }}</div>
								<div class="line-code left">{{ row.leftText }}</div>
								<div class="line-no right">{{ row.rightNo || '' }}</div>
								<div class="line-code right">{{ row.rightText }}</div>
							</div>
						</div>
					</div>
				</div>
				<div v-else class="empty-detail">
					{{ t('debug.entryDiff.selectLeftItem') }}
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '../../../stores/market/appStore';
import type { AppFullInfoLatest } from '../../../constant/constants';

type DiffStatus = 'mismatch' | 'missing-raw' | 'missing-entry';

interface DiffItem {
	key: string;
	sourceId: string;
	appName: string;
	appEntry: unknown;
	rawData: unknown;
	status: DiffStatus;
	statusLabel: string;
	diffPaths: string[];
	fieldGroups: string[];
}

interface DiffRow {
	type: 'same' | 'add' | 'remove' | 'change';
	leftNo: number | null;
	rightNo: number | null;
	leftText: string;
	rightText: string;
}

interface SourceStatsItem {
	sourceId: string;
	total: number;
	mismatch: number;
}

interface FieldOption {
	field: string;
	count: number;
}

const appStore = useAppStore();
const { t } = useI18n();
const refreshTick = ref(0);
const selectedKey = ref<string>('');
const searchQuery = ref('');
const selectedSource = ref('');
const selectedField = ref('');

const normalize = (value: unknown): unknown => {
	if (Array.isArray(value)) {
		return value.map((item) => normalize(item));
	}
	if (value && typeof value === 'object') {
		const obj = value as Record<string, unknown>;
		const sorted: Record<string, unknown> = {};
		Object.keys(obj)
			.sort()
			.forEach((k) => {
				sorted[k] = normalize(obj[k]);
			});
		return sorted;
	}
	return value;
};

const isDeepEqual = (a: unknown, b: unknown): boolean => {
	return JSON.stringify(normalize(a)) === JSON.stringify(normalize(b));
};

const collectDiffPaths = (a: unknown, b: unknown, basePath = ''): string[] => {
	if (isDeepEqual(a, b)) {
		return [];
	}

	const path = basePath || 'root';
	if (
		!a ||
		!b ||
		typeof a !== 'object' ||
		typeof b !== 'object' ||
		Array.isArray(a) !== Array.isArray(b)
	) {
		return [path];
	}

	if (Array.isArray(a) && Array.isArray(b)) {
		if (a.length !== b.length) {
			return [path];
		}
		let paths: string[] = [];
		for (let i = 0; i < a.length; i++) {
			paths = paths.concat(collectDiffPaths(a[i], b[i], `${path}[${i}]`));
		}
		return paths.length ? paths : [path];
	}

	const objA = a as Record<string, unknown>;
	const objB = b as Record<string, unknown>;
	const allKeys = new Set([...Object.keys(objA), ...Object.keys(objB)]);
	let paths: string[] = [];
	allKeys.forEach((key) => {
		paths = paths.concat(
			collectDiffPaths(
				objA[key],
				objB[key],
				basePath ? `${basePath}.${key}` : key
			)
		);
	});
	return paths.length ? paths : [path];
};

const normalizeFieldPath = (path: string): string => {
	if (!path) return '';
	return path.replace(/^root\.?/, '').replace(/\[\d+\]/g, '');
};

const pathToFieldGroup = (path: string): string => {
	const normalized = normalizeFieldPath(path);
	if (!normalized) return 'root';
	return normalized.split('.')[0] || 'root';
};

const getRawDataCandidate = (full: AppFullInfoLatest): unknown => {
	const root = full as any;
	const pkg = root?.raw_package as any;
	return (
		root?.raw_data ??
		root?.app_info?.raw_data ??
		pkg?.raw_data ??
		pkg?.app_info?.raw_data ??
		pkg?.app_entry ??
		pkg?.app_info?.app_entry ??
		null
	);
};

const diffItems = computed<DiffItem[]>(() => {
	refreshTick.value;
	const list: DiffItem[] = [];
	appStore.appFullInfoMap.forEach(
		(fullInfo: AppFullInfoLatest, key: string) => {
			const [sourceId = ''] = key.split('_');
			const appEntry = fullInfo?.app_info?.app_entry ?? null;
			const rawData = getRawDataCandidate(fullInfo);
			const appName =
				(appEntry as any)?.name ?? fullInfo?.app_simple_info?.app_name ?? '';

			if (!appEntry) {
				list.push({
					key,
					sourceId,
					appName,
					appEntry,
					rawData,
					status: 'missing-entry',
					statusLabel: t('debug.entryDiff.missingAppEntry'),
					diffPaths: ['app_info.app_entry'],
					fieldGroups: []
				});
				return;
			}

			if (!rawData) {
				list.push({
					key,
					sourceId,
					appName,
					appEntry,
					rawData,
					status: 'missing-raw',
					statusLabel: t('debug.entryDiff.missingRaw'),
					diffPaths: ['raw_data'],
					fieldGroups: []
				});
				return;
			}

			const paths = collectDiffPaths(appEntry, rawData);
			if (paths.length > 0) {
				const fieldGroups = Array.from(
					new Set(paths.map((path) => pathToFieldGroup(path)))
				).sort();
				list.push({
					key,
					sourceId,
					appName,
					appEntry,
					rawData,
					status: 'mismatch',
					statusLabel: t('debug.entryDiff.fieldMismatch'),
					diffPaths: paths,
					fieldGroups
				});
			}
		}
	);
	return list.sort((a, b) => a.key.localeCompare(b.key));
});

const mismatchItems = computed(() => diffItems.value);

const sourceOptions = computed(() => {
	const options = new Map<string, SourceStatsItem>();
	appStore.appFullInfoMap.forEach(
		(_fullInfo: AppFullInfoLatest, key: string) => {
			const [sourceId = ''] = key.split('_');
			if (!sourceId) return;
			if (!options.has(sourceId)) {
				options.set(sourceId, { sourceId, total: 0, mismatch: 0 });
			}
			options.get(sourceId)!.total += 1;
		}
	);
	mismatchItems.value.forEach((item) => {
		if (!item.sourceId) return;
		if (!options.has(item.sourceId)) {
			options.set(item.sourceId, {
				sourceId: item.sourceId,
				total: 0,
				mismatch: 0
			});
		}
		options.get(item.sourceId)!.mismatch += 1;
	});
	return Array.from(options.values()).sort((a, b) =>
		a.sourceId.localeCompare(b.sourceId)
	);
});

const sourceScopedMismatchItems = computed(() => {
	return mismatchItems.value.filter((item) => {
		if (item.status !== 'mismatch') return false;
		if (selectedSource.value && item.sourceId !== selectedSource.value) {
			return false;
		}
		return true;
	});
});

const fieldOptions = computed<FieldOption[]>(() => {
	const map = new Map<string, number>();
	sourceScopedMismatchItems.value.forEach((item) => {
		const dedup = new Set(item.fieldGroups);
		dedup.forEach((field) => {
			map.set(field, (map.get(field) || 0) + 1);
		});
	});
	return Array.from(map.entries())
		.map(([field, count]) => ({
			field,
			count
		}))
		.sort((a, b) => b.count - a.count);
});

watch(
	() => [selectedSource.value, fieldOptions.value] as const,
	() => {
		if (!selectedField.value) return;
		const exists = fieldOptions.value.some(
			(item) => item.field === selectedField.value
		);
		if (!exists) {
			selectedField.value = '';
		}
	}
);

const filteredMismatchItems = computed(() => {
	const q = searchQuery.value.trim().toLowerCase();
	return mismatchItems.value.filter((item) => {
		if (selectedSource.value && item.sourceId !== selectedSource.value) {
			return false;
		}
		if (
			selectedField.value &&
			!item.fieldGroups.includes(selectedField.value)
		) {
			return false;
		}
		if (!q) return true;
		if (item.key.toLowerCase().includes(q)) return true;
		if ((item.sourceId || '').toLowerCase().includes(q)) return true;
		if ((item.appName || '').toLowerCase().includes(q)) return true;
		if (item.fieldGroups.some((field) => field.toLowerCase().includes(q))) {
			return true;
		}
		return item.diffPaths.some((path) => path.toLowerCase().includes(q));
	});
});

const stats = computed(() => {
	const total = appStore.appFullInfoMap.size;
	const mismatchTotal = diffItems.value.length;
	const missingRaw = diffItems.value.filter(
		(item) => item.status === 'missing-raw'
	).length;
	const missingEntry = diffItems.value.filter(
		(item) => item.status === 'missing-entry'
	).length;
	return {
		total,
		mismatchTotal,
		missingRaw,
		missingEntry,
		comparable: Math.max(total - missingRaw - missingEntry, 0)
	};
});

const selectedSourceStats = computed(() => {
	if (!selectedSource.value) return null;
	return (
		sourceOptions.value.find(
			(item) => item.sourceId === selectedSource.value
		) ?? null
	);
});

const selectedItem = computed(() => {
	return (
		filteredMismatchItems.value.find(
			(item) => item.key === selectedKey.value
		) ?? null
	);
});

const toJsonLines = (value: unknown): string[] => {
	const text = stringify(value);
	return text.split('\n');
};

const getLineOps = (left: string[], right: string[]) => {
	const m = left.length;
	const n = right.length;
	const dp: number[][] = Array.from({ length: m + 1 }, () =>
		Array(n + 1).fill(0)
	);

	for (let i = m - 1; i >= 0; i--) {
		for (let j = n - 1; j >= 0; j--) {
			if (left[i] === right[j]) {
				dp[i][j] = dp[i + 1][j + 1] + 1;
			} else {
				dp[i][j] = Math.max(dp[i + 1][j], dp[i][j + 1]);
			}
		}
	}

	const ops: Array<
		| { type: 'same'; text: string }
		| { type: 'delete'; text: string }
		| { type: 'add'; text: string }
	> = [];

	let i = 0;
	let j = 0;
	while (i < m && j < n) {
		if (left[i] === right[j]) {
			ops.push({ type: 'same', text: left[i] });
			i++;
			j++;
		} else if (dp[i + 1][j] >= dp[i][j + 1]) {
			ops.push({ type: 'delete', text: left[i] });
			i++;
		} else {
			ops.push({ type: 'add', text: right[j] });
			j++;
		}
	}
	while (i < m) {
		ops.push({ type: 'delete', text: left[i++] });
	}
	while (j < n) {
		ops.push({ type: 'add', text: right[j++] });
	}

	return ops;
};

const buildDiffRows = (left: string[], right: string[]): DiffRow[] => {
	const ops = getLineOps(left, right);
	const rows: DiffRow[] = [];
	let leftNo = 1;
	let rightNo = 1;

	for (let i = 0; i < ops.length; i++) {
		const op = ops[i];
		if (op.type === 'same') {
			rows.push({
				type: 'same',
				leftNo,
				rightNo,
				leftText: op.text,
				rightText: op.text
			});
			leftNo++;
			rightNo++;
			continue;
		}

		if (op.type === 'delete') {
			const deletes: string[] = [];
			while (i < ops.length && ops[i].type === 'delete') {
				deletes.push(ops[i].text);
				i++;
			}
			const adds: string[] = [];
			while (i < ops.length && ops[i].type === 'add') {
				adds.push(ops[i].text);
				i++;
			}
			i--;

			const maxLen = Math.max(deletes.length, adds.length);
			for (let k = 0; k < maxLen; k++) {
				const leftText = deletes[k] ?? '';
				const rightText = adds[k] ?? '';
				const type: DiffRow['type'] =
					leftText && rightText ? 'change' : leftText ? 'remove' : 'add';
				rows.push({
					type,
					leftNo: leftText ? leftNo++ : null,
					rightNo: rightText ? rightNo++ : null,
					leftText,
					rightText
				});
			}
			continue;
		}

		rows.push({
			type: 'add',
			leftNo: null,
			rightNo: rightNo++,
			leftText: '',
			rightText: op.text
		});
	}

	return rows;
};

const selectedDiffRows = computed<DiffRow[]>(() => {
	if (!selectedItem.value) return [];
	const left = toJsonLines(selectedItem.value.appEntry);
	const right = toJsonLines(selectedItem.value.rawData);
	return buildDiffRows(left, right);
});

const refreshData = () => {
	refreshTick.value += 1;
	if (!selectedItem.value && mismatchItems.value.length > 0) {
		selectedKey.value = mismatchItems.value[0].key;
	}
};

const getTimestamp = () => {
	const now = new Date();
	const pad = (n: number) => String(n).padStart(2, '0');
	return `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(
		now.getDate()
	)}_${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
};

const buildReportData = () => {
	const visible = filteredMismatchItems.value;
	const visibleMismatch = visible.filter((item) => item.status === 'mismatch');
	const visibleMissing = visible.filter((item) => item.status !== 'mismatch');

	const bySource = new Map<
		string,
		{ total: number; mismatch: number; missing: number }
	>();
	visible.forEach((item) => {
		const key = item.sourceId || 'unknown';
		if (!bySource.has(key)) {
			bySource.set(key, { total: 0, mismatch: 0, missing: 0 });
		}
		const row = bySource.get(key)!;
		row.total += 1;
		if (item.status === 'mismatch') row.mismatch += 1;
		else row.missing += 1;
	});

	const byField = new Map<string, number>();
	visibleMismatch.forEach((item) => {
		new Set(item.fieldGroups).forEach((field) => {
			byField.set(field, (byField.get(field) || 0) + 1);
		});
	});

	const diffPathCounts = new Map<string, number>();
	visibleMismatch.forEach((item) => {
		item.diffPaths.forEach((path) => {
			diffPathCounts.set(path, (diffPathCounts.get(path) || 0) + 1);
		});
	});

	const topDiffPaths = Array.from(diffPathCounts.entries())
		.map(([path, count]) => ({ path, count }))
		.sort((a, b) => b.count - a.count)
		.slice(0, 30);

	const getTopLevelFieldObject = (obj: unknown, field: string) => {
		if (field === 'root') return obj;
		if (!obj || typeof obj !== 'object') return null;
		return (obj as Record<string, unknown>)[field] ?? null;
	};

	const fieldExamples = Array.from(byField.keys())
		.map((field) => {
			const sample = visibleMismatch.find((item) =>
				item.fieldGroups.includes(field)
			);
			if (!sample) return null;
			return {
				field,
				key: sample.key,
				sourceId: sample.sourceId,
				appName: sample.appName,
				appEntry: getTopLevelFieldObject(sample.appEntry, field),
				rawData: getTopLevelFieldObject(sample.rawData, field),
				diffPaths: sample.diffPaths.filter(
					(path) => pathToFieldGroup(path) === field
				)
			};
		})
		.filter((item): item is NonNullable<typeof item> => !!item)
		.sort((a, b) => a.field.localeCompare(b.field));

	const list = visible.map((item) => ({
		key: item.key,
		sourceId: item.sourceId,
		appName: item.appName,
		status: item.status,
		fieldGroups: item.fieldGroups,
		diffPathCount: item.diffPaths.length,
		diffPathsPreview: item.diffPaths.slice(0, 10)
	}));

	return {
		generatedAt: new Date().toISOString(),
		filters: {
			searchQuery: searchQuery.value,
			selectedSource: selectedSource.value || null,
			selectedField: selectedField.value || null
		},
		summary: {
			appFullInfoTotal: stats.value.total,
			currentViewCount: visible.length,
			currentViewMismatchCount: visibleMismatch.length,
			currentViewMissingCount: visibleMissing.length
		},
		bySource: Array.from(bySource.entries())
			.map(([sourceId, value]) => ({ sourceId, ...value }))
			.sort((a, b) => b.total - a.total),
		byField: Array.from(byField.entries())
			.map(([field, count]) => ({ field, count }))
			.sort((a, b) => b.count - a.count),
		topDiffPaths,
		fieldExamples,
		items: list
	};
};

const downloadTextFile = (
	content: string,
	filename: string,
	mime = 'text/plain;charset=utf-8'
) => {
	const blob = new Blob([content], { type: mime });
	const link = document.createElement('a');
	link.href = URL.createObjectURL(blob);
	link.download = filename;
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
	URL.revokeObjectURL(link.href);
};

const exportReportJson = () => {
	const data = buildReportData();
	const stamp = getTimestamp();
	downloadTextFile(
		JSON.stringify(data, null, 2),
		`app-entry-diff-report_${stamp}.json`,
		'application/json;charset=utf-8'
	);
};

const exportReportMarkdown = () => {
	const data = buildReportData();
	const stamp = getTimestamp();
	const lines: string[] = [];
	lines.push('# App Entry Diff Report');
	lines.push('');
	lines.push(`- ${t('debug.entryDiff.generatedAt')}: ${data.generatedAt}`);
	lines.push(
		`- ${t('debug.entryDiff.filters')}: source=\`${
			data.filters.selectedSource || 'ALL'
		}\`, field=\`${data.filters.selectedField || 'ALL'}\`, search=\`${
			data.filters.searchQuery || ''
		}\``
	);
	lines.push('');
	lines.push(t('debug.entryDiff.report.overview'));
	lines.push('');
	lines.push(t('debug.entryDiff.report.metricValue'));
	lines.push('| --- | ---: |');
	lines.push(
		`| ${t('debug.entryDiff.report.appFullInfoTotal')} | ${
			data.summary.appFullInfoTotal
		} |`
	);
	lines.push(
		`| ${t('debug.entryDiff.report.currentViewCount')} | ${
			data.summary.currentViewCount
		} |`
	);
	lines.push(
		`| ${t('debug.entryDiff.report.currentViewMismatchCount')} | ${
			data.summary.currentViewMismatchCount
		} |`
	);
	lines.push(
		`| ${t('debug.entryDiff.report.currentViewMissingCount')} | ${
			data.summary.currentViewMissingCount
		} |`
	);
	lines.push('');

	lines.push(t('debug.entryDiff.report.bySource'));
	lines.push('');
	lines.push(t('debug.entryDiff.report.bySourceHeader'));
	lines.push('| --- | ---: | ---: | ---: |');
	if (data.bySource.length === 0) {
		lines.push('| (empty) | 0 | 0 | 0 |');
	} else {
		data.bySource.forEach((item) => {
			lines.push(
				`| ${item.sourceId} | ${item.total} | ${item.mismatch} | ${item.missing} |`
			);
		});
	}
	lines.push('');

	lines.push(t('debug.entryDiff.report.byField'));
	lines.push('');
	lines.push(t('debug.entryDiff.report.byFieldHeader'));
	lines.push('| --- | ---: |');
	if (data.byField.length === 0) {
		lines.push('| (empty) | 0 |');
	} else {
		data.byField.forEach((item) => {
			lines.push(`| ${item.field} | ${item.count} |`);
		});
	}
	lines.push('');

	lines.push(t('debug.entryDiff.report.topDiffPaths'));
	lines.push('');
	lines.push(t('debug.entryDiff.report.pathCountHeader'));
	lines.push('| --- | ---: |');
	if (data.topDiffPaths.length === 0) {
		lines.push('| (empty) | 0 |');
	} else {
		data.topDiffPaths.forEach((item) => {
			lines.push(`| \`${item.path}\` | ${item.count} |`);
		});
	}
	lines.push('');

	lines.push(t('debug.entryDiff.report.fieldExamples'));
	lines.push('');

	if (!data.fieldExamples || data.fieldExamples.length === 0) {
		lines.push(t('debug.entryDiff.report.noFieldExamples'));
		lines.push('');
	} else {
		data.fieldExamples.forEach((example) => {
			lines.push(
				`### ${t('debug.entryDiff.report.fieldLabel')}: ${example.field}`
			);
			lines.push('');
			lines.push(
				`- key: \`${example.key}\`, source: \`${example.sourceId}\`, app: \`${example.appName}\``
			);
			lines.push('');
			lines.push(
				`- ${t('debug.entryDiff.report.relatedPaths')}: ${
					example.diffPaths.join(', ') || '(none)'
				}`
			);
			lines.push('');
			lines.push(t('debug.entryDiff.report.appEntryBlock'));
			lines.push('');
			lines.push('```json');
			lines.push(JSON.stringify(example.appEntry, null, 2));
			lines.push('```');
			lines.push('');
			lines.push(t('debug.entryDiff.report.rawDataBlock'));
			lines.push('');
			lines.push('```json');
			lines.push(JSON.stringify(example.rawData, null, 2));
			lines.push('```');
			lines.push('');
		});
	}

	lines.push(t('debug.entryDiff.report.itemsCurrentView'));
	lines.push('');
	lines.push(t('debug.entryDiff.report.itemsHeader'));
	lines.push('| --- | --- | --- | --- | --- | ---: | --- |');
	if (data.items.length === 0) {
		lines.push('| (empty) | - | - | - | - | 0 | - |');
	} else {
		data.items.forEach((item) => {
			lines.push(
				`| \`${item.key}\` | ${item.sourceId} | ${item.appName || '-'} | ${
					item.status
				} | ${item.fieldGroups.join(', ') || '-'} | ${item.diffPathCount} | ${
					item.diffPathsPreview.join(', ') || '-'
				} |`
			);
		});
	}

	downloadTextFile(
		lines.join('\n'),
		`AppEntry-RawData-Diff-${stamp}.md`,
		'text/markdown;charset=utf-8'
	);
};

const stringify = (value: unknown) => {
	try {
		return JSON.stringify(value, null, 2);
	} catch (error) {
		return String(value);
	}
};
</script>

<style scoped>
.entry-diff-viewer {
	height: 100%;
	display: flex;
	flex-direction: column;
	background: #fff;
}

.toolbar {
	height: 48px;
	background: #fff;
	border-bottom: 1px solid #e2e8f0;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 16px;
	flex-shrink: 0;
}

.title {
	font-size: 14px;
	font-weight: 600;
	color: #1e293b;
}

.actions {
	display: flex;
	gap: 8px;
}

.btn {
	padding: 6px 12px;
	border: 1px solid #e2e8f0;
	border-radius: 6px;
	background: #fff;
	font-size: 11px;
	cursor: pointer;
	transition: all 0.15s;
}

.btn:hover {
	border-color: #6366f1;
	color: #6366f1;
}

.btn.primary {
	background: #4f46e5;
	border-color: #4f46e5;
	color: #fff;
}

.btn.primary:hover {
	background: #4338ca;
	border-color: #4338ca;
	color: #fff;
}

.summary {
	display: flex;
	gap: 8px;
	padding: 10px 16px;
	border-bottom: 1px solid #f1f5f9;
	flex-shrink: 0;
}

.filter-row {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 8px 16px;
	border-bottom: 1px solid #f1f5f9;
}

.search-input {
	flex: 1;
	height: 32px;
	border: 1px solid #dbe3ef;
	border-radius: 8px;
	padding: 0 10px;
	font-size: 12px;
	color: #0f172a;
	outline: none;
}

.search-input:focus {
	border-color: #6366f1;
}

.source-select {
	height: 32px;
	border: 1px solid #dbe3ef;
	border-radius: 8px;
	padding: 0 8px;
	font-size: 12px;
	color: #0f172a;
	background: #fff;
	outline: none;
}

.source-select:focus {
	border-color: #6366f1;
}

.clear-btn {
	height: 32px;
	padding: 0 10px;
	border: 1px solid #dbe3ef;
	border-radius: 8px;
	background: #fff;
	font-size: 12px;
	color: #334155;
	cursor: pointer;
}

.clear-btn:hover {
	border-color: #6366f1;
	color: #6366f1;
}

.source-stats {
	padding: 6px 16px 10px;
	border-bottom: 1px solid #f1f5f9;
	font-size: 12px;
	color: #475569;
}

.source-stats strong {
	color: #0f172a;
}

.category-row {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 8px 16px;
	border-bottom: 1px solid #f1f5f9;
	overflow-x: auto;
}

.category-chip {
	height: 30px;
	padding: 0 10px;
	border: 1px solid #dbe3ef;
	border-radius: 999px;
	background: #fff;
	font-size: 12px;
	color: #334155;
	cursor: pointer;
	white-space: nowrap;
}

.category-chip.active {
	border-color: #6366f1;
	background: #eef2ff;
	color: #4338ca;
}

.summary-item {
	display: flex;
	align-items: center;
	gap: 8px;
	font-size: 12px;
	padding: 6px 10px;
	background: #f8fafc;
	border-radius: 8px;
}

.summary-item .label {
	color: #64748b;
}

.summary-item .value {
	color: #0f172a;
	font-weight: 700;
}

.summary-item.warn {
	background: #fffbeb;
}

.summary-item.warn .value {
	color: #b45309;
}

.summary-item.danger {
	background: #fef2f2;
}

.summary-item.danger .value {
	color: #b91c1c;
}

.main {
	flex: 1;
	display: flex;
	min-height: 0;
}

.left {
	width: 420px;
	border-right: 1px solid #e2e8f0;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.right {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.panel-title {
	height: 40px;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0 12px;
	font-size: 12px;
	font-weight: 700;
	color: #334155;
	border-bottom: 1px solid #f1f5f9;
}

.list {
	flex: 1;
	overflow: auto;
	padding: 8px;
}

.list-item {
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	padding: 8px;
	margin-bottom: 8px;
	cursor: pointer;
	background: #fff;
}

.list-item.active {
	border-color: #6366f1;
	background: #eef2ff;
}

.line {
	font-size: 11px;
	color: #334155;
}

.line.key {
	font-weight: 600;
	word-break: break-all;
}

.line.meta {
	margin-top: 4px;
	display: flex;
	justify-content: flex-end;
	align-items: center;
	gap: 6px;
	color: #64748b;
}

.source-badge {
	font-size: 10px;
	padding: 2px 6px;
	border-radius: 10px;
	background: #e0e7ff;
	color: #4338ca;
}

.category-badge {
	font-size: 10px;
	padding: 2px 6px;
	border-radius: 10px;
	background: #f1f5f9;
	color: #334155;
}

.line.diff {
	margin-top: 6px;
	color: #b45309;
	font-family: 'SF Mono', 'Fira Code', monospace;
	word-break: break-all;
}

.badge {
	font-size: 10px;
	padding: 2px 6px;
	border-radius: 10px;
}

.badge.mismatch {
	background: #fff7ed;
	color: #c2410c;
}

.badge.missing-raw,
.badge.missing-entry {
	background: #fef2f2;
	color: #b91c1c;
}

.detail-wrap {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

.detail-head {
	padding: 10px 12px;
	border-bottom: 1px solid #f1f5f9;
	font-size: 12px;
	color: #334155;
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.paths {
	word-break: break-all;
}

.diff-wrap {
	flex: 1;
	display: flex;
	flex-direction: column;
	padding: 8px;
	min-height: 0;
}

.diff-header {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 8px;
	margin-bottom: 8px;
}

.diff-col-title {
	height: 32px;
	display: flex;
	align-items: center;
	padding: 0 10px;
	font-size: 12px;
	font-weight: 700;
	color: #334155;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	background: #fff;
}

.diff-body {
	flex: 1;
	overflow: auto;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	background: #fff;
}

.diff-row {
	display: grid;
	grid-template-columns: 56px minmax(0, 1fr) 56px minmax(0, 1fr);
	border-bottom: 1px solid #f1f5f9;
	font-family: 'SF Mono', 'Fira Code', monospace;
	font-size: 11px;
}

.line-no {
	background: #f8fafc;
	color: #94a3b8;
	text-align: right;
	padding: 4px 8px;
	border-right: 1px solid #f1f5f9;
	user-select: none;
}

.line-code {
	white-space: pre;
	overflow-x: auto;
	padding: 4px 8px;
	color: #0f172a;
}

.line-code.left {
	border-right: 1px solid #e2e8f0;
}

.line-no.right {
	border-left: 1px solid #f1f5f9;
}

.diff-row.add .line-code.right,
.diff-row.add .line-no.right {
	background: #ecfdf3;
}

.diff-row.remove .line-code.left,
.diff-row.remove .line-no.left {
	background: #fef2f2;
}

.diff-row.change .line-code.left,
.diff-row.change .line-no.left {
	background: #fff1f2;
}

.diff-row.change .line-code.right,
.diff-row.change .line-no.right {
	background: #ecfdf3;
}

.json-card {
	display: flex;
	flex-direction: column;
	min-height: 0;
	border: 1px solid #e2e8f0;
	border-radius: 8px;
	overflow: hidden;
	background: #fff;
}

.json-title {
	height: 36px;
	display: flex;
	align-items: center;
	padding: 0 10px;
	font-size: 12px;
	font-weight: 700;
	color: #334155;
	border-bottom: 1px solid #f1f5f9;
}

pre {
	flex: 1;
	margin: 0;
	padding: 10px;
	overflow: auto;
	background: #f8fafc;
	color: #0f172a;
	font-size: 11px;
	line-height: 1.45;
}

.empty,
.empty-detail {
	padding: 20px;
	text-align: center;
	color: #94a3b8;
	font-size: 12px;
}
</style>
