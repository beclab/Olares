<template>
	<div>
		<input
			ref="uploadInput"
			type="file"
			style="display: none"
			accept=".tgz,.gz"
			@change="handleFileSelection"
		/>
		<div @click="openFileInput">
			<slot />
		</div>
	</div>
</template>

<script lang="ts" setup>
import LocalUploadDialog from '../../pages/market/me/LocalUploadDialog.vue';
import { bus, BUS_EVENT } from '../../utils/bus';
import { useQuasar } from 'quasar';
import { ref } from 'vue';

const uploadInput = ref();
const selectedFile = ref(null);
const $q = useQuasar();
const emit = defineEmits(['onSuccess']);

function openFileInput() {
	uploadInput.value.value = null;
	uploadInput.value.click();
}

function handleFileSelection(event: any) {
	console.log(event);
	const file = event.target.files[0];
	if (file) {
		selectedFile.value = file;
		$q.dialog({
			component: LocalUploadDialog,
			componentProps: {
				file: selectedFile.value
			}
		})
			.onOk(() => {
				selectedFile.value = null;
			})
			.onDismiss(() => {
				selectedFile.value = null;
			});
	} else {
		bus.emit(BUS_EVENT.APP_BACKEND_ERROR, 'File selection failed.');
	}
}
</script>

<style scoped lang="scss"></style>
