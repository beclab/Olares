<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('app.uninstall')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		@onSubmit="onOK"
	>
		<div v-if="step === 1" class="column">
			<div class="text-ink-2 text-body3">
				{{
					t('my.sure_to_uninstall_the_app', {
						title: appName
					})
				}}
			</div>

			<bt-check-box
				v-if="showCheckbox"
				:label="t('Also uninstall the shared server (affects all users)')"
				check-img="market/check_box.svg"
				uncheck-img="market/uncheck_box.svg"
				class="q-mt-md"
				:model-value="selected"
				@update:model-value="onUpdate"
			/>
		</div>

		<div v-if="step === 2" class="column">
			<div class="text-negative text-body2" style="white-space: pre-line">
				{{ t('Warning! Uninstalling the shared server will:', { appName }) }}
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import BtCheckBox from '../rss/BtCheckBox.vue';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();
const customRef = ref();
const step = ref(1);

const props = defineProps({
	modelValue: {
		type: Boolean,
		default: false,
		required: true
	},
	appName: {
		type: String,
		required: true
	},
	showCheckbox: {
		type: Boolean,
		default: false,
		required: false
	}
});

const selected = ref(props.modelValue);

const onUpdate = (status) => {
	selected.value = status;
};

const onOK = () => {
	if (!props.showCheckbox || step.value === 2) {
		customRef.value.onDialogOK(selected.value);
	} else if (props.showCheckbox && step.value === 1) {
		if (selected.value) {
			step.value = 2;
		} else {
			customRef.value.onDialogOK(selected.value);
		}
	}
};
</script>

<style scoped lang="scss"></style>
