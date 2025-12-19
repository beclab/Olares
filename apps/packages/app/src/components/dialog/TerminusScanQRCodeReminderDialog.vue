<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="btnTitle"
		:cancel="false"
		:size="'small'"
		:platform="'mobile'"
		@onSubmit="onSubmit"
	>
		<UserStatusCommonContent :message="message" :btnRedefined="true">
			<template v-slot:more>
				<div
					class="row items-center justify-between q-mt-lg"
					style="width: 100%"
				>
					<q-img :src="reminderImg" noSpinner />
				</div>
			</template>
		</UserStatusCommonContent>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import UserStatusCommonContent from '../userStatusDialog/UserStatusCommonContent.vue';
import { getRequireImage } from '../../utils/imageUtils';
import { ref } from 'vue';
import { useUserStore } from '../../stores/user';
import { i18n } from '../../boot/i18n';

defineProps({
	title: String,
	message: String,
	navigation: String,
	btnTitle: String
});

const userStore = useUserStore();

const currentLanguage = ref(userStore.locale || i18n.global.locale.value);

const reminderImg = ref(
	currentLanguage.value == 'zh-CN'
		? getRequireImage('setting/active_qr_code_reminder_img_cn.png')
		: getRequireImage('setting/active_qr_code_reminder_img.png')
);

const CustomRef = ref();

const onSubmit = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped></style>
