<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('delete')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="small"
		@onSubmit="submit"
	>
		<div class="prompt-name q-mb-xs">{{ t('delete_vault_message') }}</div>
		<q-input
			class="prompt-input"
			v-model="promptModel"
			borderless
			dense
			no-error-icon
			:placeholder="t('type_delete_to_confirm')"
			:rules="[
				(val) =>
					(!val || val.toLowerCase() != 'delete') && t('type_delete_to_confirm')
			]"
		/>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

defineProps({
	item: {
		type: Object,
		required: false
	},
	shared_length: {
		type: Number,
		require: false
	}
});

const { t } = useI18n();

const promptModel = ref();
const CustomRef = ref();

const submit = async () => {
	CustomRef.value.onDialogOK(promptModel.value);
};
</script>

<style lang="scss" scoped>
.prompt-name {
	color: rgba(173, 173, 173, 1);
	font-size: 12px;
	line-height: 16px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.prompt-input {
	border: 1px solid rgba(235, 235, 235, 1);
	border-radius: 8px;
	padding: 0 10px;
}
</style>
