<template>
	<div class="upload-history-root">
		<terminus-select-all
			ref="terminusSelect"
			:items="data"
			:lockEvent="lockEvent"
			:hasDate="true"
			@show-select-mode="showSelectMode"
		>
			<template v-slot="{ file }">
				<file-transfer-date-item
					v-if="checkTransferStatus(file) === TransferMobileItem.DATE"
					:date="file.date"
				/>

				<file-transfer-processing-item
					v-if="checkTransferStatus(file) === TransferMobileItem.PROCESSING"
					:file="(file as number)"
				/>

				<file-transfer-complete-item
					v-if="checkTransferStatus(file) === TransferMobileItem.COMPLETED"
					:file="(file as number)"
				/>

				<!-- <file-transfer-error-item
					v-if="checkTransferStatus(file) === TransferMobileItem.ERROR"
					:file="(file as number)"
				/> -->
			</template>
		</terminus-select-all>
	</div>
</template>

<script setup lang="ts">
import { computed, PropType, ref, watch } from 'vue';
import {
	useTransfer2Store,
	TransferMobileItem
} from '../../../stores/transfer2';

import FileTransferDateItem from './FileTransferDateItem.vue';
import FileTransferProcessingItem from './FileTransferProcessingItem.vue';
import FileTransferCompleteItem from './FileTransferCompleteItem.vue';
import TerminusSelectAll from './../../../components/common/TerminusSelectAll.vue';
import {
	TransferFront,
	TransferStatus
} from '../../../utils/interface/transfer';

const props = defineProps({
	activeStatus: {
		type: Object as PropType<TransferStatus>,
		required: true
	},
	transferFront: {
		type: Object as PropType<TransferFront>,
		required: true
	},
	lockEvent: {
		type: Boolean,
		required: true
	}
});

interface DateMarker {
	date: string;
	ids: number[];
}

const transferStore = useTransfer2Store();

const ids = ref();
const terminusSelect = ref();
const selectedIds = ref();

const emits = defineEmits(['showSelectMode']);

const showSelectMode = (value) => {
	selectedIds.value = value;
	emits('showSelectMode', value);
};

const handleClose = () => {
	if (terminusSelect.value) {
		terminusSelect.value.handleClose();
	}
};

const handleSelectAll = () => {
	if (terminusSelect.value) {
		terminusSelect.value.toggleSelectAll();
	}
};

const intoCheckedMode = () => {
	if (terminusSelect.value) {
		terminusSelect.value.intoCheckedMode();
	}
};

const handleRemove = () => {
	const cancelIds = selectedIds.value.filter(
		(e) => checkTransferStatus(e) === TransferMobileItem.PROCESSING
	);

	const removeIds = selectedIds.value.filter(
		(e) => checkTransferStatus(e) !== TransferMobileItem.PROCESSING
	);

	transferStore.bulkCancel(cancelIds);
	transferStore.bulkRemove(removeIds);
	if (terminusSelect.value) {
		terminusSelect.value.handleRemove();
	}
};

defineExpose({
	handleClose,
	handleSelectAll,
	handleRemove,
	intoCheckedMode
});

// const clearUploadData = () => {
// 	transferStore.bulkRemove(transferStore.uploadComplete);
// };

function isNumber(value) {
	return typeof value === 'number' && !isNaN(value);
}

const checkTransferStatus = (file: number | DateMarker): TransferMobileItem => {
	if (!isNumber(file)) {
		return TransferMobileItem.DATE;
	}

	if (
		transferStore.transferMap[file as number] &&
		transferStore.transferMap[file as number].status !=
			TransferStatus.Completed &&
		transferStore.transferMap[file as number].status != TransferStatus.Canceled
	) {
		return TransferMobileItem.PROCESSING;
	}

	if (
		transferStore.transferMap[file as number] &&
		[TransferStatus.Completed, TransferStatus.Canceled].includes(
			transferStore.transferMap[file as number].status
		)
	) {
		return TransferMobileItem.COMPLETED;
	}

	if (
		transferStore.transferMap[file as number] &&
		transferStore.transferMap[file as number].status === TransferStatus.Error
	) {
		return TransferMobileItem.ERROR;
	}

	return TransferMobileItem.EMPTY;
};

const sortTransfer = (items: number[]): number[] => {
	return items.sort((a, b) => {
		if (
			transferStore.transferMap[a].status === TransferStatus.Running &&
			transferStore.transferMap[b].status !== TransferStatus.Running
		) {
			return 1;
		}
		if (
			transferStore.transferMap[a].status !== TransferStatus.Running &&
			transferStore.transferMap[b].status === TransferStatus.Running
		) {
			return -1;
		}

		return (
			(transferStore.transferMap[a].id as number) -
			(transferStore.transferMap[b].id as number)
		);
	});
};

const groupByDate = (items: number[]): (number | DateMarker)[] => {
	const result: (number | DateMarker)[] = [];
	let currentGroup: DateMarker | null = null;

	sortTransfer(items);

	for (const item of items) {
		const date = new Date(transferStore.transferMap[item].startTime as number)
			.toISOString()
			.split('T')[0];

		if (currentGroup === null || currentGroup.date !== date) {
			if (currentGroup) {
				result.push(currentGroup);
			}
			currentGroup = {
				date,
				ids: [item]
			};
		} else {
			currentGroup.ids.push(item);
		}

		result.push(item);
	}

	if (currentGroup) {
		result.push(currentGroup);
	}

	result.reverse();

	return result;
};

const data = computed(() => {
	let transferData: number[] = [];
	if (props.activeStatus === TransferStatus.All) {
		transferData =
			props.transferFront === TransferFront.upload
				? transferStore.upload
				: transferStore.download;
	} else if (props.activeStatus === TransferStatus.Completed) {
		transferData =
			props.transferFront === TransferFront.upload
				? transferStore.uploadComplete
				: transferStore.downloadComplete;
	} else if (props.activeStatus === TransferStatus.Running) {
		transferData =
			props.transferFront === TransferFront.upload
				? transferStore.uploading
				: transferStore.downloading;
	}

	return groupByDate(transferData);
});

watch(
	() => data.value,
	(newVal) => {
		ids.value = newVal.filter((item) => !item.date);
		// console.log('watchwatch', ids.value);
	},
	{
		immediate: true
	}
);
</script>

<style scoped lang="scss">
.upload-history-root {
	width: 100%;
	height: 100%;
}
</style>
