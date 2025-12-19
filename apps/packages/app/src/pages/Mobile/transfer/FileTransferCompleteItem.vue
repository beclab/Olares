<template>
	<q-item
		class="history q-px-none row items-center justify-between"
		clickable
		@click="itemClick(file)"
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
			<q-item-label class="text-subtitle2 text-ink-1 content">{{
				transferStore.transferMap[file].name
			}}</q-item-label>
			<q-item-label class="text-body3 text-ink-3 content">
				<span>{{
					format.formatFileSize(transferStore.transferMap[file].size)
				}}</span>
				<span
					class="q-ml-sm"
					v-if="transferStore.transferMap[file].front === TransferFront.upload"
					>{{
						t('Upload to {address}', {
							address: formatFilePath(transferStore.transferMap[file])
						})
					}}</span
				>
			</q-item-label>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { useTransfer2Store } from '../../../stores/transfer2';

import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import { useI18n } from 'vue-i18n';
// import { useQuasar } from 'quasar';
import { dataAPIs } from '../../../api';
import { useDataStore } from '../../../stores/data';
import { useFilesStore } from '../../../stores/files';
import { format } from '../../../utils/format';
import { TransferFront } from '../../../utils/interface/transfer';

defineProps({
	file: {
		type: Number,
		required: true
	}
});

// const $q = useQuasar();

const transferStore = useTransfer2Store();

const filesStore = useFilesStore();

const { t } = useI18n();

const store = useDataStore();

const itemClick = async (id: number) => {
	if (!transferStore.transferMap[id]) {
		return false;
	}
	const dataAPI = dataAPIs(transferStore.transferMap[id].driveType);

	console.log('itemClick transferMap', transferStore.transferMap[id]);
	const cur_file = dataAPI.formatTransferToFileItem(
		transferStore.transferMap[id]
	);
	console.log('itemClick cur_file', cur_file);

	if (store.preview.isShow) {
		return;
	}
	filesStore.openPreviewDialog(cur_file);
};

const formatFilePath = (file) => {
	const dataAPI = dataAPIs(file.driveType);
	return dataAPI.formatTransferPath(file);
};

// const emits = defineEmits(['itemClick']);
</script>

<style scoped lang="scss">
.history {
	width: 100%;
	height: 72px;
	border-bottom: 1px solid $separator;
}
.content {
	text-overflow: ellipsis;
	white-space: nowrap;
	overflow: hidden;
}
</style>
