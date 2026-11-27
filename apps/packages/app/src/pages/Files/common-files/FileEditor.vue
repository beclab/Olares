<template>
	<div
		v-show="store.preview.isEditing"
		class="editor-area"
		:class="{ 'editor-area--split': showSplitPreview }"
	>
		<code-editor
			v-model:value="liveContent"
			:name="previewName"
			:readonly="isReadonly"
			:autocomplete="false"
			class="editor-area__editor"
			@init="onEditorInit"
		/>
		<FilesMarkdownPreview
			v-if="showSplitPreview"
			ref="previewRef"
			class="editor-area__preview"
			:content="liveContent"
		/>
	</div>

	<component
		v-if="!store.preview.isEditing && contentLoaded"
		:content="mdContent"
		:name="previewName"
		:is="currentView"
	>
	</component>
</template>

<script lang="ts" setup>
import {
	ref,
	onMounted,
	watch,
	onBeforeUnmount,
	computed,
	nextTick
} from 'vue';
import type { Ace } from 'ace-builds';

import { useQuasar } from 'quasar';
import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { dataAPIs } from '../../../api';
import CodeEditor from './TextPreviews/CodeEditor.vue';
import FilesMarkdownPreview from './TextPreviews/FilesMarkdownPreview.vue';
import FilesNormalTxtPreview from './TextPreviews/FilesNormalTxtPreview.vue';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const $q = useQuasar();
const filesStore = useFilesStore();

const item = filesStore.previewItem[props.origin_id];
const dataAPI = dataAPIs(item.driveType);

const currentView = ref();
const liveContent = ref('');
const initialContent = ref('');
const contentLoaded = ref(false);
const previewRef = ref();
const aceEditor = ref<Ace.Editor | null>(null);

const deviceStore = useDeviceStore();

const isMarkdown = computed(
	() => filesStore.previewItem[props.origin_id]?.extension === '.md'
);

const showSplitPreview = computed(
	() => isMarkdown.value && !deviceStore.isMobile
);

const isReadonly = computed(
	() => filesStore.previewItem[props.origin_id]?.type === 'textImmutable'
);

let isSyncingScroll = false;

const getPreviewEl = (): HTMLElement | undefined => {
	const el = previewRef.value?.$el ?? previewRef.value;
	return el instanceof HTMLElement ? el : undefined;
};

const getEditorMaxScroll = (): number => {
	const editor = aceEditor.value;
	if (!editor) return 0;
	const lineHeight = editor.renderer.lineHeight || 16;
	const totalHeight = editor.session.getLength() * lineHeight;
	const viewHeight = (editor.renderer as any).$size?.scrollerHeight || 0;
	return Math.max(0, totalHeight - viewHeight);
};

const syncToPreview = (scrollTop: number) => {
	if (isSyncingScroll || !showSplitPreview.value) return;
	const previewEl = getPreviewEl();
	if (!previewEl) return;

	const editorMax = getEditorMaxScroll();
	if (editorMax <= 0) return;

	const previewMax = previewEl.scrollHeight - previewEl.clientHeight;
	if (previewMax <= 0) return;

	isSyncingScroll = true;
	previewEl.scrollTop = (scrollTop / editorMax) * previewMax;
	requestAnimationFrame(() => {
		isSyncingScroll = false;
	});
};

const onPreviewScroll = () => {
	if (isSyncingScroll || !aceEditor.value) return;
	const previewEl = getPreviewEl();
	if (!previewEl) return;

	const previewMax = previewEl.scrollHeight - previewEl.clientHeight;
	if (previewMax <= 0) return;

	const editorMax = getEditorMaxScroll();
	if (editorMax <= 0) return;

	isSyncingScroll = true;
	aceEditor.value.session.setScrollTop(
		(previewEl.scrollTop / previewMax) * editorMax
	);
	requestAnimationFrame(() => {
		isSyncingScroll = false;
	});
};

const attachScrollSync = async () => {
	const editor = aceEditor.value;
	if (!editor) return;
	editor.session.on('changeScrollTop', syncToPreview);
	await nextTick();
	const previewEl = getPreviewEl();
	previewEl?.addEventListener('scroll', onPreviewScroll, { passive: true });
};

const detachScrollSync = () => {
	const previewEl = getPreviewEl();
	previewEl?.removeEventListener('scroll', onPreviewScroll);
	aceEditor.value?.session.off('changeScrollTop', syncToPreview);
};

const onEditorInit = async (editor: Ace.Editor) => {
	aceEditor.value = editor;
	if (!showSplitPreview.value) return;
	await attachScrollSync();
};

watch(showSplitPreview, async (next) => {
	if (!aceEditor.value) return;
	if (next) {
		await attachScrollSync();
	} else {
		detachScrollSync();
	}
});

onMounted(async () => {
	const previewItem = filesStore.previewItem[props.origin_id];
	contentLoaded.value = false;
	currentView.value =
		previewItem.extension === '.md'
			? FilesMarkdownPreview
			: FilesNormalTxtPreview;

	if (previewItem.content == undefined) {
		await dataAPI.formatFileContent(previewItem);
	}

	const fileContent = `${
		(typeof previewItem.content === 'object'
			? JSON.stringify(previewItem.content, null, 2)
			: previewItem.content) || ''
	}`;

	initialContent.value = fileContent;
	liveContent.value = fileContent;
	contentLoaded.value = true;
	store.preview.isDirty = false;
});

watch(liveContent, (next) => {
	store.preview.isDirty = next !== initialContent.value;
});

onBeforeUnmount(() => {
	detachScrollSync();
	store.preview.isSaving = false;
	store.preview.isEditing = false;
	store.preview.isDirty = false;
});

const save = async () => {
	const editorValue = liveContent.value.trim();
	try {
		const previewItem = JSON.parse(
			JSON.stringify(filesStore.previewItem[props.origin_id])
		);
		if (previewItem.front) {
			previewItem.path =
				previewItem.path + encodeURIComponent(previewItem.name);
		}
		await dataAPI.onSaveFile(
			previewItem,
			editorValue,
			$q.platform.is.nativeMobile
		);
		filesStore.previewItem[props.origin_id].content = editorValue;
		// Do not write `editorValue` (trimmed) back into `liveContent`,
		// otherwise the editor visibly strips the user's leading/trailing
		// whitespace right after save. Use the editor's current value as
		// the new dirty-tracking baseline instead.
		initialContent.value = liveContent.value;
		store.preview.isDirty = false;
		store.preview.isSaving = false;
		store.preview.isEditing = false;
	} catch (e) {
		// Error toast is handled globally by `CommonFetch` response
		// interceptor (`handleError`). Showing it again here would
		// double-toast the user. We only reset the saving flag so the
		// editor stays open with the dirty content for retry.
		store.preview.isSaving = false;
	}
};

watch(
	() => store.preview.isSaving,
	(saving) => {
		if (saving) save();
	}
);

const mdContent = computed(
	() => filesStore.previewItem[props.origin_id].content
);

const previewName = computed(
	() => filesStore.previewItem[props.origin_id]?.name || ''
);
</script>

<style lang="scss" scoped>
.editor-area {
	width: 100%;
	height: 100%;

	&__editor {
		width: 100%;
		height: 100%;
	}

	&--split {
		display: flex;
		flex-direction: row;
		align-items: stretch;

		.editor-area__editor {
			flex: 1 1 50%;
			min-width: 0;
		}

		.editor-area__preview {
			flex: 1 1 50%;
			min-width: 0;
			height: 100%;
			overflow-y: auto;
			border-left: 1px solid var(--separator, rgba(0, 0, 0, 0.08));
		}
	}
}
</style>
