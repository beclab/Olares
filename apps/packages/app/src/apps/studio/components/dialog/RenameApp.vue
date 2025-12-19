<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('dialog_rename_app')"
		:cancel="t('btn_cancel')"
		:ok="t('btn_save')"
		:okLoading="loading ? $t('loading') : false"
		size="medium"
		@onSubmit="submit"
	>
		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('dialog_rename_title') }} <span class="text-red-default">*</span>
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					ref="titleRef"
					dense
					borderless
					no-error-icon
					v-model="appTitle"
					autofocus
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="t('dialog_rename_placeholder')"
					:rules="[
						(val) => (val && val.length > 0) || t('dialog_rename_required')
					]"
				>
				</q-input>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	currentTitle: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const CustomRef = ref();
const titleRef = ref();
const loading = ref(false);
const appTitle = ref(props.currentTitle || '');

const submit = async () => {
	console.log('click submit');
	titleRef.value.validate();
	if (titleRef.value.hasError) return;

	loading.value = true;

	try {
		loading.value = false;
		CustomRef.value.onDialogOK(appTitle.value);
	} catch (error) {
		loading.value = false;
	}
};
</script>

<style lang="scss" scoped>
.form-item {
	.form-item-key {
		width: 100px;
		height: 40px;
		line-height: 40px;
		text-align: center;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
