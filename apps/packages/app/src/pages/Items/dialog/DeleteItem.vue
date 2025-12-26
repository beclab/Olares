<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="small"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="onDialogOK"
	>
		<q-card-section class="row q-px-none">
			{{ content }}
		</q-card-section>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../../boot/i18n';

defineProps({
	title: {
		type: String,
		default: i18n.global.t('message'),
		required: false
	},
	content: {
		type: String,
		default: i18n.global.t('this_is_a_meeage'),
		required: false
	}
});

const { t } = useI18n();

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 400px;
		border-radius: 12px;
	}
}
</style>
