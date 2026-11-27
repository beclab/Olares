<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="entryName"
		:ok="false"
		:cancel="t('cancel')"
		size="large"
		:platform="deviceStore.isMobile ? 'mobile' : 'web'"
		:position="deviceStore.isMobile ? 'bottom' : 'standard'"
		:content-pending="!deviceStore.isMobile"
	>
		<div class="archive-entry-preview">
			<div
				v-if="loading"
				class="archive-entry-preview__loading row items-center justify-center"
			>
				<q-spinner-dots color="primary" size="3em" />
			</div>
			<component
				v-else
				:is="currentView"
				v-bind="previewProps"
				class="archive-entry-preview__content"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import {
	computed,
	onBeforeUnmount,
	onMounted,
	PropType,
	ref,
	shallowRef
} from 'vue';
import type { Component } from 'vue';
import { useI18n } from 'vue-i18n';
import { getFileIcon, getFilePreviewType, PreviewType } from '@bytetrade/core';
import { FileItem, useFilesStore } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import ArchiveAPI from '../../../api/files/v2/common/archive';
import { CommonFetch } from '../../../api';
import { resolveFilePreview } from '../../../pages/Files/common-files/previewResolver';
import FileEpubPreview from '../../../pages/Files/common-files/FileEpubPreview.vue';
import FilePreview from '../../../pages/Files/common-files/FilePreview.vue';
import FilePDFPreview from '../../../pages/Files/common-files/FilePDFPreview.vue';
import FileUnavailable from '../../../pages/Files/common-files/FileUnavailable.vue';
import FilesMarkdownPreview from '../../../pages/Files/common-files/TextPreviews/FilesMarkdownPreview.vue';
import FilesNormalTxtPreview from '../../../pages/Files/common-files/TextPreviews/FilesNormalTxtPreview.vue';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps({
	source: {
		type: String,
		required: true
	},
	entry: {
		type: Object as PropType<{
			name: string;
			path?: string;
			size?: number;
			modified?: number | string;
			lastModified?: number | string;
			type?: string;
		}>,
		required: true
	},
	node: {
		type: String,
		default: ''
	},
	driveType: {
		type: String as PropType<DriveType>,
		default: DriveType.Drive
	},
	password: {
		type: String,
		default: ''
	}
});

const { t } = useI18n();
const filesStore = useFilesStore();
const deviceStore = useDeviceStore();
const dialogRef = ref();
const loading = ref(false);
const textContent = ref('');
const objectUrl = ref('');
const currentView = shallowRef<Component>(FileUnavailable);

const originId = Date.now();
const entryName = computed(() => props.entry.name || props.entry.path || '');
const entryPath = computed(() => props.entry.path || props.entry.name || '');
const previewType = computed(() => getFilePreviewType(entryName.value));
const streamUrl = computed(() =>
	ArchiveAPI.getPreviewStreamUrl(
		{
			source: props.source,
			path: entryPath.value
		},
		props.node
	)
);

const previewProps = computed(() => {
	if (previewType.value === PreviewType.TXT) {
		return {
			content: textContent.value,
			name: entryName.value
		};
	}

	if (currentView.value !== FileUnavailable) {
		return {
			origin_id: originId
		};
	}

	return {};
});

const setPreviewItem = () => {
	const extension = entryName.value.includes('.')
		? `.${entryName.value.split('.').pop() || ''}`
		: '';
	const rawUrl = objectUrl.value || streamUrl.value;
	filesStore.previewItem[originId] = {
		name: entryName.value,
		path: entryPath.value,
		parentPath: '',
		size: props.entry.size || 0,
		modified: props.entry.modified || props.entry.lastModified || '',
		isDir: false,
		type: getFileIcon(entryName.value),
		driveType: props.driveType,
		rawUrl,
		previewUrl: rawUrl,
		extension,
		content: textContent.value
	} as FileItem & { rawUrl: string; previewUrl: string };
};

const fetchEntryStream = async () => {
	const headers: Record<string, string> = {};
	if (props.password) {
		headers['X-Archive-Password'] = props.password;
	}
	const response = await CommonFetch.getStream(streamUrl.value, { headers });
	if (!response.ok) {
		const data = await response
			.clone()
			.json()
			.catch(() => ({
				error: `Archive entry preview failed: ${response.status}`
			}));
		throw data;
	}
	return response;
};

const loadTextContent = async () => {
	const response = await fetchEntryStream();
	textContent.value = await response.text();
};

const loadObjectUrl = async () => {
	const response = await fetchEntryStream();
	const blob = await response.blob();
	objectUrl.value = URL.createObjectURL(blob);
};

const getReadonlyTextPreview = (item: FileItem) => {
	return item.name.toLowerCase().endsWith('.md')
		? FilesMarkdownPreview
		: FilesNormalTxtPreview;
};

const resolveCurrentView = () => {
	const preview = resolveFilePreview(
		{
			name: entryName.value,
			path: entryPath.value,
			parentPath: '',
			size: props.entry.size || 0,
			modified: props.entry.modified || props.entry.lastModified || '',
			isDir: false,
			type: getFileIcon(entryName.value),
			driveType: props.driveType,
			extension: entryName.value.includes('.')
				? `.${entryName.value.split('.').pop() || ''}`
				: ''
		} as FileItem,
		{
			text: FilesNormalTxtPreview,
			ebook: FileEpubPreview,
			pdf: FilePDFPreview,
			media: FilePreview,
			unavailable: FileUnavailable
		},
		{
			usePdfComponent: true,
			textComponentResolver: getReadonlyTextPreview
		}
	);
	return preview.component;
};

onMounted(async () => {
	try {
		const view = resolveCurrentView();
		loading.value = view !== FileUnavailable;
		if (previewType.value === PreviewType.TXT) {
			await loadTextContent();
		} else if (view !== FileUnavailable) {
			await loadObjectUrl();
		}
		currentView.value = view;
		setPreviewItem();
	} catch (error) {
		console.warn('Failed to preview archive entry', error);
		currentView.value = FileUnavailable;
	} finally {
		loading.value = false;
	}
});

onBeforeUnmount(() => {
	if (objectUrl.value) {
		URL.revokeObjectURL(objectUrl.value);
	}
	delete filesStore.previewItem[originId];
});
</script>

<style scoped lang="scss">
.archive-entry-preview {
	width: 760px;
	max-width: 100%;
	height: 520px;
	background: $background-2;
	border-radius: 8px;
	overflow: hidden;

	&__loading,
	&__content {
		width: 100%;
		height: 100%;
	}
}
</style>
