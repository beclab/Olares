<template>
	<div class="uploadItem row items-center justify-between q-py-sm">
		<!-- {{ Math.round(item.progress() * 100) }} {{ uploadState }} -->
		<img class="fileIcon" :src="filesIcon(item.name)" />
		<div class="content">
			<div class="file-name text-ink-1">
				{{ item.name }}
			</div>
			<div class="text-ink-2">
				{{ format.formatFileSize(item.size) }}
			</div>
		</div>

		<div
			class="row items-center justify-center"
			v-if="item.status === TransferStatus.Error"
		>
			<span class="text-red"> {{ t('error.index') }} </span>
		</div>

		<div
			class="row items-center justify-center"
			v-else-if="item.status === TransferStatus.Completed"
		>
			<q-icon
				class="forword"
				rounded
				name="sym_r_search"
				size="sm"
				@click="forWord(item)"
			></q-icon>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onUnmounted, defineProps } from 'vue';
import { fileList } from '../../utils/constants';
import { useI18n } from 'vue-i18n';
import { common } from '../../api';
// import { StatusEnum } from './../../stores/uploader';
import { useDataStore } from '../../stores/data';
import { format } from '../../utils/format';

import { useFilesStore, FilesIdType } from '../../stores/files';
import { DriveType } from '../../utils/interface/files';
import { TransferStatus } from '../../utils/interface/transfer';

const props = defineProps({
	item: Object as any,
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

// const emits = defineEmits(['onUploadCancel']);

const { t } = useI18n();
const store = useDataStore();
const filesStore = useFilesStore();

const showUpload = ref(true);

onUnmounted(() => {
	showUpload.value = false;
});

const forWord = async (item: any) => {
	console.log('forword item', item);

	const itemFile = await insertItemFileList(item);

	if (!itemFile) {
		return false;
	}

	if (store.preview.isShow) {
		return;
	}

	filesStore.openPreviewDialog(itemFile, props.origin_id);
};

const insertItemFileList = async (item) => {
	const driveType = common().formatUrltoDriveType(item.path) || DriveType.Drive;
	const splitUrl = item.path.split('?');

	const key = filesStore.registerUniqueKey(splitUrl[0], driveType, splitUrl[1]);

	const fileList = filesStore.cached[key];

	const curFile = fileList.find((file) => file.name === item.newFileName);

	if (curFile) {
		return curFile;
	} else {
		return undefined;
	}
};

// const onUploadCancel = (item) => {
// 	emits('onUploadCancel', item);
// };

const filesIcon = (name: string) => {
	const h = name?.substring(name?.lastIndexOf('.') + 1);
	let src = '/img/';
	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/';
	}

	const hasFile = fileList.find((item: any) => item === h);
	if (hasFile) {
		src = src + h + '.png';
	} else {
		src = src + 'blob.png';
	}
	return src;
};
</script>

<style scoped lang="scss">
.uploadItem {
	padding-left: 20px;
	padding-right: 20px;

	.fileIcon {
		width: 24px;
	}

	.content {
		flex: 1;
		padding: 0 12px;
		overflow: hidden;

		div {
			width: 200px;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}

	.forword {
		cursor: pointer;
	}

	&:hover {
		background-color: $background-hover;
	}
}
</style>
