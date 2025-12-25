<template>
	<adaptive-layout auto>
		<template v-slot:mobile>
			<div
				class="mobile-status-circle row items-center justify-center"
				:style="{
					'--top': size - 6 + 'px',
					'--left': size - 6 + 'px'
				}"
			>
				<div class="mobile-status-content" :class="backgroundClass" />
			</div>
		</template>
		<template v-slot:pc>
			<div class="row items-center">
				<div class="pc-status-circle" :class="backgroundClass" />
				<div class="text-body2 pc-application-status">
					{{ appStatusText }}
				</div>
			</div>
		</template>
	</adaptive-layout>
</template>

<script setup lang="ts">
import { getApplicationStatus, getEntranceStatus } from 'src/constant';
import { APP_STATUS, ENTRANCE_STATUS } from 'src/constant/constants';
import AdaptiveLayout from '../AdaptiveLayout.vue';
import { computed } from 'vue';

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
	}
});

const backgroundClass = computed(() => {
	switch (props.status) {
		case APP_STATUS.MODEL.INSTALLED:
		case APP_STATUS.RUNNING:
		case ENTRANCE_STATUS.RUNNING:
			return 'bg-green-default';
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
			return 'bg-red-default';
		case APP_STATUS.PENDING.DEFAULT:
		case APP_STATUS.INSTALL.DEFAULT:
		case APP_STATUS.DOWNLOAD.DEFAULT:
		case APP_STATUS.INITIALIZE.DEFAULT:
		case APP_STATUS.RESUME.DEFAULT:
		case APP_STATUS.UPGRADE.DEFAULT:
		case APP_STATUS.PENDING.CANCELING:
		case APP_STATUS.DOWNLOAD.CANCELING:
		case APP_STATUS.INITIALIZE.CANCELING:
		case APP_STATUS.INSTALL.CANCELING:
		case APP_STATUS.UPGRADE.CANCELING:
		case APP_STATUS.RESUME.CANCELING:
			return 'bg-yellow-default';
		default:
			return 'bg-red-default';
	}
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
	width: 12px;
	height: 12px;
	margin-right: 6px;
	border-radius: 50%;
}
.pc-application-status {
	text-align: right;
	color: $ink-2;
	text-transform: capitalize;
	margin-right: 4px;
}

.mobile-status-circle {
	position: absolute;
	left: var(--left);
	top: var(--top);
	width: 8px;
	height: 8px;
	// padding: 2px;
	border-radius: 4px;
	background-color: $background-1;

	.mobile-status-content {
		width: 6px;
		height: 6px;
		border-radius: 3px;
	}
}
</style>
