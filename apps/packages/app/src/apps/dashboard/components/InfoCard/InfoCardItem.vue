<template>
	<div
		class="info-card-container"
		@mouseenter="() => (hovered = true)"
		@mouseleave="() => (hovered = false)"
	>
		<div class="info-card-wrapper q-pa-lg">
			<div class="row justify-between items-center">
				<div class="icon-wrapper row inline items-center justify-center">
					<q-img :src="props.img" width="20px" ratio="1" no-spinner />
				</div>
				<q-btn class="arrow-icon-wrapper" flat padding="0px">
					<q-img :src="arrowRightIcon2" width="20px" ratio="1" no-spinner />
				</q-btn>
			</div>
			<div class="text-subtitle2 text-ink-3 q-mt-lg">
				<q-skeleton v-if="loading" type="text" width="64px" />
				<template v-else>
					<span>{{ name }}&nbsp;</span>
					<span>{{ _unit === 'core' ? $t('core') : _unit }}</span>
				</template>
			</div>
			<slot v-if="$slots.network" name="network"></slot>
			<slot v-else-if="$slots.fan" name="fan"></slot>
			<template v-else>
				<div
					class="row items-center info-ratio text-h6 text-ink-1 value-wrapper q-mt-md"
				>
					<q-skeleton v-if="loading" type="text" width="88px" />
					<template v-else-if="!isNaN(_used) && !isNull(_total)">
						<span>{{ _used }}</span>
						<span>/</span>
						<span class="ratio-foolter">{{ _total || '-' }}</span>
						<q-icon v-if="info" name="sym_r_info" color="ink-3" class="q-ml-xs">
						</q-icon>
						<q-tooltip v-if="info" anchor="top middle" self="center right">{{
							info
						}}</q-tooltip>
					</template>
				</div>
				<div
					class="row items-center justify-between q-gutter-x-md info-footer"
					v-if="!isNaN(_used) && !isNull(_total)"
				>
					<span class="text-subtitle3" :class="textColorClass"
						>{{ percentCmp }}%</span
					>
					<q-linear-progress
						style="flex: 1"
						rounded
						size="4px"
						:value="ratio"
						:color="activeColor"
						:trackColor="hovered ? 'background-1' : 'background-3'"
						class="into-progress-wrapper"
					/>
				</div>
			</template>
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { computed, ref, toRefs } from 'vue';
import { isNaN, isNull, isNumber, round } from 'lodash';
import arrowRightIcon2 from '@apps/dashboard/assets/arrow-right2.svg';
import { RouteLocationRaw } from 'vue-router';
import { resourceStatusColor } from '@apps/dashboard/src/utils/status';

export interface ListItem {
	used?: string | number;
	total?: string | number;
	percent?: number;
	name: string;
	active?: boolean;
	unitType?: any;
	unit?: string;
	img?: string;
	img_active?: string;
	loading?: boolean;
	info?: string;
	id?: string;
}

export type InfoCardItemProps = ListItem & { route: RouteLocationRaw };
const hovered = ref(false);

const props = withDefaults(defineProps<ListItem>(), {
	active: false,
	title: ''
});

const { total, used, unitType, unit } = toRefs(props);
const _unit = computed(
	() =>
		getSuitableUnit(total.value || used.value, unitType.value as any) ||
		unit?.value
);
const _used = computed(() =>
	getValueByUnit(String(used.value), _unit.value, 2)
);
const _total = computed(() =>
	getValueByUnit(String(total.value), _unit.value, 2)
);

const percentCmp = computed(() => {
	return round(ratio.value * 100, 2);
});

const ratio = computed(() => {
	return isNumber(props.percent)
		? props.percent
		: _total.value
		? _used.value / _total.value
		: 0;
});

const activeColor = computed(() => resourceStatusColor(percentCmp.value));

const textColorClass = computed(() => `text-${activeColor.value}`);
</script>

<style lang="scss" scoped>
.info-card-container {
	cursor: default;
	.info-card-wrapper {
		position: relative;
		border-radius: 20px;
		border: 1px solid $separator;
		background: $background-1;
		&:hover {
			border-color: $light-blue-6;
			background: linear-gradient(
				125deg,
				$background-1 4.57%,
				$light-blue-soft 87.85%
			);
		}
		.icon-wrapper {
			width: 32px;
			height: 32px;
			border-radius: 8px;
			border: 1px solid $separator-2;
			background: $background-1;
		}
		.arrow-icon-wrapper {
			margin-right: -4px;
			border-radius: 6px;
		}
		.info-item {
			font-size: 12px;
			line-height: 12px;
			color: #5c5551;
			line-height: 16px;
			&.info-item-active {
				color: #ffffff;
			}
		}
		.value-wrapper {
			white-space: nowrap;
		}
	}
	.into-progress-wrapper {
		::v-deep(.q-linear-progress__track) {
			opacity: 1;
		}
	}
	.info-footer {
		margin-top: 16px;
	}
}
</style>
