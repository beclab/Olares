<template>
	<div
		class="notification_page"
		:class="notificationStore.showNotification ? 'moveRight' : ''"
	>
		<NotificationContents>
			<template v-slot:more>
				<div
					class="row items-center justify-center q-mt-lg"
					v-if="notificationStore.hadMore"
					@click="notificationStore.loadMore"
				>
					<span class="load-more text-body3 text-ink2">
						{{ $t('load more') }}
					</span>
				</div>
			</template>
		</NotificationContents>
	</div>
</template>

<script lang="ts" setup>
import { watch } from 'vue';
import { useNotificationStore } from '../../stores/desktop/notification';

import NotificationContents from './NotificationContents.vue';

const notificationStore = useNotificationStore();

watch(
	() => notificationStore.showNotification,
	(newVal) => {
		console.log('showNotification', newVal);
	}
);
</script>

<style lang="scss" scoped>
.notification_page {
	position: fixed;
	top: 0;
	right: -382px;
	z-index: 10;
	width: 360px;
	height: 100vh;
	overflow: scroll;
	padding: 32px 10px;
	transition: all 0.5s linear;

	&::-webkit-scrollbar {
		display: none;
	}

	.load-more {
		padding: 0px 12px;
		height: 24px;
		border-radius: 12px;
		line-height: 22px;
		background: linear-gradient(
				0deg,
				rgba(246, 246, 246, 0.5),
				rgba(246, 246, 246, 0.5)
			),
			linear-gradient(0deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.2));

		border: 1px solid #ffffff33;
	}
}

.moveRight {
	right: 22px;
}
</style>
