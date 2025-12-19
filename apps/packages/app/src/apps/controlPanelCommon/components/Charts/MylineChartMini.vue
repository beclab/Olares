<template>
	<div class="mylinechartmini-container">
		<span>{{ lastValue }}</span>
		<span class="subtitle">
			{{ unit }}

			{{ _capitalize(props.data.title) }}
		</span>
	</div>
	<v-chart
		class="chart"
		:option="option"
		:updateOptions="updateOptions"
		:theme="theme"
		autoresize
	/>
</template>

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
import { capitalize, get, isArray, last } from 'lodash';
import { _capitalize } from '@apps/control-panel-common/src/utils/index';

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

interface Props {
	data: {
		title?: string;
		unit?: string;
		percent?: boolean;
		data: Array<[string, number]>;
		legend?: string[];
	};
}
const props = withDefaults(defineProps<Props>(), {});

const newData = computed(() =>
	props.data.data.map((item) =>
		item.map((child: any) => ({
			name: child[0],
			value: child[1]
		}))
	)
);

const xKey = 'name';
const series = Object.keys(newData.value[0][0] || {}).filter(
	(key) => key !== xKey
);

const lastValue = computed(() => {
	const key = get(series, [0], '');
	// eslint-disable-next-line @typescript-eslint/ban-ts-comment
	// @ts-ignore
	return last(newData.value[0])?.[key] ?? 0;
});

const unit = props.data.unit ?? '';
const legend = isArray(props.data.legend)
	? props.data.legend.map(capitalize)
	: [];

const updateOptions = {
	notMerge: false
};

const option = computed(() => ({
	title: {
		show: false
	},
	grid: {
		top: 2,
		left: 1,
		right: 1,
		bottom: 1,
		containLabel: true
	},
	tooltip: {
		trigger: 'axis',
		valueFormatter: (value: number) => `${value} ${unit}`,
		axisPointer: {
			type: 'line',
			label: {
				color: '#414141',
				backgroundColor: '#eee',
				precision: 2,
				formatter: (params: any) => params.seriesData[0].name
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
			show: false
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
			show: false
		}
	},
	series: newData.value.map((item, index) => ({
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
	height: 70px;
}
.mylinechartmini-container {
	margin-bottom: 16px;
	font-weight: 500;
	font-size: 16px;
	line-height: 16px;
	color: #414141;
	.subtitle {
		font-weight: 400;
		font-size: 12px;
		color: #7e7e7e;
		margin-left: 8px;
	}
}
</style>
