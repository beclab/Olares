<template>
	<div class="my-linechart2-container relative-position">
		<q-skeleton v-if="loading" style="height: 100%" />
		<v-chart
			v-else
			class="chart"
			:option="option"
			:updateOptions="updateOptions"
			:theme="theme"
			autoresize
		/>
	</div>
</template>
<script lang="ts">
export interface LineProps {
	xAxisLabel?: boolean;
	data: {
		title?: string;
		unit?: string;
		percent?: boolean;
		data: Array<any>;
		legend?: string[];
	};
	/**
	 * yAxis splitNumber
	 */
	splitNumberY?: number;
	loading?: boolean;
	colorStops?: number[];
}
</script>
<script lang="ts" setup>
import { use, graphic } from 'echarts/core';
import {
	GridComponent,
	TitleComponent,
	DatasetComponent,
	LegendComponent,
	TooltipComponent
} from 'echarts/components';
import { LineChart, BarChart, PictorialBarChart } from 'echarts/charts';
import { UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import VChart, { THEME_KEY } from 'vue-echarts';
import { provide, computed, ref } from 'vue';
import { theme } from './theme';
import { capitalize, isArray, get } from 'lodash';
import { colors } from 'quasar';
import {
	chartEntervalOfWidth,
	dateFormate,
	formatter,
	labelMaxWidth
} from './utils';
import './tooltip.scss';
import { useColor } from '@bytetrade/ui';

const { color: ink1 } = useColor('ink-1');
const { color: ink2 } = useColor('ink-2');
const { color: ink3 } = useColor('ink-3');
const { color: background1 } = useColor('background-1');
const { color: background2 } = useColor('background-2');
const { color: background3 } = useColor('background-3');
const { color: lightBlueDefault } = useColor('light-blue-default');

const { changeAlpha } = colors;

use([
	GridComponent,
	LineChart,
	BarChart,
	PictorialBarChart,
	CanvasRenderer,
	UniversalTransition,
	TitleComponent,
	DatasetComponent,
	LegendComponent,
	TooltipComponent
]);

provide(THEME_KEY, theme);

/**
 const chartData = {
  title: 'CPU_USAGE',
  unit: 'm',
  legend: ['USAGE'],
  data: [
    [
      ['03:34:40', 0],
      ['03:44:40', 0],
      ['03:54:40', 0],
      ['04:04:40', 0],
    ],
  ],
}

<MyLineChart :data="chartData" />
 */

const props = withDefaults(defineProps<LineProps>(), {
	xAxisLabel: true,
	unit: '',
	legend: [],
	colorStops: () => [0, 0.6, 0.8]
});
const chartInterval = ref<number | 'auto'>(2);

const unit = computed(() => props.data?.unit ?? '');

const title = computed(() =>
	!props.data?.title
		? ''
		: unit.value
		? `${props.data.title} (${unit.value})`
		: `${props.data.title}`
);

const legend = isArray(props.data.legend)
	? props.data.legend.map(capitalize)
	: [];
const legendShow = isArray(props.data?.legend) && props.data?.legend.length > 1;
const gridTop = title.value || legendShow ? 54 : 0;

const updateOptions = {
	notMerge: false
};

const seriesData = computed(() => get(props.data, 'data', []));

const linearColors = ['#29CC5F', '#FEBE01', '#FF4D4D'];
const itemStyleColorStops = computed(() => {
	return props.colorStops.map((item, index) => ({
		offset: item,
		color: linearColors[index]
	}));
});

const option = computed(() => {
	var data = legend;
	var value = props.data?.data[0] || [0];
	var total = [1];
	const pictorialBarSymbolMargin = 11;
	return {
		grid: {
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
			containLabel: false
		},
		tooltip: {
			show: false,
			axisPointer: {
				type: 'line',
				lineStyle: {
					color: lightBlueDefault.value
				}
			},
			backgroundColor: background2.value,
			textStyle: {
				color: ink1.value
			},
			borderWidth: 0,
			renderMode: 'html',
			className: 'echart-tooltip-container'
		},
		xAxis: {
			type: 'value',
			min: 0,
			max: 1,
			axisLine: { show: false },
			splitLine: { show: false },
			axisLabel: { show: false },
			axisTick: { show: false }
		},
		yAxis: {
			//show: false,
			type: 'category',
			inverse: true,
			splitLine: { show: false },
			axisLine: { show: false },
			axisLabel: {
				show: true,
				interval: 0,
				margin: 10,
				textStyle: {
					color: background1.value,
					fontSize: 16,
					fontWeight: 'bold'
				}
			},
			axisTick: { show: false },
			data: data
		},
		series: [
			{
				type: 'bar',
				barWidth: '100%',
				itemStyle: {
					normal: {
						borderWidth: 0,
						color: {
							type: 'linear',
							x: 0,
							y: 0,
							x2: value[0] ? 1 / value[0] : 1,
							y2: 0,
							colorStops: itemStyleColorStops.value
						}
					}
				},
				label: {
					show: false
				},
				data: value,
				z: 1
			},
			{
				type: 'bar',
				barWidth: '100%',
				barGap: '-100%',
				silent: true,
				animation: false,
				itemStyle: {
					normal: {
						borderWidth: 0,
						color: background3.value
					}
				},
				data: total,
				z: 0
			},
			{
				type: 'pictorialBar',
				barWidth: '100%',
				symbol: 'rect',
				symbolMargin: pictorialBarSymbolMargin,
				symbolSize: [1, '100%'],
				symbolRepeat: true,
				animation: false,
				itemStyle: {
					normal: {
						color: background1.value
					}
				},
				data: total,
				z: 2
			}
		]
	};
});
</script>

<style lang="scss" scoped>
.my-linechart2-container {
	height: 8px;
	.chart {
		height: 100%;
	}
}
</style>
<style lang="scss">
.echart-tooltip-container {
	border-radius: 12px !important;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2) !important;
}
</style>
