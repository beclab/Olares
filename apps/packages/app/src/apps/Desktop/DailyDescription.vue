<template>
	<div class="description_box">
		<div class="description_weather">
			<div class="description_time">{{ state.time }}</div>
			<div class="description_daily">
				<div class="description_singapore">
					<p class="description_week">{{ state.week }}</p>
					<p class="description_day">
						{{ state.date }}
					</p>
				</div>
			</div>
		</div>
		<div class="description_thickness">
			<div
				class="description_track"
				v-for="(item, index) in monitorStore.usages"
				:key="`d` + index"
			>
				<q-knob
					readonly
					v-model="item.ratio"
					font-size="80px"
					size="24px"
					:thickness="0.5"
					:color="item.color"
					track-color="grey-4"
				></q-knob>
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

import { useMonitorStore } from '../../stores/desktop/monitor';

const monitorStore = useMonitorStore();
const watchTimeTask = ref();
const state = reactive({
	date: '',
	time: '',
	week: '',
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
	monitorStore.loadMonitor();
});

onUnmounted(() => {
	clearInterval(watchTimeTask.value);
});
</script>

<style lang="scss">
.description_box {
	position: absolute;
	bottom: 122px;
	right: 165px;
	.description_weather {
		height: 72px;
		display: flex;
		.description_time {
			font-size: 70px;
			font-family: Roboto-Bold, Roboto;
			font-weight: bold;
			color: #ffffff;
			line-height: 72px;
			text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.16);
		}
		.description_daily {
			display: flex;
			padding-top: 14px;
			margin-left: 14px;
			p {
				margin: 0;
			}
			.description_singapore {
				.description_week {
					font-size: 20px;
					font-family: Roboto-Bold, Roboto;
					font-weight: bold;
					color: #ffffff;
					text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.16);
				}
				.description_day {
					font-size: 12px;
					font-family: Roboto-Regular, Roboto;
					font-weight: 400;
					color: #ffffff;
					text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.16);
				}
			}
		}
	}
	.description_thickness {
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
				font-size: 12px;
				font-family: Roboto-Regular, Roboto;
				font-weight: 400;
				color: #ffffff;
				line-height: 12px;
				text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.16);
				margin-left: 8px;
			}
		}
	}
}
</style>
