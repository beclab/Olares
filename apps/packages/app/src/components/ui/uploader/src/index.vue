<template>
	<div style="display: inline-block" :style="{ width: width, height: height }">
		<div class="wrap" :style="{ width: width, height: height }">
			<input
				v-show="!loading"
				class="quploader"
				@change="selectChange($event)"
				:accept="accept"
				type="file"
				:multiple="false"
				title=""
			/>
			<slot class="slot" />
		</div>

		<Cropper
			v-if="showCropper"
			:dialogVisible="showCropper"
			:cropper-img="cropperImg"
			@colse-dialog="closeDialog"
			@upload-img="uploadImg"
		/>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, type PropType } from 'vue';
import axios, { AxiosRequestConfig, isAxiosError } from 'axios';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { IMAGES_UPLOAD_V1_ACCEPT } from '../../../../utils/upload/imagesUploadV1Formats';
import Cropper from './Cropper.vue';

function shouldCheckMinImageDimensions(minW: unknown, minH: unknown): boolean {
	return (
		typeof minW === 'number' &&
		Number.isFinite(minW) &&
		minW > 0 &&
		typeof minH === 'number' &&
		Number.isFinite(minH) &&
		minH > 0
	);
}

function readImageNaturalDimensions(file: File): Promise<{
	w: number;
	h: number;
}> {
	return new Promise((resolve, reject) => {
		const url = URL.createObjectURL(file);
		const img = new Image();
		img.onload = () => {
			const w = img.naturalWidth;
			const h = img.naturalHeight;
			URL.revokeObjectURL(url);
			resolve({ w, h });
		};
		img.onerror = () => {
			URL.revokeObjectURL(url);
			reject(new Error('decode'));
		};
		img.src = url;
	});
}

function parseUploadAxiosMessage(error: unknown): string {
	if (typeof error === 'string') {
		return error;
	}
	if (isAxiosError(error)) {
		const data = error.response?.data as unknown;
		if (typeof data === 'string' && data) {
			return data;
		}
		if (data && typeof data === 'object') {
			const o = data as Record<string, unknown>;
			if (typeof o.message === 'string' && o.message) {
				return o.message;
			}
			if (typeof o.msg === 'string' && o.msg) {
				return o.msg;
			}
		}
		if (error.message) {
			return error.message;
		}
	}
	return '';
}

export default defineComponent({
	name: 'BtUploader',
	components: {
		Cropper
	},
	props: {
		width: {
			type: String,
			required: false,
			default: '100px'
		},
		height: {
			type: String,
			required: false,
			default: '100px'
		},
		action: {
			type: String,
			required: true
		},
		accept: {
			type: String,
			required: false,
			default: IMAGES_UPLOAD_V1_ACCEPT
		},
		size: {
			type: Number,
			required: false,
			default: 5
		},
		type: {
			type: String,
			required: false,
			default: 'img'
			// validator(value) {
			//   return ['img', 'avator'].includes(value)
			// }
		},
		fileName: {
			type: String,
			required: false,
			default: 'file'
		},
		formData: {
			type: FormData,
			required: false
		},
		config: {
			type: Object as () => AxiosRequestConfig,
			required: false
		},
		parmas: {
			type: Object,
			required: false
		},
		notifyFail: {
			type: Boolean,
			required: false,
			default: true
		},
		minImageWidth: {
			type: Number,
			required: false,
			default: undefined
		},
		minImageHeight: {
			type: Number,
			required: false,
			default: undefined
		},
		fileGuard: {
			type: Function as PropType<(file: File) => string | null>,
			required: false,
			default: undefined
		}
	},

	setup(props: any, context) {
		const { t } = useI18n();
		const cropperImg = ref('');
		const showCropper = ref(false);
		const uploadFile = ref();
		const loading = ref(false);

		const notifyUser = (message: string) => {
			if (!props.notifyFail) {
				return;
			}
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message
			});
		};

		const dispatchFail = (message: string) => {
			context.emit('fail', message);
			notifyUser(message);
		};

		const runFileGuard = (payload: Blob | File): boolean => {
			const guard = props.fileGuard as
				| ((file: File) => string | null)
				| undefined;
			if (!guard) {
				return true;
			}
			const file =
				payload instanceof File
					? payload
					: new File([payload], 'image.png', {
							type: payload.type || 'image/png'
					  });
			const msg = guard(file);
			if (msg) {
				dispatchFail(msg);
				return false;
			}
			return true;
		};

		const factoryFn = async () => {
			let formData: FormData, config;
			if (props.formData) {
				formData = props.formData;
				formData.append(props.fileName, uploadFile.value);
			} else {
				formData = new FormData();
				formData.append(props.fileName, uploadFile.value);
				if (props.parmas) {
					Object.keys(props.parmas).forEach((key) => {
						formData.append(key, props.parmas[key]);
					});
				}
			}

			if (props.config) {
				config = props.config;
			} else {
				config = {
					headers: { 'Content-Type': 'multipart/form-data' }
				};
			}
			const axiosInstance = axios.create(config);
			await axiosInstance
				.post(props.action, formData)
				.then(function (response) {
					context.emit('ok', response.data);
					showCropper.value = false;
					loading.value = false;
					context.emit('loading', false);
				})
				.catch(function (error: unknown) {
					const parsed =
						parseUploadAxiosMessage(error) || t('bt_uploader_upload_failed');
					dispatchFail(parsed);
					loading.value = false;
					context.emit('loading', false);
				});
		};

		const selectChange = async (e: any) => {
			const raw = e?.target?.files?.[0];
			if (!raw) {
				return;
			}
			await openCropper(raw);
			e.target.value = '';
		};

		const openCropper = async (files: Blob) => {
			const exceedsMaxBytes = files.size > props.size << 20;
			if (exceedsMaxBytes) {
				dispatchFail(t('bt_uploader_image_too_large', { maxMb: props.size }));
				return;
			}

			if (!runFileGuard(files)) {
				return;
			}

			if (
				files instanceof File &&
				shouldCheckMinImageDimensions(props.minImageWidth, props.minImageHeight)
			) {
				try {
					const { w, h } = await readImageNaturalDimensions(files);
					if (w < props.minImageWidth || h < props.minImageHeight) {
						dispatchFail(
							t('bt_uploader_image_dimensions_too_small', {
								minWidth: props.minImageWidth,
								minHeight: props.minImageHeight
							})
						);
						return;
					}
				} catch {
					dispatchFail(t('bt_uploader_read_file_failed'));
					return;
				}
			}

			if (props.type === 'img') {
				uploadImg(files as any);
				return;
			}

			const reader = new FileReader();
			reader.onerror = () => {
				dispatchFail(t('bt_uploader_read_file_failed'));
			};
			reader.onload = (e) => {
				let data;
				if (typeof e?.target?.result === 'object') {
					data =
						e.target.result &&
						window.URL.createObjectURL(new Blob([e.target.result]));
				} else {
					data = e?.target?.result;
				}
				cropperImg.value = data || '';
				if (!cropperImg.value) {
					dispatchFail(t('bt_uploader_read_file_failed'));
					return;
				}
				showCropper.value = true;
			};
			reader.readAsArrayBuffer(files);
		};

		const closeDialog = () => {
			showCropper.value = false;
			cropperImg.value = '';
		};

		const uploadImg = async (file: string | Blob | File) => {
			if (file instanceof Blob) {
				if (!runFileGuard(file)) {
					return;
				}
			}

			uploadFile.value = file;
			loading.value = true;
			context.emit('loading', true);
			factoryFn();
		};

		return {
			cropperImg,
			showCropper,
			uploadFile,
			loading,
			closeDialog,
			uploadImg,
			selectChange
		};
	}
});
</script>

<style lang="scss" scoped>
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
