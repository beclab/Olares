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
import { LineChart, BarChart } from 'echarts/charts';
import { UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import VChart, { THEME_KEY } from 'vue-echarts';
import { provide, computed } from 'vue';
import tinycolor from 'tinycolor2';
import { theme } from './theme';
import { capitalize, get, indexOf, isArray } from 'lodash';
import { date } from 'quasar';

import { colors } from 'quasar';
import { useColor } from '@bytetrade/ui';
import { formatter } from './utils';
import './tooltip.scss';
const { color: ink1 } = useColor('ink-1');
const { color: ink2 } = useColor('ink-2');
const { color: ink3 } = useColor('ink-3');
const { color: separator } = useColor('separator');
const { color: separator2 } = useColor('separator-2');
const { color: background2 } = useColor('background-2');
const { color: lightBlueDefault } = useColor('light-blue-default');
const { getPaletteColor, changeAlpha } = colors;

use([
	GridComponent,
	LineChart,
	BarChart,
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
	legend: []
});

const unit = computed(() => props.data?.unit ?? '');

const title = computed(() =>
	!props.data?.title
		? ''
		: unit.value
		? `${props.data.title} (${unit.value})`
		: `${props.data.title}`
);

const legend = computed(() =>
	isArray(props.data.legend) ? props.data.legend.map(capitalize) : []
);

const legendShow = isArray(props.data?.legend) && props.data?.legend.length > 1;
const gridTop = title.value || legendShow ? 48 : 6;
const gridBottom = props.xAxisLabel ? 0 : 4;

const updateOptions = {
	notMerge: false
};

const seriesData = computed(() => get(props.data, 'data', []));

const option = computed(() => {
	const dataFlat = props.data.data
		.flat()
		.map((item) => item[1].toString().replace('.', '').length);
	const dataMax = Math.max(...dataFlat);
	const yAxisLabelMargin = (dataMax.toString().length + 1) * 7 + 16;
	return {
		title: {
			text: title.value,
			left: -6,
			textStyle: {
				color: ink1.value,
				fontSize: 12,
				fontWeight: 'normal'
			}
		},
		grid: {
			top: gridTop,
			left: yAxisLabelMargin,
			right: 2,
			bottom: 32,
			containLabel: false
		},
		legend: {
			show: legendShow,
			left: 'left',
			padding: 0,
			icon: 'circle',
			itemWidth: 8,
			itemHeight: 8,
			textStyle: {
				color: ink2.value
			}
		},
		tooltip: {
			trigger: 'axis',
			formatter: (params: any, ticket: string) => {
				const data = params.map((item: any) => ({
					marker: item.marker,
					seriesName: item.seriesName,
					data: seriesData.value[item.seriesIndex][item.dataIndex][1],
					unit: unit.value[item.seriesIndex],
					axisValueLabel: params[0].axisValueLabel
				}));
				return formatter(data, unit.value);
			},
			axisPointer: {
				type: 'line',
				lineStyle: {
					color: lightBlueDefault.value
				}
			},
			padding: 12,
			backgroundColor: background2.value,
			textStyle: {
				color: ink1.value
			},
			borderWidth: 0,
			renderMode: 'html',
			className: 'echart-tooltip-container'
		},
		xAxis: {
			type: 'category',
			// onZero: false,
			axisLine: {
				show: true,
				lineStyle: {
					color: separator2.value
				}
			},
			axisTick: {
				show: false
			},
			axisLabel: {
				show: props.xAxisLabel,
				color: ink3.value,
				margin: 20
			}
		},
		yAxis: {
			type: 'value',
			// min: 'dataMin',
			splitNumber: props.splitNumberY ? props.splitNumberY : 5,
			// onZero: false,
			splitLine: {
				lineStyle: {
					type: 'dashed',
					color: separator.value
				}
			},
			axisLabel: {
				show: true,
				margin: yAxisLabelMargin,
				color: ink3.value,
				align: 'left'
			}
		},
		// dataset: props.data,
		series: props.data.data.map((item, index) => ({
			type: 'line',
			name: legend.value[index],
			smooth: true,
			symbol: 'none',
			clip: false,
			lineStyle: {
				width: 3
			},
			data: item,
			areaStyle: {
				color: new graphic.LinearGradient(0, 0, 0, 1, [
					{
						offset: 0,
						color: changeAlpha(theme.color[index], 0.2)
					},
					{
						offset: 1,
						color: changeAlpha(theme.color[index], 0)
					}
				])
			}
		}))
	};
});
</script>

<style lang="scss" scoped>
.my-linechart2-container {
	height: 132px;
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
