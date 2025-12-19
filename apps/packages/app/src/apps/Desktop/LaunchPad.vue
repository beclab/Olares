<template>
	<div class="launch_pad_page in_center_page" @click="dismiss">
		<div class="launch_page_mask"></div>
		<div class="launch_pad_box in-center" ref="launchpadPage">
			<div class="launch_pad_search" ref="searchBox">
				<q-input
					dense
					stack-label
					class="launch_search"
					v-model="searchVal"
					@focus="focusSearch"
					@blur="blurSearch"
					@click.stop
					@update:model-value="updateSearch"
					input-style="color: #ffffff; height: 32px;"
				>
					<template v-slot:prepend>
						<q-icon class="search_icon" name="search" size="16px" />
					</template>
					<template v-slot:append>
						<img
							v-if="searchVal"
							class="search_clean cursor-pointer"
							src="../../assets/desktop/cancel.svg"
							style="width: 20px"
							@click.stop="cleanSearchVal"
						/>
						<div v-else class="search_input">
							{{ t('launch_input_placeholder') }}
						</div>
					</template>
				</q-input>
			</div>
			<template
				v-if="appStore.launchPadApps && appStore.launchPadApps.length > 0"
			>
				<q-carousel
					v-model="slide"
					transition-prev="slide-right"
					transition-next="slide-left"
					swipeable
					animated
					class="bg-grey-1 shadow-2 rounded-borders pad_carousel"
					ref="carousel"
				>
					<q-carousel-slide
						v-for="(appList, indexList) in appStore.launchPadApps"
						:key="'desk_' + indexList"
						:name="indexList"
						class="column_launch_pad_apps column no-wrap column_none"
					>
						<div
							v-for="(element, index) in appStore.launchPadApps[indexList]"
							:key="'app_' + index"
							:style="`
									border-radius: 16px;
									position:absolute;
									z-index: 1;
									top:${appStore.desktopApps[element].top}px;
									left:${appStore.desktopApps[element].left}px;
									width:${appStore.DESKTOP_APP_SIZE}px;`"
							v-touch-hold:1200.mouse="handleHold"
							:capture="!isInDrag"
							:mouse="!isInDrag"
							:mouseCapture="!isInDrag"
							:id="appStore.desktopApps[element].id"
							@click="openWindow(appStore.desktopApps[element])"
							@dragstart="onDragStart"
							:draggable="true"
							@dragenter="onDragEnter(appStore.desktopApps[element].id, $event)"
							@dragleave="onDragLeave"
							@dragend="onDragEnd"
						>
							<div class="dragMask"></div>
							<div
								:style="
									isDisplay && !appStore.desktopApps[element].isSysApp
										? 'display: block;'
										: 'display: none;'
								"
								class="delete_launch"
								@click.stop="
									deleteLaunch(
										appStore.desktopApps[element].icon,
										appStore.desktopApps[element].title,
										element,
										$event
									)
								"
							></div>

							<div
								class="install_loading_status"
								:style="`
										width:${appStore.DESKTOP_APP_SIZE}px;
										height:${appStore.DESKTOP_APP_SIZE}px;
										border-radius: ${borderRadiusFormat(
											appStore.DESKTOP_APP_SIZE,
											appStore.DESKTOP_APP_SIZE
										)}px;
										`"
								v-if="isDoingState(appStore.desktopApps[element].fatherState)"
							>
								<svg viewBox="0 0 32 32" id="install_loading_speed">
									<circle r="16" cx="16" cy="16" />
								</svg>
							</div>
							<img
								:key="appStore.desktopApps[element].title"
								crossorigin="anonymous"
								:src="appStore.desktopApps[element].icon"
								:style="`width:${appStore.DESKTOP_APP_SIZE}px;height:${
									appStore.DESKTOP_APP_SIZE
								}px;border-radius: ${borderRadiusFormat(
									appStore.DESKTOP_APP_SIZE,
									appStore.DESKTOP_APP_SIZE
								)}px;${
									appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.NOT_READY ||
									appStore.desktopApps[element].fatherState ==
										APP_STATUS.UPGRADE.FAILED
										? 'filter: grayscale(100%) brightness(0.8)'
										: 'filter: grayscale(0%)'
								}`"
							/>
							<div
								class="launch_pad_apps_name"
								:data-index="appStore.desktopApps[element].id"
							>
								<span
									class="app_state q-mr-sm suspend_color"
									v-if="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.STOPPED
									"
								></span>
								<span
									class="app_state q-mr-sm crash_color"
									v-if="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.NOT_READY
									"
								></span>
								{{
									systemApps.includes(appStore.desktopApps[element].name)
										? t(`system_app.${appStore.desktopApps[element].title}`)
										: appStore.desktopApps[element].title
								}}
							</div>
						</div>
					</q-carousel-slide>

					<template v-slot:control>
						<q-carousel-control
							class="row items-center justify-center full-width"
							v-if="appStore.launchPadApps.length > 1"
						>
							<span
								class="carousel_dot q-mx-sm"
								:class="slide === index ? 'active' : ''"
								v-for="(dot, index) in appStore.launchPadApps"
								:key="index"
								@click.stop="goto(index)"
							></span>
						</q-carousel-control>
					</template>
				</q-carousel>
			</template>
			<div class="data_empty" v-else>{{ t('launch_no_result') }}</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { DesktopAppInfo, DesktopPosition, AppClickInfo } from './type/types';
import { useI18n } from 'vue-i18n';
import { useAppStore, systemApps } from '../../stores/desktop/app';
import ConfirmDialog from './components/ConfirmDialog.vue';
import { borderRadiusFormat } from '../../utils/desktop/utils';
import { ENTRANCE_STATUS, APP_STATUS } from '../../constant/constants';
import { isDoingState, isUninstallableState } from 'src/constant/config';

defineProps({
	isShowLaunch: {
		type: Boolean,
		required: false
	}
});

const emits = defineEmits(['appClick', 'dismiss', 'drag_launch_app']);

const { t } = useI18n();
const launchpadPage = ref<HTMLElement>();
const searchBox = ref<HTMLElement>();
onMounted(() => {
	let grid_x: number = Math.floor(launchpadPage.value!.offsetWidth / 20);
	const grid_y: number = Math.floor(launchpadPage.value!.offsetHeight / 11.5);
	grid_x = grid_x < grid_y ? grid_x : grid_y;

	appStore.DESKTOP_APP_SIZE = grid_x;
	appStore.DESKTOP_APP_X_GAP = grid_x * 1.2;
	appStore.DESKTOP_APP_Y_GAP = grid_x * 1 + 16; // maybe have some issue

	appStore.DOCKER_APP_START_Y_GAP =
		Math.floor(
			launchpadPage.value!.offsetHeight -
				appStore.DESKTOP_APP_Y_NUM * grid_x -
				(appStore.DESKTOP_APP_Y_NUM - 1) * appStore.DESKTOP_APP_Y_GAP
		) * 0.15;

	appStore.DOCKER_APP_START_X_GAP =
		Math.floor(
			launchpadPage.value!.offsetWidth -
				appStore.DESKTOP_APP_X_NUM * grid_x -
				(appStore.DESKTOP_APP_X_NUM - 1) * appStore.DESKTOP_APP_X_GAP
		) * 0.5;
	appStore.DOCKER_APP_END_X_GAP = appStore.DOCKER_APP_START_X_GAP;

	appStore.resize();
});

const $q = useQuasar();
const appStore = useAppStore();
const searchVal = ref<string>('');
const isFocus = ref<boolean>(false);

const isDisplay = ref<boolean>(false);

let slide = ref(0);
let isDelete = false;
let isInDrag = ref(false);
let chooseInUse = ref(-1);
let drag_launch_app: string | null = null;
let choose = -1;
let isAnimation = false;
let lastDragFinishTime = 0;
const carousel = ref();

const openWindow = async (item: DesktopAppInfo) => {
	if (isDoingState(item.fatherState)) {
		return;
	}
	if (
		item.state === ENTRANCE_STATUS.STOPPED ||
		item.fatherState === APP_STATUS.UPGRADE.FAILED
	) {
		$q.dialog({
			component: ConfirmDialog,
			componentProps: {
				title: t('confirmation'),
				message: t('message_desktop.suspended'),
				icon: item.icon
			}
		});
		return false;
	}

	if (item.state === ENTRANCE_STATUS.NOT_READY) {
		$q.dialog({
			component: ConfirmDialog,
			componentProps: {
				title: t('confirmation'),
				message: t('message_desktop.crashed'),
				icon: item.icon
			}
		});
		return false;
	}

	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
};

const dismiss = () => {
	if (isDelete) {
		isDelete = false;
		for (var i = 0; i < appStore.desktopApps.length; i++) {
			const draggedEl: any = document.getElementById(
				appStore.desktopApps[i].id
			);

			draggedEl?.classList.remove('vibrate-1');
			isDisplay.value = false;
		}
	} else {
		if (isFocus.value) {
			isFocus.value = false;
			searchVal.value = '';
		}
		appStore.dismiss_search_result();
		emits('dismiss');
	}
};

const handleHold = () => {
	if (isInDrag.value) {
		return;
	}

	if (isAnimation) {
		return;
	}

	let now = new Date().getTime();
	let diff = now - lastDragFinishTime;
	if (diff < 1000) {
		return;
	}

	isDelete = true;

	for (var i = 0; i < appStore.desktopApps.length; i++) {
		const draggedEl: any = document.getElementById(appStore.desktopApps[i].id);

		draggedEl?.classList.add('vibrate-1');
		isDisplay.value = true;
	}
};

function focusSearch() {
	isFocus.value = true;
}

function blurSearch() {
	isFocus.value = false;
}

function deleteLaunch(
	icon: string,
	launchTitle: string,
	index: number,
	e: any
) {
	console.log('deleteLaunch e', e);
	const message = t('message_desktop.delete_app', {
		appName: launchTitle
	});
	$q.dialog({
		component: ConfirmDialog,
		componentProps: {
			title: t('delete'),
			icon,
			message,
			showCancel: true
		}
	}).onOk(async () => {
		e.target.parentNode.classList.add('uninstallAni');
		const fatherName = appStore.desktopApps[index].fatherName;

		setTimeout(async () => {
			await appStore.uninstall_application(fatherName);
			e.target.parentNode.classList.remove('uninstallAni');
		}, 500);
	});
}

const getAppIndexInDesktop = (id: string) => {
	for (var i = 0; i < appStore.desktopApps.length; i++) {
		const draggedEl: any = document.getElementById(appStore.desktopApps[i].id);
		if (draggedEl?.id == id) {
			return i;
		}
	}
	return -1;
};

const onDragStart = async (e: any) => {
	if (isDelete) {
		return;
	}

	e.dataTransfer.setData('text', e.target.id);
	e.dataTransfer.dropEffect = 'move';

	isInDrag.value = true;
	choose = getAppIndexInDesktop(e.target.id);
	chooseInUse.value = choose;

	e.dataTransfer.effectAllowed = 'move';
	window.setTimeout(() => {
		e.target.classList.add('drag-start');
	}, 10);

	drag_launch_app = e.target.id;
	emits('drag_launch_app', e.target.id);
};

function animate(sources: number[], targets: DesktopPosition[]) {
	if (sources.length != targets.length) return;
	if (sources.length == 0) return;

	isAnimation = true;
	let p: any[] = [];
	for (var i = 0; i < targets.length - 1; ++i) {
		p.push(
			new Promise((resolve) => {
				let source = sources[i];
				let target = targets[i];

				function finish() {
					clearInterval(moveY);

					appStore.desktopApps[source].index = target.index;
					appStore.desktopApps[source].page = target.page;
					appStore.desktopApps[source].page_num = target.page_num;
					appStore.desktopApps[source].left = target.left;
					appStore.desktopApps[source].top = target.top;

					resolve(true);
				}

				let moveY = setInterval(function () {
					if (isAnimation == false) {
						finish();
						return;
					}

					if (target.left > appStore.desktopApps[source].left) {
						appStore.desktopApps[source].left += 160;
						if (target.left <= appStore.desktopApps[source].left) {
							finish();
						}
					} else if (target.left < appStore.desktopApps[source].left) {
						appStore.desktopApps[source].left -= 160;
						if (target.left >= appStore.desktopApps[source].left) {
							finish();
						}
					} else {
						finish();
					}
				}, 4);
			})
		);
	}

	Promise.all(p).then(() => {
		appStore.desktopApps[sources[targets.length - 1]].index =
			targets[targets.length - 1].index;
		appStore.desktopApps[sources[targets.length - 1]].page =
			targets[targets.length - 1].page;
		appStore.desktopApps[sources[targets.length - 1]].page_num =
			targets[targets.length - 1].page_num;
		appStore.desktopApps[sources[targets.length - 1]].left =
			targets[targets.length - 1].left;
		appStore.desktopApps[sources[targets.length - 1]].top =
			targets[targets.length - 1].top;

		isAnimation = false;
	});
}

const checkMove = (id: string) => {
	if (isAnimation) {
		return;
	}

	let dragIndex = getAppIndexInDesktop(id);

	if (choose == dragIndex || dragIndex == -1) {
		return;
	}

	let sources: number[] = [];
	let targets: DesktopPosition[] = [];

	if (
		appStore.desktopApps[choose].page != appStore.desktopApps[dragIndex].page
	) {
		return;
	}

	if (
		appStore.desktopApps[choose].index < appStore.desktopApps[dragIndex].index
	) {
		for (var i = 0; i < appStore.desktopApps.length; i++) {
			if (i == choose) {
				continue;
			}

			if (appStore.desktopApps[i].page != appStore.desktopApps[choose].page) {
				continue;
			}

			if (
				appStore.desktopApps[i].page_num >
					appStore.desktopApps[choose].page_num &&
				appStore.desktopApps[i].page_num <=
					appStore.desktopApps[dragIndex].page_num
			) {
				sources.push(i);
				targets.push(
					appStore.get_desktop_position(appStore.desktopApps[i].index - 1)
				);
			}
		}
	} else {
		for (let i = 0; i < appStore.desktopApps.length; i++) {
			if (i == choose) {
				continue;
			}

			if (appStore.desktopApps[i].page != appStore.desktopApps[choose].page) {
				continue;
			}

			if (
				appStore.desktopApps[i].page_num <
					appStore.desktopApps[choose].page_num &&
				appStore.desktopApps[i].page_num >=
					appStore.desktopApps[dragIndex].page_num
			) {
				sources.push(i);
				targets.push(
					appStore.get_desktop_position(appStore.desktopApps[i].index + 1)
				);
			}
		}
	}

	sources.push(choose);
	targets.push(
		appStore.get_desktop_position(appStore.desktopApps[dragIndex].index)
	);

	if (targets.length > 0) {
		animate(sources, targets);
	}
};

const onDragEnter = async (id: string, e: any) => {
	e.preventDefault();

	if (!drag_launch_app) {
		return;
	}

	checkMove(id);
};

const onDragLeave = async (e: any) => {
	e.target.classList.remove('drag-enter');
};

const onDragEnd = async () => {
	isInDrag.value = false;
	lastDragFinishTime = new Date().getTime();
	chooseInUse.value = -1;

	emits('drag_launch_app', null);

	if (drag_launch_app) {
		const draggedEl: any = document.getElementById(drag_launch_app);
		draggedEl?.classList.remove('drag-start');
		drag_launch_app = null;
	}
};

const updateSearch = (val: string) => {
	appStore.update_search_result(val);
};

const cleanSearchVal = () => {
	searchVal.value = '';
};

const goto = (value: number) => {
	carousel.value.goTo(value);
};
</script>

<style lang="scss" scoped>
.drag-enter {
	border: 1px dashed white;
}

.drag-start {
	opacity: 0;
}

.dialog_box {
	.dialog_card {
		width: 426px;
		min-height: 155px;
		border-radius: 8px;
	}
	.launch_dialog_span {
		width: 300px;
	}
	.launch_dialog_btn {
		position: absolute;
		bottom: 0px;
		right: 0px;
	}
	.launch_pad_dialog {
		width: 70px;
		height: 70px;
		border-radius: 16px;
	}
}
.pad_carousel {
	background: transparent !important;
	box-shadow: none !important;
	height: 90% !important;
	overflow: hidden !important;
}

.dragMask {
	width: 140%;
	height: 120%;
	transform: translate(-13%, -10%);
	position: absolute;
	top: 0;
	left: 0;
	z-index: 2;
	overflow: hidden;
	cursor: pointer;
}

.column_none {
	overflow: hidden !important;
}
.column_launch_pad_apps {
	.launch_pad_apps_name {
		width: 140%;
		transform: translateX(-15%);
		text-align: center;
		font-size: 14px;
		font-family: Roboto-Medium, Roboto;
		font-weight: 500;
		color: #ffffff;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		margin-top: 12px;
		.app_state {
			display: inline-block;
			width: 8px;
			height: 8px;
			border-radius: 4px;
			&.suspend_color {
				background-color: $warning;
			}
			&.crash_color {
				background-color: $negative;
			}
			&.running_color {
				background-color: $positive;
			}
			&.upgrade_error_color {
				background-color: $blue;
			}
		}
	}
	.delete_launch {
		width: 20px;
		height: 20px;
		background-image: url('../../assets/desktop/delete_app_icon.svg');
		background-repeat: no-repeat;
		background-position: center;
		background-size: contain;
		position: absolute;
		top: -5px;
		left: -5px;
		z-index: 100;
		cursor: pointer;
	}
	.install_loading_status {
		position: absolute;
		top: 0px;
		left: 0px;
		z-index: 99;
		background-image: url('../../assets/desktop/installing.svg');
		background-position: center;
		background-size: 100% 100%;
		background-repeat: no-repeat;
		display: flex;
		align-items: center;
		justify-content: center;
		svg {
			transform: rotate(-90deg);
			border-radius: 50%;
			height: 40%;
		}
		circle {
			fill: rgba(0, 0, 0, 0.5);
			stroke: rgba(255, 255, 255, 1);
			stroke-width: 32;
			stroke-dasharray: 0 100;
			animation: fillup 5s linear infinite;
		}
		@keyframes fillup {
			to {
				stroke-dasharray: 158 158;
				opacity: 0;
			}
		}
	}
}
.launch_pad_page {
	width: 100%;
	height: 100%;
	z-index: 9;
	position: relative;
	top: 0px;
	left: 0px;
	.launch_page_mask {
		position: absolute;
		left: 0;
		height: 0;
		z-index: -1;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(10px);
	}
	.launch_search {
		width: 240px;
		height: 32px !important;
		line-height: 32px !important;
		border-radius: 8px;
		position: relative;
		font-size: 12px !important;
		padding-left: 8px;

		border: 1px solid rgba(255, 255, 255, 0.2);
		background: rgba(246, 246, 246, 0.1);
		box-shadow: 0px 0px 40px 0px rgba(0, 0, 0, 0.2),
			0px 0px 2px 0px rgba(0, 0, 0, 0.4);
		.search_icon {
			margin-bottom: 8px;
			color: rgba(255, 255, 255, 0.8);
		}
		.search_clean {
			margin-bottom: 10px;
			color: rgba(255, 255, 255, 0.8);
		}
		.search_input {
			position: absolute;
			top: 0px;
			left: 20px;
			color: rgba(255, 255, 255, 0.6);
			width: 100%;
			height: 32px;
			margin-bottom: 2px;
			font-size: 14px;
			font-weight: 500;
			z-index: -1;
		}
	}
}
.launch_pad_box {
	width: calc(100% - 108px);
	height: 100%;
	box-shadow: none;
	margin-left: 108px;
	padding-top: 36px;
	overflow: hidden;
	.launch_pad_search {
		width: 100%;
		height: 44px;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.data_empty {
		color: #e5e5e5;
		font-size: 20px;
		margin-top: calc(50% - 180px);
		text-align: center;
	}
}

.carousel_dot {
	display: inline-block;
	width: 8px;
	height: 8px;
	background-color: rgba(255, 255, 255, 0.3);
	border-radius: 4px;
	cursor: pointer;
	&.active {
		background-color: rgba(255, 255, 255, 1);
	}
}
.in_center_page {
	-webkit-animation: puff_in_center_page 0.6s;
	animation: puff_in_center_page 0.6s;
}
@-webkit-keyframes puff_in_center_page {
	0% {
		opacity: 0;
	}
	100% {
		opacity: 1;
	}
}
@keyframes puff_in_center_page {
	0% {
		opacity: 0;
	}
	100% {
		opacity: 1;
	}
}
.in-center {
	-webkit-animation: puff-in-center 0.6s;
	animation: puff-in-center 0.6s;
}

@-webkit-keyframes puff-in-center {
	0% {
		-webkit-transform: scale(0);
		transform: scale(0);
		opacity: 0;
	}
	50% {
		-webkit-transform: scale(1.03);
		transform: scale(1.03);
		opacity: 0;
	}
	60% {
		-webkit-transform: scale(1.02);
		transform: scale(1.02);
		opacity: 0.6;
	}
	100% {
		-webkit-transform: scale(1);
		transform: scale(1);
		opacity: 1;
	}
}
@keyframes puff-in-center {
	0% {
		-webkit-transform: scale(0);
		transform: scale(0);
		opacity: 0;
	}
	50% {
		-webkit-transform: scale(1.03);
		transform: scale(1.03);
		opacity: 0;
	}
	60% {
		-webkit-transform: scale(1.02);
		transform: scale(1.02);
		opacity: 0.6;
	}

	100% {
		-webkit-transform: scale(1);
		transform: scale(1);
		opacity: 1;
	}
}
:global(.q-field--standard .q-field__control:before) {
	border: none;
}
:global(.q-field--standard .q-field__control:after) {
	height: 0px;
}
:global(.q-field__native) {
	font-family: Roboto-Medium, Roboto;
	font-weight: 500;
	// color: #ffffff;
}
:global(.q-field__marginal) {
	height: 100%;
	padding-right: 4px !important;
}
.q-field__marginal :global(.q-field__control) {
	height: 44px;
}
.ghost {
	opacity: 0 !important;
}
.vibrate-1 {
	-webkit-animation: vibrate-1 0.3s linear infinite both;
	animation: vibrate-1 0.3s linear infinite both;
}

@-webkit-keyframes vibrate-1 {
	0% {
		transform: rotate(4deg);
		-webkit-transform: rotate(4deg);
	}

	50% {
		transform: rotate(-3deg);
		-webkit-transform: rotate(-3deg);
	}
	100% {
		transform: rotate(4deg);
		-webkit-transform: rotate(4deg);
	}
}
@keyframes vibrate-1 {
	0% {
		transform: rotate(4deg);
		-webkit-transform: rotate(4deg);
	}

	50% {
		transform: rotate(-3deg);
		-webkit-transform: rotate(-3deg);
	}
	100% {
		transform: rotate(4deg);
		-webkit-transform: rotate(4deg);
	}
}

.uninstallAni {
	display: inline-block;
	animation: scaleFadeOut 0.5s forwards ease-in-out;
}

@keyframes scaleFadeOut {
	0% {
		transform: scale(1);
		opacity: 1;
	}
	20% {
		transform: scale(1.2);
		opacity: 1;
	}
	100% {
		transform: scale(0);
		opacity: 0;
	}
}

:global(.q-carousel__navigation-icon) {
	font-size: 5px !important;
}
</style>
