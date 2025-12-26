<template>
	<div class="pdf-preview-view">
		<div
			v-if="loading"
			style="width: 100%; height: 100%"
			class="row items-center justify-center"
		>
			<q-spinner-dots color="primary" size="3em" />
		</div>
		<div id="termipass-pdf-preview"></div>
	</div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { axiosInstanceProxy } from '../../../platform/httpProxy';

import { useFilesStore, FilesIdType } from '../../../stores/files';
import { dataAPIs } from './../../../api';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const filesStore = useFilesStore();
let pdfInstance;
const loading = ref(false);

const rawUrl = () => {
	const dataAPI = dataAPIs(filesStore.previewItem[props.origin_id].driveType);

	return dataAPI.getDownloadURL(
		filesStore.previewItem[props.origin_id],
		true,
		false
	);
};

onMounted(async () => {
	const config = {
		headers: {
			'Content-Type': 'application/json',
			Accept: 'application/json'
		},
		responseType: 'arraybuffer'
	} as any;
	const instance = axiosInstanceProxy(config);
	instance.interceptResponse((res) => {
		if (typeof res.data === 'string') {
			res.data = atob(res.data);
		}
		return res;
	});
	try {
		loading.value = true;
		console.log(rawUrl());

		const response = await instance.get(rawUrl(), config);
		loading.value = false;
		const { default: Pdfh5 } = await import('pdfh5');
		await import('pdfh5/css/pdfh5.css');
		pdfInstance = new Pdfh5('#termipass-pdf-preview', {
			data: response.data
		});
	} catch (error) {
		loading.value = false;
	}
});
onBeforeUnmount(() => {
	if (pdfInstance) {
		pdfInstance.destroy();
	}
});
</script>

<style scoped lang="scss">
@import 'pdfh5/css/pdfh5.css';

.pdf-preview-view {
	width: 100%;
	height: 100%;
}
</style>
