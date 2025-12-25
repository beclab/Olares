<template>
	<div class="my-linechart2-container relative-position">
		<q-skeleton v-if="loading" style="height: 100%" />
		<div v-else class="full-height column no-wrap">
			<div
				v-if="$slots.extra"
				class="row justify-between align-center absolute-top z-top"
			>
				<div class="text-h6 text-ink-1">
					{{ title }}
				</div>
				<div class="text-body2 text-ink-2">
					<slot name="extra"></slot>
				</div>
			</div>
			<v-chart
				class="chart"
				:option="option"
				:updateOptions="updateOptions"
				:theme="theme_config"
				autoresize
			/>
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
	theme?: 'theme1' | 'theme2';
	lineWidth?: number;
	legendHide?: boolean;
	titleFormat?: string[];
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
import { provide, computed, useSlots } from 'vue';
import { theme as theme1, theme2 } from './theme';
import { capitalize, isArray, get } from 'lodash';
import { firstToUpperWith_ } from '@apps/control-panel-common/src/constant';
import { colors } from 'quasar';
import { useColor } from '@bytetrade/ui';
import { dateFormate, formatter, labelMaxWidth } from './utils';
import './tooltip.scss';

const { color: ink1 } = useColor('ink-1');
const { color: ink2 } = useColor('ink-2');
const { color: ink3 } = useColor('ink-3');
const { color: background2 } = useColor('background-2');
const { color: lightBlueDefault } = useColor('light-blue-default');
const { changeAlpha } = colors;
const slots = useSlots();

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
	theme: 'theme1',
	lineWidth: 3,
	legendHide: false,
	titleFormat: () => ['title', 'unit']
});
console.log('teataaa', props.theme);
const theme_config = computed(() => {
	switch (props.theme) {
		case 'theme1':
			return theme1;
		case 'theme2':
			return theme2;

		default:
			return theme1;
	}
});
provide(THEME_KEY, theme_config.value);

const unit = computed(() => props.data?.unit ?? '');

const title = computed(() =>
	!props.data?.title
		? ''
		: unit.value && props.titleFormat.includes('unit')
		? `${firstToUpperWith_(props.data.title)} (${unit.value})`
		: `${firstToUpperWith_(props.data.title)}`
);

const legend = isArray(props.data.legend)
	? props.data.legend.map(capitalize)
	: [];
const legendShow =
	isArray(props.data?.legend) &&
	props.data?.legend.length > 1 &&
	!props.legendHide;
const gridTop = title.value || legendShow ? 54 : 0;

const updateOptions = {
	notMerge: false
};

const option = computed(() => {
	const dataFlat = props.data.data.flat();
	const maxWidth = labelMaxWidth(dataFlat);

	const titleOPtions =
		title.value && !slots.extra
			? {
					text: title.value,
					left: 0,
					padding: 0,
					textStyle: {
						color: ink1.value,
						fontSize: 16,
						fontWeight: 700
					}
			  }
			: undefined;
	return {
		animationEasingUpdate: 'exponentialOut',
		title: titleOPtions,
		grid: {
			top: gridTop,
			left: maxWidth,
			right: 0,
			bottom: 32,
			containLabel: false
		},
		legend: {
			show: legendShow,
			left: 'right',
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
			valueFormatter: (value: any) =>
				`${isNaN(value) ? '-' : value} ${unit.value}`,
			formatter: (params: any, ticket: string) => formatter(params, unit.value),
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
				alignMaxLabel: 'right',
				showMinLabel: true,
				showMaxLabel: true,
				hideOverlap: true,
				formatter(value: string) {
					return dateFormate(value, 'HH:mm');
				}
			},
			data: get(props.data.data, '[0]', []).map((item) => item[0])
		},
		yAxis: {
			type: 'value',
			boundaryGap: false,
			splitNumber: props.splitNumberY ? props.splitNumberY : 5,
			onZero: true,
			splitLine: {
				show: false
			},
			axisLabel: {
				show: true,
				margin: maxWidth,
				color: ink3.value,
				align: 'left',
				verticalAlign: 'top',
				fontFamily: 'Roboto'
			}
		},
		// dataset: props.data,
		series: props.data.data.map((item, index) => ({
			type: 'line',
			name: legend[index],
			smooth: true,
			symbol: 'none',
			clip: false,
			lineStyle: {
				width: props.lineWidth
			},
			data: item.map((item) => item[1]),
			areaStyle: {
				color: new graphic.LinearGradient(0, 0, 0, 1, [
					{
						offset: 0,
						color: changeAlpha(theme_config.value.color[index], 0.2)
					},
					{
						offset: 1,
						color: changeAlpha(theme_config.value.color[index], 0)
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
		flex: 1;
	}
}
</style>
