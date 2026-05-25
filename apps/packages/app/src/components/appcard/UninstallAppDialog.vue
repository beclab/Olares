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
				v-if="showCSAppCheckbox"
				:label="t('Also uninstall the shared server (affects all users)')"
				:link-label="t('Learn more')"
				link="https://docs.olares.com/manual/olares/market/market.html#uninstall-shared-applications"
				check-img="market/check_box.svg"
				uncheck-img="market/uncheck_box.svg"
				class="q-mt-md"
				:model-value="allRef"
				@update:model-value="onUpdateAll"
			/>

			<bt-check-box
				:label="t('Also remove all local data')"
				:link-label="t('Learn more')"
				link="https://docs.olares.com/manual/olares/market/market.html#uninstall-applications"
				check-img="market/check_box.svg"
				uncheck-img="market/uncheck_box.svg"
				class="q-mt-md"
				:model-value="clearDataRef"
				@update:model-value="onUpdateClear"
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
	appName: {
		type: String,
		required: true
	},
	showCSAppCheckbox: {
		type: Boolean,
		default: false
	},
	all: {
		type: Boolean,
		default: true,
		required: true
	},
	clearData: {
		type: Boolean,
		default: true,
		required: true
	}
});

const allRef = ref(props.all);
const clearDataRef = ref(props.clearData);

const onUpdateAll = (status: boolean) => {
	allRef.value = status;
};

const onUpdateClear = (status: boolean) => {
	clearDataRef.value = status;
};

const onOK = () => {
	if (step.value === 2) {
		customRef.value.onDialogOK({
			all: allRef.value,
			clearData: clearDataRef.value
		});
	} else if (step.value === 1) {
		if (props.showCSAppCheckbox && allRef.value) {
			step.value = 2;
		} else {
			customRef.value.onDialogOK({
				all: allRef.value,
				clearData: clearDataRef.value
			});
		}
	}
};
</script>

<style scoped lang="scss"></style>
