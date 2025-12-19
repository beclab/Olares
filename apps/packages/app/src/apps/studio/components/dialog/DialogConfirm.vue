<template>
	<bt-custom-dialog
		persistent
		ref="CustomRef"
		:title="title"
		:cancel="t('btn_cancel')"
		:ok="t('btn_confirm')"
		size="small"
		@onSubmit="submit"
	>
		<div class="dialog-desc">{{ message }}</div>

		<div class="dialog-checkbox" v-if="showAgain">
			<q-checkbox
				v-model="checkbox"
				color="green"
				label="Do not show this confirmation again."
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, defineProps } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const CustomRef = ref();
const checkbox = ref(false);

defineProps({
	title: {
		type: String,
		required: false,
		default: 'Alert'
	},
	message: {
		type: String,
		required: false,
		default: 'Some message'
	},
	showAgain: {
		type: Boolean,
		required: false,
		default: false
	},
	okText: {
		type: String,
		required: false,
		default: 'Confirm'
	}
});

const submit = () => {
	CustomRef.value.onDialogOK(checkbox.value);
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-desc {
			padding: 12px 20px;
		}
		.dialog-checkbox {
			padding: 0px 20px;
		}
	}
}
</style>
