<template>
	<div class="my-gauge-chart-container">
		<q-responsive :ratio="1.194" class="my-gauge-chart">
			<div class="gauge-chart-content">
				<div class="absolute-full">
					<v-chart
						class="chart"
						:option="option"
						:updateOptions="updateOptions"
						:theme="theme"
						autoresize
					/>
				</div>
			</div>
		</q-responsive>
		<div class="gauge-detail-title text-body3 text-ink-1 q-mt-sm" v-if="title">
			<div class="gauge-detail-title-wrapper">
				{{ title }}
			</div>
		</div>
	</div>
</template>
<script lang="ts">
export interface LineProps {
	xAxisLabel?: boolean;
	progressWidth?: number;
	detailLabelWrap?: boolean;
	gaugeMax?: number;
	colorStops?: number[];
	data: {
		title?: string;
		max?: number;
		unit?: string;
		percent?: boolean;
		data: Array<any>;
		legend?: string[];
	};
}
</script>
<script lang="ts" setup>
import { use } from 'echarts/core';
import {
	GridComponent,
	TitleComponent,
	DatasetComponent,
	LegendComponent,
	TooltipComponent
} from 'echarts/components';
import { GaugeChart } from 'echarts/charts';
import { UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import VChart, { THEME_KEY } from 'vue-echarts';
import { provide, computed } from 'vue';
import { theme } from './theme';
import { useColor } from '@bytetrade/ui';
const { color: ink3 } = useColor('ink-3');
const { color: ink1 } = useColor('ink-1');
const { color: background3 } = useColor('background-3');
const { color: negative } = useColor('negative');
const { color: warning } = useColor('warning');
const { color: positive } = useColor('positive');

use([
	GridComponent,
	GaugeChart,
	CanvasRenderer,
	UniversalTransition,
	TitleComponent,
	DatasetComponent,
	LegendComponent,
	TooltipComponent
]);

provide(THEME_KEY, theme);

const props = withDefaults(defineProps<LineProps>(), {
	xAxisLabel: true,
	unit: '',
	legend: [],
	gaugeMax: 100,
	colorStops: () => [0, 0.3, 0.8]
});

const unit = props.data?.unit ?? '%';

const title = props.data?.title;

const updateOptions = {
	notMerge: false
};

const dataValue = computed(() => props.data?.data[0][0]);

var seriesData = computed(() => [
	{
		name: title,
		value: dataValue.value
	}
]);

const activeColor = computed(() =>
	isNaN(dataValue.value) || Number(dataValue.value) === 0
		? 'rgba(0,0,0,0)'
		: Number(dataValue.value) > props.colorStops[2] * props.gaugeMax
		? negative.value
		: Number(dataValue.value) < props.colorStops[1] * props.gaugeMax
		? positive.value
		: warning.value
);

const option = computed(() => {
	const progressWidth = props?.progressWidth || 7;
	return {
		grid: {
			top: 0,
			bottom: 0,
			left: 0,
			right: 0
		},
		series: [
			{
				type: 'gauge',
				radius: '100%',
				startAngle: 220,
				endAngle: -40,
				min: 0,
				splitNumber: 14,
				max: props.gaugeMax,
				itemStyle: {
					color: activeColor.value
				},
				progress: {
					show: true,
					roundCap: true,
					width: progressWidth,
					clip: true
				},
				pointer: {
					show: false
				},
				axisLine: {
					roundCap: true,
					lineStyle: {
						width: progressWidth,
						color: [[1, background3.value]]
					}
				},
				axisTick: {
					distance: 5,
					length: 4,
					splitNumber: 1,
					lineStyle: {
						color: ink3.value
					}
				},
				title: {
					show: false
				},

				splitLine: {
					show: false
				},
				axisLabel: {
					show: false
				},
				detail: {
					formatter: (value) => {
						const wrap = props?.detailLabelWrap ? '\n' : '';
						return `${value}${wrap}${unit}`;
					},
					color: ink1.value,
					fontWeight: 700,
					fontSize: 16,
					fontFamily: 'Roboto',
					lineHeight: 24,
					offsetCenter: [0, 0]
				},
				data: seriesData.value
			}
		]
	};
});
</script>

<style lang="scss" scoped>
.my-gauge-chart-container {
	display: flex;
	flex-direction: column;
	justify-content: center;
	align-items: center;
	width: 100px;
	.my-gauge-chart {
		width: 100%;
		.chart {
			height: 100%;
			width: 100%;
		}
		.gauge-chart-content {
			position: relative;
			padding-bottom: 100%;
		}
	}
	.gauge-detail-title {
		padding: 4px 12px;
		gap: 10px;
		border-radius: var(--8, 8px);
		background: $background-3;
		white-space: nowrap;
	}
}
</style>
