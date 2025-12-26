<template>
	<div class="wrap" :style="{ width: `${width}px`, height: `${height}px` }">
		<input
			class="quploader"
			@change="selectChange($event)"
			accept="image/*"
			type="file"
			:multiple="false"
		/>
		<slot class="slot" />
	</div>
</template>

<script lang="ts" setup>
const props = defineProps({
	width: {
		type: Number,
		required: false,
		default: 100
	},
	height: {
		type: Number,
		required: false,
		default: 100
	},
	size: {
		type: Number,
		required: false,
		default: 5
	}
});

const emit = defineEmits(['onFileSelected']);

const selectChange = async (e: any) => {
	emit('onFileSelected', e.target.files[0]);
	e.target.value = '';
};

const convertFile = (file: any) => {
	var files = file;
	let isLt5M = files.size > props.size << 20;
	if (isLt5M) {
		return `Please upload pictures within ${props.size}M`;
	}
	var reader = new FileReader();
	reader.onload = (e) => {
		let data;
		if (typeof e?.target?.result === 'object') {
			data =
				e.target.result &&
				window.URL.createObjectURL(new Blob([e.target.result]));
		} else {
			data = e?.target?.result;
		}
		emit('onFileSelected', data || '');
	};
	// Convert to base64
	// reader.readAsDataURL(file)
	// Convert to blob
	reader.readAsArrayBuffer(files);
};
</script>

<style lang="scss">
.wrap {
	position: relative;

	.quploader {
		display: inline-block;
		box-shadow: none;
		opacity: 0;
		width: 100%;
		height: 100%;
		cursor: pointer;
		position: absolute;
		left: 0;
		top: 0;
		z-index: 1;
	}

	.slot {
		position: absolute;
		left: 0;
		top: 0;
	}
}
</style>
