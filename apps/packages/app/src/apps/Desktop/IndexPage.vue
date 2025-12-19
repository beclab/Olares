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
			<div class="bg-container">
				<img
					v-if="tokenStore.config.bg"
					fit="fill"
					class="desktop-bg"
					:src="bgSrc"
				/>
				<img v-else fit="fill" class="desktop-bg" src="/desktop/bg/0.jpg" />
			</div>

			<DailyDescription />

			<dock-component
				ref="dockRef"
				:window_infos="window_infos"
				@appClick="onDockerClick"
				:drag_launch_app="drag_launch_app"
				@changeSearchDialog="changeSearchDialog"
				@onClose="onWindowClose"
				@mini="onWindowMiniSize"
			/>

			<launch-pad
				v-if="isShowLaunchPad"
				:isShowLaunch="isShowLaunchPad"
				@appClick="onLaunchPadAppClick"
				@dismiss="onLaunchPadDismiss"
				@drag_launch_app="onDragLaunchApp"
			/>

			<NotificationPopup />
		</div>

		<div
			style="width: 100%; height: 100vh"
			@click="notificationStore.showNotification = false"
			ref="window_parent"
		>
			<div
				:style="`position: fixed;width: 300%;height: 200%; left: -100%; top: 0;`"
			>
				<template v-for="window_info in window_infos" :key="window_info.id">
					<BasicWindow
						v-if="window_info.is_show"
						:modelValue="window_info"
						@update:modelValue="onWindowUpdate"
						@close="onWindowClose"
						@mini="onWindowMiniSize"
						@onTop="onWindowRequestOnTop"
						@full="onWindowFullSize"
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
					@appClick="onLaunchPadAppClick"
				/>
			</div>
		</transition>

		<UpgradeComponent v-if="upgradeFlag" @closeUpgrade="closeUpgrade" />
	</q-page>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue';
import { useQuasar, Notify } from 'quasar';
import {
	IntentFilter,
	Action,
	Category,
	Encoder,
	bytetrade,
	DeviceType
} from '@bytetrade/core';
import { WindowInfo, AppClickInfo, MessageData } from './type/types';
import { bus } from '../../utils/bus';
import { useAppStore } from '../../stores/desktop/app';
import { useTokenStore } from '../../stores/desktop/token';
import { useUpgradeStore } from '../../stores/desktop/upgrade';
import { expiresStorage } from '../../utils/desktop/location';
import DockComponent from './DockComponent.vue';
import LaunchPad from './LaunchPad.vue';
import NotificationPopup from './NotificationPopup.vue';
import Search from '../../components/search/Desktop/IndexPage.vue';
import DailyDescription from './DailyDescription.vue';
import BasicWindow from './components/BasicWindow.vue';
import UpgradeComponent from './components/UpgradeComponent.vue';
import { useNotificationStore } from 'src/stores/desktop/notification';

const $q = useQuasar();
const appStore = useAppStore();
const tokenStore = useTokenStore();
const upgradeStore = useUpgradeStore();
const window_parent = ref<HTMLElement>();
const dockRef = ref<any>();
const isShowLaunchPad = ref(false);
const upgradeFlag = ref(false);
const messageSavePath = ref<MessageData[]>([]);
const window_infos = ref<WindowInfo[]>([]);
const drag_launch_app = ref('');
const showSearchDialog = ref(false);
const screenWidth = ref(window.innerWidth);
const widthDiff = ref(0);
const notificationStore = useNotificationStore();

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
	window_infos.value.map((info: WindowInfo) => {
		if (window_info.id === info.id) {
			info.is_show = false;
		}
	});
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

const onWindowFullSize = async (window_info: WindowInfo) => {
	let index = window_infos.value.findIndex(
		(window) => window.id == window_info.id
	);

	if (index < 0) {
		return;
	}

	if (window_infos.value[index].isResizable) {
		window_infos.value[index].width = window_parent.value?.offsetWidth
			? window_parent.value?.offsetWidth - 108
			: 800;
		await nextTick();
		await nextTick();
		window_infos.value[index].left = 108 + screenWidth.value;
		await nextTick();
		await nextTick();
	}

	await nextTick();
	await nextTick();
	window_infos.value[index].height = window_parent.value?.offsetHeight
		? window_parent.value?.offsetHeight
		: 600;
	await nextTick();
	await nextTick();
	window_infos.value[index].top = 0;
	await nextTick();
	await nextTick();
};

const closeUpgrade = () => {
	upgradeFlag.value = false;
};

const listenerMessage = (e: any) => {
	let data = e.data;
	if (!data) {
		return false;
	}
	switch (data.type) {
		case 'locationHref':
			if (data.message) {
				updateWindowUrl(data.message);
			}
			break;
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

const updateWindowUrl = (message: string) => {
	const href = new URL(message);
	window_infos.value.map((item) => {
		if (item.url.indexOf(href.host) >= 0) {
			item.redirectUrl = message.startsWith('https://')
				? message.slice(6)
				: message.startsWith('http://')
				? message.slice(5)
				: message;
		}
	});
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
		item.max_width = getWindowParentInfo().width;
		item.max_height = getWindowParentInfo().height;
	});
};

const getWindowParentInfo = () => {
	return {
		width: window_parent.value?.firstElementChild?.offsetWidth || 600,
		height: window_parent.value?.firstElementChild?.offsetHeight || 800
	};
};

const __beforeunload = (event: any) => {
	// event.preventDefault();
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
	nextTick(() => {
		const window_infos_string = expiresStorage.getItem('window_infos_string');
		if (window_infos_string && JSON.parse(window_infos_string)) {
			window_infos.value = JSON.parse(window_infos_string);
		}
	});

	bytetrade.observeUrlChange.parentEventListener(listenerMessage);

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
			} else if (intent.categories.indexOf(Category.CATEGORY_DEFAULT) >= 0) {
				onAppClick({
					appid: intent.router_id,
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

const launchPadClick = async () => {
	if (isShowLaunchPad.value) {
		isShowLaunchPad.value = false;
	} else {
		isShowLaunchPad.value = true;
	}
};

const onAppClick = async (click: AppClickInfo) => {
	console.log('click ===>', click);

	let rid = click.appid;
	if (rid.startsWith('bdock:')) {
		rid = rid.substring(6);
	} else if (rid.startsWith('bdesk:')) {
		rid = rid.substring(6);
	}

	if (rid == 'launchpad') {
		launchPadClick();
		return;
	}

	if (isShowLaunchPad.value) {
		isShowLaunchPad.value = false;
	}

	let app = appStore.myApps.find((app: any) => app.id == rid);
	if (rid == 'home') {
		app = {
			id: rid,
			appid: rid,
			icon: '',
			title: '',
			target: '',
			name: '',
			namespace: '',
			url: '',
			state: 'running',
			fatherName: null,
			openMethod: 'default'
		};
		if (click.path) {
			app.url = tokenStore.url + click.path;
			if (app.url?.startsWith('http://')) {
				app.url = app.url.substring(7);
			} else if (app.url?.startsWith('https://')) {
				app.url = app.url.substring(8);
			}
		}
	}

	if (app) {
		let url = app.url;
		if (tokenStore.islocal) {
			url = tokenStore.getAppLocalUrl(app.id);
		}
		if (
			[DeviceType.MOBILE, DeviceType.TABLET].includes(
				tokenStore.deviceInfo.device
			)
		) {
			window.open('//' + url);
			return false;
		}

		if (click.data.path) {
			url = url + '/' + click.data.path;
			url = url.replaceAll('//', '/');
		}

		if (app.openMethod == 'window') {
			window.open('//' + url);
			return;
		}

		let w = window_infos.value.find((window) => window.id == rid);

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
		} else {
			let width = window_parent.value?.offsetWidth
				? window_parent.value?.offsetWidth * 0.9 - 108
				: 800;
			let height = window_parent.value?.offsetHeight
				? window_parent.value?.offsetHeight * 0.9
				: 600;

			if (app.id.startsWith('settings')) {
				height = window_parent.value?.offsetHeight
					? window_parent.value?.offsetHeight * 0.8
					: 600;
				// width = (height / 544) * 720;
				width = 858;
			}

			let left = window_parent.value?.offsetWidth
				? (window_parent.value?.offsetWidth - width) / 2 + 54
				: 0;
			let top = window_parent.value?.offsetHeight
				? (window_parent.value?.offsetHeight - height) / 2
				: 0;

			let obj: WindowInfo = {
				width: width,
				height: height,
				max_height: getWindowParentInfo().height,
				max_width: getWindowParentInfo().width,
				min_width: 400,
				min_height: 200,
				left: left + screenWidth.value,
				top: top,
				id: app.id,
				url: '//' + url,
				ssl_type: 'http',
				title: app.title,
				is_show: true,
				z: window_infos.value.length,
				icon: app.icon,
				active: true,
				isResizable: true
			};

			if (app.id.startsWith('settings')) {
				obj.isResizable = false;
				obj.min_width = obj.width;
				obj.min_height = obj.height;
			} else if (app.id.startsWith('appstore')) {
				obj.min_width = obj.width / 2;
				obj.min_height = obj.height / 2;
			} else {
				obj.left = obj.left + window_infos.value.length * 25;
				obj.top = obj.top + window_infos.value.length * 25;
			}

			window_infos.value.push(obj);

			for (let i = 0; i < window_infos.value.length - 1; ++i) {
				window_infos.value[i].active = false;
			}

			if (rid != 'home') {
				dockRef.value?.handleOpenApp(rid);
			}
		}
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

const onLaunchPadAppClick = async (click: AppClickInfo) => {
	isShowLaunchPad.value = false;
	showSearchDialog.value = false;
	onAppClick(click);
};

const onLaunchPadDismiss = async () => {
	isShowLaunchPad.value = false;
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

const bgSrc = computed(() => {
	if (tokenStore.config.bg.startsWith('http')) {
		return tokenStore.config.bg;
	}
	return '/desktop' + tokenStore.config.bg;
});
</script>
<style lang="scss" scoped>
.bg-container {
	width: 100%;
	height: 100%;
	position: fixed;
	left: 0;
	right: 0;
	top: 0;
	bottom: 0;
	z-index: -1;
	display: flex;
	justify-content: center;
	align-items: center;
	overflow: hidden;
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
