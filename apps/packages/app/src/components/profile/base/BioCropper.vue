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
	img: '', // 裁剪图片的地址
	size: 1, // 裁剪生成图片的质量
	full: false, // 是否输出原图比例的截图 默认false
	outputType: 'png', // 裁剪生成图片的格式 默认jpg
	canMove: true, // 上传图片是否可以移动
	fixedBox: true, // 固定截图框大小 不允许改变
	original: false, // 上传图片按照原始比例渲染
	canMoveBox: true, // 截图框能否拖动
	autoCrop: true, // 是否默认生成截图框
	// 只有自动截图开启 宽度高度才生效
	// autoCropWidth: 182, // 默认生成截图框宽度
	// autoCropHeight: 182, // 默认生成截图框高度
	centerBox: true, // 截图框是否被限制在图片里面
	high: false, // 是否按照设备的dpr 输出等比例图片
	enlarge: 1, // 图片根据截图框输出比例倍数
	mode: 'contain', // 图片默认渲染方式
	maxImgSize: 2000, // 限制图片最大宽度和高度
	limitMinSize: [100, 100], // 更新裁剪框最小属性
	infoTrue: false, // true 为展示真实输出图片宽高 false 展示看到的截图框宽高
	fixed: true, // 是否开启截图框宽高固定比例  (默认:true)
	fixedNumber: [1, 1] // 截图框的宽高比例
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
