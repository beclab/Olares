<template>
	<div class="desktop-box">
		<div class="desktop_app_box">
			<div
				class="desktop_app"
				:style="`height:${DOCKER_APP_TOTAL_HEIGHT + 230}px;`"
			>
				<div class="desktop_user_avatar">
					<div class="desktop_avatar" @click="openProfile">
						<TerminusAvatar :info="tokenStore.terminus" :size="36" />
					</div>

					<div class="desktop_avatar_border"></div>
				</div>
				<div
					class="column_apps_all"
					id="app_dock_list"
					@dragleave="onDockDragLeave"
				>
					<template v-for="element in appStore.dockerApps" :key="element.id">
						<img
							ref="imgRef"
							@click="openWindow(element)"
							:id="element.id"
							:style="`position:absolute;top:${element.top}px;width:${
								appStore.DOCKER_APP_SIZE
							}px;height:${appStore.DOCKER_APP_SIZE}px;left:${
								(60 - appStore.DOCKER_APP_SIZE) / 2
							}px;border-radius: ${borderRadiusFormat(
								appStore.DOCKER_APP_SIZE,
								appStore.DOCKER_APP_SIZE
							)}px`"
							draggable="true"
							@dragstart="onDragStart"
							@dragenter="onDragEnter"
							@dragleave="onDragLeave"
							@dragover="onDragOver"
							@dragend="onDragEnd"
							@drop="onDrop"
							:src="element.icon"
							@contextmenu.prevent="onContextmenu(element)"
						/>
						<div
							:class="element.show_dot ? 'img_parent' : ''"
							:style="`position:absolute;top:${
								element.top + 17
							}px;right: -3px;`"
						></div>
						<div
							:class="contextmenuId == element.id ? 'menu_open' : 'menu_hide'"
							:style="`position:absolute;top:${element.top}px;left:70px;`"
						>
							<div
								class="item"
								v-if="!curOpenApp.includes(element.id) && !element.show_dot"
								@click="openWindow(element)"
							>
								{{ t('dock_operation.open') }}
							</div>

							<div
								class="item"
								v-if="!curOpenApp.includes(element.id) && element.show_dot"
								@click="onShow(element)"
							>
								{{ t('dock_operation.show') }}
							</div>

							<div
								class="item"
								v-if="curOpenApp.includes(element.id) && element.show_dot"
								@click="mini(element)"
							>
								{{ t('dock_operation.hide') }}
							</div>

							<div
								class="item"
								v-if="element.show_dot"
								@click="updateAppTemp(element)"
							>
								<span>{{ t('dock_operation.keepInDock') }}</span>
								<span class="q-ml-md" style="width: 30px"
									><q-icon
										v-if="!element.is_temp"
										class="icon"
										name="check"
										size="12px"
								/></span>
							</div>

							<q-separator class="q-my-xs q-mx-sm" color="#d9d9d9" />

							<div
								class="item red_font"
								v-if="!element.is_temp && !element.show_dot"
								@click="removeAppInDock(element.id)"
							>
								{{ t('dock_operation.removeFromDock') }}
							</div>

							<div
								class="item red_font"
								v-if="element.show_dot"
								@click="onClose(element)"
							>
								{{ t('dock_operation.quit') }}
							</div>
							<div class="context_triangle"></div>
						</div>
					</template>
				</div>

				<div class="desktop_message desktop_user_avatar">
					<div class="desktop_avatar_border"></div>
					<div class="dock_icon" @click="toSearch">
						<img src="../../assets/desktop/dock-search.svg" alt="search" />
					</div>

					<div class="dock_icon" @click="showNotification">
						<img src="../../assets/desktop/dock-notify.svg" alt="notify" />
						<div class="badge" v-if="notificationStore.newMessage" />
					</div>
					<div class="dock_icon" @click="onLogout">
						<img src="../../assets/desktop/dock-logout.svg" alt="logout" />
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { Loading, Notify, useQuasar } from 'quasar';
import { DockerAppInfo, AppClickInfo } from './type/types';
import { useTokenStore } from '../../stores/desktop/token';
import { useAppStore } from '../../stores/desktop/app';
import { borderRadiusFormat } from '../../utils/desktop/utils';
import ConfirmDialog from './components/ConfirmDialog.vue';
import { useI18n } from 'vue-i18n';

import { useNotificationStore } from '../../stores/desktop/notification';

const props = defineProps({
	drag_launch_app: {
		type: String,
		required: true
	},
	window_infos: {
		type: Array,
		required: false,
		default: () => []
	}
});

const emits = defineEmits([
	'changeSearchDialog',
	'appClick',
	'onClose',
	'mini'
]);

const tokenStore = useTokenStore();
const appStore = useAppStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();
const contextmenuId = ref<string>('');
let origin_app: any = null;
let choose: any = null;
let remove_top = -1;
let remove_id: string | null = null;
const $q = useQuasar();

const onLogout = async () => {
	$q.dialog({
		component: ConfirmDialog,
		componentProps: {
			title: t('Log out'),
			message: t('Your changes may not be saved.'),
			showCancel: true
		}
	}).onOk(async () => {
		try {
			Loading.show();
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
	});
};

const app_height_delta = ref(0);

const DOCKER_APP_TOTAL_HEIGHT = computed(() => {
	let t =
		appStore.DOCKER_APP_START_GAP +
		appStore.dockerApps.length * appStore.DOCKER_APP_SIZE +
		(appStore.dockerApps.length - 1) * appStore.DOCKER_APP_GAP +
		appStore.DOCKER_APP_END_GAP;

	return t + app_height_delta.value;
});

const curOpenApp = ref<string[]>([]);

watch(
	() => props.window_infos.length,
	() => {
		const dockerApps = appStore.dockerApps;
		curOpenApp.value = dockerApps
			.filter((app) =>
				props.window_infos.some((info) => 'bdock:' + info.id === app.id)
			)
			.map((app) => app.id);
	},
	{
		immediate: true
	}
);

const openWindow = async (item: DockerAppInfo) => {
	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
};

const onShow = (item: DockerAppInfo) => {
	curOpenApp.value.push(item.id);
	openWindow(item);
};

const mini = (item: DockerAppInfo) => {
	curOpenApp.value = curOpenApp.value.filter((app) => app !== item.id);
	const curItem = JSON.parse(JSON.stringify(item));
	if (curItem.id.startsWith('bdock:')) {
		curItem.id = curItem.id.substring(6);
	}
	emits('mini', curItem);
};

const onClose = (item: DockerAppInfo) => {
	const curItem = JSON.parse(JSON.stringify(item));
	if (curItem.is_temp) {
		appStore.remove_app_on_docker(curItem.id, true);
	}
	if (curItem.id.startsWith('bdock:')) {
		curItem.id = curItem.id.substring(6);
	}
	emits('onClose', curItem);
};

const openProfile = async () => {
	emits('appClick', {
		appid: appStore.profile_id,
		data: {}
	} as AppClickInfo);
};

const onContextmenu = (element: DockerAppInfo) => {
	if (element.id === 'bdock:launchpad') {
		return false;
	}
	contextmenuId.value = element.id;
};

const getAppIndexInDock = (id: string) => {
	let index = -1;
	for (var i = 0; i < appStore.dockerApps.length; i++) {
		const draggedEl: any = document.getElementById(appStore.dockerApps[i].id);
		if (draggedEl.id == id) {
			index = i;
		}
	}
	return index;
};

const onDragStart = async (e: any) => {
	e.dataTransfer.setData('text', e.target.id);

	origin_app = appStore.dockerApps[getAppIndexInDock(e.target.id)];
	choose = getAppIndexInDock(e.target.id);

	remove_top = -1;
	remove_id = null;

	e.dataTransfer.dropEffect = 'move';
	e.dataTransfer.effectAllowed = 'move';
	e.target.classList.add('drag-start');
};

let isAnimation = false;
function animate(
	targets: any,
	type: number,
	source_index: any,
	source_target_y: any,
	target_top: number
) {
	isAnimation = true;
	if (type == 0) {
		appStore.dockerApps[source_index].top = source_target_y;
	}

	let p = [];
	for (var i = 0; i < targets.length; ++i) {
		p.push(
			new Promise((resolve) => {
				let targetY = targets[i].targetY;
				let index = targets[i].index;

				function finish() {
					clearInterval(moveY);
					appStore.dockerApps[index].top = targetY;
					resolve(true);
				}

				let moveY = setInterval(function () {
					if (isAnimation == false) {
						finish();
						return;
					}

					if (targetY > appStore.dockerApps[index].top) {
						appStore.dockerApps[index].top += 2;
						if (targetY <= appStore.dockerApps[index].top) {
							finish();
						}
					} else if (targetY < appStore.dockerApps[index].top) {
						appStore.dockerApps[index].top -= 2;
						if (targetY >= appStore.dockerApps[index].top) {
							finish();
						}
					} else {
						finish();
					}
				}, 1);
			})
		);
	}

	if (type == 1 || type == 2 || type == 3) {
		p.push(
			new Promise((resolve) => {
				function finish() {
					clearInterval(moveY);
					resolve(true);
				}

				let moveY = setInterval(function () {
					if (isAnimation == false) {
						finish();
						return;
					}

					if (source_target_y > 0) {
						app_height_delta.value += 4;
						if (source_target_y <= app_height_delta.value) {
							finish();
						}
					} else if (source_target_y < 0) {
						app_height_delta.value -= 4;
						if (source_target_y >= app_height_delta.value) {
							finish();
						}
					} else {
						finish();
					}
				}, 2);
			})
		);
	}

	Promise.all(p).then(() => {
		if (type == 1) {
			for (var i = 0; i < appStore.dockerApps.length; i++) {
				const draggedEl: any = document.getElementById(
					appStore.dockerApps[i].id
				);
				draggedEl.classList.remove('drag-start');
			}
			appStore.remove_app_on_docker(source_index, false);
			app_height_delta.value = 0;
		} else if (type == 2) {
			appStore.add_app_on_docker(source_index, target_top, false, false);
			app_height_delta.value = 0;
			remove_top = -1;
			remove_id = null;
		} else if (type == 3) {
			appStore.add_app_on_docker(source_index, target_top, false, false);
			app_height_delta.value = 0;

			origin_app = appStore.dockerApps[appStore.dockerApps.length - 1];
			choose = appStore.dockerApps.length - 1;

			remove_top = -1;
			remove_id = null;
		}
		isAnimation = false;
	});
}

const checkMove = (e: any) => {
	if (isAnimation) {
		return;
	}

	if (origin_app) {
		if (e.target.id == origin_app.id) {
			return;
		}
	}

	let dragIndex = getAppIndexInDock(e.target.id);

	let source_top = 0;
	if (remove_top >= 0 || !origin_app) {
		const draggedEl: any = document.getElementById(
			appStore.dockerApps[dragIndex].id
		);
		if (draggedEl) {
			let target_id = null;
			if (origin_app) {
				target_id = origin_app.id;
			} else {
				target_id = props.drag_launch_app.substring(6);
			}
			if (!appStore.is_app_in_docker(target_id)) {
				if (e.clientY * 2 > draggedEl.y) {
					checkAdd(appStore.dockerApps[dragIndex].top, 3, target_id);
				} else {
					checkAdd(
						appStore.dockerApps[dragIndex].top +
							appStore.DOCKER_APP_SIZE +
							appStore.DOCKER_APP_GAP,
						3,
						target_id
					);
				}
			}
		}
	} else {
		source_top = appStore.dockerApps[choose].top;

		let targets = [];

		if (source_top < appStore.dockerApps[dragIndex].top) {
			for (var i = 0; i < appStore.dockerApps.length; i++) {
				if (i == choose) {
					continue;
				}

				if (
					appStore.dockerApps[i].top > source_top &&
					appStore.dockerApps[i].top <= appStore.dockerApps[dragIndex].top
				) {
					targets.push({
						index: i,
						targetY:
							appStore.dockerApps[i].top -
							appStore.DOCKER_APP_SIZE -
							appStore.DOCKER_APP_GAP
					});
				}
			}
		} else if (source_top > appStore.dockerApps[dragIndex].top) {
			for (let i = 0; i < appStore.dockerApps.length; i++) {
				if (i == choose) {
					continue;
				}

				if (
					appStore.dockerApps[i].top >= appStore.dockerApps[dragIndex].top &&
					appStore.dockerApps[i].top < source_top
				) {
					targets.push({
						index: i,
						targetY:
							appStore.dockerApps[i].top +
							appStore.DOCKER_APP_SIZE +
							appStore.DOCKER_APP_GAP
					});
				}
			}
		}

		if (targets.length > 0) {
			animate(targets, 0, choose, appStore.dockerApps[dragIndex].top, 0);
		}
	}
};

const onDragEnter = async (e: any) => {
	e.preventDefault();
	if (e.target.draggable === true) {
		checkMove(e);
	}
};

const onDragLeave = async (e: any) => {
	e.preventDefault();
};

const removeAppInIndex = (app_remove_index: number, rid: string) => {
	let targets = [];

	for (var i = 0; i < appStore.dockerApps.length; i++) {
		if (i == app_remove_index) {
			continue;
		}

		if (
			appStore.dockerApps[i].top > appStore.dockerApps[app_remove_index].top
		) {
			targets.push({
				index: i,
				targetY:
					appStore.dockerApps[i].top -
					appStore.DOCKER_APP_SIZE -
					appStore.DOCKER_APP_GAP
			});
		}
	}

	remove_top = appStore.dockerApps[app_remove_index].top;
	remove_id = appStore.dockerApps[app_remove_index].id;

	animate(
		targets,
		1,
		rid,
		-appStore.DOCKER_APP_SIZE - appStore.DOCKER_APP_GAP,
		0
	);
};

const checkRemove = () => {
	if (isAnimation) {
		return;
	}

	if (!appStore.find_app(origin_app.id)) {
		return;
	}
	removeAppInIndex(choose, origin_app.id);
};

const handleRemoveApp = (appid: string) => {
	let rid = appid;
	if (rid.startsWith('bdock:') == false) {
		rid = 'bdock:' + rid;
	}

	let cc = getAppIndexInDock(rid);
	if (cc < 0) return;
	if (appStore.dockerApps[cc].is_temp) {
		removeAppInIndex(cc, rid);
	}
	appStore.dockerApps[cc].show_dot = false;

	appStore.updateAppInDock(appStore.dockerApps[cc]);
};

const updateAppTemp = (value: DockerAppInfo) => {
	const app = JSON.parse(JSON.stringify(value));
	app.is_temp = !app.is_temp;
	appStore.updateAppInDock(app);
};

const removeAppInDock = (id: string) => {
	appStore.remove_app_on_docker(id, true);
};

const handleOpenApp = (appid: string) => {
	let rid = appid;
	if (rid.startsWith('bdock:') == false) {
		rid = 'bdock:' + rid;
	}

	let cc = getAppIndexInDock(rid);
	if (cc < 0) {
		appStore.add_app_on_docker_bottom(rid);
	} else {
		appStore.dockerApps[cc].show_dot = true;
	}
};

const checkAdd = (target_top: number, type: number, app_id: string) => {
	if (isAnimation) {
		return;
	}

	let targets = [];
	for (var i = 0; i < appStore.dockerApps.length; i++) {
		if (appStore.dockerApps[i].top >= target_top) {
			targets.push({
				index: i,
				targetY:
					appStore.dockerApps[i].top +
					appStore.DOCKER_APP_SIZE +
					appStore.DOCKER_APP_GAP
			});
		}
	}

	animate(
		targets,
		type,
		app_id,
		appStore.DOCKER_APP_SIZE + appStore.DOCKER_APP_GAP,
		target_top
	);
};

const onDockDragLeave = async (e: any) => {
	e.preventDefault();
	if (e.target.id == 'app_dock_list') {
		checkRemove();
	}
};

const onDragOver = async (e: any) => {
	e.preventDefault();
	if (origin_app) {
		for (var i = 0; i < appStore.dockerApps.length; i++) {
			if (appStore.dockerApps[i].id == origin_app.id) {
				const draggedEl: any = document.getElementById(
					appStore.dockerApps[i].id
				);
				draggedEl.classList.add('drag-start');
			}
		}
	}
	checkMove(e);
};

const onDrop = async () => {
	isAnimation = false;

	const draggedEl: any = document.getElementById(origin_app.id);
	draggedEl?.classList.remove('drag-start');

	if (!remove_id) {
		origin_app = null;
		choose = -1;
	}
};

const onDragEnd = async () => {
	isAnimation = false;
	if (!origin_app) {
		return;
	}

	if (remove_id) {
		checkAdd(remove_top, 2, origin_app.id);
	} else {
		const draggedEl: any = document.getElementById(origin_app.id);
		draggedEl?.classList.remove('drag-start');

		origin_app = null;
		choose = -1;
	}
};

const toSearch = () => {
	emits('changeSearchDialog', true);
};

const showNotification = () => {
	notificationStore.toggleNotificaitonDisplay();
};

defineExpose({
	handleRemoveApp,
	handleOpenApp
});

const handleDocumentClick = () => {
	contextmenuId.value = '';
};

onMounted(() => {
	document.addEventListener('click', handleDocumentClick);
});

onUnmounted(() => {
	document.removeEventListener('click', handleDocumentClick);
});
</script>

<style lang="scss" scoped>
.drag-enter {
	border: 1px dashed white;
}
.drag-start {
	opacity: 0;
}

.desktop-box {
	.desktop_app_box {
		width: 60px;
		height: 100%;
		display: flex;
		align-items: center;
		position: absolute;
		top: 0px;
		left: 24px;
		z-index: 999;
		.desktop_app {
			width: 100%;
			max-height: 94%;
			border: 1px solid rgba(255, 255, 255, 0.2);
			border-radius: 16px;
			padding-top: 20px;
			padding-bottom: 20px;

			background: rgba(246, 246, 246, 0.3);
			border: 1px solid rgba(255, 255, 255, 0.2);
			filter: drop-shadow(0px 0px 40px rgba(0, 0, 0, 0.2))
				drop-shadow(0px 0px 2px rgba(0, 0, 0, 0.4));
			backdrop-filter: blur(120px);
			display: flex;
			flex-direction: column;

			.desktop_avatar_border {
				width: 36px;
				height: 0px;
				opacity: 0.3;
				border-bottom: 1px solid rgba(31, 24, 20, 1);
			}
			.desktop_user_avatar {
				display: flex;
				flex-wrap: wrap;
				justify-content: center;
				.desktop_avatar {
					width: 36px;
					height: 36px;
					box-shadow: 0px 0px 11px 0px rgba(0, 0, 0, 0.2);
					border-radius: 50%;
					margin-bottom: 12px;
					cursor: pointer;
					overflow: hidden;
				}
				.dock_icon {
					margin-top: 14px;
					width: 32px;
					height: 32px;
					border-radius: 16px;
					// overflow: hidden;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					border-radius: 18px;
					background: rgba(255, 255, 255, 0.2);
					display: flex;
					align-items: center;
					justify-content: center;
					justify-content: center;
					position: relative;
					cursor: pointer;
					img {
						width: 16px;
						height: 16px;
					}

					.badge {
						width: 8px;
						height: 8px;
						border-radius: 4px;
						position: absolute;
						right: 0px;
						top: 0px;
						background-color: #fa473b;
					}
				}
				.desktop_search {
					margin-top: 14px;
					width: 100%;
					width: 32px;
					height: 32px;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					img {
						cursor: pointer;
					}
				}
				.desktop_notify {
					margin-top: 14px;
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					img {
						cursor: pointer;
					}
				}
				.desktop_logout {
					margin-top: 14px;
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					.desktop_message_icons {
						right: -17px;
					}
					img {
						cursor: pointer;
					}
				}
			}
			.column_apps_all {
				width: 100%;
				flex: 1;
				// min-height: 200px;
				// max-height: 555px;
				// overflow: hidden;
				position: relative;
				img {
					cursor: pointer;
				}
				.img_parent {
					display: flex;
					&::after {
						content: '';
						display: block;
						width: 4px;
						height: 4px;
						background: rgba(255, 255, 255, 0.8);
						border-radius: 50%;
						margin-right: 6px;
					}
				}
				.menu_open {
					background: $background-1;
					// backdrop-filter: blur(120px);
					border-radius: 6px;
					padding: 4px;
					opacity: 0;
					animation: fadeIn 0.5s forwards;
					.context_triangle {
						position: absolute;
						top: 10px;
						left: -8px;
						width: 0;
						height: 0;
						border-top: 8px solid transparent;
						border-bottom: 8px solid transparent;
						border-right: 8px solid $background-1;
					}
					.item {
						font-size: 12px;
						line-height: 16px;
						font-family: Roboto-Regular, Roboto;
						font-weight: 400;
						color: $ink-1;
						text-align: left;
						margin: 0px;
						cursor: pointer;
						white-space: nowrap;
						padding: 4px 8px;
						border-radius: 4px;
						&:hover {
							background-color: rgba(0, 0, 0, 0.1);
						}

						&.red_font {
							color: $negative;
						}
					}
				}

				@keyframes fadeIn {
					to {
						opacity: 1;
					}
				}
				.menu_hide {
					display: none;
					opacity: 1;
					animation: fadeOut 0.5s forwards;
				}

				@keyframes fadeOut {
					to {
						opacity: 0;
					}
				}

				.column_myapps {
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					padding-top: 15px;
					cursor: pointer;
					.column_myapps_logo {
						width: 40px;
						height: 40px;
						overflow: hidden;
						box-shadow: 0px 2px 10px 0px rgba(0, 0, 0, 0.2);
						border-radius: 5px;
						margin-bottom: 16px;
						background-position: center;
						background-repeat: no-repeat;
						background-size: contain;
					}
				}
			}
		}
	}
}
</style>
