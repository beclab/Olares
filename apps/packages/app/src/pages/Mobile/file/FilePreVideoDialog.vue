<template>
	<q-dialog
		v-model="dialog"
		persistent
		:maximized="maximizedToggle"
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<div id="video-previewer">
			<div
				class="header row items-center justify-between q-px-sm"
				:class="{ hidden: isTitleHidden }"
			>
				<div class="row items-center justify-between flex-1" @click="close">
					<q-icon name="sym_r_keyboard_arrow_left" size="28px" color="white" />
					<div class="title text-ink-on-brand text-h5 flex-1">
						{{ filesStore.previewItem[origin_id].name }}
					</div>
				</div>

				<!-- <div>
					<q-btn dense flat color="white" icon="sym_r_cast" v-close-popup>
					</q-btn>
					<q-btn dense flat color="white" icon="sym_r_music_note" v-close-popup>
					</q-btn>
					<q-btn dense flat color="white" icon="sym_r_more_vert" v-close-popup>
					</q-btn>
				</div> -->
			</div>
			<div class="content">
				<terminus-video-player
					v-if="raw"
					:req="filesStore.previewItem[origin_id]"
					:path="path"
					id="videoPlayer"
					platform="mobile"
					:src="common().videoPlayUrl(raw, node)"
					@holdShowTitle="holdShowTitle"
				/>
			</div>
		</div>
	</q-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useDataStore } from '../../../stores/data';
import { useFilesStore } from '../../../stores/files';
import TerminusVideoPlayer from '../../../components/common/TerminusVideoPlayer.vue';
import { FilesIdType } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';

import { onMounted, ref, onBeforeMount, onBeforeUnmount } from 'vue';
import { StatusBar } from '@capacitor/status-bar';
import { useQuasar } from 'quasar';
import { ScreenOrientation } from '@capacitor/screen-orientation';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';
import iOSPlugins from 'src/platform/interface/capacitor/ios/iOSRegisterPlugins';
import { busOn, busOff } from '../../../utils/bus';
import { common } from 'src/api';
import { getNativeAppPlatform } from 'src/application/platform';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const dialog = ref(false);

const maximizedToggle = ref(true);

const store = useDataStore();
const filesStore = useFilesStore();
const $q = useQuasar();

const isDark = ref(false);

const path = computed(function () {
	return (
		filesStore.previewItem[props.origin_id].path +
		'_' +
		filesStore.previewItem[props.origin_id].name
	);
});

const raw = computed(function () {
	console.log('filesStorepreviewItem', filesStore.previewItem[props.origin_id]);
	// if (
	// 	filesStore.previewItem[props.origin_id].type === 'video' &&
	// 	filesStore.previewItem[props.origin_id].driveType === DriveType.Sync
	// ) {
	// 	const url = new URL(filesStore.previewItem[props.origin_id].url);
	// 	return '/Seahub' + url.pathname;
	// }
	const path = filesStore.previewItem[props.origin_id].path;
	if (!path || path.length == 0) {
		return '';
	}

	return filesStore.getPreviewURL(
		filesStore.previewItem[props.origin_id],
		'big'
	);
});

const node = computed(() => {
	if (!filesStore.nodes || filesStore.nodes.length == 0) {
		return '';
	}
	if (!filesStore.previewItem[props.origin_id]) {
		return filesStore.masterNode;
	}

	if (
		filesStore.previewItem[props.origin_id].driveType == DriveType.Cache ||
		filesStore.previewItem[props.origin_id].driveType == DriveType.External
	) {
		return filesStore.previewItem[props.origin_id].fileExtend;
	}
	return filesStore.masterNode;
});

const isTitleHidden = ref(false);
const hideTimeout = ref();

const holdShowTitle = (flag: boolean) => {
	if (flag) {
		clearTimeout(hideTimeout.value);
		isTitleHidden.value = false;
		return false;
	}
	isTitleHidden.value = false;
	clearTimeout(hideTimeout.value);
	hideTimeout.value = setTimeout(() => {
		isTitleHidden.value = true;
	}, 2500);
};

onBeforeMount(() => {
	store.preview.isShow = true;
});

onMounted(() => {
	isDark.value = false;
	if ($q.platform.is.nativeMobile) {
		// ScreenOrientation.unlock({
		// 	orientation:
		// });
		ScreenOrientation.lock({
			orientation: 'landscape'
		});
		StatusBar.hide();

		if ($q.platform.is.android) {
			AndroidPlugins.AndroidUniversal.hideNavigationBar();
			AndroidPlugins.EdgeToEdge.disable();
		} else if ($q.platform.is.ios) {
			iOSPlugins.IOSUniversal.hideHomeIndicator();
		}
	}

	busOn('dialogDismiss', () => {
		close();
	});
});

const close = () => {
	dialog.value = false;
};

onBeforeUnmount(() => {
	store.preview.fullSize = false;
	store.preview.isShow = false;
	busOff('dialogDismiss');
	if ($q.platform.is.nativeMobile) {
		getNativeAppPlatform().resetOrientationLockType();
		StatusBar.show();
		if ($q.platform.is.android) {
			AndroidPlugins.AndroidUniversal.showNavigationBar();
			AndroidPlugins.EdgeToEdge.enable();
		} else if ($q.platform.is.ios) {
			iOSPlugins.IOSUniversal.showHomeIndicator();
		}
	}
	filesStore.isInPreview[props.origin_id] = '';
});
</script>

<style scoped lang="scss">
#video-previewer {
	position: relative;
	background-color: rgba(0, 0, 0, 1);
	height: 100vh;
	max-height: 100vh !important;
}

#video-previewer .preview {
	text-align: center;
	height: 100%;
}

#video-previewer .content {
	height: 100%;
	overflow: hidden;
}

.header {
	position: absolute;
	width: 100%;
	height: 56px;
	z-index: 1000;
	opacity: 1;
	transition: opacity 1s ease;
	padding-left: env(safe-area-inset-left);
	padding-right: env(safe-area-inset-right);
	.title {
		width: calc(
			100vw - 60px - env(safe-area-inset-left) - env(safe-area-inset-right)
		);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.hidden {
	opacity: 0;
	pointer-events: none;
}
</style>
