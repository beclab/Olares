<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('app.resume')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		@onSubmit="onOK"
	>
		<div class="text-ink-2 text-body3">
			{{
				t(
					'Are you sure you want to resume the shared app "{app}" ?（affects all users）',
					{
						app: appName
					}
				)
			}}
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();
const customRef = ref();

defineProps({
	appName: {
		type: String,
		required: true
	}
});

const onOK = () => {
	customRef.value.onDialogOK(true);
};
</script>

<style scoped lang="scss"></style>
