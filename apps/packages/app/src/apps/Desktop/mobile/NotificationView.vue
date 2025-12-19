<template>
	<!-- <q-page> -->
	<div class="Launch_pad_page in-center-page">
		<div class="launch_pad_box in-center" ref="launchpadPage">
			<div class="q-mt-xl notification-title text-white q-mb-lg">
				{{ $t('permission.notifications') }}
			</div>
			<NotificationPopup
				v-if="notificationStore.data && notificationStore.data.length > 0"
			/>
			<div
				v-else
				class="full-width row items-center justify-center notification-empty"
			>
				{{ $t('There is currently no news') }}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import NotificationPopup from '../NotificationMobilePopup.vue';
import { useNotificationStore } from '../../../stores/desktop/notification';
import { T } from 'ts-toolbelt';
const notificationStore = useNotificationStore();
</script>

<style lang="scss" scoped>
.drag-enter {
	border: 1px dashed white;
}

.drag-start {
	opacity: 0;
}

.Launch_pad_page {
	width: 100%;
	height: 100%;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(10px);
	z-index: 6;
	position: absolute;
	top: 0px;
	left: 0px;
	touch-action: none;
}

.launch_pad_box {
	width: 100%;
	height: 100%;

	.notification-title {
		font-family: Inter;
		font-weight: 700;
		font-style: Bold;
		font-size: 18px;
		line-height: 16px;
		letter-spacing: 0%;
		vertical-align: middle;
		padding: 0 30px;
	}

	.notification-empty {
		font-family: Inter;
		font-weight: 400;
		font-style: Regular;
		font-size: 18px;
		line-height: 100%;
		letter-spacing: 0%;
		text-align: center;
		vertical-align: middle;
		color: #ffffff;
		margin-top: 300px;
	}
}
</style>
