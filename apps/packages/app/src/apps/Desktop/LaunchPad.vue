<template>
	<div class="launch_pad_page" @click="dismiss">
		<div class="launch_page_mask in_center_page"></div>
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
					:swipeable="!isInDrag"
					:animated="carouselAnimated"
					class="bg-grey-1 shadow-2 rounded-borders pad_carousel"
					ref="carousel"
				>
					<q-carousel-slide
						v-for="(appList, indexList) in appStore.launchPadApps"
						:key="'desk_' + indexList"
						:name="indexList"
						class="column_launch_pad_apps column no-wrap column_none"
					>
						<div class="lp-slide-surface" :data-lp-slide="indexList">
							<div
								v-for="element in appStore.launchPadApps[indexList]"
								:key="appStore.desktopApps[element].id"
								class="lp-launch-pad-slot"
								:class="[
									isDisplay && !isInDrag ? 'vibrate-1' : '',
									dragLaunchAppId === appStore.desktopApps[element].id &&
									launchPadDragSlotHidden
										? 'drag-start'
										: '',
									dragLaunchAppId === appStore.desktopApps[element].id
										? 'dragging-pause-shake'
										: ''
								]"
								:style="`
									top:${appStore.desktopApps[element].top}px;
									left:${appStore.desktopApps[element].left}px;
									width:${appStore.DESKTOP_APP_SIZE}px;
									transition: left ${iconPosTransitionMs}ms ease, top ${iconPosTransitionMs}ms ease;`"
								v-touch-hold:1200.mouse="handleHold"
								:capture="!isInDrag"
								:mouse="!isInDrag"
								:mouseCapture="!isInDrag"
								:id="appStore.desktopApps[element].id"
								@mousedown.stop
								@click="openWindow(appStore.desktopApps[element])"
								@dragstart="onDragStart"
								:draggable="true"
								@dragenter="
									onDragEnter(appStore.desktopApps[element].id, $event)
								"
								@dragend="onDragEnd"
							>
								<LaunchpadAppTile
									:app-id="appStore.desktopApps[element].id"
									:kind="appStore.desktopApps[element].kind"
									:icon="appStore.desktopApps[element].icon"
									:label="getAppDisplayTitle(appStore.desktopApps[element], t)"
									:tile-size-px="appStore.DESKTOP_APP_SIZE"
									:show-delete-button="
										isDisplay &&
										!appStore.desktopApps[element].isSysApp &&
										(appStore.desktopApps[element].kind === 'bookmark' ||
											(!isDoingState(
												appStore.desktopApps[element].fatherState
											) &&
												isUninstallableState(
													appStore.desktopApps[element].fatherState
												)))
									"
									:show-install-loading="
										isDoingState(appStore.desktopApps[element].fatherState)
									"
									:show-stopped-state="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.STOPPED
									"
									:show-crash-state="
										appStore.desktopApps[element].state ==
										ENTRANCE_STATUS.NOT_READY
									"
									:icon-is-grayscale="
										appStore.desktopApps[element].state ==
											ENTRANCE_STATUS.NOT_READY ||
										appStore.desktopApps[element].fatherState ==
											APP_STATUS.UPGRADE.FAILED
									"
									@delete-click="
										deleteLaunch(appStore.desktopApps[element], $event)
									"
								/>
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
								@dragenter.prevent="onDotDragEnter(index)"
								@dragleave.prevent="onDotDragLeave"
								@dragover.prevent
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
import { ref, computed, onBeforeUnmount } from 'vue';
import { useQuasar } from 'quasar';
import { DesktopAppInfo, AppClickInfo } from './type/types';
import { useI18n } from 'vue-i18n';
import {
	useApplicationStore,
	getAppDisplayTitle
} from 'src/stores/desktop/app';
import { BOOKMARK_RAW_PREFIX } from 'src/types/desktop-layout';
import ConfirmDialog from './components/ConfirmDialog.vue';
import LaunchpadAppTile from './LaunchpadAppTile.vue';
import { ENTRANCE_STATUS, APP_STATUS } from 'src/constant/constants';
import { isDoingState, isUninstallableState } from 'src/constant/config';
import { AppService } from 'src/services/market/appOperation/appService';
import { useAppStore } from 'src/stores/market/appStore';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useLaunchpadSearch } from 'src/application/launchpadSearch';
import { useLaunchpadMetrics } from 'src/application/useLaunchpadMetrics';
import { useLaunchpadDragReorder } from 'src/application/useLaunchpadDragReorder';

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

const $q = useQuasar();
const appStore = useApplicationStore();

const slide = ref(0);

const {
	searchVal,
	isFocus,
	carouselAnimated,
	focusSearch,
	blurSearch,
	updateSearch,
	cleanSearchVal,
	exitSearchForDrag,
	applyPendingSlideRestoreAfterDrag,
	clearPendingSlideAfterSearchDrag,
	resumeCarouselAfterSearchDrag
} = useLaunchpadSearch(slide);
carouselAnimated.value = false;

const carousel = ref();

const { iconPosTransitionMs, dispose: disposeLaunchpadMetrics } =
	useLaunchpadMetrics(launchpadPage, appStore, carouselAnimated);

const {
	dragLaunchAppId,
	launchPadDragSlotHidden,
	isInDrag,
	onDragStart,
	onDragEnd,
	onDragEnter,
	onDotDragEnter,
	onDotDragLeave,
	goto,
	lastDragFinishTime,
	dispose: disposeLaunchpadDrag
} = useLaunchpadDragReorder({
	slide,
	carousel,
	carouselAnimated,
	launchpadMount: launchpadPage,
	appStore,
	exitSearchForDrag,
	applyPendingSlideRestoreAfterDrag,
	clearPendingSlideAfterSearchDrag,
	resumeCarouselAfterSearchDrag,
	onEmitDrag: (id) => emits('drag_launch_app', id)
});

onBeforeUnmount(() => {
	disposeLaunchpadDrag();
	appStore.launchpadShakeMode = false;
});
onBeforeUnmount(() => {
	disposeLaunchpadMetrics();
});

const isDisplay = computed<boolean>({
	get: () => appStore.launchpadShakeMode,
	set: (v) => {
		appStore.launchpadShakeMode = v;
	}
});

const openWindow = async (item: DesktopAppInfo) => {
	if (isDisplay.value) {
		isDisplay.value = false;
		return;
	}
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
				icon: item.icon,
				okTitle: t('i_got_it')
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
				icon: item.icon,
				okTitle: t('i_got_it')
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
	if (isDisplay.value) {
		isDisplay.value = false;
		return;
	}
	if (isFocus.value) {
		isFocus.value = false;
		searchVal.value = '';
	}
	appStore.dismiss_search_result();
	emits('dismiss');
};

const handleHold = () => {
	if (isInDrag.value) {
		return;
	}

	const now = new Date().getTime();
	const diff = now - lastDragFinishTime.value;
	if (diff < 300) {
		return;
	}

	isDisplay.value = true;
};

async function deleteLaunch(appInfo: DesktopAppInfo, e: MouseEvent) {
	if (appInfo.kind === 'bookmark') {
		const bookmarkId = appInfo.id.startsWith('bdesk:' + BOOKMARK_RAW_PREFIX)
			? appInfo.id.slice(('bdesk:' + BOOKMARK_RAW_PREFIX).length)
			: '';
		if (!bookmarkId) {
			return;
		}
		$q.dialog({
			component: ConfirmDialog,
			componentProps: {
				title: t('desktop_settings.delete_bookmark_title'),
				message: t('desktop_settings.delete_bookmark_msg', {
					name: appInfo.title || appInfo.name
				}),
				showCancel: true,
				persistent: true
			}
		}).onOk(async () => {
			await appStore.deleteBookmarkFromDesktop(bookmarkId);
		});
		return;
	}

	const marketAppStore = useAppStore();
	const marketApp = marketAppStore.findAppByName(
		appInfo && appInfo.fatherName ? appInfo.fatherName : ''
	);

	if (
		isDoingState(appInfo.fatherState) ||
		!isUninstallableState(appInfo.fatherState)
	) {
		return;
	}

	if (marketApp) {
		const succeeded = await AppService.uninstallApp(
			marketApp.status,
			{
				app_name: marketApp.appId,
				source: marketApp.sourceId,
				version: marketApp.version,
				deleteData: false
			},
			$q
		);
		if (succeeded) {
			appStore.uninstall_application(appInfo.fatherName);
		}
	} else {
		notifyFailed(t('error.app_info_get_failure'));
	}
}
</script>

<style lang="scss" scoped>
$ease-out-cubic: cubic-bezier(0.25, 0.1, 0.25, 1);
$launch-pad-animation-duration: 0.6s;

.drag-enter {
	border: 1px dashed white;
}

.drag-start {
	opacity: 0;
}

.dragging-pause-shake {
	animation: none !important;
	-webkit-animation: none !important;
	transform: none !important;
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

.column_none {
	overflow: hidden !important;
}

.column_launch_pad_apps {
	position: relative;

	.lp-slide-surface {
		position: relative;
		flex: 1 1 auto;
		width: 100%;
		min-height: 0;
	}
}

.lp-launch-pad-slot {
	border-radius: 16px;
	position: absolute;
	z-index: 1;
	box-sizing: border-box;
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
	animation: puff_in_center_page $launch-pad-animation-duration $ease-out-cubic
		forwards;
}

@keyframes puff_in_center_page {
	0% {
		background: rgba(0, 0, 0, 0);
		backdrop-filter: blur(0px);
	}
	100% {
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(10px);
	}
}

.in-center {
	transform-origin: center;
	animation: puff-in-center $launch-pad-animation-duration $ease-out-cubic
		forwards;
}

@keyframes puff-in-center {
	0% {
		transform: scale(1.03);
		opacity: 0;
	}

	100% {
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
