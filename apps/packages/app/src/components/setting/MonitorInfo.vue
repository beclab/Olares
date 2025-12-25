<template>
	<div class="terminus-bind__system-info monitor-wrapper monitor-wrapper-gap">
		<div
			class="info-item column justify-top q-pt-md no-wrap"
			:style="`background:${
				$q.dark.isActive
					? monitorStore.dark_background[index]
					: monitorStore.background[index]
			}`"
			v-for="(item, index) in monitorStore.usages"
			:key="`d` + index"
		>
			<div>
				<q-img :src="iconsComputed(index)" width="24px" />
			</div>
			<div class="name text-subtitle2 text-ink-1">
				{{ item.name }}
			</div>
			<div class="text-overline text-ink-3 q-mt-xs">
				{{ $t('monitoring_used') }}:
				<span class="text-ink-2" v-if="index == 0"> {{ item.ratio }}% </span>
				<span class="text-ink-2" v-else>
					{{ item.usage.toFixed(1) }}
					{{ (item as any).unit ? $t((item as any).unit) : '' }}
				</span>
			</div>
			<div class="text-overline text-ink-3 q-mt-xs">
				<!-- <div v-if="index == 0">
					Temp:
					<span class="text-green"> 30°C </span>
				</div>
				<div v-else> -->
				{{ $t('monitoring_remain') }}:
				<span class="text-ink-2">
					{{ (item.total - item.usage).toFixed(1) }}
					<!-- {{ $t((item as any).unit) }} -->
					{{ (item as any).unit ? $t((item as any).unit) : '' }}
				</span>
				<!-- </div> -->
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useMonitorStore } from '../../stores/monitor';
import { getRequireImage } from '../../utils/imageUtils';
import { useQuasar } from 'quasar';
const $q = useQuasar();
const iconsComputed = (index: number) => {
	if ($q.dark.isActive) {
		return getRequireImage(`setting/${monitorStore.icons_dark[index]}`);
	} else {
		return getRequireImage(`setting/${monitorStore.icons[index]}`);
	}
};

const monitorStore = useMonitorStore();
</script>

<style scoped lang="scss">
.terminus-bind {
	width: 100%;
	padding: 0px 16px;

	&__system-info {
		width: 100%;

		.info-item {
			height: 108px;
			border-radius: 20px;
			border: 1px solid $separator;
			padding: 12px 8px 12px 12px;
			white-space: nowrap;

			.amount {
				text-align: left;
				margin-top: 12px;
			}

			.name {
				text-align: left;
				margin-top: 6px;
				// background-color: red;
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
		grid-gap: 12px;
	}
}
</style>
