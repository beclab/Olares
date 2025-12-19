<template>
	<div class="terminus-bind__system-info monitor-wrapper monitor-wrapper-gap">
		<div
			class="info-item q-pt-md no-wrap"
			v-for="(item, index) in monitorStore.usages"
			:key="`d` + index"
		>
			<div class="row items-center justify-between">
				<div class="text-overline text-ink-2">
					{{ item.name }}
				</div>
				<div class="text-caption text-ink-3">{{ item.ratio }}%</div>
			</div>
			<q-linear-progress
				rounded
				class="q-mt-sm"
				size="4px"
				:value="item.ratio / 100"
				:color="item.color"
				track-color="separator"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useMonitorStore } from '../../stores/monitor';
const monitorStore = useMonitorStore();
</script>

<style scoped lang="scss">
.terminus-bind {
	width: 100%;
	padding: 0px 16px;

	&__system-info {
		width: 100%;

		.info-item {
			white-space: nowrap;
			.name {
				text-align: left;
				margin-top: 6px;
			}

			.amount {
				text-align: right;
				margin-top: 12px;
			}
		}
	}
}

.app-version {
	text-align: right;
}

.monitor-wrapper {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	grid-gap: 20px;
}
@media (width < 400px) {
	.monitor-wrapper-gap {
		grid-gap: 8px;
	}
}
@media (400px <= width <= 430px) {
	.monitor-wrapper-gap {
		grid-gap: 20px;
	}
}
</style>
