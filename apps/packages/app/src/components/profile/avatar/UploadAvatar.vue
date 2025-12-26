<template>
	<div class="upload-background-root column justify-center items-center">
		<BtUploader
			class="upload"
			width="120px"
			height="120px"
			:size="1"
			fileName="image"
			accept=".jpg, .jpeg, .png, .gif, image/*"
			action="/images/upload/v1"
			:parmas="{ policy: 'public' }"
			type="avator"
			@ok="ok"
			@fail="fail"
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
import { onMounted, ref } from 'vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useQuasar } from 'quasar';
import { bus } from '@apps/profile/src/utils/bus';
import { useI18n } from 'vue-i18n';

defineProps({
	modelValue: {
		type: Object,
		required: true
	}
});

const $q = useQuasar();
const userStore = useUserStore();
const { t } = useI18n();

const currentPath = ref();
onMounted(async () => {
	console.log(userStore.user);
});

const ok = async (response: any) => {
	console.log('ok ');
	console.log(response.data);

	if (response.code !== 200) {
		$q.notify(response.message);
		return;
	}

	currentPath.value = response.data.imageUrl;
	bus.emit('choice', {
		imageUrl: currentPath.value,
		avatar: currentPath.value
	});
};

const fail = (response: unknown) => {
	console.log('fail', response);
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
