<template>
	<div
		class="resource-limit-root row justify-between"
		:class="!deviceStore.isMobile ? 'resource-border' : 'resource-background'"
	>
		<div class="column justify-between">
			<div
				class="resource-label"
				:class="deviceStore.isMobile ? 'text-subtitle1' : 'text-subtitle1-m'"
			>
				{{ label }}
			</div>
			<div
				class="text-h5 resource-value"
				:class="deviceStore.isMobile ? 'text-h5' : 'text-h5-m'"
			>
				{{ usageFormat }}/{{ totalFormat }}
			</div>
		</div>
		<q-knob
			readonly
			v-model="percent"
			show-value
			size="56px"
			:thickness="0.4"
			:color="progressColor"
			track-color="background-3"
			class="text-subtitle3 knob-class"
			:style="{ '--textColor': textColor }"
		>
			{{ percent }}%
		</q-knob>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { getValueByUnit, getSuitableUnit } from 'src/utils/settings/monitoring';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps({
	total: Number,
	usage: Number,
	label: String,
	unitKey: String
});

const deviceStore = useDeviceStore();

const numTotal = computed(() => Number(props.total) || 0);
const numUsage = computed(() => Number(props.usage) || 0);

const totalFormat = computed(() => {
	if (props.unitKey == 'cpu') {
		if (!numTotal.value) {
			return 0;
		}
		return Number(numTotal.value.toFixed(2));
	}
	const _unit = getSuitableUnit(
		numTotal.value || numUsage.value,
		props.unitKey as any
	);
	return getValueByUnit(`${props.total}`, _unit);
});

const usageFormat = computed(() => {
	if (props.unitKey == 'cpu') {
		if (!numUsage.value) {
			return 0;
		}
		return Number(numUsage.value.toFixed(2));
	}
	const _unit = getSuitableUnit(
		numTotal.value || numUsage.value,
		props.unitKey as any
	);
	return getValueByUnit(`${props.usage}`, _unit);
});

const percent = computed(() => {
	if (!numTotal.value || !numUsage.value) {
		return 0;
	}
	return Number(((numUsage.value / numTotal.value) * 100).toFixed(0));
});

const progressColor = computed(() => {
	if (percent.value <= 50) {
		return 'color-knob-green';
	} else if (percent.value <= 80) {
		return 'color-knob-yellow';
	} else {
		return 'color-knob-red';
	}
});

const textColor = computed(() => {
	if (percent.value <= 50) {
		return '#29CC5F';
	} else if (percent.value <= 80) {
		return '#FEBE01';
	} else {
		return '#FF4D4D';
	}
});
</script>

<style scoped lang="scss">
.resource-limit-root {
	width: 100%;
	border-radius: 12px;
	padding: 20px;
	height: 96px;

	.resource-label {
		color: $grey;
	}

	.resource-value {
		margin-top: 4px;
		color: $ink-1;
	}

	.knob-class {
		color: $ink-2;
	}
}

.resource-border {
	border: 1px solid $separator;
}

.resource-background {
	background: $background-2;
}
</style>
