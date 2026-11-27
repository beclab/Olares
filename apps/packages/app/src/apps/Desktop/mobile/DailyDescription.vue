<template>
	<div v-if="widgetPrefsStore.showWeight" class="description_box">
		<div class="clock_widget">
			<div class="clock_cluster">
				<div class="datetime_row">
					<span class="date">{{ state.date }}</span>
					<span class="week">{{ state.week }}</span>
				</div>
				<div class="time_row">
					<span class="time">{{ state.time }}</span>
					<span v-if="!widgetPrefsStore.is24HourFormat" class="meridiem">
						{{ state.isAM ? 'AM' : 'PM' }}
					</span>
				</div>
			</div>
			<div v-if="widgetPrefsStore.showDashboard" class="dashboard_row">
				<div
					class="usage_item"
					v-for="(item, index) in moniterStore.usages"
					:key="`d` + index"
				>
					<q-knob
						readonly
						v-model="item.ratio"
						size="32px"
						:thickness="0.5"
						:color="item.color"
						track-color="grey-7"
					/>
					<div class="usage_text">
						<p class="usage_label text-uppercase">{{ item.name }}</p>
						<p class="usage_value">{{ item.ratio }}%</p>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { useWidgetPreferencesStore } from 'src/stores/settings/widgetPreferences';
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue';
import { useMonitorStore } from 'src/stores/desktop/monitor';

const moniterStore = useMonitorStore();
const widgetPrefsStore = useWidgetPreferencesStore();
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

const updateDateTime = async () => {
	const { date, time, week, year, month, day, isAM } =
		widgetPrefsStore.formatNow();

	state.date = date;
	state.time = time;
	state.week = week;
	state.year = year;
	state.month = month;
	state.day = day;
	state.isAM = isAM;
	state.show = false;
	await nextTick();
	state.show = true;
};

const watchTime = () => {
	watchTimeTask.value = setInterval(() => {
		updateDateTime();
	}, 1000);
	updateDateTime();
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
	display: flex;
	justify-content: center;
	font-family: Inter, Roboto, sans-serif;
	.clock_widget {
		width: 291px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 20px;
		.clock_cluster {
			width: 100%;
			display: flex;
			flex-direction: column;
			align-items: center;
			.datetime_row {
				display: flex;
				align-items: flex-start;
				gap: 12px;
				.date,
				.week {
					font-size: 16px;
					font-weight: 500;
					line-height: 20px;
					color: rgba(255, 255, 255, 0.8);
					text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.1);
				}
			}
			.time_row {
				width: 100%;
				display: flex;
				align-items: flex-end;
				justify-content: center;
				gap: 4px;
				.time {
					font-size: 92px;
					font-weight: 700;
					line-height: 100px;
					color: rgba(255, 255, 255, 0.8);
					text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.1);
					font-feature-settings: 'tnum' 1, 'lnum' 1;
				}
				.meridiem {
					padding-bottom: 12px;
					font-size: 16px;
					font-weight: 500;
					line-height: 20px;
					color: rgba(255, 255, 255, 0.8);
					text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.1);
				}
			}
		}
		.dashboard_row {
			width: 100%;
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 0 8px;
			.usage_item {
				display: flex;
				align-items: center;
				gap: 9px;
				.usage_text {
					display: flex;
					flex-direction: column;
					p {
						margin: 0;
					}
					.usage_label {
						font-size: 12px;
						font-weight: 400;
						line-height: 20px;
						color: rgba(255, 255, 255, 0.8);
						text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.1);
					}
					.usage_value {
						font-size: 12px;
						font-weight: 700;
						line-height: 20px;
						color: rgba(255, 255, 255, 0.8);
						text-shadow: 0px 2px 6px rgba(0, 0, 0, 0.1);
					}
				}
			}
		}
	}
	:deep(.q-circular-progress__track) {
		color: rgba(255, 255, 255, 0.2) !important;
	}
}
</style>
