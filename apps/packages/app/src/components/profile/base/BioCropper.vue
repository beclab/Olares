<template>
	<div :style="{ height: `${imageSize}px`, width: `${imageSize}px` }">
		<vue-cropper
			ref="cropper"
			:img="cropperImg"
			:output-size="option.size"
			:output-type="option.outputType"
			:info="true"
			:full="option.full"
			:can-move="option.canMove"
			:can-move-box="option.canMoveBox"
			:fixed-box="option.fixedBox"
			:original="option.original"
			:auto-crop="option.autoCrop"
			:auto-crop-width="cropSize"
			:auto-crop-height="cropSize"
			:center-box="option.centerBox"
			:high="option.high"
			:info-true="option.infoTrue"
			:enlarge="option.enlarge"
			:fixed="option.fixed"
			:fixed-number="option.fixedNumber"
			@realTime="realTime"
		/>
	</div>
</template>

<script lang="ts" setup>
import 'vue-cropper/dist/index.css';
import { ref } from 'vue';
import { VueCropper } from 'vue-cropper';

const props = defineProps({
	imageSize: {
		type: Number,
		default: 182
	},
	imgType: {
		type: String,
		default: 'blob'
	},
	cropperImg: {
		type: String,
		default: ''
	},
	cropSize: {
		type: Number,
		default: 182
	}
});

const previews = ref({
	div: '',
	url: '',
	img: ''
});
const option = ref({
	img: '', // Source image URL
	size: 1, // Output image quality
	full: false, // Export crop with original aspect ratio (default: false)
	outputType: 'png', // Output format (default: jpg)
	canMove: true, // Whether uploaded image can be moved
	fixedBox: true, // Keep crop box size fixed
	original: false, // Render uploaded image with original ratio
	canMoveBox: true, // Whether crop box can be dragged
	autoCrop: true, // Create crop box by default
	// Width/height options only work when autoCrop is enabled.
	// autoCropWidth: 182, // Default crop box width
	// autoCropHeight: 182, // Default crop box height
	centerBox: true, // Keep crop box inside image bounds
	high: false, // Output with device DPR scale
	enlarge: 1, // Output scale multiplier based on crop box
	mode: 'contain', // Default image fit mode
	maxImgSize: 2000, // Max image width/height
	limitMinSize: [100, 100], // Min crop box size
	infoTrue: false, // true: real output size; false: visible crop box size
	fixed: true, // Keep fixed crop ratio (default: true)
	fixedNumber: [1, 1] // Crop box width/height ratio
});

const cropper = ref();

const realTime = (data: any) => {
	console.log('realTime', data);
	previews.value = data;
};

const emit = defineEmits(['upload-img']);

const getCropImg = () => {
	if (props.imgType === 'blob') {
		cropper.value.getCropBlob((data: any) => {
			console.log('blobdata', data);
			emit('upload-img', data);
		});
	} else {
		cropper.value.getCropData((data: any) => {
			emit('upload-img', data);
		});
	}
};

defineExpose({ getCropImg });
</script>

<style lang="scss" scoped></style>
