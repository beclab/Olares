<template>
	<q-page class="flex flex-center">
		<div
			class="desktop-box"
			@dragstart="onDragStart"
			@dragenter="onDragEnter"
			@dragleave="onDragLeave"
			@dragover="onDragOver"
			@dragend="onDragEnd"
			@drop="onDrop"
		>
			<div class="desktop-avatar" @click="onLogout">
				<!-- <AvatarComponent :width="48" :height="48" /> -->
			</div>

			<DailyDescription />

			<dock-component
				ref="dockRef"
				:drag_launch_app="drag_launch_app"
				@appClick="onDockerClick"
				@changeSearchDialog="changeSearchDialog"
			/>

			<launch-pad
				v-if="isShowLauncchPad"
				:isShowLaunc="isShowLauncchPad"
				@appClick="onLaunhPadAppClick"
				@dismiss="onLaunchPadDismiss"
				@drag_launch_app="onDragLaunchApp"
			/>

			<!-- <NotificationPopup /> -->
		</div>

		<div style="width: 100%; height: 100vh" ref="window_parent">
			<div
				:style="`position: relative;width: 100%;height: 100%; left: 0; top: 0;`"
			>
				<template v-for="window_info in window_infos" :key="window_info.id">
					<BasicWindow
						v-if="window_info.is_show"
						:modelValue="window_info"
						@update:modelValue="onWindowUpdate"
						@close="onWindowClose"
						@mini="onWindowMiniSize"
						@onTop="onWindowRequestOnTop"
						:style="`z-index:${window_info.z};`"
					/>
				</template>
			</div>
		</div>
		<transition
			appear
			enter-active-class="animated fadeIn"
			leave-active-class="animated fadeOut"
		>
			<div
				class="search_mask"
				v-if="showSearchDialog"
				@click.self="changeSearchDialog(false)"
			>
				<Search
					class="search"
					v-if="showSearchDialog"
					@hide="changeSearchDialog"
					@appClick="onLaunhPadAppClick"
				/>
			</div>
		</transition>

		<UpgradeComponent v-if="upgradeFlag" @closeUpgrade="closeUpgrade" />
	</q-page>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useQuasar, Notify, Loading } from 'quasar';
import {
	IntentFilter,
	Action,
	Category,
	Encoder,
	bytetrade
} from '@bytetrade/core';
import { WindowInfo, AppClickInfo, MessageData } from '../type/types';
import { bus } from '../../../utils/bus';
import { useAppStore } from '../../../stores/desktop/app';
import { useTokenStore } from '../../../stores/desktop/token';
import { useUpgradeStore } from '../../../stores/desktop/upgrade';
import { expiresStorage } from '../../../utils/desktop/location';
import Search from '../../../components/search/Desktop/IndexPage.vue';
import AvatarComponent from '../components/AvatarComponent.vue';
import BasicWindow from './BasicWindow.vue';
import UpgradeComponent from '../components/UpgradeComponent.vue';

import DockComponent from './DockComponent.vue';
import LaunchPad from './LaunchPad.vue';
import DailyDescription from './DailyDescription.vue';

const $q = useQuasar();
const appStore = useAppStore();
const tokenStore = useTokenStore();
const upgradeStore = useUpgradeStore();
const window_parent = ref<HTMLElement>();
const dockRef = ref<any>();
const isShowLauncchPad = ref(false);
const upgradeFlag = ref(false);
const messageSavePath = ref<MessageData[]>([]);
const window_infos = ref<WindowInfo[]>([]);
const drag_launch_app = ref('');
const showSearchDialog = ref(false);
const screenWidth = ref(window.innerWidth);
const widthDiff = ref(0);

let need_save_window = false;

const onWindowUpdate = (window_info: WindowInfo) => {
	window_infos.value.forEach((info: WindowInfo) => {
		if (info.id === window_info.id) {
			Object.assign(info, window_info);
			need_save_window = true;
		}
	});
};

const onWindowClose = (window_info: WindowInfo) => {
	window_infos.value = window_infos.value.filter(
		(info: WindowInfo) => info.id !== window_info.id
	);

	messageSavePath.value = messageSavePath.value.filter(
		(info: MessageData) => info.type !== window_info.title
	);

	dockRef.value?.handleRemoveApp(window_info.id);

	for (let i = 0; i < window_infos.value.length; ++i) {
		if (window_infos.value[i].z > window_info.z) {
			window_infos.value[i].z--;
		}
	}

	need_save_window = true;
};

const onWindowMiniSize = (window_info: WindowInfo) => {
	window_info.is_show = false;
	need_save_window = true;
};

const onWindowRequestOnTop = (window_info: WindowInfo) => {
	let w = window_infos.value.find((window) => window.id == window_info.id);

	if (w) {
		w.is_show = true;
		for (let i = 0; i < window_infos.value.length; ++i) {
			if (window_infos.value[i].z > w.z) {
				window_infos.value[i].z--;
				window_infos.value[i].active = false;
			}
		}
		w.z = window_infos.value.length;
		w.active = true;
	}

	need_save_window = true;
};

// const onWindowFullSize = async (window_info: WindowInfo) => {
// 	let index = window_infos.value.findIndex(
// 		(window) => window.id == window_info.id
// 	);

// 	if (index < 0) {
// 		return;
// 	}

// 	if (window_infos.value[index].isResizable) {
// 		window_infos.value[index].width = window_parent.value?.offsetWidth
// 			? window_parent.value?.offsetWidth - 108
// 			: 800;
// 		await nextTick();
// 		await nextTick();
// 		window_infos.value[index].left = 108 + screenWidth.value;
// 		await nextTick();
// 		await nextTick();
// 	}

// 	await nextTick();
// 	await nextTick();
// 	window_infos.value[index].height = window_parent.value?.offsetHeight
// 		? window_parent.value?.offsetHeight
// 		: 600;
// 	await nextTick();
// 	await nextTick();
// 	window_infos.value[index].top = 0;
// 	await nextTick();
// 	await nextTick();
// };

const closeUpgrade = () => {
	upgradeFlag.value = false;
};

const listenerMessage = (e: any) => {
	let data = e.data;
	if (!data) {
		return false;
	}
	switch (data.type) {
		case 'Files':
			if (data.message) {
				data.message = data.message.slice(6);
				let hasMessageIndex = messageSavePath.value.findIndex(
					(item: { type: string }) => item.type === data.type
				);
				if (hasMessageIndex >= 0) {
					messageSavePath.value[hasMessageIndex] = data;
				} else {
					messageSavePath.value.push(data);
				}
			}
			break;
		case 'dev_mode':
			if (data.message === 'true') {
				localStorage.setItem('dev_mode', data.message);
			} else {
				localStorage.removeItem('dev_mode');
			}
			break;
		default:
			break;
	}
};

watch(
	() => upgradeStore.state,
	(newVal, oldVal) => {
		if (newVal && newVal !== 'not_running') {
			if (newVal === 'completed' && oldVal === 'not_running') {
				upgradeFlag.value = false;
			} else {
				upgradeFlag.value = true;
				window_infos.value = [];
			}
		} else {
			upgradeFlag.value = false;
		}
	}
);

watch(
	() => screenWidth.value,
	() => {
		computeWidowLeft();
	}
);

const computeWidowLeft = () => {
	window_infos.value.map((item) => {
		item.left = item.left + widthDiff.value;
	});
};

const __beforeunload = (event: any) => {
	event.preventDefault();
	const window_infos_parse = JSON.parse(JSON.stringify(window_infos.value));
	for (let i = 0; i < window_infos_parse.length; ++i) {
		let hasMessage = messageSavePath.value.find(
			(item: { type: string }) => item.type === window_infos_parse[i].title
		);
		if (hasMessage) {
			window_infos_parse[i].url = hasMessage.message;
		}
	}
	const window_infos_string = JSON.stringify(window_infos_parse);
	expiresStorage.setItem('window_infos_string', window_infos_string, 10000);
};

let window_update_interval: any = null;
onMounted(async () => {
	// const content = '<div class="content"><strong style="font-size: 14px;">Trading information</strong><div style="font-size: 12px;">Your transaction on Sep 6 was successful, please go to the Dapp to view the transaction detials.</div></div>';
	// $q.notify({
	// 	message: content,
	// 	html: true,
	// 	position: 'top-right',
	// 	timeout: 2000,
	// 	avatar: './icons/favicon-128x128.ico',
	// });

	nextTick(() => {
		const window_infos_string = expiresStorage.getItem('window_infos_string');
		if (window_infos_string && JSON.parse(window_infos_string)) {
			window_infos.value = JSON.parse(window_infos_string);
		}
	});

	bytetrade.observeUrlChange.parentEventListener(listenerMessage);
	//updateDesktopList();

	window_update_interval = setInterval(() => {
		if (need_save_window) {
			need_save_window = false;
		}
	}, 5 * 1000);

	window.addEventListener('beforeunload', __beforeunload);
	window.addEventListener('keydown', keydownEnter);

	bus.on('intent', (intent: IntentFilter) => {
		if (!intent) return;
		if (
			intent.actions.indexOf(Action.ACTION_VIEW) >= 0 &&
			intent.categories.indexOf(Category.CATEGORY_CONTAINER_LOG) >= 0
		) {
			let path = '';
			if (intent.data.deployment) {
				path =
					'container/logs/deployment/' +
					intent.data.deployment +
					'/' +
					intent.data.container;
			} else if (intent.data.statefulset) {
				path =
					'container/logs/statefulset/' +
					intent.data.statefulset +
					'/' +
					intent.data.container;
			} else if (intent.data.daemonset) {
				path =
					'container/logs/daemonset/' +
					intent.data.daemonset +
					'/' +
					intent.data.container;
			}
			onAppClick({
				appid: intent.router_id!,
				data: null,
				path
			});
		} else if (intent.actions.indexOf(Action.ACTION_VIEW) >= 0) {
			if (intent.categories.indexOf(Category.CATEGORY_VIDEO) >= 0) {
				if (intent.data.path) {
					let path =
						'/video/stream/file/' +
						Encoder.bytesToString(Encoder.base64UrlToBytes(intent.data.path));
					onAppClick({
						appid: intent.router_id!,
						data: null,
						path
					});
				}
			} else if (intent.categories.indexOf(Category.CATEGORY_LAUNCHER) >= 0) {
				onAppClick({
					appid: intent.data.appid,
					data: {
						path: intent.data.path
					}
				});
			}
		} else if (intent.actions.indexOf(Action.ACTION_EDIT) >= 0) {
			if (intent.categories.indexOf(Category.CATEGORY_LAUNCHER) >= 0) {
				tokenStore.updateDesktopConfig(intent.data);
			}
		}
	});

	bus.on('notification', (notification: any) => {
		$q.notify({
			message: notification.body,
			caption: notification.title,
			type: 'info',
			position: 'top',
			timeout: 2000
		});
	});

	window.onresize = () => {
		return (() => {
			widthDiff.value = window.innerWidth - screenWidth.value;
			screenWidth.value = window.innerWidth;
		})();
	};
});

onUnmounted(() => {
	bus.off('intent');
	bus.off('notification');
	window.removeEventListener('beforeunload', __beforeunload);
	if (window_update_interval) {
		clearInterval(window_update_interval);
	}
});

const keydownEnter = (event: any) => {
	if (event.shiftKey && event.keyCode === 32) {
		event.stopPropagation();
		showSearchDialog.value = !showSearchDialog.value;
	}
};

const launchPadclick = async () => {
	if (isShowLauncchPad.value) {
		isShowLauncchPad.value = false;
	} else {
		isShowLauncchPad.value = true;
	}
};

const onAppClick = async (click: AppClickInfo) => {
	let rid = click.appid;
	if (rid.startsWith('bdock:')) {
		rid = rid.substring(6);
	} else if (rid.startsWith('bdesk:')) {
		rid = rid.substring(6);
	}

	if (rid == 'launchpad') {
		launchPadclick();
		return;
	}

	if (isShowLauncchPad.value) {
		isShowLauncchPad.value = false;
	}

	let app = appStore.myApps.find((app: any) => app.id == rid);

	if (app) {
		let url = app.url;
		if (tokenStore.islocal) {
			url = tokenStore.getAppLocalUrl(app.id);
		}
		window.open('//' + url);
	} else {
		Notify.create({
			type: 'negative',
			message: rid
		});
	}
};

const onDockerClick = async (click: AppClickInfo) => {
	onAppClick(click);
};

const onLaunhPadAppClick = async (click: AppClickInfo) => {
	isShowLauncchPad.value = false;
	showSearchDialog.value = false;
	onAppClick(click);
};

const onLaunchPadDismiss = async () => {
	isShowLauncchPad.value = false;
};

const onDragLaunchApp = async (id: string) => {
	drag_launch_app.value = id;
};

const onDragStart = async (e: any) => {
	e.dataTransfer.setData('text', e.target.id);
	e.dataTransfer.dropEffect = 'move';
};

const onDragEnter = async () => {
	console.log('desktop onEnter');
};

const onDragLeave = async (e: any) => {
	e.target.classList.remove('drag-enter');
};

const onDragOver = async (e: any) => {
	e.preventDefault();
};

const onDrop = async (e: any) => {
	e.preventDefault();
	console.log('desktop onDrop');
};

const onDragEnd = async (e: any) => {
	e.preventDefault();
	console.log('desktop onDragEnd');
};

const changeSearchDialog = (value: boolean) => {
	showSearchDialog.value = value;
};

const onLogout = async () => {
	Loading.show();
	try {
		await tokenStore.logout();
		const auth_url = tokenStore.getAuthURL() + '?logout=1';
		window.location.replace(auth_url);
	} catch (err) {
		Notify.create({
			type: 'negative',
			message: (err as Error).message
		});
	} finally {
		Loading.hide();
	}
};
</script>
<style lang="scss" scoped>
.desktop-avatar {
	position: fixed;
	right: 20px;
	top: 20px;
	z-index: 1;
}
.desktop-box {
	width: 100%;
	height: 100%;
	overflow: hidden;
	position: absolute;
	top: 0px;
	left: 0px;
}

.desktop-bg {
	width: auto;
	min-width: 100%;
	height: 100%;
	object-fit: cover;
}

.app-link {
	width: 96px;
	height: 96px;
}

.fade-none {
	display: block;
}

.fade-in {
	-webkit-animation: fade-in 0.8s cubic-bezier(0.39, 0.575, 0.565, 1) both;
	animation: fade-in 0.8s cubic-bezier(0.39, 0.575, 0.565, 1) both;
}

@-webkit-keyframes fade-in {
	0% {
		opacity: 1;
	}

	100% {
		opacity: 0;
		display: none;
	}
}

@keyframes fade-in {
	0% {
		opacity: 1;
	}

	100% {
		opacity: 0;
		display: none;
	}
}

.search_mask {
	position: absolute;
	top: 0;
	left: 0;
	width: 100vw;
	height: 100vh;
	z-index: 1000;
	background-color: rgba($color: #000000, $alpha: 0.5);
}
</style>
