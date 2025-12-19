<template>
	<bt-scroll-area class="upload-only">
		<div
			class="drop-zone"
			:class="{ 'drag-over': isDragOver }"
			@dragover.prevent="handleDragOver"
			@dragleave.prevent="handleDragLeave"
			@drop.prevent="handleDrop"
			@click.stop="uploadFiles"
		>
			<div class="drop-zone-content column items-center justify-center">
				<q-icon name="sym_r_file_copy" size="44px" color="ink-3" />
				<p class="q-mt-md text-body1 text-ink-3">
					{{ t('share.Drag and drop files or directories here') }}
				</p>
			</div>
		</div>
		<input
			ref="fileInput"
			type="file"
			multiple
			style="display: none"
			@change="handleFileSelect"
		/>
		<div class="operations row items-center justify-center">
			<div class="action-btn text-body3 row items-center">
				<q-icon name="sym_r_upload" class="text-ink-1 q-mr-xs" size="24px" />
				<span>
					{{ t('buttons.upload') }}
				</span>
				<q-menu class="popup-menu bg-background-2" :offset="[0, 5]">
					<q-list dense padding style="min-width: 150px">
						<q-item
							class="text-ink-2"
							style="height: 40px; padding: 0 6px; border-radius: 4px"
							clickable
							v-close-popup
							@click="uploadFiles"
						>
							<div class="row items-center">
								<div>
									<q-icon name="sym_r_upload_file" color="ink-2" size="24px" />
								</div>
								<div class="text-body1 text-ink-2 q-ml-sm">
									{{ t('files_popup_menu.upload_file') }}
								</div>
							</div>
						</q-item>

						<q-item
							class="text-ink-2"
							style="height: 40px; padding: 0 6px; border-radius: 4px"
							clickable
							v-close-popup
							@click="uploadFolder"
						>
							<div class="row items-center">
								<div>
									<q-icon
										name="sym_r_drive_folder_upload"
										color="ink-2"
										size="24px"
									/>
								</div>
								<div class="text-body1 text-ink-2 q-ml-sm">
									{{ t('files_popup_menu.upload_folder') }}
								</div>
							</div>
						</q-item>
					</q-list>
				</q-menu>
			</div>
			<div
				class="clear-btn q-ml-lg text-body3 row items-center"
				@click="clearAction"
			>
				<span>
					{{ t('transmission.all_clear') }}
				</span>
			</div>
		</div>
		<div v-if="pagination.rowsNumber > 0">
			<div class="header-progress column justify-center">
				<div class="row items-center justify-between">
					<div
						class="text-ink-1 text-subtitle1"
						v-if="uploadedList.length < pagination.rowsNumber"
					>
						{{ t('share.Uploading files') + '...' }}
						<span class="text-subtitle2 text-light-blue-default">
							{{ format.formatFileSize(uploadSpeed || 0) }} /s</span
						>
					</div>
					<div v-else class="text-ink-1 text-subtitle1">
						{{ t('vault_t.upload_complete') }}
					</div>
					<div>
						{{ uploadedList.length }}/{{ pagination.rowsNumber }}
						{{ t('files.items') }}
					</div>
				</div>
				<q-linear-progress
					class="q-mt-sm"
					size="4px"
					rounded
					:value="uploadedList.length / pagination.rowsNumber"
					color="positive"
					trackColor="background3"
				/>
			</div>
			<q-separator />
			<QTableStyle2>
				<q-table
					:rows="datas"
					:columns="columnsProcess"
					color="primary"
					row-key="id"
					hide-header
					v-model:pagination="pagination"
					@request="onRequest"
					flat
					binary-state-sort
					:loading="false"
					:rows-per-page-label="$t('share.Records per page')"
					:tableRowStyleFn="
						() => {
							return 'height: 64px';
						}
					"
				>
					<template v-slot:body-cell-name="props">
						<q-td :props="props" no-hover>
							<div class="row items-center">
								<terminus-file-icon
									class="q-mr-sm"
									:name="props.row.name"
									:type="props.row.type"
									:path="props.row.path"
									:driveType="props.row.driveType"
									:modified="0"
									:is-dir="props.row.isFolder"
									:iconSize="32"
								/>
								<div
									style="
										width: calc(100% - 40px);
										overflow: hidden;
										white-space: nowrap;
										text-overflow: ellipsis;
									"
								>
									{{ props.row.name }}
								</div>
							</div>
						</q-td>
					</template>
					<template v-slot:body-cell-size="props">
						<q-td :props="props" no-hover>
							{{ format.formatFileSize(props.row.size) }}
						</q-td>
					</template>

					<template v-slot:body-cell-status="props">
						<q-td :props="props" no-hover>
							<div
								class="row items-center"
								v-if="
									TransferStatus.Running === props.row.status ||
									TransferStatus.Pending === props.row.status
								"
							>
								<div class="column q-mr-lg" style="flex: 1">
									<q-linear-progress
										class="q-mt-sm"
										size="4px"
										rounded
										:value="
											props.row.isFolder
												? props.row.folderCompletedCount /
												  props.row.folderTotalCount
												: props.row.progress
										"
										color="positive"
										trackColor="background3"
									/>
									<div
										class="text-light-blue-default text-body3 q-mt-xs"
										v-if="TransferStatus.Running === props.row.status"
									>
										{{ t('files.leftTimes') }}
										{{ formatLeftTimes(props.row.leftTime) }}
									</div>
									<div
										class="text-light-blue-default text-body3 q-mt-xs"
										v-else
									>
										{{ t(`transferStatus.${props.row.status}`) }}
									</div>
								</div>
								<div
									style="flex: 0 0 32; cursor: pointer"
									class="row items-center justify-center"
									@click="deleteItem(props.row)"
								>
									<q-icon name="sym_r_cancel" size="20px" color="negative" />
								</div>
							</div>
							<div v-else>
								<div class="text-positive text-body3">
									{{ t(`transferStatus.${props.row.status}`) }}
								</div>
							</div>
						</q-td>
					</template>
				</q-table>
			</QTableStyle2>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useFilesStore } from 'src/stores/files';
import { useShareStore } from 'src/stores/share/share';
import { DriveType } from 'src/utils/interface/files';
import QTableStyle2 from 'src/apps/controlPanelCommon/components/QTableStyle2.vue';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useTransfer2Store, formatLeftTimes } from 'src/stores/transfer2';
import TerminusFileIcon from '../../components/common/TerminusFileIcon.vue';
import { format } from '../../utils/format';
import { TransferItem, TransferStatus } from '../../utils/interface/transfer';
import { scanFiles } from '../../utils/upload';
import { getApplication } from 'src/application/base';

const { t } = useI18n();

const fileInput = ref<any>(null);

const fileStore = useFilesStore();
const shareStore = useShareStore();

const isDragOver = ref(false);

const handleDragOver = (event: any) => {
	isDragOver.value = true;
	event.dataTransfer.dropEffect = 'copy';
};

const handleDragLeave = (event: {
	currentTarget: { getBoundingClientRect: () => any };
	clientX: any;
	clientY: any;
}) => {
	const rect = event.currentTarget.getBoundingClientRect();
	const x = event.clientX;
	const y = event.clientY;

	if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
		isDragOver.value = false;
	}
};

const handleDrop = async (event: { dataTransfer: any }) => {
	isDragOver.value = false;

	let dt = event.dataTransfer;

	if (dt.files.length <= 0) return;

	let files = await scanFiles(dt);

	const dataTransfer = new DataTransfer();
	const remove: File[] = [];
	files.forEach((file: any) => {
		if (file instanceof File) {
			if (
				shareStore.share?.upload_size_limit == undefined ||
				file.size <= shareStore.share.upload_size_limit
			) {
				dataTransfer.items.add(file);
			} else {
				remove.push(file);
			}
		}
	});

	if (getApplication().filesUploadConfig?.toastDeleteFiles) {
		getApplication().filesUploadConfig?.toastDeleteFiles!(remove);
	}

	fileStore.uploadSelectFile(
		{
			target: {
				files: dataTransfer.files
			}
		},
		{
			isDir: true,
			path: `/Share/${shareStore.share?.id}/`,
			driveType: DriveType.PublicShare,
			param: ''
		}
	);
};

const handleFileSelect = (event: { target: { files: any } }) => {
	const selectedFiles = event.target.files;
	if (selectedFiles && selectedFiles.length > 0) {
		if (getApplication().filesUploadConfig?.filesFilter) {
			const files = getApplication().filesUploadConfig?.filesFilter!(
				event.target.files
			);
			if (files) {
				event.target.files = files;
				fileStore.uploadSelectFile(event, {
					isDir: true,
					path: `/Share/${shareStore.share?.id}/`,
					driveType: DriveType.PublicShare,
					param: ''
				});
			}
		}
	}
};

const uploadFiles = () => {
	if (fileInput.value) {
		fileInput.value.value = '';
		fileInput.value.removeAttribute('webkitdirectory', 'false');
		fileInput.value.click();
	}
};

const uploadFolder = () => {
	if (fileInput.value) {
		fileInput.value.value = '';
		fileInput.value.setAttribute('webkitdirectory', 'true');
		fileInput.value.click();
	}
};

const transferStore = useTransfer2Store();

const pagination = ref({
	page: 1,
	rowsPerPage: 10,
	rowsNumber: shareStore.path_id
		? transferStore.shareUploadList(shareStore.path_id).length
		: 0
});

const onRequest = (props: any) => {
	pagination.value = props.pagination;
};

watch(
	() => transferStore.transfers.length,
	() => {
		if (!shareStore.path_id) {
			return;
		}
		pagination.value.rowsNumber = transferStore.shareUploadList(
			shareStore.path_id
		).length;
	}
);

const uploadedList = computed(() => {
	if (!shareStore.path_id) {
		return [];
	}
	return transferStore.shareUploadedList(shareStore.path_id);
});

const datas = computed(() => {
	if (!shareStore.path_id) {
		return [];
	}
	const list = transferStore
		.shareUploadList(shareStore.path_id)
		.slice(
			(pagination.value.page - 1) *
				(pagination.value.rowsPerPage > 0
					? pagination.value.rowsPerPage
					: pagination.value.rowsNumber),
			pagination.value.page *
				(pagination.value.rowsPerPage > 0
					? pagination.value.rowsPerPage
					: pagination.value.rowsNumber)
		);

	const i = mapItems(list, transferStore.transferMap);

	return i;
});

const uploadSpeed = computed(() => {
	if (!shareStore.path_id) {
		return 0;
	}
	let speed = 0;
	datas.value.forEach((item) => {
		if (item.status == TransferStatus.Running && !item.isPaused)
			speed += item.speed;
	});
	return speed;
});

function mapItems(items: any[], mapStore: Record<string, any>) {
	return items.map((id) => ({
		id,
		...mapStore[id]
	}));
}

const columnsProcess = [
	{
		name: 'name',
		required: true,
		label: t('files.name'),
		align: 'left',
		field: (row: { name: any }) => row.name,
		format: (val: any) => `${val}`,
		sortable: false
	},
	{
		name: 'size',
		label: t('files.size'),
		align: 'left',
		sortable: false,
		style: 'width: 120px'
	},
	{
		name: 'status',
		label: t('transmission.status'),
		sortable: false,
		align: 'left',
		field: (row: { status: any }) => row.status,
		format: (val: any) => `${val}`,
		style: 'width: 400px'
	}
];

const deleteItem = async (item: TransferItem) => {
	transferStore.cancel(item);
};

const clearAction = async () => {
	if (!shareStore.path_id) {
		return;
	}

	transferStore.bulkCancel(
		transferStore.shareUploadingList(shareStore.path_id)
	);

	setTimeout(() => {
		transferStore.bulkRemove(
			transferStore.shareUploadedList(shareStore.path_id!)
		);
	}, 300);
};
</script>

<style scoped lang="scss">
.upload-only {
	height: calc(100% - 120px);
	padding: 20px;
	.drop-zone {
		height: 240px;
		background-color: $background-3;
		border: 1px dashed $separator;
		border-radius: 12px;
		text-align: center;
		transition: all 0.3s ease;
		cursor: pointer;
	}

	.drop-zone.drag-over {
		background-color: #e8f4f8;
		border-color: #007aff;
	}
	.drop-zone-content {
		height: 100%;
	}

	.operations {
		height: 80px;
		.action-btn {
			border: 1px solid $yellow;
			background-color: $yellow;
			display: inline-block;
			color: $grey-10;
			padding: 7px 12px;
			border-radius: 8px;
			cursor: pointer;

			&:hover {
				background-color: $yellow-3;
			}
		}

		.clear-btn {
			border: 1px solid $separator;
			display: inline-block;
			padding: 12px 12px;
			border-radius: 8px;
			cursor: pointer;

			&:hover {
				background-color: $background-3;
			}
		}
	}

	.header-progress {
		height: 76px;
	}
}
</style>
