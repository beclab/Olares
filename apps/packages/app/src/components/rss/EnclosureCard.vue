<template>
	<div class="enclosure-root q-mt-lg">
		<div v-if="showReading" class="column full-width">
			<slot name="complete" :path="filePath" />
			<div
				class="text-orange-default text-underline text-right text-body3 q-mt-sm cursor-pointer"
				@click="userShowCard = true"
			>
				{{ t('base.switch_to_the_card_view') }}
			</div>
		</div>
		<div v-else class="enclosure-card q-px-md row justify-start items-center">
			<q-img class="enclosure-type-img" fit="cover" :src="entryImage">
				<template v-slot:loading>
					<q-skeleton width="140px" height="88px" />
				</template>
				<template v-slot:error>
					<q-img
						class="enclosure-type-img"
						fit="cover"
						:src="getRequireImage('entry_default_img.svg')"
					/>
				</template>
			</q-img>
			<div
				class="enclosure-card-right row justify-between items-center q-ml-sm"
			>
				<div class="enclosure-card-text column justify-start items-center">
					<div class="row full-width justify-between">
						<div
							class="enclosure-title text-subtitle1 text-ink-1"
							:style="{
								maxWidth:
									enclosureStatus === TransferStatus.Running
										? 'calc(100% - 80px)'
										: '100%'
							}"
							:class="
								enclosureStatus === DOWNLOAD_RECORD_STATUS.LOSS
									? 'text-ink-3 text-line-through'
									: 'text-ink-1'
							"
						>
							{{ enclosureName }}
						</div>
						<div
							v-if="enclosureStatus === TransferStatus.Running"
							class="row text-right items-center text-info card-right"
						>
							<q-icon size="20px" name="sym_r_download" />
							<div class="q-ml-xs text-body3">
								{{ Math.round(transferItem.progress * 100) }}%
							</div>
						</div>
					</div>

					<q-linear-progress
						v-if="enclosureStatus === TransferStatus.Running"
						class="q-mt-xs"
						rounded
						size="4px"
						:value="
							transferItem ? Math.round(transferItem.progress * 100) / 100 : 0
						"
						color="positive"
						track-color="separator"
					/>

					<div
						class="row justify-start full-width"
						:class="
							enclosureStatus === TransferStatus.Error
								? 'text-negative'
								: 'text-ink-3'
						"
					>
						<q-icon
							v-if="enclosureStatus === TransferStatus.Error"
							size="12px"
							class="q-mr-xs"
							name="sym_r_error"
						/>

						<div v-if="enclosureDesc" class="text-overline enclosure-title">
							{{ enclosureDesc }}
						</div>
					</div>
				</div>

				<q-icon
					v-if="operateIcon"
					class="text-ink-1 cursor-pointer"
					size="20px"
					:name="operateIcon"
					@click="enclosureOperate"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { TransferStatus } from '../../utils/interface/transfer';
import { DOWNLOAD_RECORD_STATUS } from '../../utils/rss-types';
import { FILE_TYPE, Enclosure } from '../../utils/rss-types';
import { useTransferStore } from '../../stores/rss-transfer';
import { useTransfer2Store } from '../../stores/transfer2';
import { getRequireImage } from '../../utils/rss-utils';
import TransferClient from '../../services/transfer';
import { PropType, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	enclosure: {
		type: Object as PropType<Enclosure>,
		require: true
	}
});

const { t } = useI18n();
const userShowCard = ref(false);
const showReading = computed(() => {
	return (
		enclosureStatus.value === TransferStatus.Completed &&
		filePath.value &&
		!userShowCard.value
	);
});
const transferStore = useTransferStore();
const transfer2Store = useTransfer2Store();

const transferItem = computed(() => {
	console.log(transferStore.enclosureMaps);
	const identify = TransferClient.client.clouder?.taskIdentify(
		String(transferStore.enclosureMaps[props.enclosure.id])
	);
	console.log(identify);
	let transferId = transfer2Store.filesCloudTransferMap[identify] || -1;
	if (transferId > 0) {
		const record = transfer2Store.transferMap[transferId];
		console.log(record);
		if (record) {
			return record;
		}
	}

	return null;
});

const enclosureName = computed(() => {
	let name = props.enclosure.name;
	if (!name && transferItem.value) {
		name = transferItem.value.name;
	}

	if (!name) {
		name = t('download.fetching_data_please_wait');
	}
	return name;
});

const enclosureStatus = computed(() => {
	if (transferItem.value) {
		return transferItem.value.status;
	}
	return DOWNLOAD_RECORD_STATUS.LOSS;
});

const filePath = computed(() => {
	if (transferItem.value) {
		return transferItem.value.path;
	}
	return '';
});

const enclosureDesc = computed(() => {
	if (transferItem.value && transferItem.value.isPaused) {
		return (
			t('download.paused') +
			': ' +
			Math.round(transferItem.value.progress * 100) +
			'%'
		);
	}
	switch (enclosureStatus.value) {
		case TransferStatus.Pending:
			return t('download.starting_download_task');
		case TransferStatus.Canceled:
			return t('download.cancelled');
		case TransferStatus.Error:
			return t('base.download_failed');
		case TransferStatus.Completed:
		case TransferStatus.Removed:
			return transferItem.value
				? transferItem.value.path
				: props.enclosure.local_file_path
				? props.enclosure.local_file_path.replace(
						'/' + props.enclosure.name,
						''
				  )
				: '';
		default:
			return '';
	}
});

const operateIcon = computed(() => {
	if (transferItem.value && transferItem.value.isPaused) {
		return 'sym_r_download';
	}
	if (transferItem.value) {
		switch (enclosureStatus.value) {
			case TransferStatus.Pending:
				return 'sym_r_cancel';
			case TransferStatus.Running:
				return 'sym_r_pause_circle';
			case TransferStatus.Removed:
				return 'sym_r_download';
			case TransferStatus.Error:
				return 'sym_r_autorenew';
			case TransferStatus.Completed:
				return 'sym_r_play_circle';
			default:
				return '';
		}
	} else {
		return '';
	}
});

const enclosureOperate = () => {
	if (transferItem.value && transferItem.value.isPaused) {
		transfer2Store.resume(transferItem.value);
		return;
	}
	if (transferItem.value) {
		switch (enclosureStatus.value) {
			case TransferStatus.Pending:
				TransferClient.client.clouder?.cancelTask(transferItem.value);
				transfer2Store.cancel(transferItem.value);
				break;
			case TransferStatus.Running:
				transfer2Store.pause(transferItem.value);
				break;
			case TransferStatus.Removed:
				TransferClient.client.clouder?.removeTask(transferItem.value.id, false);
				transfer2Store.remove(transferItem.value.id);
				break;
			case TransferStatus.Error:
				transfer2Store.resume(transferItem.value);
				break;
			case TransferStatus.Completed:
				userShowCard.value = false;
				break;
		}
	}
};

const entryImage = computed(() => {
	if (!props.enclosure || !props.enclosure?.mime_type) {
		return getRequireImage('entry_default_img.svg');
	}
	switch (props.enclosure?.mime_type) {
		case FILE_TYPE.VIDEO:
			return getRequireImage('filetype/video.svg');
		case FILE_TYPE.AUDIO:
			return getRequireImage('filetype/radio.svg');
		case FILE_TYPE.PDF:
			return getRequireImage('filetype/pdf.svg');
		case FILE_TYPE.EBOOK:
			return getRequireImage('filetype/ebook.svg');
		case FILE_TYPE.GENERAL:
			return getRequireImage('filetype/general.svg');
		default:
			return getRequireImage('entry_default_img.svg');
	}
});
</script>

<style scoped lang="scss">
.enclosure-root {
	width: 100%;
	height: auto;
	padding-left: var(--padding);
	padding-right: var(--padding);

	.enclosure-card {
		width: 100%;
		height: 64px;
		border-radius: 12px;
		border: 1px solid $separator;

		.enclosure-type-img {
			width: 32px;
			height: 32px;
		}

		.enclosure-card-right {
			width: calc(100% - 40px);

			.enclosure-card-text {
				width: calc(100% - 32px);
				max-width: calc(100% - 32px);

				.enclosure-title {
					overflow: hidden;
					display: -webkit-box;
					-webkit-line-clamp: 1;
					-webkit-box-orient: vertical;
					text-overflow: ellipsis;
				}
			}
		}
	}
}
</style>
