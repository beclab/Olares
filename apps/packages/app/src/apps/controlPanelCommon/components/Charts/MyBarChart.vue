<template>
	<v-chart
		class="chart"
		:option="option"
		:updateOptions="updateOptions"
		:theme="theme"
		autoresize
	>
	</v-chart>
	<Empty center v-show="!hasData"></Empty>
</template>
<script lang="ts">
import Empty from '../Empty.vue';
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
import { capitalize, get, indexOf, isArray } from 'lodash';
import { date } from 'quasar';
import { firstToUpper } from '@apps/control-panel-common/src/constant';

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
      ['03:34:40', 32],
      ['03:44:40', 20],
      ['03:54:40', 50],
      ['04:04:40', 20],
    ],
  ],
}

<MyBarChart :data="chartData" />
 */

const props = withDefaults(defineProps<LineProps>(), {
	xAxisLabel: true,
	unit: '',
	legend: []
});

const unit = computed(() => props.data?.unit ?? '');

const title = !props.data?.title
	? ''
	: unit.value
	? `${firstToUpper(props.data.title)} (${unit.value})`
	: `${firstToUpper(props.data.title)}`;
const legend = computed(() =>
	isArray(props.data?.legend) ? props.data.legend.map(capitalize) : []
);
const legendShow = isArray(props.data?.legend) && props.data?.legend.length > 1;
const gridTop = title || legendShow ? 40 : 6;
const gridBottom = props.xAxisLabel ? 0 : 4;

const updateOptions = {
	notMerge: false
};

const hasData = computed(() => {
	return get(props.data, 'data[0]', []).length > 0;
});

const option = computed(() => ({
	title: {
		text: title,
		textStyle: {
			color: '#414141',
			fontSize: 12,
			fontWeight: 'normal'
		}
	},
	grid: {
		top: gridTop,
		left: 0,
		right: 0,
		bottom: 0,
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
		title: title,
		valueFormatter: (value: any) =>
			`${isNaN(value) ? '-' : value} ${unit.value}`,
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
		axisLine: {
			show: true,
			lineStyle: {
				color: '#F1F1F1'
			}
		},
		axisTick: false,
		axisLabel: {
			show: props.xAxisLabel,
			color: '#BDBDBD'
		}
	},
	yAxis: {
		type: 'value',
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
	series: props.data?.data.map((item, index) => ({
		type: 'bar',
		name: legend.value[index],
		data: item,
		itemStyle: {
			borderRadius: [3, 3, 0, 0]
		},
		barMaxWidth: 32
	}))
}));
</script>

<style lang="scss" scoped>
.chart {
	height: 200px;
}
</style>
