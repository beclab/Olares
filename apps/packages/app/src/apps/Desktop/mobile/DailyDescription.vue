<template>
	<div class="description_box">
		<div class="description_weather">
			<div class="date">{{ state.month }}<span>/</span>{{ state.day }}</div>
			<div class="year_week">
				<div class="week q-mb-xs">
					{{ state.week }}
				</div>
				<div class="year">
					{{ state.year }}
				</div>
			</div>
		</div>
		<div class="description_thickness">
			<div
				class="description_track"
				v-for="(item, index) in moniterStore.usages"
				:key="`d` + index"
			>
				<q-knob
					readonly
					v-model="item.ratio"
					font-size="80px"
					size="32px"
					:thickness="0.5"
					:color="item.color"
					track-color="grey-4"
				/>
				<div class="description_track_txt">
					<p class="text-uppercase">{{ item.name }}</p>
					<p>{{ item.ratio }}%</p>
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue';

import { useMonitorStore } from '../../../stores/desktop/monitor';

const moniterStore = useMonitorStore();
const watchTimeTask = ref();
const state = reactive({
	date: '',
	time: '',
	week: '',
	year: '',
	month: '',
	day: '',
	showIndex: 0,
	isAM: false,
	show: true
});

const getTime = async () => {
	var myDate = new Date();
	let hour = myDate.getHours().toString().padStart(2, '0');
	let minutes = myDate.getMinutes().toString().padStart(2, '0');
	let year = myDate.getFullYear().toString();
	let month = (myDate.getMonth() + 1).toString().padStart(2, '0');
	let day = myDate.getDate().toString().padStart(2, '0');

	state.date = `${year}/${month}/${day}`;
	state.year = year;
	state.month = month;
	state.day = day;
	state.time = hour + ':' + minutes;
	state.isAM = myDate.getHours() <= 12;
	state.show = false;
	await nextTick();
	state.show = true;
};

const getWeekDate = () => {
	var now = new Date();
	var day = now.getDay();
	var weeks = [
		'Sunday',
		'Monday',
		'Tuesday',
		'Wednesday',
		'Thursday',
		'Friday',
		'Saturday'
	];
	state.week = weeks[day];
};

const watchTime = () => {
	watchTimeTask.value = setInterval(() => {
		getWeekDate();
		getTime();
	}, 1000 * 1);
	getWeekDate();
	getTime();
};

onMounted(() => {
	watchTime();
	moniterStore.loadMonitor();
});

onUnmounted(() => {
	clearInterval(watchTimeTask.value);
});
</script>

<style lang="scss" scoped>
.description_box {
	width: calc(100vw - 100px);
	position: absolute;
	top: 116px;
	left: 0;
	right: 0;
	margin: auto;
	.description_weather {
		width: 100%;
		height: 72px;
		display: flex;
		align-items: center;
		font-style: normal;
		font-weight: 700;
		line-height: normal;
		color: #ffffff;
		display: flex;
		align-items: center;
		justify-content: center;
		.date {
			display: inline-block;
			height: 100%;
			font-size: 64px;
			font-weight: 700;
			span {
				color: rgba(255, 255, 255, 0.5);
			}
		}
		.year_week {
			margin-left: 20px;
			display: inline-block;
			height: 100%;
			display: flex;
			align-items: start;
			justify-content: center;
			flex-direction: column;
			.year {
				font-size: 18px;
				line-height: 22px;
			}
			.week {
				font-size: 18px;
				font-weight: 500;
				line-height: 22px;
			}
		}
	}
	.description_thickness {
		width: 100%;
		height: 40px;
		display: flex;
		margin-top: 15px;
		justify-content: space-between;
		.q-circular-progress__track {
			color: rgba(255, 255, 255, 0.46) !important;
		}
		.description_track {
			display: flex;
			opacity: 0.8;
			p {
				margin: 0px;
			}
			.description_track_txt {
				height: 32px;
				font-size: 12px;
				font-family: Roboto-Regular, Roboto;
				font-weight: 400;
				color: #ffffff;
				line-height: 12px;
				text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.16);
				margin-left: 8px;
				display: flex;
				align-items: start;
				justify-content: space-around;
				flex-direction: column;
			}
		}
	}
}
</style>
