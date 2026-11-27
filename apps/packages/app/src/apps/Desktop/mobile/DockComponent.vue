<template>
	<div class="desktop-box">
		<div
			class="desktop-app-box"
			:class="{ dockAni: appStore.dockerApps.length > 0 }"
		>
			<div class="dock-track">
				<template v-for="element in appStore.dockerApps" :key="element.id">
					<div class="launch-app column items-center">
						<div
							class="app-img"
							@click="openWindow(element)"
							:id="element.id"
							:style="{
								width: element.id === 'bdock:launchpad' ? '56px' : '52px',
								height: element.id === 'bdock:launchpad' ? '56px' : '52px'
							}"
						>
							<img :src="element.icon" />
							<div class="overlay"></div>
						</div>

						<div
							class="app-name q-mt-xs"
							v-if="element.id !== 'bdock:launchpad'"
						>
							{{ appTitle(element) }}
						</div>
					</div>
				</template>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { DockerAppInfo, AppClickInfo } from '../type/types';
import { useApplicationStore } from '../../../stores/desktop/app';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import ConfirmDialog from '../components/ConfirmDialog.vue';
import { ENTRANCE_STATUS, APP_STATUS } from 'src/constant/constants';
import { isDoingState } from 'src/constant/config';
import { getAppDisplayTitle } from 'src/stores/desktop/app';

defineProps({
	drag_launch_app: {
		type: String,
		required: true
	}
});

const emits = defineEmits(['appClick']);

const appStore = useApplicationStore();
const $q = useQuasar();
const { t } = useI18n();

const appTitle = (element: DockerAppInfo) => {
	const app = appStore.resolveAppForOpen(element.id);
	return getAppDisplayTitle(
		{ name: element.name, title: element.title, isSysApp: app?.isSysApp },
		t
	);
};

const openWindow = async (item: DockerAppInfo) => {
	// A dock entry only stores layout data, not the live app state. Resolve the
	// backing app so a paused / crashed / installing app opened from the dock
	// shows the same dialog the launchpad shows, instead of opening a window
	// that fails to load.
	const app = appStore.resolveAppForOpen(item.id);
	if (app) {
		if (isDoingState(app.fatherState)) {
			return;
		}
		if (
			app.state === ENTRANCE_STATUS.STOPPED ||
			app.fatherState === APP_STATUS.UPGRADE.FAILED
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
			return;
		}
		if (app.state === ENTRANCE_STATUS.NOT_READY) {
			$q.dialog({
				component: ConfirmDialog,
				componentProps: {
					title: t('confirmation'),
					message: t('message_desktop.crashed'),
					icon: item.icon,
					okTitle: t('i_got_it')
				}
			});
			return;
		}
	}

	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
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

defineExpose({
	handleOpenApp
});
</script>

<style lang="scss" scoped>
.desktop-box {
	.desktop-app-box {
		width: 100%;
		display: flex;
		position: fixed;
		bottom: 76px;
		left: 0;
		right: 0;
		z-index: 9;
		// Allow horizontal scrolling when the shared dock layout (synced with PC)
		// holds more apps than fit one mobile row, instead of overflowing the
		// viewport and overlapping icons/names.
		overflow-x: auto;
		overflow-y: hidden;
		-webkit-overflow-scrolling: touch;
		scrollbar-width: none;
		&::-webkit-scrollbar {
			display: none;
		}

		// `flex: 1 0 auto` lets the track stretch to fill the viewport width when
		// the dock holds few apps (so icons spread edge-to-edge via
		// space-between), while its min-content min-width still lets it grow
		// beyond the viewport and scroll horizontally when there are many apps.
		.dock-track {
			display: flex;
			align-items: flex-start;
			justify-content: space-between;
			gap: 12px;
			flex: 1 0 auto;
			padding: 0 16px;
		}

		.launch-app {
			flex: 0 0 auto;
			width: 56px;

			.app-img {
				width: 52px;
				height: 52px;
				border-radius: 14px;
				-webkit-touch-callout: none;
				-webkit-user-select: none;
				-webkit-user-drag: none;
				overflow: hidden;
				position: relative;
				img {
					width: 100%;
					height: 100%;
				}
				.overlay {
					width: 100%;
					height: 100%;
					position: absolute;
					top: 0;
					left: 0;
					right: 0;
					bottom: 0;
					background: transparent;
				}
			}
			.app-name {
				max-width: 56px;
				height: 18px;
				border-radius: 14px;
				font-size: 14px;
				font-style: normal;
				font-weight: 400;
				line-height: 16px;
				color: #fff;
				text-align: center;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}
	}
}

// .dockAni {
// 	// animation: DockAni 0.5s ease-in-out;
// }

// @keyframes DockAni {
// 	from {
// 		bottom: 66px;
// 		opacity: 0;
// 	}
// 	to {
// 		bottom: 76px;
// 		opacity: 1;
// 	}
// }
</style>
