<template>
	<div class="info-card-container">
		<div
			class="row items-center no-wrap justify-center info-card-wrapper"
			:class="[props.active ? 'info-card-wrapper-active' : '']"
			v-ripple
		>
			<q-knob
				:min="0"
				:max="_total"
				:title="name"
				v-model="_used"
				size="32px"
				:thickness="0.28"
				:color="props.active ? 'teal-13' : 'primary'"
				:track-color="props.active ? 'teal-12' : 'teal-14'"
				readonly
				class="info-circel-wrapper"
			/>
			<div>
				<div
					class="info-item"
					:class="[props.active ? 'info-item-active' : '']"
				>
					{{ _capitalize(name) }}&nbsp;
					<span>{{ _unit === 'core' ? $t('core') : _unit }}</span>
				</div>
				<div
					class="info-ratio"
					:class="[props.active ? 'info-ratio-active' : '']"
				>
					<span>{{ _used }}</span>
					<span class="ratio-foolter">/{{ _total }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	getSuitableUnit,
	getValueByUnit
} from '@apps/control-hub/src/utils/monitoring';
import { ref, toRefs } from 'vue';
import { _capitalize } from '@apps/control-hub/src/utils/index';

export interface InfoCardItemProps {
	used: string;
	total: string;
	name?: string;
	active?: boolean;
	unitType?: any;
	unit?: string;
}

const props = withDefaults(defineProps<InfoCardItemProps>(), {
	active: false,
	title: ''
});
const { total, used, unitType, unit } = toRefs(props);
const _unit =
	getSuitableUnit(total.value || used.value, unitType.value as any) ||
	unit?.value;
const _used = getValueByUnit(used.value, _unit as string);
const _total = getValueByUnit(total.value, _unit as string);
</script>

<style lang="scss" scoped>
.info-card-container {
	padding-bottom: 5px;
	width: 180px;

	cursor: default;
	.info-card-wrapper {
		position: relative;
		padding: 12px;
		&.info-card-wrapper-active {
			background: $primary;
			position: relative;
			border-radius: 4px;
			&::after {
				position: absolute;
				content: '';
				left: 50%;
				bottom: 0;
				border-left: 5px solid transparent;
				border-right: 5px solid transparent;
				border-top: 5px solid $primary;
				transform: translateY(100%);
			}
		}
		.info-circel-wrapper {
			margin-right: 12px;
		}
		.info-item {
			font-size: 12px;
			line-height: 12px;
			color: #7e7e7e;
			&.info-item-active {
				color: rgba(255, 255, 255, 0.8);
			}
		}
		.info-ratio {
			font-weight: 500;
			font-size: 16px;
			line-height: 16px;
			color: #414141;
			margin-top: 8px;
			&.info-ratio-active {
				color: #ffffff;
			}
			.ratio-foolter {
				font-weight: 400;
				font-size: 12px;
			}
		}
	}
}
</style>
