<template>
	<div class="rss-video-preview bg-black full-width">
		<terminus-video-player
			v-if="videoRealPath"
			id="videoPlayer"
			:played-time="playedTime"
			:path="path"
			:progress="!src"
			:src="videoRealPath"
		/>
	</div>
	<full-content-reader v-if="!src" :margin-top="true" />
</template>

<script lang="ts" setup>
import TerminusVideoPlayer from '../../../../components/common/TerminusVideoPlayer.vue';
import SimpleWaiter from '../../../../utils/simpleWaiter';
import FullContentReader from './FullContentReader.vue';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { useReaderStore } from '../../../../stores/rss-reader';
import { common } from '../../../../api';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

const props = defineProps({
	src: {
		type: String,
		require: false
	}
});

const transferStore = useTransferStore();
const readerStore = useReaderStore();
const waiter = new SimpleWaiter();
const videoRealPath = ref();

const path = computed(function () {
	return props.src || readerStore.readingEntry.local_file_path;
});

const node = computed(() => {
	if (!transferStore.nodes || transferStore.nodes.length == 0) {
		return '';
	}
	return transferStore.nodes[0].name;
});

const raw = computed(function () {
	return transferStore.getVideoUrl(path.value, 'video');
});

onMounted(() => {
	waiter.waitForCondition(
		() => {
			return raw.value && node.value;
		},
		() => {
			videoRealPath.value = common().videoPlayUrl(raw.value, node.value);
		}
	);
});

onBeforeUnmount(() => {
	waiter.clear();
});

const playedTime = computed(() => {
	const readerStore = useReaderStore();
	if (readerStore.readingEntry && !props.src) {
		return readerStore.readingEntry.played_time;
	} else {
		return 0;
	}
});
</script>

<style lang="scss" scoped>
.rss-video-preview {
	height: 400px;
}
</style>
