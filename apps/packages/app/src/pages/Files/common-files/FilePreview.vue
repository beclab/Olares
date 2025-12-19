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
			></ExtendedImage>
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
					:src="raw"
					controls
					:autoplay="autoPlay"
					@play="autoPlay = true"
				></audio>
			</div>

			<object
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'pdf'
				"
				class="pdf"
				:data="raw"
			></object>
			<vue-office-docx
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'word'
				"
				:src="raw"
			/>
			<vue-office-excel
				v-else-if="
					filesStore.previewItem[origin_id]?.type?.toLowerCase() == 'excel'
				"
				:src="raw"
			/>
		</div>
	</template>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useDataStore } from '../../../stores/data';
import ExtendedImage from '../../../components/files/ExtendedImage.vue';
import { FilesIdType } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import { useFilesStore } from './../../../stores/files';
import VueOfficeDocx from '@vue-office/docx';
import VueOfficeExcel from '@vue-office/excel';

import '@vue-office/docx/lib/index.css';
import '@vue-office/excel/lib/index.css';

import { shallowRef } from 'vue';

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
}
</style>
