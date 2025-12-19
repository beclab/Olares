<template>
	<div class="column my-linechart2-container relative-position">
		<div class="row justify-between q-pb-md">
			<div
				v-for="(item, index) in legend"
				:key="item"
				@click="legendHandler(item)"
			>
				<div class="row items-center cursor-pointer">
					<div
						class="legend-symbol"
						:style="{ background: theme.color[index] }"
					></div>
					<span class="q-ml-sm text-body2 text-ink-2">{{ item }}</span>
					<div class="q-ml-md text-subtitle2 text-ink-1">
						<span>{{ lastValue(index) }}</span>
						<span>&nbsp;</span>
						<span>{{ unit[index] }}</span>
					</div>
				</div>
			</div>
		</div>
		<div style="flex: 1" class="full-width full-height">
			<q-skeleton v-if="loading" class="full-width full-height" />
			<v-chart
				v-else
				class="chart"
				:option="option"
				:updateOptions="updateOptions"
				:theme="theme"
				autoresize
				ref="chartRef"
			/>
			<q-resize-observer @resize="onResize" />
		</div>
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
import { provide, computed, ref } from 'vue';
import tinycolor from 'tinycolor2';
import { theme } from '@apps/control-panel-common/src/components/Charts/theme';
import { capitalize, get, indexOf, isArray, last } from 'lodash';
import { date } from 'quasar';
import { colors } from 'quasar';
import { getValue } from '@apps/control-panel-common/src/utils/yaml';
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

const chartInterval = ref<number | 'auto'>(2);
const chartRef = ref();

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
const legendShow = computed(() => legend.value.length > 0);

const updateOptions = {
	notMerge: false
};
const seriesData = computed(() => get(props.data, 'data', []));
const option = computed(() => {
	const dataLeft = get(props.data, 'data[0]', []);
	const dataRight = get(props.data, 'data[1]', []);
	const maxWidthLeft = labelMaxWidth(dataLeft);
	const maxWidthRight = labelMaxWidth(dataRight);
	return {
		title: {
			text: title.value,
			left: 0,
			padding: 0,
			textStyle: {
				color: ink1.value,
				fontSize: 16,
				fontWeight: 700
			}
		},
		grid: {
			top: 6,
			left: maxWidthLeft,
			right: maxWidthRight,
			bottom: 32,
			containLabel: false
		},
		legend: {
			show: false,
			left: 'left',
			right: 'right',
			padding: 0,
			icon: 'circle',
			itemWidth: 8,
			itemHeight: 8,
			textStyle: {
				color: ink2.value
			},
			formatter: function (name: string, data: any) {
				return 'Legend ' + name;
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
			backgroundColor: background2.value,
			padding: 12,
			textStyle: {
				color: ink1.value
			},
			borderWidth: 0,
			renderMode: 'html',
			className: 'echart-tooltip-container'
		},
		xAxis: {
			type: 'category',
			onZero: true,
			boundaryGap: false,
			axisLine: {
				show: false
			},
			axisTick: {
				show: false
			},
			axisLabel: {
				show: props.xAxisLabel,
				color: ink3.value,
				margin: 20,
				showMinLabel: true,
				showMaxLabel: true,
				hideOverlap: true,
				formatter(value: string) {
					return dateFormate(value, 'HH:mm');
				}
			},
			data: get(props.data.data, '[0]', []).map((item) => item[0])
		},
		yAxis: [
			{
				type: 'value',
				splitNumber: props.splitNumberY ? props.splitNumberY : 5,
				onZero: true,
				splitLine: {
					show: false
				},
				axisLabel: {
					show: true,
					margin: maxWidthLeft,
					color: ink3.value,
					align: 'left',
					verticalAlign: 'top',
					fontFamily: 'Roboto'
				}
			},
			{
				type: 'value',
				splitNumber: props.splitNumberY ? props.splitNumberY : 5,
				splitLine: {
					lineStyle: {
						type: 'dashed',
						color: separator.value
					}
				},
				axisLabel: {
					show: true,
					margin: maxWidthRight,
					color: ink3.value,
					align: 'right',
					verticalAlign: 'top',
					fontFamily: 'Roboto'
				}
			}
		],
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
			yAxisIndex: index,
			data: item.map((item) => item[1]),
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

const lastValue = (index: number) => {
	return get(last(get(props.data, `data[${index}]`)), '[1]');
};

const legendHandler = (name: string) => {
	chartRef.value.dispatchAction({
		type: 'legendToggleSelect',
		name: name
	});
};

const onResize = (size: { width: number; height: number }) => {
	chartInterval.value = chartEntervalOfWidth(size.width);
};
</script>

<style lang="scss" scoped>
.my-linechart2-container {
	height: 132px;
	.chart {
		height: 100%;
	}
	.legend-symbol {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}
}
</style>
<style lang="scss">
.echart-tooltip-container {
	border-radius: 12px !important;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2) !important;
}
</style>
