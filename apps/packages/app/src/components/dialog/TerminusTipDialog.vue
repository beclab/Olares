<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="position ?? t('confirm')"
		:cancel="showCancel ? navigation ?? t('cancel') : false"
		size="small"
		platform="mobile"
		@onSubmit="onSubmit"
	>
		<div
			class="biometric-unlock-dialog__desc login-sub-title"
			:style="{ textAlign: descAlign }"
			v-html="message"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import { ref } from 'vue';

const { t } = useI18n();

defineProps({
	title: String,
	message: String,
	navigation: String,
	position: String,
	showCancel: {
		default: true,
		type: Boolean
	},
	descAlign: {
		default: 'center',
		type: String as () => 'left' | 'center' | 'right'
	}
});

const CustomRef = ref();

const onSubmit = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.biometric-unlock-dialog {
	padding: 20px;

	&__title {
		color: $ink-1;
	}

	&__desc {
		width: 100%;
		margin-top: 12px;
	}

	&__group {
		margin-top: 40px;
		width: 100%;

		&__btn {
			width: calc(50% - 6px);
		}
		.cancel {
			border: 1px solid $separator;
		}
	}
}
</style>
