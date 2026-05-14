<template>
	<div class="upload-background-root column justify-center items-center">
		<BtUploader
			class="upload"
			width="120px"
			height="120px"
			:size="5"
			:min-image-width="400"
			:min-image-height="400"
			:file-guard="imagesUploadFormatGuard"
			fileName="image"
			:accept="IMAGES_UPLOAD_V1_ACCEPT"
			action="/images/upload/v1"
			:parmas="{ policy: 'public' }"
			type="avator"
			@ok="ok"
		>
			<div class="upload-image-inner column justify-center items-center">
				<q-icon size="20px" name="sym_r_add_photo_alternate" />
				<div class="upload-image-inner-label">
					{{ t('profile.upload_image') }}
				</div>
			</div>
		</BtUploader>

		<div class="upload-image-title">
			{{ t('profile.select_local_image_desc') }}
		</div>
		<div class="upload-image-label">
			{{ t('profile.recommend_sizes') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	createImagesUploadV1FormatGuard,
	IMAGES_UPLOAD_V1_ACCEPT
} from 'src/utils/upload/imagesUploadV1Formats';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { bus } from '@apps/profile/src/utils/bus';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

defineProps({
	modelValue: {
		type: Object,
		required: true
	}
});

const userStore = useUserStore();
const { t } = useI18n();

const imagesUploadFormatGuard = createImagesUploadV1FormatGuard(t);

const currentPath = ref();
onMounted(async () => {
	console.log(userStore.user);
});

const ok = async (response: any) => {
	console.log('ok ');
	console.log(response.data);

	if (response.code !== 200) {
		notifyFailed(String(response.message ?? ''));
		return;
	}

	currentPath.value = response.data.imageUrl;
	bus.emit('choice', {
		imageUrl: currentPath.value,
		avatar: currentPath.value
	});
};
</script>

<style scoped lang="scss">
.upload-background-root {
	width: 100%;
	height: 100%;
	padding: 0 20px;

	.upload-image-inner {
		border-radius: 8px;
		border: 1px solid $separator-2;
		width: 120px;
		height: 120px;

		.upload-image-inner-label {
			margin-top: 8px;
			font-family: Roboto;
			font-size: 12px;
			font-weight: 400;
			line-height: 16px;
			letter-spacing: 0em;
			text-align: left;
			color: $ink-2;
		}
	}

	.upload-image-title {
		font-family: Roboto;
		font-size: 12px;
		margin-top: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		color: $ink-1;
	}

	.upload-image-label {
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		margin-top: 4px;
		color: $ink-3;
	}
}
</style>
