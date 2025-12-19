<template>
	<div class="notification_page">
		<NotificationContents>
			<template v-slot:more>
				<div style="margin-bottom: 200px">
					<div
						class="row items-center justify-center q-mt-lg"
						v-if="
							notificationStore.hadMore && notificationStore.data.length > 0
						"
						@click="notificationStore.loadMore"
					>
						<span class="load-more text-body3 text-ink2">
							{{ $t('load more') }}
						</span>
					</div>
				</div>
			</template>
		</NotificationContents>
	</div>
</template>

<script lang="ts" setup>
import { useNotificationStore } from '../../stores/desktop/notification';

import NotificationContents from './NotificationContents.vue';

const notificationStore = useNotificationStore();
</script>

<style lang="scss" scoped>
.notification_page {
	top: 0;
	width: 100%;
	height: 100%;
	overflow: scroll;
	padding: 32px 30px;

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
