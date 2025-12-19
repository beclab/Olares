<template>
	<div>
		<div class="bg-container">
			<img
				v-if="tokenStore.config.bg"
				fit="fill"
				class="desktop-bg"
				:src="`/desktop${tokenStore.config.bg}`"
			/>
			<img v-else fit="fill" class="desktop-bg" src="/desktop/bg/0.jpg" />
		</div>

		<q-tab-panels v-model="panel" animated swipeable class="tab-panels-bg">
			<q-tab-panel name="notification" class="tab-item">
				<NotificationView />
			</q-tab-panel>
			<q-tab-panel name="home" class="tab-item">
				<HomeView />
			</q-tab-panel>
		</q-tab-panels>

		<div
			class="column items-center bottom-tab justify-between"
			:class="{
				'bottom-shadow-bg': panel == 'notification'
			}"
		>
			<div
				class="clear row items-center justify-center"
				@click="notificationStore.deleteAll()"
				v-if="panel == 'notification' && notificationStore.data.length > 0"
			>
				<q-img src="/desktop/app-icon/clean.svg" width="20px" height="20px" />
			</div>
			<div v-else></div>
			<div class="tab-itmes row items-center items-center q-px-sm">
				<template v-if="panel == 'notification'">
					<q-img src="/desktop/app-icon/notification-actived.svg" width="8px" />
					<q-img src="/desktop/app-icon/home.svg" width="8px" class="q-ml-sm" />
				</template>
				<template v-else-if="panel == 'home'">
					<div class="row items-center justify-center notification">
						<q-img src="/desktop/app-icon/notification.svg" width="8px"></q-img>
						<div class="badge" v-if="notificationStore.newMessage"></div>
					</div>
					<q-img
						src="/desktop/app-icon/home-actived.svg"
						width="8px"
						class="q-ml-sm"
					/>
				</template>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import HomeView from './HomeView.vue';
import NotificationView from './NotificationView.vue';
import { useTokenStore } from '../../../stores/desktop/token';
import { useNotificationStore } from '../../../stores/desktop/notification';

const panel = ref('home');
const tokenStore = useTokenStore();
const notificationStore = useNotificationStore();

watch(
	() => panel.value,
	() => {
		if (panel.value == 'notification') {
			notificationStore.newMessage = false;
		}
	}
);
</script>
<style scoped lang="scss">
.bg-container {
	width: 100%;
	height: 100%;
	position: fixed;
	left: 0;
	right: 0;
	top: 0;
	bottom: 0;
	display: flex;
	justify-content: center;
	align-items: center;
	overflow: hidden;
}

.desktop-bg {
	width: auto;
	min-width: 100%;
	height: 100%;
	object-fit: cover;
}

.tab-panels-bg {
	width: 100vw;
	height: 100vh;
	padding: 0;
	background: transparent;

	.tab-item {
		width: 100%;
		height: 100%;
		padding: 0;
	}
}

.bottom-tab {
	position: absolute;
	height: 133px;
	width: 100%;
	z-index: 7;
	// background-color: red;
	bottom: 0px;
	left: 0px;

	.clear {
		height: 50px;
		width: 50px;
		border-radius: 50%;
		margin-top: 20px;
		background: #ffffffcc;
		border: 1px solid #ffffffcc;
		backdrop-filter: blur(15.699999809265137px);
	}

	.tab-itmes {
		// background-color: red;
		box-shadow: 0px 0px 4px 0px #0000002e;
		backdrop-filter: blur(50px);
		background: #ffffff66;
		border-radius: 10px;
		padding: 0 10px;
		// width: 60px;
		height: 20px;
		margin-bottom: 30px;

		.notification {
			position: relative;
			.badge {
				width: 4px;
				height: 4px;
				border-radius: 2px;
				position: absolute;
				right: -2px;
				top: 0px;
				background-color: #fa473b;
			}
		}
	}
}

.bottom-shadow-bg {
	background: linear-gradient(
		180deg,
		rgba(0, 0, 0, 0) 2.33%,
		rgba(11, 38, 63, 0.8) 53.49%
	);
}
</style>
