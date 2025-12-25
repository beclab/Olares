<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		@onSubmit="onDialogOK"
	>
		<div
			class="dialog-desc"
			:style="{ textAlign: isMobile ? 'center' : 'left' }"
		>
			{{ message }}
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { ref, defineProps } from 'vue';
import { useI18n } from 'vue-i18n';

const $q = useQuasar();
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

defineProps({
	title: String,
	message: String,
	navigation: String,
	position: String
});

const { t } = useI18n();

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-desc {
			padding-left: 20px;
		}
	}
}
</style>
