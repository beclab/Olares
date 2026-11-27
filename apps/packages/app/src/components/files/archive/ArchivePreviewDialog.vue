<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="t('files.archive.preview_archive_contents')"
		:ok="t('files.archive.extract_all')"
		:cancel="t('cancel')"
		size="large"
		:platform="deviceStore.isMobile ? 'mobile' : 'web'"
		:position="deviceStore.isMobile ? 'bottom' : 'standard'"
		:content-pending="!deviceStore.isMobile"
		@onSubmit="extractAll"
	>
		<template v-if="deviceStore.isMobile" v-slot:header>
			<div class="archive-preview-mobile-header row items-center no-wrap">
				<div class="text-subtitle1 text-ink-1 text-ellipsis">
					{{ t('files.archive.preview_archive_contents') }}
				</div>
				<q-space />
				<q-btn
					flat
					dense
					round
					icon="sym_r_close"
					size="sm"
					color="ink-3"
					@click="dialogRef.onDialogCancel()"
				/>
			</div>
		</template>
		<div class="archive-preview-dialog">
			<div class="archive-preview-body">
				<div class="archive-preview-toolbar row items-center no-wrap">
					<div class="row items-center no-wrap q-mr-sm">
						<q-btn
							icon="sym_r_arrow_back"
							flat
							dense
							:disabled="!currentArchivePath"
							@click="goParentFolder"
							class="btn-no-text btn-no-border"
							:class="[
								currentArchivePath ? 'items-no-drag' : '',
								props.origin_id === FilesIdType.PAGEID
									? 'btn-size-sm'
									: 'btn-size-xs'
							]"
							color="ink-2"
							:style="{ pointerEvents: currentArchivePath ? 'auto' : 'none' }"
						/>
						<q-btn
							icon="sym_r_arrow_forward"
							flat
							dense
							:disabled="true"
							class="btn-no-text btn-no-border"
							:class="[
								'',
								props.origin_id === FilesIdType.PAGEID
									? 'btn-size-sm'
									: 'btn-size-xs'
							]"
							color="ink-2"
							:style="{ pointerEvents: 'none' }"
						/>
					</div>
					<div class="text-body2 text-ink-2 text-ellipsis archive-path">
						{{ displayArchivePath }}
					</div>
					<div v-if="loading" class="row items-center no-wrap text-ink-3">
						<q-spinner size="16px" color="light-blue-default" class="q-mr-sm" />
						<span class="text-body2">
							{{
								entries.length
									? t('files.archive.loaded_archive_contents', {
											loaded: entries.length,
											total: totalCount || '~'
									  })
									: t('files.archive.loading_archive_contents')
							}}
						</span>
					</div>
					<div v-else class="text-body2 text-ink-3">
						{{
							t('files.archive.entry_count', {
								count: currentFolderEntryCount
							})
						}}
					</div>
				</div>

				<div class="archive-preview-table-header row items-center no-wrap">
					<div class="archive-col-name text-body3 text-ink-3">
						{{ t('files.name') }}
					</div>
					<div class="archive-col-spacer" />
					<div class="archive-col-size text-body3 text-ink-3">
						{{ t('files.size') }}
					</div>
					<div
						class="archive-col-modified row items-center text-body3 text-ink-3"
					>
						{{ t('files.lastModified') }}
						<q-icon name="sym_r_arrow_downward_alt" size="16px" />
					</div>
					<div class="archive-col-action" />
				</div>

				<bt-scroll-area class="archive-preview-list">
					<template v-if="visibleEntries.length > 0">
						<div
							v-for="entry in visibleEntries"
							:key="entry.path || entry.name"
							class="archive-preview-row row items-center no-wrap"
							:class="{ 'archive-preview-row--folder': entry.isDir }"
							@click="deviceStore.isMobile && openEntry(entry)"
							@dblclick="!deviceStore.isMobile && openEntry(entry)"
						>
							<div class="archive-col-name row items-center no-wrap">
								<terminus-file-icon
									:name="entry.name"
									:type="entry.type"
									:is-dir="entry.isDir"
									style="width: 24px; height: 24px"
								/>
								<div class="text-subtitle3 text-ink-1 text-ellipsis q-ml-md">
									{{ displayEntryName(entry) }}
								</div>
								<div class="archive-mobile-entry-meta text-body3 text-ink-3">
									{{ formatMobileEntryMeta(entry) }}
								</div>
							</div>
							<div class="archive-col-spacer" />
							<div class="archive-col-size text-body3 text-ink-2">
								{{ formatEntrySize(entry) }}
							</div>
							<div class="archive-col-modified text-body3 text-ink-2">
								{{ formatEntryModified(entry) }}
							</div>
							<div
								v-if="entry.isDir"
								class="archive-mobile-folder-arrow row items-center justify-center"
							>
								<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
							</div>
							<div class="archive-col-action row items-center justify-end">
								<!-- Entry-level archive actions are hidden. -->
							</div>
						</div>
					</template>
					<template v-if="loading">
						<div
							v-for="index in skeletonRowCount"
							:key="`skeleton-${index}`"
							class="archive-preview-row row items-center no-wrap"
						>
							<div class="archive-col-name row items-center no-wrap">
								<q-skeleton type="rect" width="24px" height="24px" />
								<q-skeleton type="text" class="q-ml-md archive-skeleton-name" />
							</div>
							<div class="archive-col-spacer" />
							<div class="archive-col-size">
								<q-skeleton type="text" />
							</div>
							<div class="archive-col-modified">
								<q-skeleton type="text" />
							</div>
							<!-- Action skeleton is hidden until entry-level actions are supported.
							<div class="archive-col-action">
								<q-skeleton type="text" />
							</div>
							-->
						</div>
					</template>
					<div
						v-if="!loading && visibleEntries.length === 0"
						class="archive-empty column items-center justify-center"
					>
						<div class="text-body2 text-ink-3">
							{{ errorMessage || t('files.lonely') }}
						</div>
					</div>
				</bt-scroll-area>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, PropType, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { getFileIcon } from '@bytetrade/core';
import { FileItem, FilesIdType, useFilesStore } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import { appendPath } from '../../../api/files/path';
import ArchiveAPI from '../../../api/files/v2/common/archive';
import { CommonFetch } from '../../../api';
import ArchiveExtractDialog from './ArchiveExtractDialog.vue';
import ArchiveEntryPreviewDialog from './ArchiveEntryPreviewDialog.vue';
import ArchivePasswordDialog from './ArchivePasswordDialog.vue';
import TerminusFileIcon from '../../common/TerminusFileIcon.vue';
import { format, formatDateFromNow } from '../../../utils/format';
import {
	getArchiveErrorCode,
	getArchiveErrorMessage,
	isArchivePasswordError
} from '../../../utils/interface/archive';
import { useDeviceStore } from 'src/stores/settings/device';

interface ArchivePreviewEntry {
	name: string;
	path?: string;
	size?: number;
	modified?: number | string;
	lastModified?: number | string;
	type?: string;
	isDir?: boolean;
	encrypted?: boolean;
}

const props = defineProps({
	item: {
		type: Object as PropType<FileItem | undefined>,
		default: undefined
	},
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	},
	driveType: {
		type: String as PropType<DriveType>,
		default: DriveType.Drive
	},
	streamUrl: {
		type: String,
		default: ''
	},
	password: {
		type: String,
		default: ''
	}
});

const dialogRef = ref();
const $q = useQuasar();
const { t } = useI18n();
const filesStore = useFilesStore();
const loading = ref(true);
const entries = ref<ArchivePreviewEntry[]>([]);
const totalCount = ref(0);
const errorMessage = ref('');
const archivePassword = ref(props.password);
const passwordPromptShowing = ref(false);
const currentArchivePath = ref('');
const deviceStore = useDeviceStore();
const skeletonRowCount = computed(() => Math.max(0, 5 - entries.value.length));

const archiveItem = computed(() => {
	if (props.item) {
		return props.item;
	}
	return filesStore.getTargetFileItem(
		filesStore.selected[props.origin_id]?.[0],
		props.origin_id
	) as FileItem | undefined;
});

const archiveSource = computed(() => {
	const item = archiveItem.value;
	if (!item) {
		return '';
	}
	return filesStore.fileApiFrontendPath(item);
});

const displayArchivePath = computed(() => {
	const archiveName = archiveItem.value?.name || '';
	return currentArchivePath.value
		? `${archiveName}/${currentArchivePath.value.replace(/\/$/, '')}`
		: archiveName;
});

const archiveNode = computed(() => {
	const item = archiveItem.value;
	return (
		item?.node ||
		filesStore.currentFileList[props.origin_id]?.node ||
		filesStore.currentNode[props.origin_id]?.name ||
		filesStore.nodes[0]?.name ||
		''
	);
});

const previewStreamUrl = computed(() => {
	if (props.streamUrl) {
		return props.streamUrl;
	}
	if (!archiveSource.value) {
		return '';
	}
	return ArchiveAPI.getPreviewContentsStreamUrl(
		{ source: archiveSource.value },
		archiveNode.value
	);
});

const formatEntrySize = (entry: ArchivePreviewEntry) => {
	if (entry.isDir) {
		return '--';
	}
	return typeof entry.size === 'number'
		? format.humanStorageSize(entry.size)
		: '--';
};

const formatEntryModified = (entry: ArchivePreviewEntry) => {
	const modified = entry.modified || entry.lastModified;
	if (!modified) {
		return '--';
	}
	const timestamp = Number(modified);
	return timestamp > 0
		? formatDateFromNow(
				timestamp < 1000000000000 ? timestamp * 1000 : timestamp
		  )
		: '--';
};

const formatMobileEntryMeta = (entry: ArchivePreviewEntry) => {
	const modified = formatEntryModified(entry);
	if (entry.isDir) {
		return modified;
	}
	const size = formatEntrySize(entry);
	return [size === '--' ? '' : size, modified === '--' ? '' : modified]
		.filter(Boolean)
		.join('  ');
};

const normalizeEntry = (entry: any): ArchivePreviewEntry | undefined => {
	const path = entry?.path || entry?.name;
	if (!path) {
		return undefined;
	}
	const name = path.split('/').filter(Boolean).pop() || path;
	const isDir = Boolean(entry.is_dir || entry.isDir || entry.type === 'folder');
	return {
		name,
		path,
		size: entry.size,
		modified: entry.modified || entry.lastModified,
		lastModified: entry.lastModified,
		type: isDir ? 'folder' : entry.type || getFileIcon(name),
		isDir,
		encrypted: Boolean(entry.encrypted)
	};
};

const normalizeArchivePath = (path?: string, isDir = false) => {
	if (!path) {
		return '';
	}
	const normalized = path.replace(/^\/+/, '').replace(/\/+$/, '');
	return isDir && normalized ? `${normalized}/` : normalized;
};

const visibleEntries = computed(() => {
	const children = new Map<string, ArchivePreviewEntry>();
	const currentPath = currentArchivePath.value;

	entries.value.forEach((entry) => {
		const entryPath = normalizeArchivePath(
			entry.path || entry.name,
			entry.isDir
		);
		if (!entryPath || !entryPath.startsWith(currentPath)) {
			return;
		}

		const relativePath = entryPath.slice(currentPath.length);
		if (!relativePath) {
			return;
		}

		const [segment] = relativePath.split('/');
		if (!segment) {
			return;
		}

		const hasNestedPath = relativePath.includes('/');
		const isDirectDir = entry.isDir && relativePath === `${segment}/`;
		const childPath = `${currentPath}${segment}${hasNestedPath ? '/' : ''}`;

		if (!hasNestedPath || isDirectDir) {
			children.set(childPath, {
				...entry,
				name: segment,
				path: childPath
			});
			return;
		}

		if (!children.has(childPath)) {
			children.set(childPath, {
				name: segment,
				path: childPath,
				size: 0,
				modified: entry.modified,
				type: 'folder',
				isDir: true,
				encrypted: entry.encrypted
			});
		}
	});

	return Array.from(children.values()).sort((a, b) => {
		if (a.isDir !== b.isDir) {
			return a.isDir ? -1 : 1;
		}
		return a.name.localeCompare(b.name);
	});
});

const currentFolderEntryCount = computed(() => visibleEntries.value.length);

const displayEntryName = (entry: ArchivePreviewEntry) =>
	entry.isDir ? `${entry.name}/` : entry.name;

const enterFolder = (entry: ArchivePreviewEntry) => {
	if (!entry.path) {
		return;
	}
	currentArchivePath.value = normalizeArchivePath(entry.path, true);
};

const previewEntry = (entry: ArchivePreviewEntry) => {
	if (!archiveSource.value || !entry.path || entry.isDir) {
		return;
	}
	$q.dialog({
		component: ArchiveEntryPreviewDialog,
		componentProps: {
			source: archiveSource.value,
			entry,
			node: archiveNode.value,
			driveType: props.driveType,
			password: archivePassword.value
		}
	});
};

const openEntry = (entry: ArchivePreviewEntry) => {
	if (entry.isDir) {
		enterFolder(entry);
		return;
	}
	// previewEntry(entry);
};

const goParentFolder = () => {
	const path = currentArchivePath.value.replace(/\/$/, '');
	const index = path.lastIndexOf('/');
	currentArchivePath.value = index === -1 ? '' : `${path.slice(0, index)}/`;
};

const appendEntry = (entry: any) => {
	const normalized = normalizeEntry(entry);
	if (!normalized) {
		return;
	}
	entries.value.push(normalized);
};

const requestArchivePassword = (errorMessage = '') => {
	passwordPromptShowing.value = true;
	return new Promise<string>((resolve, reject) => {
		let settled = false;
		$q.dialog({
			component: ArchivePasswordDialog,
			componentProps: {
				archiveName: archiveItem.value?.name || '',
				errorMessage,
				title: t('files.archive.unlock_archive'),
				okText: archivePassword.value
					? t('files.archive.try_again')
					: t('files.archive.unlock_and_preview'),
				errorPassword: archivePassword.value
			}
		})
			.onOk((password: string) => {
				settled = true;
				passwordPromptShowing.value = false;
				resolve(password);
			})
			.onCancel(() => {
				settled = true;
				passwordPromptShowing.value = false;
				reject(new Error('archive_password_cancelled'));
			})
			.onDismiss(() => {
				passwordPromptShowing.value = false;
				if (!settled) {
					reject(new Error('archive_password_cancelled'));
				}
			});
	});
};

const consumeStreamLine = (line: string) => {
	const trimmed = line.trim();
	if (!trimmed || trimmed === 'data: [DONE]' || trimmed === '[DONE]') {
		return;
	}

	const jsonString = trimmed.startsWith('data:')
		? trimmed.replace(/^data:\s*/, '')
		: trimmed;

	let payload: any;
	try {
		payload = JSON.parse(jsonString);
	} catch (error) {
		console.warn('Ignore invalid archive preview stream line', error, line);
		return;
	}
	if (payload._done) {
		totalCount.value = payload.total || entries.value.length;
		return;
	}
	if (payload._error) {
		throw {
			...payload,
			code: payload.code || payload._error?.code,
			message: getArchiveErrorMessage(payload)
		};
	}
	const businessCode = getArchiveErrorCode(payload);
	if (Number.isFinite(businessCode) && ![0, 200, 300].includes(businessCode)) {
		throw payload;
	}
	if (Array.isArray(payload)) {
		payload.forEach(appendEntry);
		return;
	}
	if (Array.isArray(payload.items)) {
		payload.items.forEach(appendEntry);
		return;
	}
	appendEntry(payload);
};

const resetArchiveContents = () => {
	entries.value = [];
	totalCount.value = 0;
	errorMessage.value = '';
	currentArchivePath.value = '';
};

const loadArchiveContents = async () => {
	const url = previewStreamUrl.value;
	if (!url) {
		loading.value = false;
		return;
	}

	try {
		loading.value = true;
		const headers: Record<string, string> = {};
		if (archivePassword.value) {
			headers['X-Archive-Password'] = archivePassword.value;
		}

		const response = await CommonFetch.getStream(url, { headers });
		if (!response.ok || !response.body) {
			const data = await response
				.clone()
				.json()
				.catch(() => ({
					msg: `Archive preview stream failed: ${response.status}`
				}));
			throw { response: { data } };
		}

		const reader = response.body.getReader();
		const decoder = new TextDecoder('utf-8');
		let left = '';

		let reading = true;
		while (reading) {
			const { done, value } = await reader.read();
			const chunk =
				left +
				decoder.decode(value || new Uint8Array(), {
					stream: !done
				});
			const lines = chunk.split('\n');
			left = done ? '' : lines.pop() || '';
			lines.forEach(consumeStreamLine);
			if (done) {
				reading = false;
				if (left) {
					consumeStreamLine(left);
				}
			}
		}
	} catch (error) {
		if (isArchivePasswordError(error)) {
			try {
				const password = await requestArchivePassword(
					getArchiveErrorMessage(error)
				);
				archivePassword.value = password;
				resetArchiveContents();
				await loadArchiveContents();
				return;
			} catch (cancelError) {
				console.warn('Archive password input cancelled', cancelError);
				dialogRef.value.onDialogCancel();
			}
		} else {
			console.warn('Failed to load archive contents', error);
			errorMessage.value =
				getArchiveErrorMessage(error) ||
				t('files.archive.load_archive_contents_failed');
		}
	} finally {
		loading.value = false;
	}
};

const extractAll = () => {
	dialogRef.value.onDialogCancel();
	$q.dialog({
		component: ArchiveExtractDialog,
		componentProps: {
			item: archiveItem.value,
			origin_id: props.origin_id,
			driveType: props.driveType,
			password: archivePassword.value
		}
	});
};

onMounted(loadArchiveContents);
</script>

<style scoped lang="scss">
.archive-preview-dialog {
	width: 760px;
	max-width: 100%;
}

.archive-preview-mobile-header,
.archive-mobile-entry-meta,
.archive-mobile-folder-arrow {
	display: none;
}

.archive-preview-body {
	position: relative;
	width: 760px;
	max-width: 100%;
	height: 400px;
	border-radius: 8px;
	overflow: hidden;
	background: $background-2;
	box-shadow: inset 1px 0 0 $separator, inset -1px 0 0 $separator;

	&::after {
		content: '';
		position: absolute;
		inset: 0;
		z-index: 2;
		pointer-events: none;
		border: 1px solid $separator;
		border-radius: 8px;
	}
}

.archive-preview-toolbar {
	height: 36px;
	padding: 8px;
	border-bottom: 1px solid $separator;
	min-width: 0;
}

.archive-path {
	flex: 1;
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.archive-preview-table-header {
	height: 32px;
	padding: 0 12px;
	border-bottom: 1px solid $separator;
	gap: 12px;
}

.archive-preview-list {
	height: 331px;

	:deep(.q-scrollarea__container),
	:deep(.q-scrollarea__content) {
		height: 100%;
		min-height: 100%;
	}

	// Lock the scroll content to the container width. Otherwise QScrollArea's
	// absolutely-positioned content shrinks to fit the longest (untruncated)
	// file name, which adds an unwanted horizontal scrollbar.
	:deep(.q-scrollarea__content) {
		width: 100%;
		max-width: 100%;
	}
}

.archive-preview-row {
	width: 100%;
	height: 40px;
	padding: 8px 12px;
	gap: 12px;

	&:hover {
		background: $background-hover;

		.archive-action-btn {
			opacity: 1;
			pointer-events: auto;
		}
	}
}

.archive-col-name {
	flex: 1;
	min-width: 0;

	// Keep the file icon at a fixed size so long names can't squeeze it.
	:deep(.file-icon) {
		flex: 0 0 auto;
	}

	.text-subtitle3 {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
}

.archive-col-spacer {
	width: 60px;
	flex: 0 0 60px;
}

.archive-col-size {
	width: 80px;
	flex: 0 0 80px;
}

.archive-col-modified {
	width: 160px;
	flex: 0 0 160px;
	justify-content: flex-end;
	text-align: right;
}

.archive-col-action {
	display: none;
	width: 80px;
	flex: 0 0 80px;
	gap: 4px;
}

.archive-action-btn {
	width: 24px;
	height: 24px;
	min-height: 24px;
	border-radius: 4px;
	background: $background-1;
	color: $ink-2;
	opacity: 0;
	pointer-events: none;
	transition: opacity 0.15s ease;

	:deep(.q-icon) {
		font-size: 16px;
	}
}

.archive-skeleton-name {
	flex: 1;
}

.archive-empty {
	height: 100%;
	min-height: 100%;
	width: 100%;
}

@media (max-width: 599px) {
	:deep(.card-container) {
		width: 100vw !important;
		max-width: 100vw !important;
		height: 700px;
		border-radius: 12px 12px 0 0;
		overflow: hidden;
	}

	:deep(.dialog-content) {
		margin: 0 !important;
		max-height: none !important;
		overflow: hidden !important;
	}

	:deep(.card-action) {
		height: 88px;
		padding: 20px;
		gap: 12px;
		background: $background-1;
	}

	.archive-preview-mobile-header {
		display: flex;
		height: 64px;
		padding: 12px 20px;
		border-radius: 12px 12px 0 0;
		backdrop-filter: blur(30px);
	}

	.archive-preview-dialog,
	.archive-preview-body {
		width: 100%;
		max-width: 100%;
	}

	.archive-preview-body {
		height: 548px;
		border-radius: 0;
		box-shadow: none;

		&::after {
			display: none;
		}
	}

	.archive-preview-toolbar {
		height: 44px;
		padding: 8px 20px;
		border-bottom-width: 0.5px;
	}

	.archive-preview-toolbar > .row:first-child {
		margin-right: 8px;
		flex-shrink: 0;
	}

	.archive-path {
		font-size: 14px;
		line-height: 20px;
		color: $ink-2;
		display: block;
	}

	.archive-preview-toolbar .text-ink-3 {
		flex-shrink: 0;
		font-size: 14px;
		line-height: 20px;
	}

	.archive-preview-table-header {
		display: none;
	}

	.archive-preview-list {
		height: 504px;
	}

	.archive-preview-row {
		height: 72px;
		padding: 12px 20px;
		gap: 12px;
		border-bottom: 0.5px solid $separator;
	}

	.archive-col-spacer {
		display: none;
	}

	.archive-col-size {
		display: none;
	}

	.archive-col-modified {
		display: none;
	}

	.archive-col-name {
		display: grid;
		grid-template-columns: 32px minmax(0, 1fr);
		grid-template-rows: 20px 16px;
		column-gap: 12px;
		align-items: center;
	}

	.archive-col-name :deep(.terminus-file-icon),
	.archive-col-name > :first-child {
		grid-row: 1 / span 2;
		width: 32px !important;
		height: 32px !important;
	}

	.archive-col-name .text-subtitle3 {
		grid-column: 2;
		margin-left: 0;
		font-size: 14px;
		font-weight: 500;
		line-height: 20px;
	}

	.archive-mobile-entry-meta {
		display: block;
		grid-column: 2;
		line-height: 16px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.archive-mobile-folder-arrow {
		display: flex;
		width: 32px;
		height: 32px;
		flex: 0 0 32px;
		color: $ink-2;
	}

	.archive-col-action {
		display: none;
	}

	.archive-skeleton-name {
		margin-left: 12px;
	}
}
</style>
