<template>
	<div class="uploadModal bg-background-2">
		<panel-header
			:processingCount="processingCount"
			:totalCount="uploadData.length"
			:showUpload="showUpload"
			@closePanel="closeUploadModal"
			@togglePanel="toggle"
		/>

		<q-scroll-area
			:thumb-style="scrollBarStyle.thumbStyle"
			:class="
				showUpload ? 'uploadContent heightFull' : 'uploadContent heightZero'
			"
		>
			<panel-item
				v-for="(file, index) in uploadData"
				:key="index"
				:file="file"
				class="q-mt-sm"
			/>
		</q-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { ref, onUnmounted, computed } from 'vue';
import { useTransfer2Store } from '../../../stores/transfer2';
import { scrollBarStyle } from '../../../utils/contact';
import { useI18n } from 'vue-i18n';

import PanelHeader from './PanelHeader.vue';
import PanelItem from './PanelItem.vue';

import {
	TransferFront,
	TransferItemInMemory,
	TransferStatus
} from '../../../utils/interface/transfer';

const emits = defineEmits(['onCloseUploadDialog']);

const transfer2Store = useTransfer2Store();

const { t } = useI18n();
const showUpload = ref(true);

const frontFilter = ref([
	TransferFront.upload,
	TransferFront.copy,
	TransferFront.move
]);

onUnmounted(() => {
	showUpload.value = false;
});

const processingCount = computed(() => {
	const res = transfer2Store.filesInDialog
		.map((id) => ({ id, ...transfer2Store.filesInDialogMap[id] }))
		.filter(
			(item) =>
				frontFilter.value.includes(item.front) &&
				(item.status === TransferStatus.Running ||
					item.status === TransferStatus.Pending)
		);

	return res.length;
});

const uploadData = computed(() => {
	const res = transfer2Store.filesInDialog
		.map((id) => ({ id, ...transfer2Store.filesInDialogMap[id] }))
		.filter((item) => frontFilter.value.includes(item.front));

	return [
		...res.filter((item) => item.status === TransferStatus.Running),
		...res.filter((item) => item.status === TransferStatus.Pending),
		...res.filter(
			(item) =>
				item.status !== TransferStatus.Running &&
				item.status !== TransferStatus.Pending
		)
	];
});

const uploadCompleteData = computed(() => {
	const res = transfer2Store.filesInDialog
		.map((id) => ({ id, ...transfer2Store.filesInDialogMap[id] }))
		.filter(
			(item) =>
				frontFilter.value.includes(item.front) &&
				item.status === TransferStatus.Completed
		);
	return res;
});

const toggle = () => {
	showUpload.value = !showUpload.value;
};

const closeUploadModal = () => {
	setTimeout(() => {
		emits('onCloseUploadDialog');

		let res = transfer2Store.filesInDialog.filter((id) => {
			return (
				transfer2Store.filesInDialogMap[id].status !==
					TransferStatus.Completed &&
				transfer2Store.filesInDialogMap[id].status !== TransferStatus.Canceled
			);
		});
		transfer2Store.filesInDialog = res;
		transfer2Store.filesInDialogMap = pickKeys(
			transfer2Store.filesInDialogMap,
			res
		);
	}, 300);
};

function pickKeys<
	T extends Record<number, TransferItemInMemory>,
	K extends keyof T
>(record: T, keysToPick: K[]): Pick<T, K> {
	return keysToPick.reduce((result, key) => {
		if (key in record) {
			result[key] = record[key];
		}
		return result;
	}, {} as Pick<T, K>);
}
</script>

<style scoped lang="scss">
.uploadModal {
	width: 400px;
	position: fixed;
	right: 20px;
	bottom: 20px;
	border-radius: 12px;
	overflow: hidden;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
	.uploadHeader {
		width: 100%;
		height: 48px;
		padding: 0 20px;

		div {
			color: $ink-1;
			font-weight: 700;

			.uploadStatus {
				width: 20px;
				margin-right: 10px;
			}
		}
	}

	.uploadContent {
		transition: height 0.3s;
		box-sizing: border-box;
	}

	.heightFull {
		height: 180px;
		border-top: 1px solid $separator;
	}

	.heightZero {
		height: 0;
	}
}
</style>
