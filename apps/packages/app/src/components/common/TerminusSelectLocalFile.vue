<template>
	<div>
		<input
			ref="uploadInput"
			type="file"
			style="display: none"
			:accept="accept"
			:multiple="multiple"
			@change="handleFileSelection"
		/>
		<div @click="openFileInput">
			<slot />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

defineProps({
	accept: {
		type: String,
		default: '*',
		required: false
	},
	multiple: {
		type: Boolean,
		required: false,
		default: false
	}
});

const uploadInput = ref();
const emit = defineEmits(['onSuccess']);

function openFileInput() {
	uploadInput.value.value = null;
	uploadInput.value.click();
}

function handleFileSelection(event: any) {
	const files = event.target.files;
	if (files) {
		emit('onSuccess', files);
	}
}
</script>

<style scoped lang="scss"></style>
