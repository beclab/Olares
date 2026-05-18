<template>
	<div
		class="Launch_pad_page in-center-page"
		@touchstart.stop
		@touchmove.stop
		@click.prevent="dismiss"
	>
		<div
			class="launch_pad_box in-center column items-center"
			ref="launchpadPage"
		>
			<div class="launch_pad_search" @click.stop>
				<q-input
					ref="searchInputRef"
					dense
					stack-label
					class="launch_search"
					v-model="searchVal"
					@focus="focusSearch"
					@blur="blurSearch"
					@update:model-value="updateSearch"
					input-style="color: var(--q-ink-on-brand) "
				>
					<template v-slot:prepend>
						<q-icon class="search_icon" name="search" size="16px" />
					</template>
					<template v-slot:append>
						<img
							v-if="searchVal"
							class="search_clean cursor-pointer"
							src="../../../assets/desktop/cancel.svg"
							style="width: 20px"
							@click="cleanSearchVal"
						/>
						<div v-else class="search_input">
							{{ t('launch_input_placeholder') }}
						</div>
					</template>
				</q-input>
			</div>
			<div
				v-if="appStore.launchPadApps && appStore.launchPadApps.length > 0"
				ref="launchPadAppsEl"
				class="launch_pad_APPs"
			>
				<q-carousel
					v-model="slide"
					transition-prev="slide-right"
					transition-next="slide-left"
					swipeable
					:animated="carouselAnimated"
					ref="carousel"
					class="bg-grey-1 shadow-2 rounded-borders q_vackgr_carousel"
				>
					<q-carousel-slide
						v-for="(appList, Indexlist) in appStore.launchPadApps"
						:key="'deskp0' + Indexlist"
						:name="Indexlist"
						class="column_launchpadapps column no-wrap column_none"
						:style="slideGridStyle"
					>
						<div
							class="row items-center justify-center pad-app-item"
							:class="isDisplay ? 'vibrate-1' : ''"
							v-for="element in appStore.launchPadApps[Indexlist]"
							:key="appStore.desktopApps[element].id"
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
							:id="appStore.desktopApps[element].id"
						>
							<div
								:style="
									isDisplay && !appStore.desktopApps[element].isSysApp
										? 'display: block;'
										: 'display: none;'
								"
								class="delete_launch"
								@click.stop="
									deleteLaunch(appStore.desktopApps[element], $event)
								"
							/>
							<div class="relative-position" style="font-size: 0px">
								<div
									class="install_loading_status"
									v-if="isDoingState(appStore.desktopApps[element].fatherState)"
								>
									<svg viewBox="0 0 32 32" id="install_loading_speed">
										<circle r="16" cx="16" cy="16" />
									</svg>
								</div>

								<img
									v-touch-hold:1200.mouse="handleHold"
									:id="appStore.desktopApps[element].id"
									@click.stop="openWindow(appStore.desktopApps[element])"
									@contextmenu.prevent
									draggable="false"
									class="pad-img"
									:key="
										appStore.desktopApps[element].id +
										'-' +
										appStore.desktopApps[element].fatherState
									"
									:src="appStore.desktopApps[element].icon"
									:style="`border-radius: 15px;${
										appStore.desktopApps[element].state ==
											ENTRANCE_STATUS.NOT_READY ||
										appStore.desktopApps[element].fatherState ==
											APP_STATUS.UPGRADE.FAILED
											? 'filter: grayscale(100%) brightness(0.8)'
											: 'filter: grayscale(0%)'
									}`"
								/>
							</div>
							<div
								class="launchpadapps_name"
								:data-index="appStore.desktopApps[element].id"
								@click.stop
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
			</div>
			<div class="no-result absolute-center" v-else>
				{{ t('launch_no_result') }}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import ConfirmDialog from '../components/ConfirmDialog.vue';
import { APP_STATUS, ENTRANCE_STATUS } from 'src/constant/constants';
import { useApplicationStore } from 'src/stores/desktop/app';
import { DesktopAppInfo, AppClickInfo } from '../type/types';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { AppService } from 'src/stores/market/appService';
import { useAppStore } from 'src/stores/market/appStore';
import { isDoingState } from 'src/constant/config';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { onBeforeUnmount, ref } from 'vue';
import { useMobileLaunchpad } from 'src/application/mobile';
import { useLaunchpadSearch } from 'src/application/launchpadSearch';
import { useStableViewportHeight } from 'src/composables/useStableViewportHeight';
import { borderRadiusFormat } from 'src/utils/desktop/utils';

defineProps({
	isShowLaunc: {
		type: Boolean,
		required: false
	}
});

const emits = defineEmits(['appClick', 'dismiss', 'drag_launch_app']);

const { t } = useI18n();
const launchpadPage = ref<HTMLElement>();
const launchPadAppsEl = ref<HTMLElement | null>(null);

const $q = useQuasar();
const appStore = useApplicationStore();

const isDisplay = ref<boolean>(false);
const carousel = ref();
const searchInputRef = ref<{ blur?: () => void } | null>(null);
let slide = ref(0);

const {
	searchVal,
	isFocus,
	carouselAnimated,
	focusSearch,
	blurSearch,
	updateSearch,
	cleanSearchVal
} = useLaunchpadSearch(slide, searchInputRef);

const { stableHeight, keyboardOpen } = useStableViewportHeight();

const { slideGridStyle } = useMobileLaunchpad(
	launchPadAppsEl,
	carousel,
	slide,
	isFocus,
	keyboardOpen
);

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
		isDisplay.value = false;
	} else {
		if (isFocus.value) {
			cleanSearchVal();
			isFocus.value = false;
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

	isDisplay.value = true;
};

function deleteLaunch(appInfo: DesktopAppInfo, e: any) {
	const marketAppStore = useAppStore();
	const marketApp = marketAppStore.findAppByName(
		appInfo && appInfo.fatherName ? appInfo.fatherName : ''
	);
	if (marketApp) {
		e.target.parentNode.classList.add('uninstallAni');
		setTimeout(async () => {
			await AppService.uninstallApp(
				marketApp.status,
				{
					app_name: marketApp.appId,
					source: marketApp.sourceId,
					version: marketApp.version
				},
				$q
			);
			e.target.parentNode.classList.remove('uninstallAni');
		}, 500);
	} else {
		notifyFailed('Failed to retrieve app information');
	}
}

const goto = (value: number) => {
	carousel.value.goTo(value);
};

onBeforeUnmount(() => {
	cleanSearchVal();
	updateSearch('');
});
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

.launch_pad_APPs {
	position: absolute;
	top: 104px;
	left: 0px;
	right: 0px;
	bottom: auto;
	height: calc(var(--stable-vh, 100svh) - 104px);
	box-sizing: border-box;

	::v-deep(.q-carousel__control) {
		margin-bottom: 43px !important;
	}
	::v-deep(.q-carousel__slide) {
		padding: 0px 6px !important;
	}
	::v-deep(.q-panel) {
		overflow: unset;
		padding-top: 10px;
	}
}

.q_vackgr_carousel {
	background: transparent !important;
	box-shadow: none !important;
	height: 100% !important;
	overflow: hidden !important;
	touch-action: pan-x;
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
	// overflow: hidden !important;
}

.column_launchpadapps {
	display: grid;
	grid-template-rows: repeat(var(--lp-rows, 4), var(--lp-row-track));
	grid-template-columns: repeat(var(--lp-cols, 4), minmax(0, 1fr));
	grid-row-gap: 30px;

	.pad-img {
		width: 58px;
		height: 58px;
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
		margin-top: 6px;

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
		inset: 0px;
		z-index: 98;
		background-image: url('../../../assets/desktop/installing.svg');
		background-position: center;
		background-size: contain;
		background-repeat: no-repeat;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 15px;

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
	min-height: var(--stable-vh, 100svh);
	height: var(--stable-vh, 100svh);

	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(10px);
	z-index: 9;
	position: absolute;
	top: 0px;
	left: 0px;
	right: 0px;
	touch-action: manipulation;
	overscroll-behavior: contain;

	.launch_search {
		width: 100% !important;
		height: 40px !important;
		line-height: 40px !important;
		border-radius: 8px;
		position: relative;
		font-size: 12px !important;
		padding-left: 8px;

		border: 1px solid rgba(255, 255, 255, 0.2);
		background: rgba(246, 246, 246, 0.1);
		box-shadow: 0px 0px 40px 0px rgba(0, 0, 0, 0.2),
			0px 0px 2px 0px rgba(0, 0, 0, 0.4);

		.search_icon {
			color: rgba(255, 255, 255, 0.8);
		}

		.search_clean {
			color: rgba(255, 255, 255, 0.8);
		}

		.search_input {
			position: absolute;
			top: 0px;
			left: 20px;
			color: rgba(255, 255, 255, 0.6);
			width: 100%;
			height: 38px;
			margin-bottom: 2px;
			font-size: 14px;
			font-weight: 500;
			z-index: -1;
		}
	}
}

.launch_pad_box {
	width: 100%;
	height: 100%;
	box-shadow: none;
	overflow: hidden;

	.launch_pad_search {
		width: 64%;
		display: flex;
		justify-content: center;
		align-items: center;
		position: relative;
		z-index: 10;
		margin-top: 32px;
	}

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
}

.close {
	position: absolute;
	top: 0;
	right: 0;
	margin: auto;
	width: 52px;
	height: 52px;
	margin-top: -20px;
	margin-right: -20px;
	border-radius: 50%;

	img {
		margin-top: 12px;
		margin-right: 12px;
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
	height: 38px;
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
