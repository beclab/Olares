<template>
	<div class="cropper-wrap">
		<div class="cropper-dialog" ref="dialogRef" :model-value="dialogVisible">
			<div class="cropper-container">
				<div>
					<div class="hearder">Cropper</div>
				</div>
				<div class="cropper">
					<div class="cropper-el">
						<VueCropper
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
							:auto-crop-width="option.autoCropWidth"
							:auto-crop-height="option.autoCropHeight"
							:center-box="option.centerBox"
							:high="option.high"
							:info-true="option.infoTrue"
							:enlarge="option.enlarge"
							:fixed="option.fixed"
							:fixed-number="option.fixedNumber"
							@realTime="realTime"
						/>
					</div>
					<div class="prive-el">
						<div
							class="prive-style"
							:style="{
								width: '200px',
								height: '200px',
								overflow: 'hidden',
								margin: '0 25px',
								display: 'flex',
								'align-items': 'center'
							}"
						>
							<div class="preview" :style="previews.div">
								<img :src="previews.url" :style="previews.img" />
							</div>
						</div>
					</div>
				</div>

				<div align="right" class="bottom">
					<profile-button
						:default-selected="false"
						class="text-body3 q-mr-lg"
						label="Cancel"
						@click="handleClose"
					/>
					<profile-button
						width="75px"
						class="text-body3 q-mr-lg"
						label="Confirm"
						@click="saveImg"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import 'vue-cropper/dist/index.css';
import { defineComponent, ref } from 'vue';
import { VueCropper } from 'vue-cropper';
import ProfileButton from './ProfileButton.vue';

export default defineComponent({
	name: 'MyCropper',
	components: {
		ProfileButton,
		VueCropper
	},
	props: {
		dialogVisible: {
			type: Boolean,
			default: false
		},
		imgType: {
			type: String,
			default: 'blob'
		},
		cropperImg: {
			type: String,
			default: ''
		}
	},

	setup(props, context) {
		const previews = ref({
			div: '',
			url: '',
			img: ''
		});
		const option = ref({
			img: '',
			size: 1,
			full: false,
			outputType: 'png',
			canMove: false,
			fixedBox: false,
			original: false,
			canMoveBox: true,
			autoCrop: true,

			autoCropWidth: 200,
			autoCropHeight: 150,
			centerBox: true,
			high: false,
			enlarge: 1,
			mode: 'contain',
			maxImgSize: 2000,
			limitMinSize: [100, 120],
			infoTrue: false,
			fixed: true,
			fixedNumber: [1, 1]
		});

		const cropper = ref();

		const realTime = (data: any) => {
			console.log('realTime', data);
			previews.value = data;
		};

		const handleClose = () => {
			context.emit('colse-dialog', false);
		};

		const saveImg = () => {
			if (props.imgType === 'blob') {
				cropper.value.getCropBlob((data: any) => {
					console.log('blobdata', data);
					const file = new File([data], 'cropped-image.png', {
						type: data.type
					});
					context.emit('upload-img', file);
				});
			} else {
				cropper.value.getCropData((data: any) => {
					context.emit('upload-img', data);
				});
			}
		};

		return {
			previews,
			option,
			cropper,
			realTime,
			handleClose,
			saveImg
		};
	}
});
</script>

<style lang="scss" scoped>
.cropper-wrap {
	width: 100vw;
	height: 100vh;
	position: fixed;
	left: 0;
	top: 0;
	background-color: rgba(0, 0, 0, 0.4);
	display: flex;
	align-items: center;
	justify-content: center;
	.cropper-dialog {
		width: 500px;
		height: 340px;
		background-color: #ffffff;
		padding: 20px;
		padding-top: 0;
		border-radius: 8px;
		.cropper-container {
			width: 100%;
			height: 100%;
			overflow: hidden;
			.hearder {
				width: 100%;
				line-height: 60px;
				font-weight: 700;
			}
		}
	}
}

.cropper {
	display: flex;
	justify-content: center;
	align-items: center;
}

.cropper-el {
	height: 200px;
	width: 50%;
}

.prive-el {
	height: 200px;
	width: 50%;
	text-align: center;
	margin-left: 20px;
	display: flex;
	align-items: center;
	.preview {
		overflow: hidden;
		background: #ededed;
	}
	.el-button {
		margin-top: 20px;
	}
}

.bottom {
	display: flex;
	align-items: center;
	justify-content: end;
	margin-top: 30px;
	button {
		width: 120px;
		padding: 10px;
		border: none;
		cursor: pointer;
		border-radius: 4px;
		font-weight: 500;
		border: 1px solid #e1e1e1;
	}
	.cancel {
	}
	.confirm {
		margin-left: 20px;
	}
}
</style>
