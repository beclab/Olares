<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="t('prompts.unsavedChangesTitle')"
		:ok="t('buttons.save')"
		:cancel="t('buttons.cancel')"
		size="medium"
		persistent
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="onSave"
	>
		<div class="dialog-desc text-body2 text-ink-2">
			{{ t('prompts.unsavedChangesMessage') }}
		</div>

		<template #footerMore>
			<q-item
				clickable
				dense
				class="but-discard row justify-center items-center"
				@click="onDiscard"
			>
				{{ t('prompts.discard') }}
			</q-item>
		</template>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const $q = useQuasar();
const dialogRef = ref();

const onSave = () => {
	dialogRef.value?.onDialogOK({ action: 'save' });
};

const onDiscard = () => {
	dialogRef.value?.onDialogOK({ action: 'discard' });
};
</script>

<style lang="scss" scoped>
.dialog-desc {
	text-align: left;
	line-height: 1.5;
}

.but-discard {
	min-width: 100px;
	border-radius: 8px;
	font-weight: 500;
	border: 1px solid $btn-stroke;
	color: $negative;
	font-size: 16px;
	padding: 7px 8px;
	line-height: 24px;
	margin-right: 12px;
}
</style>
