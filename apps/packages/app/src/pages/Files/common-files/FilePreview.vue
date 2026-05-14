<template>
	<BtLoading
		:show="true"
		v-if="store.loading"
		textColor="#4999ff"
		color="#4999ff"
		text=""
		backgroundColor="rgba(0, 0, 0, 0)"
	>
	</BtLoading>
	<template v-else>
		<div class="preview">
			<ExtendedImage
				v-if="filesStore.previewItem[origin_id].type == 'image'"
				:src="raw"
			/>

			<div
				v-else-if="filesStore.previewItem[origin_id].type == 'audio'"
				class="audio-container column justify-between items-center q-pb-lg q-px-md"
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
				<audio
					ref="player"
					:src="displayRaw"
					controls
					controlsList="nodownload noplaybackrate"
					:autoplay="autoPlay"
					@play="autoPlay = true"
				/>
			</div>

			<object
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'pdf'
				"
				class="pdf"
				:data="displayRaw"
			/>
			<vue-office-docx
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'word'
				"
				:src="displayRaw"
			/>
			<vue-office-excel
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'excel'
				"
				:src="displayRaw"
			/>
		</div>
	</template>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useDataStore } from '../../../stores/data';
import ExtendedImage from '../../../components/files/ExtendedImage.vue';
import { FilesIdType } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import { useFilesStore } from './../../../stores/files';
// Use lib/v3 explicitly: package "main" is lib/index.js which is created by postinstall;
// if postinstall skipped or Vue was not resolvable during install, lib/index.js is missing and webpack fails.
import VueOfficeDocx from '@vue-office/docx/lib/v3';
import VueOfficeExcel from '@vue-office/excel/lib/v3';

import '@vue-office/docx/lib/v3/index.css';
import '@vue-office/excel/lib/v3/index.css';

import { shallowRef } from 'vue';
import { useUserStore } from 'src/stores/user';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const filesStore = useFilesStore();

const raw = computed(function () {
	if (
		filesStore.previewItem[props.origin_id].type === 'image' &&
		!store.preview.fullSize
	) {
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
const autoPlay = shallowRef(true);

const displayRaw = ref('');

onMounted(() => {
	const userStore = useUserStore();
	if (
		userStore.current_user &&
		userStore.current_user.isLocal &&
		filesStore.previewItem[props.origin_id].type != 'image'
	) {
		loadImageContent();
	} else {
		displayRaw.value = raw.value;
	}
});

const loadImageContent = async () => {
	try {
		const userStore = useUserStore();

		const response = await fetch(raw.value, {
			headers: {
				'X-Authorization': userStore.current_user!.access_token
			}
		});
		const blob = await response.blob();
		const url = URL.createObjectURL(blob);
		displayRaw.value = url;
	} catch (error) {
		console.error('Error converting HEIC to JPEG:', error);
	}
};
</script>

<style lang="scss" scoped>
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
	audio::-webkit-media-controls-panel {
		background-color: $background-3;
	}

	audio::-webkit-media-controls-play-button {
		background-color: $primary;
		border-radius: 50%;
		margin-right: 10px;
	}

	audio::-webkit-media-controls-timeline {
		background: $background-3;
	}
	audio::-webkit-media-controls-current-time-display,
	audio::-webkit-media-controls-time-remaining-display {
		color: $ink-1;
	}

	audio::-webkit-media-controls-volume-slider {
		background: $background-3;
	}

	audio {
		accent-color: $primary;
	}

	audio::-webkit-media-controls-overflow-button {
		display: none !important;
	}
}
</style>
