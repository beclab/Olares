<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="confirmBtnTitle"
		:cancel="showCancel ? cancelBtnTitle : false"
		size="small"
		:noRouteDismiss="true"
		@onSubmit="submit"
	>
		<div class="dialog-desc text-ink-2 text-body-3" v-html="message"></div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { i18n } from '../../boot/i18n';

defineProps({
	title: {
		type: String,
		default: '',
		required: false
	},
	message: {
		type: String,
		default: '',
		required: false
	},

	confirmBtnTitle: {
		type: String,
		default: i18n.global.t('confirm'),
		required: false
	},
	cancelBtnTitle: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	},
	showCancel: {
		type: Boolean,
		default: false,
		required: false
	}
});

const CustomRef = ref();

const submit = async () => {
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
			padding-right: 20px;
		}
	}
}
</style>
