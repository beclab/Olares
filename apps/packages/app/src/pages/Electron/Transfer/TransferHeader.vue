<template>
	<q-header class="transer-header">
		<slot name="transfer-add"></slot>
		<div style="height: 48px" class="row justify-between">
			<div
				class="row items-center justify-start ellipsis text-ink-1 text-weight-medium"
				v-if="transfer2Store.activeItem === TransferFront.upload"
			>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.UPLOADING
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.UPLOADING)"
				>
					<span>{{ t('transmission.uploading') }}</span>
					<span
						v-if="uploadingNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.UPLOADING
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ uploadingNum }}</span
					>
				</div>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.UPLOADED
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.UPLOADED)"
				>
					<span>{{ t('transmission.completed') }}</span>
					<span
						v-if="uploadedNum > 0"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.UPLOADED
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ uploadedNum }}</span
					>
				</div>
			</div>

			<div
				class="row items-center justify-start ellipsis text-ink-1 text-weight-medium"
				v-if="transfer2Store.activeItem === TransferFront.download"
			>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.DOWNLOADING
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.DOWNLOADING)"
				>
					<span>{{ t('transmission.downloading') }}</span>
					<span
						v-if="downloadingNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.DOWNLOADING
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ downloadingNum }}</span
					>
				</div>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.DOWNLOADED
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.DOWNLOADED)"
				>
					<span>{{ t('transmission.completed') }}</span>
					<span
						v-if="downloadedNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.DOWNLOADED
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ downloadedNum }}</span
					>
				</div>
			</div>

			<div
				class="row items-center justify-start ellipsis text-ink-1 text-weight-medium"
				v-if="transfer2Store.activeItem === TransferFront.cloud"
			>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.CLOUDING
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.CLOUDING)"
				>
					<span>{{ t('transmission.cloud.transferring') }}</span>
					<span
						v-if="cloudingNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.CLOUDING
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ cloudingNum }}</span
					>
				</div>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.CLOUDED
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.CLOUDED)"
				>
					<span>{{ t('transmission.completed') }}</span>
					<span
						v-if="cloudedNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.CLOUDED
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ cloudedNum }}</span
					>
				</div>
			</div>

			<div
				class="row items-center justify-start ellipsis text-ink-1 text-weight-medium"
				v-if="transfer2Store.activeItem === TransferFront.copy"
			>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.COPYING
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.COPYING)"
				>
					<span>{{ t('transmission.pasting') }}</span>
					<span
						v-if="copyingNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.COPYING
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ copyingNum }}</span
					>
				</div>
				<div
					class="transfer-tab row items-center justify-center"
					:class="
						transfer2Store.transferType === TransferType.COPIED
							? 'text-ink-1'
							: 'text-ink-3'
					"
					@click="changeTransferTab(TransferType.COPIED)"
				>
					<span>{{ t('transmission.completed') }}</span>
					<span
						v-if="copiedNum"
						class="q-ml-sm num"
						:class="
							transfer2Store.transferType === TransferType.COPIED
								? 'text-ink-2'
								: 'text-ink-3'
						"
						>{{ copiedNum }}</span
					>
				</div>
			</div>

			<div
				class="row items-center justify-end"
				v-if="
					transfer2Store.transferType === TransferType.UPLOADING ||
					transfer2Store.transferType === TransferType.DOWNLOADING ||
					transfer2Store.transferType === TransferType.CLOUDING ||
					transfer2Store.transferType === TransferType.COPYING
				"
			>
				<div
					class="upload-btn text-body3 q-mr-sm text-ink-1"
					v-if="
						transfer2Store.transferType === TransferType.UPLOADING ||
						transfer2Store.transferType === TransferType.DOWNLOADING ||
						transfer2Store.transferType === TransferType.COPYING
					"
					@click="pauseAction"
					:style="{
						pointerEvents: `${!pauseEnable ? 'none' : 'auto'}`,
						opacity: `${!pauseEnable ? 0.7 : 1}`
					}"
				>
					<q-icon class="q-mr-xs" name="sym_r_pause_circle" size="20px" />
					{{ t('transmission.pause_all') }}
				</div>

				<div
					class="upload-btn text-body3 q-mr-sm text-ink-1"
					v-if="
						transfer2Store.transferType === TransferType.UPLOADING ||
						transfer2Store.transferType === TransferType.DOWNLOADING ||
						transfer2Store.transferType === TransferType.COPYING
					"
					@click="startAction"
					:style="{
						pointerEvents: `${!resumeEnable ? 'none' : 'auto'}`,
						opacity: `${!resumeEnable ? 0.7 : 1}`
					}"
				>
					<q-icon class="q-mr-xs" name="sym_r_play_circle" size="20px" />
					{{ t('transmission.all_Start') }}
				</div>

				<div
					class="upload-btn text-body3 q-mr-sm text-ink-1"
					@click="clearAction"
				>
					<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
					{{ t('transmission.all_clear') }}
				</div>
			</div>

			<div v-else class="row justify-end items-center">
				<div
					class="upload-btn text-body3 q-mr-sm text-ink-1"
					@click="clearAllHistory"
				>
					<q-icon class="q-mr-xs" name="sym_r_format_paint" size="20px" />
					{{ t('transmission.clearAllhistory') }}
				</div>
			</div>
		</div>
	</q-header>
</template>

<script setup lang="ts">
import { watch, computed } from 'vue';
import {
	MenuType,
	useTransfer2Store,
	TransferType,
	TransferItemInMemory
} from '../../../stores/transfer2';

import { useI18n } from 'vue-i18n';

import {
	TransferFront,
	TransferStatus
} from '../../../utils/interface/transfer';

import TransferClient from '../../../services/transfer';

const transfer2Store = useTransfer2Store();

const { t } = useI18n();

watch(
	() => transfer2Store.activeItem,
	(newVal) => {
		if (newVal === TransferFront.upload) {
			transfer2Store.transferType = TransferType.UPLOADING;
		} else if (newVal === TransferFront.download) {
			transfer2Store.transferType = TransferType.DOWNLOADING;
		} else if (newVal === TransferFront.cloud) {
			transfer2Store.transferType = TransferType.CLOUDING;
		} else if (newVal === TransferFront.copy) {
			transfer2Store.transferType = TransferType.COPYING;
		}
	},
	{
		immediate: true
	}
);

const changeTransferTab = (value: TransferType) => {
	transfer2Store.transferType = value;
};

const updateResumeAndPause = () => {
	let taskingData: TransferItemInMemory[] = [];
	if (transfer2Store.activeItem == TransferFront.download) {
		taskingData = Object.values(transfer2Store.transferMap).filter(
			(item) =>
				item.front === TransferFront.download &&
				item.status !== TransferStatus.Completed &&
				item.status !== TransferStatus.Canceled
		);
	} else if (transfer2Store.activeItem == TransferFront.upload) {
		taskingData = Object.values(transfer2Store.transferMap).filter(
			(item) =>
				item.front === TransferFront.upload &&
				item.status !== TransferStatus.Completed &&
				item.status !== TransferStatus.Canceled
		);
	} else if (transfer2Store.activeItem == TransferFront.cloud) {
		taskingData = Object.values(transfer2Store.transferMap).filter(
			(item) =>
				item.front === TransferFront.cloud &&
				item.status !== TransferStatus.Completed &&
				item.status !== TransferStatus.Canceled
		);
	} else if (transfer2Store.activeItem == TransferFront.copy) {
		taskingData = Object.values(transfer2Store.transferMap).filter(
			(item) =>
				(item.front === TransferFront.copy ||
					item.front === TransferFront.move) &&
				item.status !== TransferStatus.Completed &&
				item.status !== TransferStatus.Canceled
		);
	}
	return taskingData;
};

const pauseEnable = computed(() => {
	if (transfer2Store.filesInFolder.length > 0) {
		if (
			Object.values(transfer2Store.filesInFolderMap).findIndex(
				(e) => !e.isPaused
			) >= 0
		) {
			return true;
		} else {
			return false;
		}
	} else {
		const taskingData = updateResumeAndPause();
		if (taskingData && taskingData.findIndex((e) => !e.isPaused) >= 0) {
			return true;
		} else {
			return false;
		}
	}
});

const resumeEnable = computed(() => {
	if (transfer2Store.filesInFolder.length > 0) {
		if (
			Object.values(transfer2Store.filesInFolderMap).findIndex(
				(e) => e.isPaused
			) >= 0
		) {
			return true;
		} else {
			return false;
		}
	} else {
		const taskingData = updateResumeAndPause();

		if (taskingData && taskingData.findIndex((e) => e.isPaused) >= 0) {
			return true;
		} else {
			return false;
		}
	}
});

const getInFolder = (filesInFolder: number[]) => {
	let curFolder: number[] = [];
	if (
		transfer2Store.transferType === TransferType.UPLOADED ||
		transfer2Store.transferType === TransferType.DOWNLOADED
	) {
		curFolder = filesInFolder.filter(
			(id) =>
				transfer2Store.filesInFolderMap[id].status ===
					TransferStatus.Completed ||
				transfer2Store.filesInFolderMap[id].status === TransferStatus.Canceled
		);
	} else {
		curFolder = filesInFolder.filter(
			(id) =>
				transfer2Store.filesInFolderMap[id].status !==
					TransferStatus.Completed &&
				transfer2Store.filesInFolderMap[id].status !== TransferStatus.Canceled
		);
	}
	return curFolder;
};

const clearAction = async () => {
	if (transfer2Store.filesInFolder.length > 0) {
		const ids = getInFolder(transfer2Store.filesInFolder);
		transfer2Store.bulkCancel(ids);
	} else {
		if (transfer2Store.transferType == TransferType.DOWNLOADING) {
			transfer2Store.bulkCancel(transfer2Store.downloading);
		} else if (transfer2Store.transferType == TransferType.UPLOADING) {
			transfer2Store.bulkCancel(transfer2Store.uploading);
		} else if (transfer2Store.transferType == TransferType.CLOUDING) {
			transfer2Store.bulkCancel(transfer2Store.clouding);
		} else if (transfer2Store.transferType == TransferType.COPY) {
			transfer2Store.bulkCancel(transfer2Store.copying);
		}
	}
};

const clearAllHistory = async () => {
	if (transfer2Store.filesInFolder.length > 0) {
		const ids = getInFolder(transfer2Store.filesInFolder);
		transfer2Store.bulkRemove(ids);
	} else {
		if (transfer2Store.transferType == TransferType.UPLOADED) {
			transfer2Store.bulkRemove(transfer2Store.uploadComplete);
		} else if (transfer2Store.transferType == TransferType.DOWNLOADED) {
			transfer2Store.bulkRemove(transfer2Store.downloadComplete);
		} else if (transfer2Store.transferType == TransferType.CLOUDED) {
			if (TransferClient.client.clouder) {
				await TransferClient.client.clouder.removeAllTask();
			}
			transfer2Store.bulkRemove(transfer2Store.cloudComplete);
		} else if (transfer2Store.transferType == TransferType.COPIED) {
			transfer2Store.bulkRemove(transfer2Store.copyComplete);
		}
	}
};

const startAction = async () => {
	if (transfer2Store.filesInFolder.length > 0) {
		const ids = getInFolder(transfer2Store.filesInFolder);
		transfer2Store.bulkResume(ids);
	} else {
		if (transfer2Store.transferType == TransferType.DOWNLOADING) {
			transfer2Store.bulkResume(transfer2Store.downloading);
		} else {
			transfer2Store.bulkResume(transfer2Store.uploading);
		}
	}
};

const pauseAction = async () => {
	if (transfer2Store.filesInFolder.length > 0) {
		const ids = getInFolder(transfer2Store.filesInFolder);
		transfer2Store.bulkPause(ids);
	} else {
		if (transfer2Store.transferType == TransferType.DOWNLOADING) {
			transfer2Store.bulkPause(transfer2Store.downloading);
		} else {
			console.log('pauseAction uploading', transfer2Store.uploading);
			transfer2Store.bulkPause(transfer2Store.uploading);
		}
	}
};

const uploadingNum = computed(() => {
	const uploadingTotal = transfer2Store.uploading.length;

	if (uploadingTotal > 99) {
		return '99+';
	} else {
		return uploadingTotal;
	}
});

const uploadedNum = computed(() => {
	const uploadingTotal = transfer2Store.uploadComplete.length;

	if (uploadingTotal > 99) {
		return '99+';
	} else {
		return uploadingTotal;
	}
});

const downloadingNum = computed(() => {
	const downloadTotal = transfer2Store.downloading.length;

	if (downloadTotal > 99) {
		return '99+';
	} else {
		return downloadTotal;
	}
});

const downloadedNum = computed(() => {
	const downloadTotal = transfer2Store.downloadComplete.length;

	if (downloadTotal > 99) {
		return '99+';
	} else {
		return downloadTotal;
	}
});

const cloudingNum = computed(() => {
	const total = transfer2Store.clouding.length;

	if (total > 99) {
		return '99+';
	} else {
		return total;
	}
});

const cloudedNum = computed(() => {
	const total = transfer2Store.cloudComplete.length;

	if (total > 99) {
		return '99+';
	} else {
		return total;
	}
});

const copyingNum = computed(() => {
	const total = transfer2Store.copying.length;

	if (total > 99) {
		return '99+';
	} else {
		return total;
	}
});

const copiedNum = computed(() => {
	const total = transfer2Store.copyComplete.length;

	if (total > 99) {
		return '99+';
	} else {
		return total;
	}
});
</script>

<style lang="scss">
.transfer-tab {
	font-size: 14px;
	font-weight: 700;
	white-space: nowrap;
	margin-right: 32px;
	cursor: pointer;
	.num {
		width: auto;
		display: flex;
		height: 16px;
		line-height: 16px;
		padding: 0 8px;
		background-color: rgba(0, 0, 0, 0.1);
		border-radius: 8px;
		font-size: 12px;
	}
}

.upload-btn {
	border: 1px solid $btn-stroke;
	padding: 6px 8px;
	border-radius: 8px;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
}
</style>
