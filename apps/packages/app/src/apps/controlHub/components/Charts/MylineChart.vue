<template>
	<v-chart
		class="chart"
		:option="option"
		:updateOptions="updateOptions"
		:theme="theme"
		autoresize
	/>
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
import { capitalize, indexOf, isArray } from 'lodash';
import { date } from 'quasar';
import {
	firstToUpper,
	firstToUpperWith_
} from '@apps/control-hub/src/constant';

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

const props = withDefaults(defineProps<LineProps>(), {
	xAxisLabel: true,
	unit: '',
	legend: []
});

const unit = props.data?.unit ?? '';

const title = !props.data?.title
	? ''
	: unit
	? `${firstToUpperWith_(props.data.title)} (${unit})`
	: `${firstToUpperWith_(props.data.title)}`;
const legend = isArray(props.data.legend)
	? props.data.legend.map(capitalize)
	: [];
const legendShow = isArray(props.data?.legend) && props.data?.legend.length > 1;
const gridTop = title || legendShow ? 40 : 6;
const gridBottom = props.xAxisLabel ? 0 : 4;

const updateOptions = {
	notMerge: false
};

const option = computed(() => ({
	title: {
		text: title,
		left: -4,
		textStyle: {
			color: '#414141',
			fontSize: 12,
			fontWeight: 'normal'
		}
	},
	grid: {
		top: gridTop,
		left: 2,
		right: 0,
		bottom: gridBottom,
		containLabel: true
	},
	legend: {
		show: legendShow,
		left: 'right',
		icon: 'circle',
		itemWidth: 6,
		itemHeight: 6
	},
	tooltip: {
		trigger: 'axis',
		valueFormatter: (value: any) => `${isNaN(value) ? '-' : value} ${unit}`,
		axisPointer: {
			type: 'line',
			label: {
				color: '#414141',
				backgroundColor: '#eee'
			},
			crossStyle: {}
		}
	},
	xAxis: {
		type: 'category',
		boundaryGap: false,
		axisLine: {
			show: true,
			lineStyle: {
				color: '#F1F1F1'
			}
		},
		axisTick: false,
		axisLabel: {
			show: props.xAxisLabel,
			color: '#BDBDBD',
			// hideOverlap: true,
			margin: 9,
			// padding: [0, 42, 0, 0],
			borderColor: 'red'
		}
	},
	yAxis: {
		type: 'value',
		// min: 'dataMin',
		splitLine: {
			lineStyle: {
				type: 'dashed',
				color: '#F7F7F7'
			}
		},
		axisLabel: {
			show: true,
			color: '#BDBDBD'
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
			width: 1
		},
		data: item,
		areaStyle: {
			color: new graphic.LinearGradient(0, 0, 0, 1, [
				{
					offset: 0,
					color: tinycolor(theme.color[index]).setAlpha(0.19).toRgbString()
				},
				{
					offset: 1,
					color: tinycolor(theme.color[index]).setAlpha(0).toRgbString()
				}
			])
		}
	}))
}));
</script>

<style lang="scss" scoped>
.chart {
	height: 200px;
}
</style>
