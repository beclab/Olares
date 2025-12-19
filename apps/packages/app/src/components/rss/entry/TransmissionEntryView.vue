<template>
	<base-entry-view
		:name="record?.name ? record.name : t('download.fetching_data_please_wait')"
		desc=""
		:time="record?.startTime ? new Date(record.startTime).getTime() / 1000 : 0"
		image-url=""
		:file-type="upload ? getFileTypeByName(record?.name) : record.type"
		:show-read-status="false"
		:loss="false"
		:status="ENTRY_STATUS.Completed"
		:selected="selected"
		@on-hover="onHover"
		:skeleton="skeleton"
		:time-prefix="t('base.create_at')"
		:percentage="
			record?.status === TransferStatus.Running && record.progress
				? Math.round(record.progress * 100)
				: '0'
		"
	>
		<template v-slot:bottom>
			<div class="layout-pdf-other">
				<div
					v-if="record?.status === TransferStatus.Error"
					class="row text-negative"
				>
					<q-icon size="16px" name="sym_r_error" />

					<div class="q-ml-sm text-body3">
						{{ t('base.download_failed') }}
					</div>
				</div>

				<div
					v-else-if="record?.status === TransferStatus.Completed"
					class="row justify-start items-center text-ink-2 content-layout"
				>
					<q-icon size="16px" name="sym_r_folder" />

					<span class="complete-text q-ml-sm text-body3 text-ink-3">
						{{ decodeURIComponent(record?.to) }}
					</span>
				</div>

				<div
					v-else-if="record?.status === TransferStatus.Canceled"
					class="row text-ink-3"
				>
					<q-icon size="16px" name="sym_r_block" />

					<div class="q-ml-sm text-body3">
						{{ t('download.cancelled') }}
					</div>
				</div>
				<div v-else-if="record?.isPaused" class="row text-ink-3">
					<q-icon size="16px" name="sym_r_download" />

					<div class="q-ml-sm text-body3">
						{{
							record.progress && record.size
								? t('download.paused') +
								  ':' +
								  Math.round(record.progress * 100) +
								  '%'
								: t('download.paused') +
								  ':' +
								  getValueByUnit(
										record.bytes,
										getSuitableUnit(record.bytes, 'memory')
								  ) +
								  ' ' +
								  getSuitableUnit(record.bytes, 'memory')
						}}
					</div>
				</div>

				<div
					v-else-if="
						record?.status === TransferStatus.Prepare ||
						record?.status === TransferStatus.Pending ||
						record?.status === TransferStatus.Canceling ||
						record?.status === TransferStatus.Checking ||
						record?.status === TransferStatus.Resuming ||
						record?.status === TransferStatus.Removing ||
						record?.status === TransferStatus.Running
					"
					class="row"
				>
					<bt-loading :loading="true" size="16px" />

					<div class="q-ml-sm text-body3 text-ink-3">
						{{
							record.progress && record.size
								? (upload ? t('base.uploading') : t('base.downloading')) +
								  ':' +
								  Math.round(record.progress * 100) +
								  '%'
								: (upload ? t('base.uploading') : t('base.downloading')) +
								  ':' +
								  getValueByUnit(
										record.bytes,
										getSuitableUnit(record.bytes, 'memory')
								  ) +
								  ' ' +
								  getSuitableUnit(record.bytes, 'memory')
						}}
					</div>
				</div>
			</div>
		</template>
		<template v-slot:float>
			<q-btn
				v-if="record?.status !== TransferStatus.Completed"
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				icon="sym_r_more_horiz"
			>
				<bt-tooltip :label="t('base.more')" />
				<bt-popup style="width: 176px">
					<bt-popup-item
						v-if="record?.status !== TransferStatus.Error && !record.isPaused"
						v-close-popup
						icon="sym_r_pause"
						:title="t('download.pause')"
						@on-item-click="pauseOrResume"
					/>
					<bt-popup-item
						v-if="record?.status !== TransferStatus.Error && record.isPaused"
						v-close-popup
						icon="sym_r_download"
						:title="t('download.resume')"
						@on-item-click="pauseOrResume"
					/>
					<bt-popup-item
						v-if="record?.status === TransferStatus.Canceled"
						v-close-popup
						icon="sym_r_autorenew"
						:title="t('download.retry')"
						@on-item-click="cancelRetry"
					/>
					<bt-popup-item
						v-if="record?.status === TransferStatus.Error"
						v-close-popup
						icon="sym_r_autorenew"
						:title="t('download.retry')"
						@on-item-click="pauseOrResume"
					/>
					<bt-popup-item
						v-else
						v-close-popup
						icon="sym_r_block"
						:title="t('base.cancel')"
						@on-item-click="onCancel"
					/>
				</bt-popup>
			</q-btn>

			<q-btn
				v-if="props.record.status === TransferStatus.Completed"
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				icon="sym_r_folder_open"
				@click="openDir"
			>
				<bt-tooltip :label="t('base.open')" />
			</q-btn>

			<q-btn
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				@click="onRemove"
				icon="sym_r_do_not_disturb_on"
			>
				<bt-tooltip :label="t('base.remove')" />
			</q-btn>
		</template>
	</base-entry-view>
</template>

<script setup lang="ts">
import BtPopup from '../../base/BtPopup.vue';
import BaseEntryView from './BaseEntryView.vue';
import BtTooltip from '../../base/BtTooltip.vue';
import BtLoading from '../../base/BtLoading.vue';
import BtPopupItem from '../../base/BtPopupItem.vue';
import BaseCheckBoxDialog from '../../base/BaseCheckBoxDialog.vue';
import TransferClient from '../../../services/transfer';
import { useConfigStore } from '../../../stores/rss-config';
import { ENTRY_STATUS, getFileTypeByName } from '../../../utils/rss-types';
import { useTransfer2Store } from '../../../stores/transfer2';
import { uploadDeleteEntry } from '../../../api/wise/upload';
import { getSuitableUnit, getValueByUnit } from 'src/utils/monitoring';
import {
	TransferStatus,
	TransferItemInMemory
} from '../../../utils/interface/transfer';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const transfer2Store = useTransfer2Store();
const $q = useQuasar();
const configStore = useConfigStore();

const props = defineProps({
	upload: {
		type: Boolean,
		default: true
	},
	record: {
		type: Object as PropType<TransferItemInMemory>,
		require: true
	},
	selected: {
		type: Boolean,
		default: false
	},
	skeleton: {
		type: Boolean,
		default: false
	}
});

const emit = defineEmits(['onSelectedChange']);

const onHover = (hover: boolean) => {
	if (hover) {
		emit('onSelectedChange', hover);
	}
};

const openDir = () => {
	if (!props.record) {
		return;
	}
	let suffix;
	if (props.record.status === TransferStatus.Completed) {
		suffix = decodeURIComponent(
			props.record.path.replace(props.record.name, '')
		);
	}
	// else {
	// 	suffix = decodeURIComponent(props.record.parentPath);
	// }
	const configStore = useConfigStore();
	let url = configStore.getModuleSever('files', 'https:', suffix);
	url = url + encodeURIComponent(props.record.name);
	window.open(url);
};

const pauseOrResume = () => {
	if (!props.record) {
		return;
	}

	if (props.record.status == TransferStatus.Error) {
		transfer2Store.transferMap[props.record.id].status = TransferStatus.Pending;
		transfer2Store.transferMap[props.record.id].isPaused = false;
		transfer2Store.resume(props.record);
		return;
	}
	if (props.record.isPaused) {
		transfer2Store.resume(props.record);
	} else {
		transfer2Store.pause(props.record);
	}
};

const cancelRetry = () => {
	if (!props.record) {
		return;
	}

	transfer2Store.resume(props.record);
};

const onCancel = async () => {
	if (!props.record) {
		return;
	}
	if (!props.upload) {
		TransferClient.client.clouder?.cancelTask(props.record);
	}
	// transfer2Store.cancel(props.record);
};

const onRemove = () => {
	if (!props.record) {
		return;
	}
	const remove = configStore.taskRemoveWithFile;
	if (props.upload) {
		$q.dialog({
			component: BaseCheckBoxDialog,
			componentProps: {
				label: t('dialog.remove_task'),
				content: t('dialog.remove_task_desc'),
				modelValue: remove,
				showCheckbox: true,
				boxLabel: t('dialog.delete_the_files')
			}
		})
			.onOk((selected) => {
				console.log('remove task ok');
				if (props.record) {
					configStore.setTaskRemoveWithFile(selected);
					console.log(props.record);
					uploadDeleteEntry(props.record.wiseRecordId, selected);
					transfer2Store.remove(props.record.id);
				} else {
					console.log(props.record);
				}
			})
			.onCancel(() => {
				console.log('remove task cancel');
			});
	} else {
		const remove = configStore.taskRemoveWithFile;
		$q.dialog({
			component: BaseCheckBoxDialog,
			componentProps: {
				label: t('dialog.remove_task'),
				content: t('dialog.remove_task_desc'),
				modelValue: remove,
				showCheckbox: true,
				boxLabel: t('dialog.delete_the_files')
			}
		})
			.onOk((selected) => {
				console.log('remove task ok');
				if (props.record) {
					configStore.setTaskRemoveWithFile(selected);
					TransferClient.client.clouder?.removeTask(props.record, selected);
					transfer2Store.remove(props.record.id);
				} else {
					console.log(props.record);
				}
			})
			.onCancel(() => {
				console.log('remove task cancel');
			});
	}
};
</script>

<style lang="scss" scoped>
.layout-pdf-other {
	width: 100%;

	.content-layout {
		width: 100%;
		display: flex;
		white-space: nowrap;

		.complete-text {
			width: calc(100% - 40px);
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 1;
			-webkit-box-orient: vertical;
		}
	}
}
</style>
