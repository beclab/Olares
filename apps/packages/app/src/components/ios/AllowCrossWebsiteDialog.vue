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
					<q-img src="../../assets/setting/allow_cross_website.png" noSpinner />
				</div>
			</template>
		</UserStatusCommonContent>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import UserStatusCommonContent from '../userStatusDialog/UserStatusCommonContent.vue';

defineProps({
	title: String,
	message: String,
	navigation: String,
	btnTitle: String
});

const CustomRef = ref();

const onSubmit = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.user-status-common-dialog-root {
	.q-dialog__backdrop {
		background: rgba(0, 0, 0, 0.7);
	}
}
</style>
