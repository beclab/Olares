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
import { useFilesStore, FilesIdType } from './../../../stores/files';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { axiosInstanceProxy } from '../../../platform/httpProxy';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const filesStore = useFilesStore();
let pdfInstance;
const loading = ref(false);

const rawUrl = () => {
	return filesStore.getDownloadURL(
		filesStore.previewItem[props.origin_id],
		true
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
.pdf-preview-view {
	width: 100%;
	height: 100%;
}
</style>
