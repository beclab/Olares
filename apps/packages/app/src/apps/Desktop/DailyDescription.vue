<template>
	<div v-if="widgetPrefsStore.showWeight" class="description_box">
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
		<div v-if="widgetPrefsStore.showDashboard" class="description_thickness">
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
import { useWidgetPreferencesStore } from 'src/stores/settings/widgetPreferences';
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue';
import { useMonitorStore } from 'src/stores/desktop/monitor';

const monitorStore = useMonitorStore();
const widgetPrefsStore = useWidgetPreferencesStore();
const watchTimeTask = ref();
const state = reactive({
	date: '',
	time: '',
	week: '',
	showIndex: 0,
	isAM: false,
	show: true
});

const updateDateTime = async () => {
	const { date, time, week, isAM } = widgetPrefsStore.formatNow();

	state.date = date;
	state.time = time;
	state.week = week;
	state.isAM = isAM;
	state.show = false;
	await nextTick();
	state.show = true;
};

const watchTime = () => {
	watchTimeTask.value = setInterval(() => {
		updateDateTime();
	}, 1000 * 1);
	updateDateTime();
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
