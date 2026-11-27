<template>
	<div class="row items-center">
		<div
			v-if="isDotLayout"
			class="pc-status-circle"
			:class="palette.indicator"
		/>
		<div class="pc-application-status" :class="labelClasses">
			{{ appStatusText }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { getApplicationStatus, getEntranceStatus } from 'src/constant';
import { APP_STATUS, ENTRANCE_STATUS } from 'src/constant/constants';
import { computed } from 'vue';
import type { PropType } from 'vue';

type StatusPalette = 'success' | 'danger' | 'pending';

const PALETTE: Record<
	StatusPalette,
	{ indicator: string; badgeText: string; badgeBg: string }
> = {
	success: {
		indicator: 'bg-positive',
		badgeText: 'text-positive',
		badgeBg: 'bg-green-alpha'
	},
	danger: {
		indicator: 'bg-negative',
		badgeText: 'text-negative',
		badgeBg: 'bg-red-alpha'
	},
	pending: {
		indicator: 'bg-warning',
		badgeText: 'text-warning',
		badgeBg: 'bg-yellow-alpha'
	}
};

function resolveStatusPalette(status: string): StatusPalette {
	switch (status) {
		case APP_STATUS.MODEL.INSTALLED:
		case APP_STATUS.RUNNING:
		case ENTRANCE_STATUS.RUNNING:
			return 'success';
		case APP_STATUS.PENDING.CANCEL_FAILED:
		case APP_STATUS.DOWNLOAD.CANCEL_FAILED:
		case APP_STATUS.INSTALL.CANCEL_FAILED:
		case APP_STATUS.DOWNLOAD.FAILED:
		case APP_STATUS.INSTALL.FAILED:
		case APP_STATUS.UNINSTALL.FAILED:
		case APP_STATUS.UPGRADE.FAILED:
		case APP_STATUS.RESUME.FAILED:
		case APP_STATUS.STOP.FAILED:
		case APP_STATUS.UNINSTALL.DEFAULT:
		case APP_STATUS.UNINSTALL.COMPLETED:
		case APP_STATUS.STOP.COMPLETED:
		case ENTRANCE_STATUS.NOT_READY:
		case ENTRANCE_STATUS.STOPPED:
			return 'danger';
		case APP_STATUS.PENDING.DEFAULT:
		case APP_STATUS.INSTALL.DEFAULT:
		case APP_STATUS.DOWNLOAD.DEFAULT:
		case APP_STATUS.INITIALIZE.DEFAULT:
		case APP_STATUS.RESUME.DEFAULT:
		case APP_STATUS.UPGRADE.DEFAULT:
		case APP_STATUS.PENDING.CANCELING:
		case APP_STATUS.DOWNLOAD.CANCELING:
		case APP_STATUS.STOP.DEFAULT:
		case APP_STATUS.INITIALIZE.CANCELING:
		case APP_STATUS.INSTALL.CANCELING:
		case APP_STATUS.UPGRADE.CANCELING:
		case APP_STATUS.RESUME.CANCELING:
			return 'pending';
		default:
			return 'danger';
	}
}

const props = defineProps({
	status: {
		type: String,
		required: true,
		default: ''
	},
	text: {
		type: String,
		required: false,
		default: ''
	},
	size: {
		type: Number,
		required: false,
		default: 32
	},
	uiStyle: {
		type: String as PropType<'circle' | 'square'>,
		required: false,
		default: 'circle'
	}
});

const palette = computed(() => PALETTE[resolveStatusPalette(props.status)]);

const isDotLayout = computed(() => props.uiStyle !== 'square');

const labelClasses = computed(() => {
	if (isDotLayout.value) {
		return ['text-caption', 'text-ink-3'];
	}
	return [
		'text-caption',
		'pc-status-square',
		palette.value.badgeText,
		palette.value.badgeBg
	];
});

const appStatusText = computed(() => {
	if (props.text) {
		return props.text;
	}
	let status = getApplicationStatus(props.status);
	if (!status) {
		status = getEntranceStatus(props.status as ENTRANCE_STATUS);
	}
	return status;
});
</script>

<style scoped lang="scss">
.pc-status-circle {
	width: 8px;
	height: 8px;
	margin-right: 4px;
	border-radius: 50%;
}

.pc-application-status {
	text-transform: capitalize;
}

.pc-status-square {
	padding: 2px 4px;
	border-radius: 4px;
}
</style>
