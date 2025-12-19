<template>
	<div class="video-preview-root">
		<div
			v-if="filesStore.previewItem[origin_id].type == 'image'"
			class="image-container"
		>
			<ExtendedImage :src="raw" class="img-part" />
			<div class="row items-center justify-center q-mt-lg">
				<div
					class="operation-part text-body3 q-px-md q-py-xs"
					@click="fullSize = !fullSize"
				>
					{{
						!fullSize
							? t('files.view_original_image') + ' ' + size
							: t('files.view_preview_image')
					}}
				</div>
			</div>
		</div>

		<div
			v-else-if="filesStore.previewItem[origin_id].type == 'audio'"
			class="audio-container column justify-between items-center"
		>
			<div class="audio-info column items-center justify-center">
				<img
					src="../../../assets/images/file-audio.svg"
					width="96"
					height="96"
				/>
				<div class="audio-name text-body3 q-mt-md text-color-title">
					{{ filesStore.previewItem[origin_id].name }}
				</div>
			</div>
			<div
				style="height: 20%; width: 100%; padding: 20px"
				class="row items-center justify-center"
			>
				<audio
					:src="raw"
					controls
					:autoplay="autoPlay"
					@play="autoPlay = true"
					style="height: 20px"
				></audio>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import ExtendedImage from '../../../components/files/ExtendedImage.vue';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ScreenOrientation } from '@capacitor/screen-orientation';
import { StatusBar } from '@capacitor/status-bar';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { format } from '../../../utils/format';
import { getNativeAppPlatform } from 'src/application/platform';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const { humanStorageSize } = format;
const autoPlay = ref(false);
const fullSize = ref(false);
const $q = useQuasar();
const filesStore = useFilesStore();

const { t } = useI18n();

const size = ref(
	humanStorageSize(filesStore.previewItem[props.origin_id].size ?? 0)
);

const raw = computed(function () {
	if (
		filesStore.previewItem[props.origin_id].type === 'image' &&
		!fullSize.value
	) {
		console.log(
			'filesStore.previewItem[props.origin_id], ===>',
			filesStore.previewItem[props.origin_id]
		);
		return filesStore.getPreviewURL(
			filesStore.previewItem[props.origin_id],
			'big'
		);
	}
	return filesStore.getDownloadURL(
		filesStore.previewItem[props.origin_id],
		true
	);
});

onMounted(() => {
	if (
		$q.platform.is.nativeMobile &&
		filesStore.previewItem[props.origin_id].type == 'video'
	) {
		ScreenOrientation.unlock();
		StatusBar.hide();
	}
});

onUnmounted(() => {
	if (
		$q.platform.is.nativeMobile &&
		filesStore.previewItem[props.origin_id].type == 'video'
	) {
		getNativeAppPlatform().resetOrientationLockType();
		StatusBar.show();
	}
});
</script>

<style scoped lang="scss">
.video-preview-root {
	width: 100%;
	height: 100%;
	background: $shadow-color;
}

audio {
	width: 100%;
	outline: none;
}

.image-container {
	width: 100%;
	height: 100%;

	.img-part {
		height: calc(100% - 150px);
		width: 100%;
	}

	.operation-part {
		text-align: right;
		color: $grey-4;
		border: 1px solid $separator;
		border-radius: 4px;
		text-align: center;
	}
}

.audio-container {
	width: 100%;
	height: 100%;

	.audio-info {
		height: 80%;
		width: 100%;

		.audio-name {
			text-align: center;
			color: $grey-5;
		}
	}
}

.video-container {
	width: 100%;
	height: 100%;

	audio::-webkit-media-controls-panel {
		background-color: white;
		border-radius: 8px;
	}

	input[type='range'] {
		width: 100%;
		height: 5px;
		-webkit-appearance: none;
		background-color: $grey-12;
		border-radius: 5px;
		outline: none;
		margin-top: 5px;
	}
}
</style>
