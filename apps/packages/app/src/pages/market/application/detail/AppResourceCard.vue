<template>
	<div class="app-resource-card bg-background-1">
		<div class="app-resource-card__header row items-center justify-between">
			<div class="text-subtitle2 text-ink-1 app-resource-card__title">
				{{ t('detail.resource') }}
			</div>

			<div
				v-if="showModeSelect"
				class="app-resource-card__select row items-center no-wrap"
				:class="{
					'app-resource-card__select--error': selectHasError,
					'app-resource-card__select--na': isNA
				}"
			>
				<q-icon
					v-if="selectHasError"
					name="sym_r_error"
					size="16px"
					color="negative"
				/>
				<div class="app-resource-card__select-label text-body3 ellipsis">
					{{ selectLabel }}
				</div>
				<q-icon name="sym_r_expand_more" size="16px" color="ink-3" />

				<q-menu
					v-if="modes.length > 0"
					ref="menuRef"
					anchor="bottom right"
					self="top right"
					class="app-resource-card__menu"
				>
					<div class="app-resource-card__menu-list column">
						<div
							v-for="item in modes"
							:key="item.mode"
							class="app-resource-card__menu-item row items-center justify-between no-wrap"
							:class="{
								'app-resource-card__menu-item--disabled':
									item.status === 'no-matching-node',
								'app-resource-card__menu-item--active':
									item.mode === selectedMode
							}"
							@click="onSelectMode(item)"
						>
							<div class="text-body3 ellipsis">{{ modeLabel(item.mode) }}</div>
							<q-icon
								v-if="item.mode === selectedMode"
								name="sym_r_check_circle"
								size="16px"
								color="blue-default"
							/>
						</div>
					</div>
				</q-menu>
			</div>
		</div>

		<div class="app-resource-card__grid">
			<div
				v-for="metric in visibleMetrics"
				:key="metric.key"
				class="app-resource-card__metric row items-center"
			>
				<div
					class="app-resource-card__metric-icon row items-center justify-center"
				>
					<q-icon :name="metric.icon" size="20px" color="ink-2" />
				</div>
				<div class="column">
					<div class="row items-center no-wrap app-resource-card__metric-label">
						<div class="text-body3 text-ink-3">{{ metric.label }}</div>
						<q-icon
							v-if="metric.error"
							name="sym_r_error"
							size="16px"
							color="negative"
						>
							<q-tooltip
								anchor="top middle"
								self="bottom middle"
								class="text-body3"
							>
								{{ metric.tip }}
							</q-tooltip>
						</q-icon>
					</div>
					<div class="text-subtitle2 text-ink-2">
						{{ metric.value }}
						<span v-if="metric.unit" class="text-body3">{{ metric.unit }}</span>
					</div>
				</div>
			</div>
		</div>

		<div v-if="showBanner" class="app-resource-card__status row items-start">
			<q-icon name="sym_r_error" size="20px" color="negative" />
			<div class="text-body2 text-negative">{{ bannerMessage }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { AppEntry } from '../../../../constant/constants';
import { computeModeTitle } from '../../../../constant/compute';
import { usePreCheckStore } from '../../../../stores/market/preCheck';
import {
	evaluateAllResourceModes,
	formatCpuForDisplay,
	formatMemoryForDisplay,
	hasAcceleratorRequirements
} from '../../../../utils/market/appResource';
import type {
	EvaluatedResourceMode,
	ResourceShortfall
} from '../../../../utils/market/appResource';
import { computed, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	appEntry: {
		type: Object as PropType<AppEntry>,
		required: true
	}
});

const { t } = useI18n();
const preCheckStore = usePreCheckStore();

const menuRef = ref();
const selectedMode = ref('');

const modes = computed<EvaluatedResourceMode[]>(() =>
	evaluateAllResourceModes(props.appEntry, {
		user: preCheckStore.userResource,
		cluster: preCheckStore.systemResource,
		computeNodes: preCheckStore.computeNodes
	})
);

// Legacy apps (no `accelerator` field) only have a single fallback mode, so the
// mode selector carries no choice — hide it and show the metrics + banner only.
const showModeSelect = computed(() =>
	hasAcceleratorRequirements(props.appEntry)
);

watch(
	modes,
	(next) => {
		const selectable = next.filter((m) => m.status !== 'no-matching-node');
		if (selectable.length === 0) {
			selectedMode.value = '';
			return;
		}
		if (!selectable.some((m) => m.mode === selectedMode.value)) {
			selectedMode.value = selectable[0].mode;
		}
	},
	{ immediate: true }
);

const currentMode = computed(() =>
	modes.value.find((item) => item.mode === selectedMode.value)
);

/** All declared modes have no supporting node on this cluster. */
const isNA = computed(() => modes.value.length > 0 && !currentMode.value);

const modeLabel = (mode: string) => computeModeTitle(mode);

const onSelectMode = (item: EvaluatedResourceMode) => {
	if (item.status === 'no-matching-node') return;
	selectedMode.value = item.mode;
	menuRef.value?.hide();
};

const selectLabel = computed(() =>
	isNA.value ? t('detail.na') : computeModeTitle(currentMode.value!.mode)
);

const selectHasError = computed(
	() => !isNA.value && currentMode.value?.status === 'insufficient-resources'
);

const SHORTFALL_KEY: Record<ResourceShortfall, string> = {
	'cluster-cpu': 'detail.shortfall_cluster_cpu',
	'cluster-memory': 'detail.shortfall_cluster_memory',
	'cluster-disk': 'detail.shortfall_cluster_disk',
	'cluster-gpu': 'detail.shortfall_cluster_gpu'
};

const shortfallMessage = (shortfall: ResourceShortfall) =>
	t(SHORTFALL_KEY[shortfall]);

const SHORTFALL_PRIORITY: ResourceShortfall[] = [
	'cluster-gpu',
	'cluster-cpu',
	'cluster-memory',
	'cluster-disk'
];

const primaryShortfall = computed<ResourceShortfall | undefined>(() => {
	const mode = currentMode.value;
	if (!mode) return undefined;
	return SHORTFALL_PRIORITY.find((s) => mode.shortfalls.includes(s));
});

type MetricDef = {
	key: string;
	icon: string;
	label: string;
	shortfall: ResourceShortfall;
	raw?: string;
	fmt: (raw: string | number | undefined | null) => {
		value: string;
		unit: string;
	};
};

const hasValue = (raw?: string) =>
	raw !== undefined && raw !== null && String(raw).trim() !== '';

const isCustomValue = (raw?: string | number | null) =>
	raw !== undefined && raw !== null && Number(raw) === -1;

const visibleMetrics = computed(() => {
	const mode = currentMode.value;

	const defs: MetricDef[] = [
		{
			key: 'cpu',
			icon: 'sym_r_memory',
			label: t('compute.cpu_label'),
			shortfall: 'cluster-cpu',
			raw: mode?.requiredCpu,
			fmt: formatCpuForDisplay
		},
		{
			key: 'memory',
			icon: 'sym_r_memory_alt',
			label: t('compute.memory_label'),
			shortfall: 'cluster-memory',
			raw: mode?.requiredMemory,
			fmt: formatMemoryForDisplay
		},
		{
			key: 'disk',
			icon: 'sym_r_hard_drive',
			label: t('detail.require_disk'),
			shortfall: 'cluster-disk',
			raw: mode?.requiredDisk,
			fmt: formatMemoryForDisplay
		},
		{
			key: 'gpu',
			icon: 'sym_r_developer_board',
			label: t('detail.gpu_memory'),
			shortfall: 'cluster-gpu',
			raw: mode?.requiredGPUMemory,
			fmt: formatMemoryForDisplay
		}
	];

	// All modes unsupported: show every metric as N/A.
	if (isNA.value || !mode) {
		return defs.map((d) => ({
			key: d.key,
			icon: d.icon,
			label: d.label,
			value: t('detail.na'),
			unit: '',
			error: false,
			tip: ''
		}));
	}

	// Otherwise only render the metrics the mode actually declares.
	return defs
		.filter((d) => hasValue(d.raw))
		.map((d) => {
			// Custom (-1): no fixed amount, never a shortfall.
			if (isCustomValue(d.raw)) {
				return {
					key: d.key,
					icon: d.icon,
					label: d.label,
					value: t('detail.custom'),
					unit: '',
					error: false,
					tip: ''
				};
			}
			const { value, unit } = d.fmt(d.raw);
			const error = mode.shortfalls.includes(d.shortfall);
			return {
				key: d.key,
				icon: d.icon,
				label: d.label,
				value,
				unit,
				error,
				tip: error ? shortfallMessage(d.shortfall) : ''
			};
		});
});

const showBanner = computed(
	() => isNA.value || currentMode.value?.status === 'insufficient-resources'
);

const bannerMessage = computed(() => {
	if (isNA.value) {
		return t('detail.no_matching_accelerator_banner');
	}
	const shortfall = primaryShortfall.value;
	return t('detail.resource_insufficient_banner', {
		reason: shortfall ? shortfallMessage(shortfall) : ''
	});
});
</script>

<style scoped lang="scss">
.app-resource-card {
	border-radius: 12px;
	padding: 20px;
	width: 100%;
	display: flex;
	flex-direction: column;
	gap: 20px;

	&__header {
		gap: 12px;
	}

	&__title {
		flex: 1;
		min-width: 0;
	}

	&__select {
		gap: 4px;
		width: 120px;
		padding: 8px 8px 8px 12px;
		border: 1px solid $input-stroke;
		border-radius: 8px;
		cursor: pointer;
		flex-shrink: 0;

		&-label {
			flex: 1;
			min-width: 0;
			color: $ink-1;
		}

		&--error {
			.app-resource-card__select-label {
				color: $ink-1;
			}
		}

		// All modes unmatched: mirror the disabled menu-item look on the trigger
		// (greyed label) so the control reads as not actionable, while the menu
		// stays openable to inspect the unmatched modes.
		&--na {
			.app-resource-card__select-label {
				color: $ink-3;
			}
		}
	}

	&__menu-list {
		min-width: 120px;
		padding: 8px;
		gap: 4px;
	}

	&__menu-item {
		padding: 8px;
		border-radius: 4px;
		gap: 4px;
		cursor: pointer;
		color: $ink-2;

		&:hover:not(&--disabled) {
			background: $background-hover;
		}

		&--active {
			color: $blue-default;
		}

		&--disabled {
			color: $ink-3;
			cursor: not-allowed;
		}
	}

	&__grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 12px;
	}

	&__metric {
		gap: 12px;
		padding: 12px;
		border-radius: 8px;
		background: $background-6;
		min-width: 0;
	}

	&__metric-icon {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: $blue-alpha;
		flex-shrink: 0;
	}

	&__metric-label {
		gap: 4px;
	}

	&__status {
		gap: 8px;
		padding: 12px;
		border-radius: 12px;
		background: $red-soft;
	}
}
</style>
