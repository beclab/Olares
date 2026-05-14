<template>
	<div id="editor-container" v-show="store.preview.isEditing">
		<div id="editor"></div>
	</div>

	<component
		v-if="!store.preview.isEditing"
		:content="mdContent"
		:is="currentView"
	>
	</component>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, onUnmounted, computed } from 'vue';
import ace from 'ace-builds/src-min-noconflict/ace.js';
import modelist from 'ace-builds/src-min-noconflict/ext-modelist.js';
import 'ace-builds/src-min-noconflict/theme-chrome';
import 'ace-builds/src-min-noconflict/theme-tomorrow_night';
import { useDataStore } from '../../../stores/data';
import { useQuasar } from 'quasar';
import 'ace-builds/webpack-resolver';
import { useFilesStore, FilesIdType } from '../../../stores/files';

import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { dataAPIs } from '../../../api';
import FilesMarkdownPreview from './TextPreviews/FilesMarkdownPreview.vue';
import FilesNormalTxtPreview from './TextPreviews/FilesNormalTxtPreview.vue';

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
const currentView = ref();

onMounted(async () => {
	if (filesStore.previewItem[props.origin_id].extension == '.md') {
		currentView.value = FilesMarkdownPreview;
	} else {
		currentView.value = FilesNormalTxtPreview;
	}

	if (filesStore.previewItem[props.origin_id].content == undefined) {
		await dataAPI.formatFileContent(filesStore.previewItem[props.origin_id]);
	}

	const fileContent = `${
		(typeof filesStore.previewItem[props.origin_id].content == 'object'
			? JSON.stringify(filesStore.previewItem[props.origin_id].content, null, 2)
			: filesStore.previewItem[props.origin_id].content) || ''
	}`;

	try {
		editor.value = ace.edit('editor', {
			value: fileContent,
			showPrintMargin: false,
			readOnly:
				filesStore.previewItem[props.origin_id].type === 'textImmutable',
			theme: $q.dark.isActive ? 'ace/theme/tomorrow_night' : 'ace/theme/chrome',
			mode: modelist.getModeForPath(
				filesStore.previewItem[props.origin_id].name
			).mode,
			wrap: true
		});
	} catch (error) {
		/* empty */
	}
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

const mdContent = computed(() => {
	return filesStore.previewItem[props.origin_id].content;
});
</script>
