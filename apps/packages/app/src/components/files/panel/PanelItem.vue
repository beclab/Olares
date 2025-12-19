<template>
	<div class="uploadItem row items-center justify-between q-py-sm">
		<div class="icon-wrap">
			<img
				v-if="!file.isFolder"
				style="border-radius: 4px"
				class="image"
				:src="fileIcon(file.name)"
				@error.once="
					(e) => {
						e.target.src = fileIcon(file.name);
					}
				"
			/>
			<img v-else :src="folderIcon(file.name)" class="image" />
			<template
				v-if="
					file.status === TransferStatus.Completed ||
					file.status === TransferStatus.Error ||
					(file.status === TransferStatus.Pending && !file.isPaused)
				"
			>
				<div class="row items-center justify-center item-status">
					<div
						class="status-bg row items-center justify-center"
						:class="{
							'bg-negative': file.status === TransferStatus.Error,
							'bg-warning':
								file.status === TransferStatus.Pending && !file.isPaused,
							'bg-positive': file.status === TransferStatus.Completed
						}"
					>
						<q-icon
							:name="
								file.status === TransferStatus.Completed
									? 'sym_r_check'
									: file.status === TransferStatus.Error
									? 'sym_r_exclamation'
									: 'sym_r_schedule'
							"
							size="12px"
							style="color: #ffffff"
						></q-icon>
					</div>
				</div>
			</template>
		</div>

		<div class="content q-ml-sm" v-if="file.front === TransferFront.upload">
			<div class="text-body2 text-ink-1">{{ file.name }}</div>
			<div class="text-ink-2">
				<div class="row items-center">
					<div class="row items-center justify-center">
						<q-icon name="sym_r_upload" color="ink-3" size="16px" />
					</div>
					<div class="text-ink-3 q-ml-xs" style="flex: 1">
						<span v-if="file.isFolder">{{ file.folderCompletedCount }}</span>
						<span v-else>{{
							format.formatFileSize(file.size * file.progress)
						}}</span>
						<span>/</span>
						<span v-if="file.isFolder">{{ file.folderTotalCount }}</span>
						<span v-else>{{ format.formatFileSize(file.size) }}</span>
					</div>
				</div>
			</div>
		</div>

		<div class="content q-ml-sm" v-else>
			<div class="text-body2 text-ink-1">{{ file.name }}</div>
			<div class="row items-center">
				<div class="row items-center justify-center">
					<q-icon name="sym_r_content_copy" color="ink-3" size="16px" />
				</div>
				<div class="text-ink-3" style="flex: 1">
					<span>&nbsp;{{ getCopyDir(file.from, file.isFolder) }}&nbsp;</span>
					<span>
						<q-icon name="sym_r_arrow_forward" size="16px" />
					</span>
					<span style="flex: 1"
						>&nbsp;{{ getCopyDir(file.to, file.isFolder) }}</span
					>
				</div>
			</div>
		</div>

		<div
			class="row items-center justify-center"
			v-if="file.status === TransferStatus.Completed"
		>
			<q-icon
				class="forward text-ink-2"
				rounded
				name="sym_r_search"
				size="sm"
				@click="forWord(file)"
			></q-icon>
		</div>

		<template v-else>
			<div class="row items-center justify-center">
				<div v-if="TransferStatus.Running === file.status">
					<div class="row items-center justify-between" v-if="!file.isPaused">
						<div
							class="text-blue"
							style="
								overflow: hidden;
								white-space: nowrap;
								text-overflow: ellipsis;
							"
						>
							<!-- <span v-if="file.totalPhase > 1" class="q-mr-sm">
								{{ file.currentPhase }}
							</span> -->
							<span> {{ Math.floor(file.progress * 100) }}% </span>
							<!-- <q-icon
								v-if="!file.pauseDisable"
								class="forward q-ml-sm text-ink-2"
								rounded
								name="sym_r_pause"
								size="sm"
								@click="pauseOrResumeAction(file)"
							></q-icon> -->
						</div>
					</div>
					<div class="row items-center justify-center" v-else>
						<div class="text-ink-2">{{ t('download.pause') }}</div>
						<q-icon
							v-if="!file.pauseDisable"
							class="forward q-ml-sm text-ink-2"
							rounded
							name="sym_r_resume"
							size="sm"
							@click="pauseOrResumeAction(file)"
						></q-icon>
					</div>
				</div>

				<div v-else-if="TransferStatus.Error === file.status">
					<div class="row items-center justify-center">
						<!-- <div class="text-red-8"> -->
						<!-- {{ t(`transferStatus.${file.status}`) }} -->
						<!-- </div> -->
						<q-icon size="20px" name="sym_r_error" color="negative">
							<q-tooltip
								maxWidth="240px"
								anchor="top right"
								self="bottom right"
							>
								{{ file.message }}
							</q-tooltip>
						</q-icon>

						<q-icon
							v-if="
								![TransferFront.copy, TransferFront.move].includes(
									file.front
								) &&
								!(file.front === TransferFront.upload && file.currentPhase > 1)
							"
							class="forward q-ml-sm text-ink-2"
							rounded
							name="sym_r_refresh"
							size="sm"
							@click="onUploadRetry(file)"
						></q-icon>
					</div>
				</div>

				<div v-else-if="TransferStatus.Checking === file.status">
					<div class="row items-center justify-center">
						<div class="text-blue">
							{{ t(`transferStatus.${file.status}`) }}...
						</div>
					</div>
				</div>

				<div v-else>
					<div class="row items-center justify-center">
						<div
							class="row items-center justify-center"
							v-if="file.status === TransferStatus.Pending && file.isPaused"
						>
							<div class="text-ink-2">{{ t('download.pause') }}</div>
							<q-icon
								v-if="!file.pauseDisable"
								class="forward q-ml-sm text-ink-2"
								rounded
								name="sym_r_resume"
								size="sm"
								@click="pauseOrResumeAction(file)"
							></q-icon>
						</div>
						<div class="text-ink-2" v-else>
							{{ t(`transferStatus.${file.status}`) }}
						</div>
					</div>
				</div>

				<q-icon
					class="forward q-ml-sm text-ink-2"
					rounded
					name="sym_r_close_small"
					size="sm"
					@click="onUploadCancel(file)"
				></q-icon>
			</div>
		</template>
	</div>
	<div
		v-if="TransferStatus.Running === file.status"
		style="padding-left: 20px; padding-right: 20px"
	>
		<q-linear-progress
			rounded
			size="2px"
			:value="file.progress"
			color="light-blue"
			track-color="backgrund-4"
		/>
	</div>
</template>

<script lang="ts" setup>
// import { useQuasar } from 'quasar';
import { ref, onUnmounted, defineProps, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { dataAPIs } from '../../../api';
import { useTransfer2Store } from '../../../stores/transfer2';
import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { getFileIcon } from '@bytetrade/core';
import { format } from '../../../utils/format';
import {
	TransferStatus,
	TransferFront,
	TransferItemInMemory
} from '../../../utils/interface/transfer';
import {
	notifySuccess,
	notifyFailed
} from '../../../utils/notifyRedefinedUtil';

const props = defineProps({
	file: {
		type: Object as PropType<TransferItemInMemory>,
		required: true
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

// const $q = useQuasar();
const { t } = useI18n();
const store = useDataStore();
const filesStore = useFilesStore();
const transferStore = useTransfer2Store();

const showUpload = ref(true);

const operateIcon = ref<Record<string, { icon: string; bg: string }>>({
	[TransferFront.upload]: {
		icon: 'sym_r_arrow_circle_up',
		bg: '#29CC5F'
	},
	[TransferFront.copy]: {
		icon: 'sym_r_file_copy',
		bg: '#3377FF'
	},
	[TransferFront.move]: {
		icon: 'sym_r_move_up',
		bg: '#51AEFF'
	}
});

onUnmounted(() => {
	showUpload.value = false;
});

const forWord = async (file: TransferItemInMemory) => {
	if (file.isFolder) {
		let url = file.path;
		if ([TransferFront.copy, TransferFront.move].includes(file.front)) {
			url = dataAPIs(file.driveType).getPanelJumpPath(file);
		}
		filesStore.setBrowserUrl(url, file.driveType, false);
	} else {
		if (store.preview.isShow) {
			return;
		}

		const dataAPI = dataAPIs(file.driveType);
		const cur_file = dataAPI.formatTransferToFileItem(file);

		filesStore.openPreviewDialog(cur_file, props.origin_id);
	}
};

const onUploadCancel = async (file) => {
	await transferStore.cancel(file);
	let newName = file.name;
	if (newName.length > 40) {
		newName = newName.substring(0, 40) + '...';
	}

	setTimeout(() => {
		if (transferStore.filesInDialog.length === 0) {
			transferStore.isUploadProgressDialogShow = false;
		}
	}, 1000);
};

const onUploadRetry = async (item: any) => {
	if (item.status == TransferStatus.Error) {
		await transferStore.recoverErrorTransfer(item.id);
		transferStore.onFileError(item.id, item.front, '');
		return;
	}
};

const openErrorToast = (file) => {
	notifyFailed(file.message);
};

const fileIcon = (name: any) => {
	let src = '/img/file-';
	let folderSrc = '/img/file-blob.svg';

	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/file-';
		folderSrc = './img/file-blob.svg';
	}

	if (name.split('.').length > 1) {
		src = src + getFileIcon(name) + '.svg';
	} else {
		src = folderSrc;
	}

	return src;
};

const getCopyDir = (path: string, isFolder: boolean): string => {
	const cleanPath = path.includes('?') ? path.split('?')[0] : path;

	const normalizedPath = cleanPath.endsWith('/')
		? cleanPath.slice(0, -1)
		: cleanPath;

	if (path.includes('?')) {
		const lastSlashIndex = normalizedPath.lastIndexOf('/');

		if (isFolder) {
			const parentDir = normalizedPath.slice(0, lastSlashIndex);
			return parentDir.split('/').pop() || '';
		} else {
			return normalizedPath.slice(lastSlashIndex + 1);
		}
	}

	const paths = normalizedPath.split('/');
	if (paths.length >= 2) {
		return paths[paths.length - 2];
	}

	return paths.pop() || '';
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const folderIcon = (_name: any) => {
	let src = '/img/folder-';

	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/folder-';
	}

	src = src + 'default.svg';
	return src;
};

const pauseOrResumeAction = (item: TransferItemInMemory) => {
	if (item.status == TransferStatus.Error) {
		transferStore.recoverErrorTransfer(item.id);
		return;
	}
	if (item.isPaused) {
		console.log('resume111');

		transferStore.resume(item);
	} else {
		transferStore.pause(item);
	}
};
</script>

<style scoped lang="scss">
.uploadItem {
	padding-left: 20px;
	padding-right: 20px;
	width: 400px;

	.icon-wrap {
		position: relative;
		width: 40px;
		height: 40px;
		.image {
			width: 100%;
			height: 100%;
		}
		.item-status {
			position: absolute;
			right: 0;
			bottom: 0;
			background: $background-1;
			width: 16px;
			height: 16px;
			border-radius: 8px;

			.status-bg {
				width: 14px;
				height: 14px;
				border-radius: 7px;
			}
		}
	}

	.content {
		// max-width: 274px;
		flex: 1;
		overflow: hidden;

		div {
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.copy-dir {
			color: $light-blue-default;
			display: inline-block;
			max-width: 80px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	.margin-left-32 {
		margin-left: 32px;
	}

	.forward {
		cursor: pointer;
	}

	&:hover {
		background-color: $background-hover;
	}
}
</style>
