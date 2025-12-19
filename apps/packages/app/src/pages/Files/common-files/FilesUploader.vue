<template>
	<input
		id="uploader-input"
		className="upload-input"
		type="file"
		ref="uploadInput"
		multiple
		:accept="accept"
	/>
	<panel-index
		v-if="transfer2Store.isUploadProgressDialogShow && openPlatform"
		@onCloseUploadDialog="onCloseUploadDialog"
	/>
</template>

<script lang="ts" setup>
import { useTransfer2Store } from '../../../stores/transfer2';
import PanelIndex from './../../../components/files/panel/PanelIndex.vue';

import { FilesIdType } from '../../../stores/files';
import Resumable from './../../../utils/resumejs';

import { ref, onMounted } from 'vue';

const props = defineProps({
	autoBindResumable: {
		type: Boolean,
		default: true,
		required: false
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	accept: {
		type: String,
		default: ''
	}
});

const uploadInput = ref(null);

const transfer2Store = useTransfer2Store();
const fileUploader = ref();
const openPlatform = ref(
	process.env.APPLICATION === 'FILES' || process.env.APPLICATION === 'SHARE'
);

onMounted(async () => {
	if (props.autoBindResumable) {
		fileUploader.value =
			uploadInput.value &&
			Resumable.setupResumable({
				uploadInput: uploadInput.value,
				origin_id: props.origin_id
			});
	} else {
		if (uploadInput.value) {
			(uploadInput.value as any).addEventListener(
				'change',
				function (e: any) {
					emits('filesUpdate', e);
				},
				false
			);
		}
	}
	// fileUploader.value =
	// 	uploadInput.value &&
	// 	Resumable.setupResumable({
	// 		uploadInput: uploadInput.value,
	// 		origin_id: props.origin_id
	// 	}); //new Resumable({ uploadInput: uploadInput.value });
});

const onCloseUploadDialog = () => {
	Resumable.onCloseUploadDialog();
};

const emits = defineEmits(['filesUpdate']);
</script>

<style lang="scss" scoped>
.upload-input {
	display: none;
}
</style>
