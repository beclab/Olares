<template>
	<CardInfo2 v-if="loading" />
	<div class="my-card-container" v-else>
		<div class="row items-center">
			<div class="row items-baseline">
				<count-up
					class="text-h5"
					:start-val="startValue"
					:end-val="changeFormat(value).value"
					:duration="1"
				></count-up>
				<span class="text-h5 text-ink-1">{{ changeFormat(value).unit }}</span>
			</div>
			<div class="q-ml-sm">
				<q-badge
					v-show="~~change !== 0"
					:class="[change > 0 ? 'up' : 'down']"
					class="q-px-md q-py-xs rounded-borders-lg"
				>
					<div class="row justify-center full-width Subtitle3">
						{{ change > 0 ? '+' : change < 0 ? '-' : '' }}
						<count-up
							:start-val="0"
							:end-val="changeFormat(change).value"
							:duration="1"
						></count-up>
						{{ changeFormat(change).unit }}
					</div>
				</q-badge>
			</div>
		</div>
		<div class="text-subtitle3 text-ink-2 q-mt-sm">
			<span>{{ label }}</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import CountUp from 'vue-countup-v3';
import { formatNumber } from './format';
import { isUndefined, get } from 'lodash';
import CardInfo2 from '@apps/dashboard/components/Skeleton/CardInfo2.vue';

interface Props {
	value: number;
	label: string;
	change?: number;
	unit?: string;
	format?: any;
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	change: 0
});

const startValue = ref(0);

const format = !isUndefined(props.format) ? props.format : formatNumber;
const changeFormat = (num: number) => {
	const data = format(num);
	return {
		value: get(data.match(/\d+/g), 0),
		unit: get(data.match(/[a-zA-Z]+/g), 0)
	};
};
watch(
	() => props.value,
	(newValue, oldValue) => {
		startValue.value = oldValue;
	}
);
</script>

<style lang="scss" scoped>
.my-card-container {
	min-width: 135px;

	.change.plusSign::before {
		content: '+';
	}
	.up {
		background: $green-soft;
		color: $positive;
	}
	.down {
		background: $red-soft;
		color: $negative;
	}
}
</style>
