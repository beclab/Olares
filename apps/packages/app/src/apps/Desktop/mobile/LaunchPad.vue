<template>
	<div class="Launch_pad_page in-center-page" @click="dismiss">
		<div class="launch_pad_box in-center" ref="launchpadPage">
			<template
				v-if="appStore.launchPadApps && appStore.launchPadApps.length > 0"
			>
				<q-carousel
					v-model="slide"
					transition-prev="slide-right"
					transition-next="slide-left"
					swipeable
					animated
					ref="carousel"
					class="bg-grey-1 shadow-2 rounded-borders q_vackgr_carousel"
				>
					<q-carousel-slide
						v-for="(appList, Indexlist) in appStore.launchPadApps"
						:key="'deskp0' + Indexlist"
						:name="Indexlist"
						class="column_launchpadapps column no-wrap column_none"
					>
						<div
							class="row items-center justify-center"
							v-for="(element, index) in appStore.launchPadApps[Indexlist]"
							:key="'adod+index' + index"
							style="
								border-radius: 16px;
								-webkit-touch-callout: none;
								-webkit-user-select: none;
								-khtml-user-select: none;
								-moz-user-select: none;
								-ms-user-select: none;
								user-select: none;
								position: relative;
							"
							v-touch-hold:1200.mouse="handleHold"
							:id="appStore.desktopApps[element].id"
							@click="openWindow(appStore.desktopApps[element])"
							@contextmenu.prevent
						>
							<!-- <div class="dragMask"></div> -->
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
										Indexlist
									)
								"
							></div>

							<div
								class="install_loading_status"
								:style="`
										width:58px;
										height:58px;
										`"
								v-if="isDoingState(appStore.desktopApps[element].fatherState)"
							>
								<svg viewBox="0 0 32 32" id="install_loading_speed">
									<circle r="16" cx="16" cy="16" />
								</svg>
							</div>

							<img
								class="pad-img"
								:key="appStore.desktopApps[element].title"
								crossorigin="anonymous"
								:src="appStore.desktopApps[element].icon"
								:style="`border-radius: ${borderRadiusFormat(
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
								class="launchpadapps_name"
								:data-index="appStore.desktopApps[element].id"
							>
								<span
									class="app_state q-mr-xs suspend_color"
									v-if="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.STOPPED
									"
								></span>
								<span
									class="app_state q-mr-xs crash_color"
									v-if="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.NOT_READY
									"
								></span>
								{{ appStore.desktopApps[element].title }}
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
			<div class="no-result" v-else>{{ t('launch_no_result') }}</div>

			<div class="close row items-center justify-center" @click="dismiss">
				<img src="../../../assets/desktop/close-icon.svg" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { DesktopAppInfo, AppClickInfo } from '../type/types';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '../../../stores/desktop/app';
import ConfirmDialog from '../components/ConfirmDialog.vue';
import { borderRadiusFormat } from '../../../utils/desktop/utils';
import { isDoingState } from 'src/constant/config';
import { APP_STATUS, ENTRANCE_STATUS } from 'src/constant/constants';

defineProps({
	isShowLaunc: {
		type: Boolean,
		required: false
	}
});

const emits = defineEmits(['appClick', 'dismiss', 'drag_launch_app']);

const { t } = useI18n();
const launchpadPage = ref<HTMLElement>();

const $q = useQuasar();
const appStore = useAppStore();
const searchVal = ref<string>('');
const isFocus = ref<boolean>(true);

const isDisplay = ref<boolean>(false);
const carousel = ref();

let slide = ref(0);
let isDelete = false;
let lastDragFinishTime = 0;

const openWindow = async (item: DesktopAppInfo) => {
	console.log('openWindow', item);
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
			appStore.dismiss_search_result();
			isFocus.value = false;
			searchVal.value = '';
		}
		emits('dismiss');
	}
};

const handleHold = () => {
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

function deleteLaunch(
	launchlogo: string,
	launchTitle: string,
	index: number,
	Indexlist: number
) {
	const message = t('message_desktop.delete_app', {
		appName: launchTitle
	});

	$q.dialog({
		component: ConfirmDialog,
		componentProps: {
			title: t('delete'),
			icon: launchlogo,
			message,
			showCancel: true
		}
	}).onOk(async () => {
		const fatherName = appStore.desktopApps[index].fatherName;
		let categoryLaunchPadApps = appStore.launchPadApps[Indexlist];
		appStore.launchPadApps[Indexlist] = categoryLaunchPadApps.filter(
			(itme, indexs) => indexs !== index
		);

		await appStore.uninstall_application(fatherName);
		await appStore.get_my_apps_info(true);
	});
}

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
.q_vackgr_carousel {
	background: transparent !important;
	box-shadow: none !important;
	height: 90% !important;
	overflow: hidden !important;
}

.dragMask {
	width: 140%;
	height: 120%;
	transform: translate(-14%, -10%);
	position: absolute;
	top: 0;
	left: 0;
	z-index: 2;
	border-radius: 10px;
	overflow: hidden;
	cursor: pointer;
}

.column_none {
	overflow: hidden !important;
}
.column_launchpadapps {
	display: grid;
	grid-template-rows: 80px 80px 80px 80px;
	grid-template-columns: repeat(4, 25%);
	grid-row-gap: 30px;
	padding-top: 60px;
	.pad-img {
		width: 58px;
		height: 58px;
		-webkit-touch-callout: none;
		-webkit-user-select: none;
		-webkit-user-drag: none;
	}
	.launchpadapps_name {
		width: 100%;
		text-align: center;
		font-family: Roboto-Medium, Roboto;
		font-size: 12px;
		font-style: normal;
		font-weight: 400;
		line-height: 16px;
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
		width: 36px;
		height: 36px;
		background-image: url('../../../assets/desktop/delete_app_icon.svg');
		background-repeat: no-repeat;
		background-position: center;
		background-size: 60% 60%;
		position: absolute;
		top: -14px;
		left: 7px;
		z-index: 99;
		cursor: pointer;
	}
	.install_loading_status {
		position: absolute;
		top: 0px;
		left: calc((100% - 58px) / 2);
		z-index: 99;
		background-image: url('../../../assets/desktop/installing.svg');
		background-position: center;
		background-size: 100% 100%;
		background-repeat: no-repeat;
		border-radius: 16px;
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
.Launch_pad_page {
	width: 100%;
	height: 100%;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(10px);
	z-index: 9;
	position: absolute;
	top: 0px;
	left: 0px;
	touch-action: none;
}
.launch_pad_box {
	width: 100%;
	height: 100%;
	box-shadow: none;
	overflow: hidden;
	.launch_pad_APPs {
		width: 100%;
		display: flex;
		flex-wrap: wrap;
		.launch_pad_Dragg {
			width: 100%;
			display: flex;
			flex-wrap: wrap;
			align-content: flex-start;
			.launch_pad_drag {
				width: calc(100% / 7);
				margin-bottom: 48px;
				cursor: pointer;
			}
		}

		.launch_myapps_box {
			width: 70%;
			.contain_img {
				display: flex;
				justify-content: center;
				align-items: center;

				.animation_delete {
					width: 70px;
					position: relative;
				}
			}

			.launch_pad_myapps_logo {
				display: flex;
				width: 70px;
				height: 70px;
				background: #ffffff;
				box-shadow: 0px 2px 12px 0px rgba(0, 0, 0, 0.2);
				border-radius: 16px;
			}
			.launch_myapps_text {
				width: 100%;
				font-size: 14px;
				font-family: Roboto-Medium, Roboto;
				font-weight: 500;
				color: #ffffff;
				margin-top: 12px;
				text-align: center;
			}
		}
	}

	.no-result {
		color: #e5e5e5;
		font-size: 20px;
		margin-top: calc(50% - 180px);
		text-align: center;
	}

	.close {
		position: absolute;
		top: calc(100vh - 188px);
		left: 0;
		right: 0;
		margin: auto;
		width: 56px;
		height: 56px;
		border-radius: 28px;
		background-color: rgba(255, 255, 255, 0.9);
	}
}
.in-center-page {
	-webkit-animation: puff-in-center-page 0.6s;
	animation: puff-in-center-page 0.6s;
}
@-webkit-keyframes puff-in-center-page {
	0% {
		opacity: 0;
	}
	100% {
		opacity: 1;
	}
}
@keyframes puff-in-center-page {
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
		transform: rotate(6deg);
		-webkit-transform: rotate(6deg);
	}

	50% {
		transform: rotate(-5deg);
		-webkit-transform: rotate(-5deg);
	}
	100% {
		transform: rotate(6deg);
		-webkit-transform: rotate(6deg);
	}
}
@keyframes vibrate-1 {
	0% {
		transform: rotate(6deg);
		-webkit-transform: rotate(6deg);
	}

	50% {
		transform: rotate(-5deg);
		-webkit-transform: rotate(-5deg);
	}
	100% {
		transform: rotate(6deg);
		-webkit-transform: rotate(6deg);
	}
}

:global(.q-carousel__navigation-icon) {
	font-size: 5px !important;
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
</style>
