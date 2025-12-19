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
		<div class="preview q-py-md">
			<TerminusEbook :src="raw" />

			<!-- <div ref="viewer" style="width: 100%; height: 90%"></div> -->
		</div>
	</template>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useDataStore } from '../../../stores/data';
import { DriveType } from '../../../utils/interface/files';
import { useFilesStore, FilesIdType } from './../../../stores/files';
import TerminusEbook from './../../../components/common/TerminusEbook.vue';

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
