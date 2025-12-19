<template>
	<div class="transfer-page-root q-px-lg">
		<div
			class="q-ml-sm text-body-2 row items-center justify-start folder-header"
			v-if="
				[
					TransferType.UPLOADING,
					TransferType.DOWNLOADING,
					TransferType.COPYING
				].includes(transferStore.transferType) && folderName
			"
		>
			<span class="text-ink-2 cursor-pointer" @click="openFolder">{{
				transferStore.transferType === TransferType.UPLOADING
					? t('transmission.uploading')
					: transferStore.transferType === TransferType.DOWNLOADING
					? t('transmission.downloading')
					: t('transmission.copying')
			}}</span>
			<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
			<span class="folder-name"> {{ folderName }} </span>
		</div>

		<div
			class="q-ml-sm text-body-2 row items-center justify-start folder-header"
			v-if="
				[
					TransferType.UPLOADED,
					TransferType.DOWNLOADED,
					TransferType.COPIED
				].includes(transferStore.transferType) && folderName
			"
		>
			<span class="text-ink-2 cursor-pointer" @click="openFolder">{{
				t('transmission.completed')
			}}</span>
			<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
			<span>
				{{ folderName }}
			</span>
		</div>

		<div
			class="empty column items-center justify-center"
			v-if="datas.length == 0"
		>
			<img src="../../../assets/nodata.svg" alt="empty" />
			<span class="text-body2">{{ $t('files.lonely') }}</span>
		</div>

		<div
			class="table-content"
			:style="{ '--height': folderName ? '30px' : '0px' }"
			v-else
		>
			<q-table
				v-if="datas.length > 0"
				flat
				:rows="datas"
				:columns="
					[TransferType.DOWNLOADING, TransferType.UPLOADING].includes(
						transferStore.transferType
					)
						? columnsProcess
						: columnsCompleted
				"
				row-key="id"
				hide-pagination
				hide-bottom
				@row-dblclick="openFolderDetail"
				v-model:pagination="pagination"
				:style="{
					height: datas.length > 30 ? 'calc(100% - 40px)' : 'calc(100% - 0px)'
				}"
				class="custom-table"
			>
				<template v-slot:body-cell-name="props">
					<q-td :props="props">
						<div class="row items-center">
							<!-- <img
								class="q-mr-sm"
								:src="filesIcon(props.row.name, props.row.isFolder)"
								style="width: 20px; height: 20px"
							/> -->
							<terminus-file-icon
								class="q-mr-sm"
								:name="props.row.name"
								:type="props.row.type"
								:path="props.row.path"
								:driveType="props.row.driveType"
								:modified="0"
								:is-dir="props.row.isFolder"
								:iconSize="20"
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
					<q-td :props="props">
						<div class="row items-center justify-start">
							<div
								v-if="
									props.row.isFolder &&
									![TransferFront.copy, TransferFront.move].includes(
										props.row.front
									)
								"
							>
								{{ props.row.folderCompletedCount }} /
								{{ props.row.folderTotalCount }}
							</div>

							<div v-else>
								{{ format.formatFileSize(props.row.size) }}
								{{
									props.row.status !== TransferStatus.Completed
										? props.row.size === 0
											? '(0%)'
											: `(${Math.round(props.row.progress * 100) + '%'})`
										: ''
								}}
							</div>
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-status="props">
					<q-td :props="props">
						<div v-if="TransferStatus.Running === props.row.status">
							<div
								class="row items-center justify-between"
								v-if="!props.row.isPaused"
							>
								<div
									class="text-blue"
									style="
										width: 100%;
										overflow: hidden;
										white-space: nowrap;
										text-overflow: ellipsis;
									"
								>
									{{ format.formatFileSize(props.row.speed) }}
									/s
								</div>
							</div>
							<div class="row items-center justify-center" v-else>
								<div class="text-ink-1">{{ t('download.pause') }}</div>
							</div>
						</div>

						<div v-else-if="TransferStatus.Error === props.row.status">
							<div
								class="row items-center justify-start"
								@click.stop="toastTransferItemMessage(props.row.message)"
							>
								<div class="text-red-8">
									{{ t(`transferStatus.${props.row.status}`) }}
								</div>
								<q-icon
									v-if="props.row.message"
									name="sym_r_error"
									class="q-ml-sm text-red-8"
									size="20px"
								/>
							</div>
						</div>

						<div v-else-if="TransferStatus.Checking === props.row.status">
							<div class="row items-center justify-start">
								<div class="text-blue">
									{{ t(`transferStatus.${props.row.status}`) }}...
								</div>
							</div>
						</div>

						<div v-else-if="TransferStatus.Canceled === props.row.status">
							<div class="row items-center justify-start">
								<div class="text-ink-3">
									{{ t(`transferStatus.${props.row.status}`) }}
								</div>
							</div>
						</div>

						<div v-else>
							<div class="row items-center justify-start">
								<div
									class="row items-center justify-start"
									v-if="
										props.row.status === TransferStatus.Pending &&
										props.row.isPaused
									"
								>
									<div class="text-ink-1">{{ t('download.pause') }}</div>
								</div>
								<div class="text-ink-1" v-else>
									{{ t(`transferStatus.${props.row.status}`) }}
								</div>
							</div>
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-leftTimes="props">
					<q-td :props="props">
						<div
							v-if="
								[
									TransferType.DOWNLOADING,
									TransferType.UPLOADING,
									TransferType.COPYING
								].includes(transferStore.transferType)
							"
						>
							{{ formatLeftTimes(props.row.leftTime) }}
						</div>
						<div v-else>
							<div
								class="row items-center justify-start"
								style="min-width: 100px"
							>
								<div
									class="q-mr-sm"
									style="
										width: calc(100% - 30px);
										overflow: hidden;
										white-space: nowrap;
										text-overflow: ellipsis;
									"
								>
									{{
										formatStampTime(
											props.row.endTime
												? props.row.endTime
												: props.row.startTime
										)
									}}
								</div>
							</div>
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-actions="props">
					<q-td :props="props" class="text-ink-1">
						<div
							v-if="
								[
									TransferType.DOWNLOADING,
									TransferType.UPLOADING,
									TransferType.CLOUDING,
									TransferType.COPYING
								].includes(transferStore.transferType)
							"
						>
							<q-icon
								v-if="
									props.row.status == TransferStatus.Error &&
									[
										TransferType.DOWNLOADING,
										TransferType.UPLOADING,
										TransferType.CLOUDING
									].includes(transferStore.transferType)
								"
								name="sym_r_refresh"
								size="20px"
								@click.stop="pauseOrResumeAction(props.row)"
							/>

							<q-icon
								v-if="
									(props.row.status === TransferStatus.Running ||
										(props.row.status === TransferStatus.Pending &&
											props.row.isPaused == true)) &&
									!props.row.pauseDisable
								"
								:name="
									props.row.isPaused
										? 'sym_r_play_circle'
										: 'sym_r_pause_circle'
								"
								size="20px"
								@click.stop="pauseOrResumeAction(props.row)"
							/>
							<q-icon
								v-if="props.row.status !== TransferStatus.Canceling"
								name="sym_r_delete"
								size="20px"
								class="q-ml-sm"
								@click.stop="deleteItem(props.row)"
							/>
						</div>
						<div v-else>
							<q-icon
								name="sym_r_folder"
								size="20px"
								class="cursor-pointer"
								v-if="
									(props.row.isFolder &&
										((props.row.front == TransferFront.upload &&
											props.row.folderCompletedCount > 0) ||
											props.row.front != TransferFront.upload)) ||
									!props.row.isFolder
								"
								@click.stop="openItem(props.row)"
							/>
							<q-icon
								v-if="
									(props.row.isFolder &&
										(props.row.status === TransferStatus.Completed ||
											props.row.status === TransferStatus.Canceled)) ||
									!props.row.isFolder
								"
								class="cursor-pointer q-ml-sm"
								name="sym_r_delete"
								size="20px"
								@click.stop="deleteItem(props.row)"
							/>
						</div>
					</q-td>
				</template>
			</q-table>
			<div
				class="row justify-center"
				style="height: 40px"
				v-if="datas.length > 30"
			>
				<q-pagination
					v-model="pagination.page"
					:max="pagesNumber"
					input
					round
					icon-first="sym_r_keyboard_double_arrow_left"
					icon-last="sym_r_keyboard_double_arrow_right"
					icon-next="sym_r_keyboard_arrow_right"
					icon-prev="sym_r_keyboard_arrow_left"
					color="ink-3"
					input-class="text-ink-2 text-body-3"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from 'vue';
import {
	useTransfer2Store,
	TransferType,
	formatLeftTimes
} from '../../../stores/transfer2';
import { useFilesStore } from '../../../stores/files';
import { useQuasar } from 'quasar';
import { formatStampTime } from '../../../utils/utils';
import {
	TransferFront,
	TransferStatus,
	TransferItem
} from '../../../utils/interface/transfer';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useTermipassStore } from '../../../stores/termipass';
import { TermiPassStatus } from '../../../utils/termipassState';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import {
	notifyFailed,
	notifyWarning
} from '../../../utils/notifyRedefinedUtil';
import { dataAPIs } from './../../../api';
import { useMenuStore } from '../../../stores/menu';
import { LayoutMenu } from '../../../utils/constants';
import { fileList } from '../../../utils/constants';

import TransferClient from '../../../services/transfer';
import { format } from '../../../utils/format';
import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';

const filesStore = useFilesStore();
const transferStore = useTransfer2Store();
const menuStore = useMenuStore();

const $q = useQuasar();
const router = useRouter();

const { t } = useI18n();
const folderName = ref();

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
		style: 'width: 120px'
	},
	{
		name: 'leftTimes',
		label: t('files.leftTimes'),
		sortable: false,
		align: 'left',
		field: (row: { leftTimes: any }) => row.leftTimes,
		format: (val: any) => `${val}`,
		style: 'width: 100px'
	},
	{
		name: 'actions',
		label: t('transmission.action'),
		field: 'carbs',
		align: 'right',
		style: 'width: 90px'
	}
];

const columnsCompleted = [
	{
		name: 'name',
		label: t('files.name'),
		align: 'left',
		field: (row: { name: any }) => row.name,
		format: (val: any) => `${val}`
	},
	{
		name: 'size',
		label: t('files.size'),
		align: 'left',
		style: 'width: 100px'
	},
	{
		name: 'status',
		label: t('transmission.status'),
		sortable: false,
		align: 'left',
		field: (row: { status: any }) => row.status,
		format: (val: any) => `${val}`,
		style: 'width: 100px'
	},
	{
		name: 'leftTimes',
		label: t('files.time'),
		align: 'left',
		field: (row: { leftTimes: any }) => row.leftTimes,
		format: (val: any) => `${val}`,
		style: 'width: 100px'
	},

	{
		name: 'actions',
		label: t('transmission.action'),
		align: 'right',
		style: 'width: 90px'
	}
];

const pauseOrResumeAction = (item: any) => {
	const termipassStore = useTermipassStore();
	if (termipassStore.totalStatus?.status === TermiPassStatus.OfflineMode) {
		return BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('offline_message')
		});
	}
	if (item.status == TransferStatus.Error) {
		transferStore.recoverErrorTransfer(item.id);
		return;
	}
	if (item.isPaused) {
		transferStore.resume(item);
	} else {
		transferStore.pause(item);
	}
};

const filesIcon = (name: string, isFolder: boolean) => {
	const h = name?.substring(name?.lastIndexOf('.') + 1);
	let src = '/img/';
	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/';
	}

	if (isFolder) {
		src = src + 'folder.svg';
		return src;
	}

	const hasFile = fileList.find((item: any) => item === h);
	if (hasFile) {
		src = src + h + '.png';
	} else {
		src = src + 'blob.png';
	}
	return src;
};

const deleteItem = async (item: TransferItem) => {
	const termipassStore = useTermipassStore();

	if (item.front == TransferFront.cloud) {
		if (
			item.status === TransferStatus.Completed ||
			item.status === TransferStatus.Canceled
		) {
			await TransferClient.client.clouder?.removeTask(item, false);
			item.id && transferStore.remove(item.id);
		}
		await TransferClient.client.clouder?.cancelTask(item);
		transferStore.cancel(item);
		return;
	}
	if (termipassStore.totalStatus?.status === TermiPassStatus.OfflineMode) {
		return BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('offline_message')
		});
	}

	if (
		item.status === TransferStatus.Completed ||
		item.status === TransferStatus.Canceled
	) {
		item.id && transferStore.remove(item.id);
	} else {
		transferStore.cancel(item);
	}
};

const openItem = async (item: TransferItem) => {
	if ($q.platform.is.electron) {
		if (item.front == TransferFront.download && item.to) {
			const result = await window.electron.api.transfer.openFileInFolder(
				item.to
			);
			if (!result) {
				notifyFailed(t('The file does not exist or has been deleted'));
			}
			return;
		}
	}

	if (
		item.front === TransferFront.upload ||
		item.front === TransferFront.cloud ||
		item.front == TransferFront.copy ||
		item.front == TransferFront.move
	) {
		uploadPreview(item);
	}
};

const uploadPreview = async (item: TransferItem) => {
	try {
		const res = dataAPIs(item.driveType).formatTransferToFileItem(item);
		const dataAPI = dataAPIs(res.driveType);

		if (res) {
			await dataAPI.transferItemBackToFiles(item, router);
			const curLayoutMenu = LayoutMenu[0];
			menuStore.pushTerminusMenuCache(curLayoutMenu.identify);
		}
	} catch (error) {
		console.log('error', error);
		notifyWarning(t('transmission.hasDelete'));
	}
};

const detailTask = ref();

const openFolderDetail = (evt, row, index) => {
	console.group('openFolderDetail');
	console.log(evt);
	console.log(row);
	console.log(index);
	console.groupEnd();

	if (row.isFolder) {
		folderName.value = row.name;
		detailTask.value = row.id;
		transferStore.getFilesInTask(row.id);
	} else {
		if (
			row.status !== TransferStatus.Completed &&
			(row.front == TransferFront.upload || row.front == TransferFront.cloud)
		) {
			return;
		}
		console.log('itemClick transferMap', row);
		const dataAPI = dataAPIs(row.driveType);
		const cur_file = dataAPI.formatTransferToFileItem(row);
		filesStore.openPreviewDialog(cur_file);
	}
};

const openFolder = async () => {
	// const folderItem: any = datas.value[datas.value.length - 1];
	// await transferStore.getFolderInTask(folderItem);
	folderName.value = null;
	detailTask.value = null;
	transferStore.filesInFolder = [];
	transferStore.filesInFolderMap = {};
};

onUnmounted(() => {
	transferStore.filesInFolder = [];
	transferStore.filesInFolderMap = {};
});

function mapItems(items: any[], mapStore: Record<string, any>) {
	return items.map((id) => ({
		id,
		...mapStore[id]
	}));
}

const datas = computed(() => {
	if (detailTask.value) {
		const filterCondition = (status: TransferStatus) =>
			transferStore.transferType === TransferType.DOWNLOADED ||
			transferStore.transferType === TransferType.UPLOADED
				? status === TransferStatus.Completed ||
				  status === TransferStatus.Canceled
				: status !== TransferStatus.Completed;

		return mapItems(
			transferStore.filesInFolder.filter((id) =>
				filterCondition(transferStore.filesInFolderMap[id].status)
			),
			transferStore.filesInFolderMap
		);
	}

	const { activeItem, transferType } = transferStore;

	const activeItemMap = {
		[TransferFront.download]:
			transferType === TransferType.DOWNLOADING
				? transferStore.downloading
				: transferStore.downloadComplete,
		[TransferFront.upload]:
			transferType === TransferType.UPLOADING
				? transferStore.uploading
				: transferStore.uploadComplete,
		[TransferFront.cloud]:
			transferType === TransferType.CLOUDING
				? transferStore.clouding
				: transferStore.cloudComplete,
		[TransferFront.copy]:
			transferType === TransferType.COPYING
				? transferStore.copying
				: transferStore.copyComplete
	};

	const items = activeItemMap[activeItem] || [];
	return mapItems(items, transferStore.transferMap);
});

watch(
	() => transferStore.transferType,
	() => {
		folderName.value = null;
		detailTask.value = null;
		transferStore.filesInFolder = [];
		transferStore.filesInFolderMap = {};
	}
);

const pagination = ref({
	// sortBy: 'desc',
	// descending: false,
	page: 1,
	rowsPerPage: 100
});

const pagesNumber = computed(() =>
	Math.ceil(datas.value.length / pagination.value.rowsPerPage)
);

const toastTransferItemMessage = (message: string) => {
	if (!message || message === '') {
		return;
	}
	notifyFailed(message);
};
</script>

<style lang="scss">
.custom-table {
	width: 100%;
	::-webkit-scrollbar {
		width: 0;
	}
}
.custom-table .q-table {
	// height: 100%;
	width: 100%;
}

.custom-table .q-table theader tr th:nth-child(1),
.custom-table .q-table tbody tr td:nth-child(1) {
	max-width: 100px;
}

.custom-table .q-table tbody tr td {
	height: 56px;
}

.transfer-page-root {
	width: 100%;
	height: 100%;
	background: $background-1;
	overflow: hidden;

	.folder-name {
		display: inline-block;
		width: calc(100% - 200px);
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.select-box {
		width: 24px;
		height: 24px;
		border-radius: 4px;
		box-sizing: border-box;
	}
	.un-selected {
		background-color: $background-1;
		border: 1px solid $separator;
	}

	.selected {
		background-color: $yellow;
	}

	.empty {
		width: 100%;
		height: calc(100% - 80px);

		img {
			width: 240px;
			height: 240px;
			margin-bottom: 20px;
		}

		span {
			color: $grey-14;
		}
	}

	.folder-header {
		height: 30px;
	}

	.table-content {
		width: 100%;
		height: calc(100% - var(--height, 0px));
	}

	.transfer-table {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
	}
}
</style>
