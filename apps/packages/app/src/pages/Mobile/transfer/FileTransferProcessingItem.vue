<template>
	<q-item
		class="processing-history q-px-none row items-center justify-between"
		clickable
	>
		<q-item-section
			avatar
			style="max-width: 32px; max-height: 32px; object-fit: cover"
		>
			<terminus-file-icon
				:name="transferStore.transferMap[file].name"
				:type="transferStore.transferMap[file].type"
				:path="transferStore.transferMap[file].path"
				:driveType="transferStore.transferMap[file].driveType"
				:modified="0"
				:is-dir="transferStore.transferMap[file].isFolder"
			/>
		</q-item-section>

		<q-item-section>
			<q-item-label class="text-subtitle2 text-ink-1 name">{{
				transferStore.transferMap[file].name
			}}</q-item-label>
			<q-item-label class="text-body3 text-ink-3">
				<div>
					<div class="row items-center justify-between">
						<div class="text-body3 text-ink-3 item-info">
							{{ format.formatFileSize(transferStore.transferMap[file].size) }}
							/
							{{ formatLeftTimes(transferStore.transferMap[file].leftTime) }}
						</div>
						<div
							class="text-body3 info-error"
							:class="
								transferStore.transferMap[file].status == TransferStatus.Error
									? 'text-red-8'
									: 'text-light-blue-default'
							"
							@click="
								toastTransferItemMessage(
									transferStore.transferMap[file].status == TransferStatus.Error
										? transferStore.transferMap[file].message
										: ''
								)
							"
						>
							{{
								transferStore.transferMap[file].status == TransferStatus.Error
									? t('Failed')
									: transferStore.transferMap[file].status ==
											TransferStatus.Pending &&
									  !transferItemIsPaused(transferStore.transferMap[file])
									? t('pending')
									: transferItemIsPaused(transferStore.transferMap[file])
									? transferItemPausedMessage(transferStore.transferMap[file])
									: format.formatFileSize(
											transferStore.transferMap[file].speed
									  ) + '/s'
							}}
						</div>
					</div>
					<q-linear-progress
						v-if="
							transferStore.transferMap[file].status != TransferStatus.Error
						"
						stripe
						rounded
						size="5px"
						:value="transferStore.transferMap[file].progress"
						class="q-mt-xs"
						color="green"
					/>
				</div>
			</q-item-label>
		</q-item-section>

		<q-item-section side>
			<div class="row items-center justify-end" style="min-width: 64px">
				<div
					v-if="transferStore.transferMap[file].status == TransferStatus.Error"
					class="row items-center justify-center item-action"
					@click="pauseOrResume(file)"
				>
					<q-icon name="sym_r_refresh" size="20px" color="ink-2" />
				</div>
				<div
					v-else-if="
						(transferStore.transferMap[file].status ===
							TransferStatus.Running &&
							!transferStore.transferMap[file].networkOfflinePaused &&
							!transferStore.transferMap[file].onlyWifiPaused) ||
						(transferStore.transferMap[file].status ===
							TransferStatus.Pending &&
							transferStore.transferMap[file].isPaused)
					"
					class="row items-center justify-center item-action"
					@click="pauseOrResume(file)"
				>
					<q-icon
						:name="
							transferStore.transferMap[file].isPaused
								? 'sym_r_play_circle'
								: 'sym_r_pause_circle'
						"
						size="20px"
						color="ink-2"
					/>
				</div>

				<div
					class="row items-center justify-center item-action"
					@click="cancelTask(file)"
				>
					<q-icon name="sym_r_close" size="20px" color="ink-2" />
				</div>
			</div>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { formatLeftTimes, useTransfer2Store } from '../../../stores/transfer2';
import { transferItemIsPaused } from 'src/utils/interface/transfer';
import {
	TransferStatus,
	TransferItemInMemory
} from '../../../utils/interface/transfer';
import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import { format } from '../../../utils/format';
import { useI18n } from 'vue-i18n';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

defineProps({
	file: {
		type: Number,
		required: true
	}
});

const { t } = useI18n();

const transferStore = useTransfer2Store();

const pauseOrResume = async (id: number) => {
	if (!transferStore.transferMap[id]) {
		return;
	}

	if (transferStore.transferMap[id].status == TransferStatus.Error) {
		transferStore.recoverErrorTransfer(id);
		return;
	}
	if (transferStore.transferMap[id].isPaused) {
		transferStore.resume(transferStore.transferMap[id]);
	} else {
		transferStore.pause(transferStore.transferMap[id]);
	}
};

const cancelTask = async (id: number) => {
	if (!transferStore.transferMap[id]) {
		return;
	}
	transferStore.cancel(transferStore.transferMap[id]);
};

const toastTransferItemMessage = (message?: string) => {
	if (!message || message === '') {
		return;
	}
	notifyFailed(message);
};

const transferItemPausedMessage = (item: TransferItemInMemory) => {
	return item.isPaused
		? t('download.paused')
		: item.networkOfflinePaused
		? t('Network abnormal')
		: t('Non WiFi network');
};
</script>

<style scoped lang="scss">
.processing-history {
	height: 72px;
	border-bottom: 1px solid $separator;

	.name {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
	.item-action {
		width: 32px;
		height: 32px;
	}

	.item-info {
		width: 50%;
		text-align: left;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.info-error {
		width: 50%;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		text-align: right;
	}
}
</style>
