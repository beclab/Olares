<template>
	<q-item
		v-if="isMobile"
		class="archive-transfer-item q-px-none row items-center justify-between"
		clickable
	>
		<q-item-section
			avatar
			style="max-width: 32px; max-height: 32px; object-fit: cover"
		>
			<div class="icon-wrap row items-center justify-center">
				<terminus-file-icon
					:name="displayName"
					:type="isResultFolder ? 'Folder' : 'archive'"
					:path="targetPath"
					:driveType="task.driveType"
					:modified="0"
					:is-dir="isResultFolder"
					:iconSize="32"
				/>
				<div
					v-if="statusIcon"
					class="status-badge row items-center justify-center"
					:class="statusClass"
				>
					<q-icon :name="statusIcon" size="10px" color="white">
						<q-tooltip
							v-if="task.status === TransferStatus.Error && !!task.message"
							maxWidth="240px"
							anchor="top right"
							self="bottom right"
						>
							{{ task.message }}
						</q-tooltip>
					</q-icon>
				</div>
			</div>
		</q-item-section>

		<q-item-section>
			<q-item-label class="text-subtitle2 text-ink-1 name">
				{{ displayName }}
			</q-item-label>
			<q-item-label class="text-body3 text-ink-3">
				<div>
					<div class="row items-center justify-between">
						<div class="text-body3 text-ink-3 item-info">
							{{ statusText }}
						</div>
						<div
							v-if="task.status === TransferStatus.Running"
							class="text-body3 text-light-blue-default info-status"
						>
							{{ Math.floor(task.progress * 100) }}%
						</div>
					</div>
					<q-linear-progress
						v-if="task.status === TransferStatus.Running"
						rounded
						size="5px"
						:value="task.progress"
						class="q-mt-xs"
						color="green"
					/>
				</div>
			</q-item-label>
		</q-item-section>

		<q-item-section side>
			<div class="row items-center justify-end" style="min-width: 64px">
				<div
					v-if="canPause"
					class="row items-center justify-center item-action"
					@click="pauseOrResume"
				>
					<q-icon
						:name="
							task.isPaused || task.status === TransferStatus.Paused
								? 'sym_r_play_circle'
								: 'sym_r_pause_circle'
						"
						size="20px"
						color="ink-2"
					/>
				</div>
				<div
					v-if="task.status === TransferStatus.Completed"
					class="row items-center justify-center item-action"
					@click="openTarget"
				>
					<q-icon name="sym_r_search" size="20px" color="ink-2" />
				</div>
				<div
					v-else
					class="row items-center justify-center item-action"
					@click="transferStore.cancel(task)"
				>
					<q-icon name="sym_r_close" size="20px" color="ink-2" />
				</div>
			</div>
		</q-item-section>
	</q-item>

	<div v-else class="archive-panel-item">
		<div class="archive-panel-main row items-center justify-between q-py-sm">
			<div class="row items-center item-main">
				<div class="icon-wrap">
					<terminus-file-icon
						:name="displayName"
						:type="isResultFolder ? 'Folder' : 'archive'"
						:path="targetPath"
						:driveType="task.driveType"
						:modified="0"
						:is-dir="isResultFolder"
						:iconSize="40"
					/>
					<div
						v-if="statusIcon"
						class="row items-center justify-center item-status"
					>
						<div
							class="status-bg row items-center justify-center"
							:class="statusClass"
						>
							<q-icon :name="statusIcon" size="12px" color="white">
								<q-tooltip
									v-if="task.status === TransferStatus.Error && !!task.message"
									maxWidth="240px"
									anchor="top right"
									self="bottom right"
								>
									{{ task.message }}
								</q-tooltip>
							</q-icon>
						</div>
					</div>
				</div>

				<div class="q-ml-sm item-content content">
					<div class="text-body2 text-ink-1">
						{{ displayName }}
					</div>
					<div class="row items-center text-body3 text-ink-3">
						<q-icon name="sym_r_archive" size="16px" class="q-mr-xs" />
						<span>{{ statusText }}</span>
					</div>
				</div>
			</div>

			<div class="row items-center justify-end item-actions">
				<span
					v-if="task.status === TransferStatus.Running"
					class="text-body3 text-light-blue-default q-mr-sm"
				>
					{{ Math.floor(task.progress * 100) }}%
				</span>

				<q-icon
					v-if="canPause"
					class="cursor-pointer text-ink-2 q-mr-sm"
					:name="
						task.isPaused || task.status === TransferStatus.Paused
							? 'sym_r_play_circle'
							: 'sym_r_pause_circle'
					"
					size="20px"
					@click="pauseOrResume"
				/>
				<q-icon
					v-if="task.status === TransferStatus.Completed"
					class="cursor-pointer text-ink-2"
					name="sym_r_search"
					size="20px"
					@click="openTarget"
				/>
				<q-icon
					v-else
					class="cursor-pointer text-ink-2"
					name="sym_r_close"
					size="20px"
					@click="transferStore.cancel(task)"
				/>
			</div>
		</div>

		<div v-if="task.status === TransferStatus.Running" class="progress-wrap">
			<q-linear-progress
				rounded
				size="2px"
				:value="task.progress"
				color="positive"
				track-color="background-4"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, PropType } from 'vue';
import TerminusFileIcon from '../../common/TerminusFileIcon.vue';
import { FilesIdType, useFilesStore } from '../../../stores/files';
import { useTransfer2Store } from '../../../stores/transfer2';
import { format } from '../../../utils/format';
import {
	TransferFront,
	TransferItemInMemory,
	TransferStatus
} from '../../../utils/interface/transfer';
import { useI18n } from 'vue-i18n';
import { dataAPIs } from 'src/api';

const props = defineProps({
	task: {
		type: Object as PropType<TransferItemInMemory>,
		required: true
	},
	mode: {
		type: String as PropType<'desktop' | 'mobile'>,
		required: false,
		default: 'desktop'
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();
const filesStore = useFilesStore();
const transferStore = useTransfer2Store();

const isMobile = computed(() => props.mode === 'mobile');
const displayName = computed(() => props.task.name);
const isResultFolder = computed(
	() =>
		props.task.front === TransferFront.extract &&
		props.task.status === TransferStatus.Completed
);
const targetPath = computed(() => props.task.to || props.task.path);
const canPause = computed(
	() =>
		!props.task.pauseDisable &&
		[
			TransferStatus.Running,
			TransferStatus.Pending,
			TransferStatus.Paused
		].includes(props.task.status)
);

const statusIcon = computed(() => {
	if (props.task.status === TransferStatus.Completed) {
		return 'sym_r_check';
	}
	if (props.task.status === TransferStatus.Error) {
		return 'sym_r_exclamation';
	}
	if (
		props.task.status === TransferStatus.Pending ||
		props.task.status === TransferStatus.Paused ||
		props.task.isPaused
	) {
		return 'sym_r_schedule';
	}
	return '';
});

const statusClass = computed(() => ({
	'bg-positive': props.task.status === TransferStatus.Completed,
	'bg-negative': props.task.status === TransferStatus.Error,
	'bg-warning':
		props.task.status === TransferStatus.Pending ||
		props.task.status === TransferStatus.Paused ||
		props.task.isPaused
}));

const operationText = computed(() =>
	props.task.front === TransferFront.compress
		? t('files.archive.compressing')
		: t('files.archive.extracting')
);

const completedText = computed(() =>
	props.task.front === TransferFront.compress
		? t('files.archive.compressed')
		: t('files.archive.extracted')
);

const statusText = computed(() => {
	if (props.task.status === TransferStatus.Completed) {
		return `${completedText.value} ${formatProgressSize(
			props.task.size,
			props.task.size
		)}`;
	}
	if (props.task.status === TransferStatus.Error) {
		return isMobile.value
			? props.task.message || t('Failed')
			: t(`transferStatus.${TransferStatus.Error}`);
	}
	if (props.task.status === TransferStatus.Paused || props.task.isPaused) {
		return `${t('download.paused')} ${formatProgressSize(
			props.task.bytes || 0,
			props.task.size
		)}`;
	}
	if (props.task.status === TransferStatus.Pending) {
		return `${t('pending')} ${formatProgressSize(
			props.task.bytes || 0,
			props.task.size
		)}`;
	}
	return `${operationText.value} ${formatProgressSize(
		props.task.bytes || 0,
		props.task.size
	)}`;
});

const formatProgressSize = (bytes: number, size: number) => {
	if (size <= 0) {
		return '--';
	}

	return `${format.formatFileSize(bytes)} / ${format.formatFileSize(size)}`;
};

const withTrailingSlash = (path: string) => {
	return path.endsWith('/') ? path : `${path}/`;
};

const openTarget = () => {
	if (isResultFolder.value) {
		filesStore.setBrowserUrl(
			withTrailingSlash(targetPath.value),
			props.task.driveType,
			false
		);
	} else {
		const dataAPI = dataAPIs(props.task.driveType);
		const cur_file = dataAPI.formatTransferToFileItem(props.task);

		if (isMobile.value) {
			filesStore.openPreviewDialog(cur_file);
			return;
		}

		filesStore.openPreviewDialog(cur_file, props.origin_id);
	}
};

const pauseOrResume = () => {
	if (props.task.isPaused || props.task.status === TransferStatus.Paused) {
		transferStore.resume(props.task);
	} else {
		transferStore.pause(props.task);
	}
};
</script>

<style scoped lang="scss">
.archive-transfer-item {
	height: 72px;
	border-bottom: 1px solid $separator;

	.icon-wrap {
		position: relative;
		width: 32px;
		height: 32px;
	}

	.status-badge {
		position: absolute;
		right: -2px;
		bottom: -2px;
		width: 14px;
		height: 14px;
		border-radius: 50%;
		border: 1px solid $background-1;
	}

	.name,
	.item-info,
	.info-status {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.item-info {
		width: 70%;
		text-align: left;
	}

	.info-status {
		width: 30%;
		text-align: right;
	}

	.item-action {
		width: 32px;
		height: 32px;
	}
}

.archive-panel-item {
	width: 400px;

	.archive-panel-main {
		padding-left: 20px;
		padding-right: 20px;
		width: 400px;
	}

	.item-main {
		min-width: 0;
		flex: 1;
	}

	.item-content {
		min-width: 0;
	}

	.item-actions {
		min-width: 76px;
	}

	.icon-wrap {
		position: relative;
		width: 40px;
		height: 40px;
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

	.content {
		flex: 1;
		overflow: hidden;

		div,
		span {
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}

	.progress-wrap {
		padding: 0 20px 8px;
	}

	&:hover {
		background-color: $background-hover;
	}
}
</style>
