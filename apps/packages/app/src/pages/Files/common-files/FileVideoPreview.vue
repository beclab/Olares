<template>
	<div class="preview bg-black">
		<terminus-video-player
			v-if="raw"
			:req="filesStore.previewItem"
			:path="path"
			id="videoPlayer"
			platform="web"
			:src="common().videoPlayUrl(raw, node)"
			@holdShowTitle="holdShowTitle"
		/>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useFilesStore } from '../../../stores/files';
import { useDataStore } from '../../../stores/data';
import { FilesIdType } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import { common } from 'src/api';
import TerminusVideoPlayer from '../../../components/common/TerminusVideoPlayer.vue';

const emit = defineEmits(['holdShowTitle']);

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const filesStore = useFilesStore();
const dataStore = useDataStore();

const path = computed(function () {
	return (
		filesStore.previewItem[props.origin_id].path +
		'_' +
		filesStore.previewItem[props.origin_id].name
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

const raw = computed(function () {
	if (!filesStore.previewItem[props.origin_id]) {
		return '';
	}
	const path = filesStore.previewItem[props.origin_id].path;
	if (!path || path.length == 0) {
		return '';
	}
	return filesStore.getPreviewURL(
		filesStore.previewItem[props.origin_id],
		'big'
	);
});

const holdShowTitle = (flag: boolean) => {
	emit('holdShowTitle', flag);
};
</script>

<style lang="scss" scoped></style>
