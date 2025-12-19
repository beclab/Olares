<template>
	<q-dialog
		v-model="dialog"
		persistent
		:maximized="maximizedToggle"
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<div
			id="video-previewer"
			:style="{
				'padding-top':
					$q.platform.is.electron && $q.platform.is.win ? '34px' : '12px'
			}"
		>
			<div
				class="header row items-center justify-between"
				:class="{ hidden: isTitleHidden }"
			>
				<div class="title text-ink-on-brand text-h6">
					{{ filesStore.previewItem[origin_id].name }}
				</div>
				<div>
					<q-btn
						dense
						flat
						color="white"
						icon="sym_r_close"
						@click="close"
						v-close-popup
						style="width: 32px"
					>
					</q-btn>
				</div>
			</div>
			<div class="content">
				<file-video-preview
					:origin_id="origin_id"
					@holdShowTitle="holdShowTitle"
				/>
			</div>
		</div>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import FileVideoPreview from '../common-files/FileVideoPreview.vue';

import { onMounted, onUnmounted, ref, onBeforeMount } from 'vue';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const dialog = ref(true);

const maximizedToggle = ref(true);

const store = useDataStore();
const filesStore = useFilesStore();

const isDark = ref(false);

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
});

const close = () => {
	dialog.value = false;
	setTimeout(() => {
		filesStore.isInPreview[props.origin_id] = '';
	}, 500);
};

onUnmounted(() => {
	store.preview.fullSize = false;
	store.preview.isShow = false;
});
</script>

<style scoped lang="scss">
#video-previewer {
	position: relative;
	background-color: rgba(0, 0, 0, 0);
	padding: 12px;
	height: 100%;
}

#video-previewer .preview {
	text-align: center;
	height: 100%;
}

#video-previewer .content {
	height: 100%;
	border-radius: 12px;
	overflow: hidden;
	background-color: rgba(0, 0, 0, 1);
}

.header {
	position: absolute;
	width: calc(100% - 24px);
	height: 56px;
	z-index: 1;
	padding: 0 20px;
	opacity: 1;
	transition: opacity 1s ease;
}

.hidden {
	opacity: 0;
	pointer-events: none;
}
</style>
