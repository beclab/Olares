<template>
	<div class="files-transfer-root">
		<Terminus-user-header
			v-if="selectIds === null"
			:title="t('transport_mgnt')"
		>
			<template v-slot:add>
				<q-btn
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_checklist"
					text-color="ink-2"
					@click="intoCheckedMode"
				>
				</q-btn>
			</template>
		</Terminus-user-header>
		<terminus-select-header
			v-else
			:selectIds="selectIds"
			@handle-close="handleClose"
			@handle-select-all="handleSelectAll"
			@handle-remove="handleRemove"
		/>
		<TerminusUserHeaderReminder v-if="showOnlyWifiTransfer">
			<template v-slot:errorcontent>
				<div>
					{{
						t('It is set to transmit only under WiFi, want to close it?').split(
							t('want to close it?')
						)[0]
					}}
					<span class="jump-subline-item" @click="updateWifiSetting">
						{{ t('want to close it?') }}
					</span>
					{{
						t('It is set to transmit only under WiFi, want to close it?').split(
							t('want to close it?')
						)[1] || ''
					}}
				</div>
			</template>
			<!-- It is set to transmit only under WiFi, want to modify it? -->
		</TerminusUserHeaderReminder>
		<div class="content">
			<div class="row items-center justify-between">
				<q-tabs
					v-model="tab"
					dense
					no-caps
					inline-label
					active-color="ink-1"
					class="text-ink-2 tabsss text-subtitle2"
					:class="{ isSelectedMode: selectIds }"
					indicator-color="yellow"
					narrow-indicator
					align="left"
					:breakpoint="0"
				>
					<q-tab
						:name="TransferFront.upload"
						icon="sym_r_upload"
						:label="t('transmission.upload.title')"
					/>
					<q-tab
						:name="TransferFront.download"
						icon="sym_r_download"
						:label="t('transmission.download.title')"
					/>
				</q-tabs>
				<!-- <q-btn
					dense
					flat
					icon="sym_r_error"
					class="q-mr-md"
					color="red-6"
					@click="clickHandler"
				>
				</q-btn> -->
			</div>
			<div class="tabs-separatar"></div>

			<div
				class="transfer-status row items-center justify-start"
				:class="{ isSelectedMode: selectIds }"
			>
				<div
					class="text-overline-m q-mr-sm"
					:class="
						activeStatus === TransferStatus.All
							? 'text-grey-10 bg-yellow-default'
							: 'text-ink-3 bg-background-3'
					"
					@click="handlestatus(TransferStatus.All)"
				>
					<span>{{ t('files.all') }}</span>
				</div>
				<div
					class="text-overline-m q-mr-sm"
					:class="
						activeStatus === TransferStatus.Completed
							? 'text-grey-10 bg-yellow-default'
							: 'text-ink-3 bg-background-3'
					"
					@click="handlestatus(TransferStatus.Completed)"
				>
					<span>{{ t('completed') }}</span>
					<span class="q-ml-xs">{{ completedTotal }}</span>
				</div>
				<div
					class="text-overline-m q-mr-sm"
					:class="
						activeStatus === TransferStatus.Running
							? 'text-grey-10 bg-yellow-default'
							: 'text-ink-3 bg-background-3'
					"
					@click="handlestatus(TransferStatus.Running)"
				>
					<sapn>{{ t('Ongoing') }}</sapn>
					<span class="q-ml-xs">{{ progressTotal }}</span>
				</div>
			</div>

			<q-tab-panels
				v-model="tab"
				animated
				swipeable
				class="transfer-page-tabs-panel q-pa-none q-mt-md"
				@update:model-value="updateTab"
			>
				<q-tab-panel
					:name="TransferFront.upload"
					class="transfer-page-tab-panel q-pa-none tables-padding"
				>
					<file-transfer-history
						ref="fileTransfer"
						:transferFront="TransferFront.upload"
						:activeStatus="activeStatus"
						:lockEvent="lockEvent"
						@show-select-mode="showSelectMode"
					/>
				</q-tab-panel>

				<q-tab-panel
					:name="TransferFront.download"
					class="transfer-page-tab-panel q-pa-none tables-padding"
				>
					<file-transfer-history
						ref="fileTransfer"
						:transferFront="TransferFront.download"
						:activeStatus="activeStatus"
						:lockEvent="lockEvent"
						@show-select-mode="showSelectMode"
					/>
				</q-tab-panel>
			</q-tab-panels>
		</div>

		<div
			v-if="hasProcessingItems"
			class="pause row items-center justify-center text-ink-2 text-body3 bg-background-1"
			@click="handlePause"
		>
			<q-icon v-if="!isPauseAll" name="sym_r_pause_circle" size="16px" />
			<span v-if="!isPauseAll" class="q-ml-xs">
				{{ t('transmission.pause_all') }}</span
			>

			<q-icon v-if="isPauseAll" name="sym_r_play_circle" size="16px" />
			<span v-if="isPauseAll" class="q-ml-xs">{{
				t('transmission.all_Start')
			}}</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import TerminusUserHeader from '../../../components/common/TerminusUserHeader.vue';
import TerminusSelectHeader from '../../../components/common/TerminusSelectHeader.vue';
import TerminusUserHeaderReminder from '../../../components/common/TerminusUserHeaderReminder.vue';

import FileTransferHistory from './FileTransferHistory.vue';
import { useI18n } from 'vue-i18n';
import { useTransfer2Store } from '../../../stores/transfer2';
import { useDeviceStore } from 'src/stores/device';
import {
	TransferFront,
	TransferStatus
} from '../../../utils/interface/transfer';
import { useTermipassStore } from 'src/stores/termipass';
import { UserStatusActive } from '../../../utils/checkTerminusState';
import { useUserStore } from 'src/stores/user';
import { useQuasar } from 'quasar';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';

const transferStore = useTransfer2Store();

const { t } = useI18n();

const tab = ref(TransferFront.upload);
const activeStatus = ref(TransferStatus.All);
const selectIds = ref(null);
const fileTransfer = ref();
const lockEvent = ref(false);
const termipassStore = useTermipassStore();
const userStore = useUserStore();
const deviceStore = useDeviceStore();
const $q = useQuasar();

const updateTab = () => {
	lockEvent.value = true;
	setTimeout(() => {
		lockEvent.value = false;
		handleClose();
	}, 500);
};

const showSelectMode = (value) => {
	selectIds.value = value;
};

const handleClose = () => {
	if (fileTransfer.value) {
		fileTransfer.value.handleClose();
	}
};

const handleSelectAll = () => {
	if (fileTransfer.value) {
		fileTransfer.value.handleSelectAll();
	}
};

const handleRemove = () => {
	if (fileTransfer.value) {
		fileTransfer.value.handleRemove();
	}
	setTimeout(() => {
		handleClose();
	}, 300);
};

const intoCheckedMode = () => {
	if (fileTransfer.value) {
		fileTransfer.value.intoCheckedMode();
	}
};

const handlestatus = (value: TransferStatus) => {
	activeStatus.value = value;
};

const uploadingNum = computed(() => {
	const uploadingTotal = transferStore.uploading.length;

	if (uploadingTotal > 99) {
		return '99+';
	} else {
		return uploadingTotal;
	}
});

const downloadingNum = computed(() => {
	const downloadTotal = transferStore.downloading.length;

	if (downloadTotal > 99) {
		return '99+';
	} else {
		return downloadTotal;
	}
});

const completedTotal = computed(() => {
	if (tab.value === TransferFront.upload) {
		return transferStore.uploadComplete.length;
	}
	if (tab.value == TransferFront.download) {
		return transferStore.downloadComplete.length;
	}
	return 0;
});

const progressTotal = computed(() => {
	if (tab.value === TransferFront.upload) {
		return uploadingNum.value;
	} else if (tab.value === TransferFront.download) {
		return downloadingNum.value;
	} else {
		return 0;
	}
});

const isPauseAll = computed(() => {
	if (tab.value == TransferFront.upload) {
		const isNotPause = transferStore.uploading.find(
			(item) => !transferStore.transferMap[item].isPaused
		);
		if (isNotPause) {
			return false;
		} else {
			return true;
		}
	}

	if (tab.value == TransferFront.download) {
		const isNotPause = transferStore.downloading.find(
			(item) => !transferStore.transferMap[item].isPaused
		);
		if (isNotPause) {
			return false;
		} else {
			return true;
		}
	}

	return false;
});

const handlePause = async () => {
	if (tab.value == TransferFront.upload) {
		if (isPauseAll.value) {
			transferStore.bulkResume(transferStore.uploading);
		} else {
			transferStore.bulkPause(transferStore.uploading);
		}
	} else if (tab.value == TransferFront.download) {
		if (isPauseAll.value) {
			transferStore.bulkResume(transferStore.downloading);
		} else {
			transferStore.bulkPause(transferStore.downloading);
		}
	}
};

const hasProcessingItems = computed(() => {
	if (tab.value == TransferFront.upload) {
		return transferStore.uploading.length > 0;
	}
	if (tab.value == TransferFront.download) {
		return transferStore.downloading.length > 0;
	}
	return 0;
});

const showOnlyWifiTransfer = computed(() => {
	if (termipassStore.totalStatus?.isError === UserStatusActive.error) {
		return false;
	}
	if (
		transferStore.uploading.length == 0 &&
		transferStore.downloading.length == 0
	) {
		return false;
	}

	return deviceStore.connectType == 'cellular' && userStore.transferOnlyWifi;
});

const updateWifiSetting = () => {
	$q.dialog({
		component: TerminusTipDialog,
		componentProps: {
			title: t('Only transfer files over wifi'),
			message: t(
				'Are you sure you want to turn off file transfer only under WiFi?'
			)
		}
	}).onOk(async () => {
		userStore.updateTransferOnlyWifiStatus(!userStore.transferOnlyWifi);
	});
};
</script>

<style scoped lang="scss">
.files-transfer-root {
	width: 100%;
	height: 100%;

	.isSelectedMode {
		opacity: 0.5;
		pointer-events: none;
	}

	.content {
		width: 100%;
		height: calc(100% - 66px);
		// padding-bottom: 20px;
		margin-top: 10px;
		// padding: 20px;

		.tabsss {
			width: calc(100%);
			height: 36px;
		}

		.tabs-separatar {
			height: 1px;
			width: 100%;
			border-bottom: 1px solid $separator;
		}

		.transfer-page-tabs-panel {
			width: 100%;
			height: calc(100% - 80px);
			::-webkit-scrollbar {
				width: 0 !important;
			}
		}

		.tables-padding {
			padding-left: 20px;
			padding-right: 20px;
		}

		.transfer-status {
			width: calc(100% - 40px);
			margin: 12px 20px;
			div {
				padding: 4px 12px;
				border-radius: 4px;
				cursor: pointer;
			}
		}
	}

	.func-menu {
		width: 40px;
		height: 40px;
	}

	.pause {
		position: fixed;
		bottom: calc(env(safe-area-inset-bottom) + 80px);
		right: 20px;
		padding: 0 12px;
		height: 32px;
		line-height: 32px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 8px;
	}
}
</style>
