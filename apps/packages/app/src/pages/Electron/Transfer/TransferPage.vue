<template>
	<div class="container">
		<header>
			<transfer-header>
				<template v-slot:transfer-add>
					<transfer-add-header
						@add-cloud-task="addCloudTask"
						@add-upload-task="addUploadTask"
					/>
				</template>
			</transfer-header>
		</header>

		<main>
			<transfer-table />
		</main>
		<files-uploader :autoBindResumable="false" @files-update="filesUpdate" />
	</div>
</template>

<script setup lang="ts">
import TransferTable from './TransferTable.vue';
import TransferHeader from './TransferHeader.vue';
import TransferAddHeader from './TransferAddHeader.vue';
import FilesUploader from '../../Files/common-files/FilesUploader.vue';
import TransferCloudAddDialog from './TransferCloudAddDialog.vue';
import TransferUploadAddDialog from './TransferUploadAddDialog.vue';
import { useQuasar } from 'quasar';
import { FilePath, useFilesStore } from '../../../stores/files';
import { ref } from 'vue';

const filesUpdate = (event: any) => {
	filesRef.value = event.target.files;
	targetRef.value = event;
	showUploadDialog();
};

const $q = useQuasar();

const fileStore = useFilesStore();

const filesRef = ref();
const targetRef = ref();

const addCloudTask = () => {
	$q.dialog({
		component: TransferCloudAddDialog
	});
};

const addUploadTask = async () => {
	const event = await fileStore.selectSystemFile();
	if (!event) {
		return;
	}
	filesRef.value = event.files;
	targetRef.value = event.target;
	showUploadDialog();
};

const showUploadDialog = () => {
	if (!filesRef.value) {
		return;
	}
	$q.dialog({
		component: TransferUploadAddDialog,
		componentProps: {
			files: filesRef.value
		}
	}).onOk((fileSavePath: FilePath) => {
		fileStore.uploadSelectFile(targetRef.value, fileSavePath);
	});
};
</script>

<style lang="scss" scoped>
.container {
	width: 100%;
	height: 100%;
	position: absolute;
	left: 0;
	top: 0;
	display: flex;
	flex-direction: column;
	background: $background-1;
	overflow: hidden;

	header,
	footer {
		flex: 0 0 auto;
	}

	main {
		flex: 0.99 0.99 auto;
		overflow-y: auto;
		height: calc(100% - 108px);
		// border-left: solid 1px $separator;
	}
}
</style>
