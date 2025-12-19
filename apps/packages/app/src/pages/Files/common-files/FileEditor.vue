<template>
	<div id="editor-container" v-show="store.preview.isEditing">
		<div id="editor"></div>
	</div>
	<div
		class="info-content text-subtitle3 text-ink-1"
		:class="$q.platform.is.mobile ? 'q-pa-md' : 'q-pa-lg'"
		v-if="!store.preview.isEditing"
	>
		{{ filesStore.previewItem[origin_id].content }}
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, onUnmounted } from 'vue';
import ace from 'ace-builds/src-min-noconflict/ace.js';
import modelist from 'ace-builds/src-min-noconflict/ext-modelist.js';
import { useDataStore } from '../../../stores/data';
import { useQuasar } from 'quasar';
import 'ace-builds/webpack-resolver';
import { useFilesStore, FilesIdType } from '../../../stores/files';

import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { dataAPIs } from '../../../api';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const editor = ref();
const $q = useQuasar();
const filesStore = useFilesStore();

const item = filesStore.previewItem[props.origin_id];
const dataAPI = dataAPIs(item.driveType);

onMounted(async () => {
	if (filesStore.previewItem[props.origin_id].content == undefined) {
		await dataAPI.formatFileContent(filesStore.previewItem[props.origin_id]);
	}

	const fileContent = `${
		filesStore.previewItem[props.origin_id].content || ''
	}`;

	editor.value = ace.edit('editor', {
		value: fileContent,
		showPrintMargin: false,
		readOnly: filesStore.previewItem[props.origin_id].type === 'textImmutable',
		theme: 'ace/theme/chrome',
		mode: modelist.getModeForPath(filesStore.previewItem[props.origin_id].name)
			.mode,
		wrap: true
	});
});

onUnmounted(() => {
	store.preview.isSaving = false;
	store.preview.isEditing = false;
});

const save = async () => {
	if (!editor.value) {
		return;
	}
	let editorValue = editor.value.getValue();
	try {
		editorValue = (editorValue as string).trim();
		const previewItem = JSON.parse(
			JSON.stringify(filesStore.previewItem[props.origin_id])
		);
		if (previewItem.front) {
			previewItem.path =
				previewItem.path + encodeURIComponent(previewItem.name);
		}
		dataAPI.onSaveFile(previewItem, editorValue, $q.platform.is.nativeMobile);
		filesStore.previewItem[props.origin_id].content = editorValue;
		store.preview.isSaving = false;
		store.preview.isEditing = false;
	} catch (e) {
		store.preview.isSaving = false;
		notifyFailed(e.message);
	}
};
watch(
	() => store.preview.isSaving,
	() => {
		if (store.preview.isSaving) {
			save();
		}
	}
);
</script>
<style scoped lang="scss">
.file-editor-root {
	width: 100%;
	height: 100%;
}

.info-content {
	overflow-y: scroll;
	height: 100%;
	white-space: pre-wrap;
	word-wrap: break-word;
}
</style>
