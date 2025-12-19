<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Reset SSH Password')"
		:skip="false"
		:ok="t('ok')"
		size="medium"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		:okDisabled="!enableSave"
		@onSubmit="updatePassword"
	>
		<terminus-edit
			v-model="newPassword"
			:label="t('password')"
			style="width: 100%"
			class="q-mt-lg"
			:show-password-img="true"
			:is-error="newPassword.length > 0 && passwordRule(newPassword).length > 0"
			:error-message="passwordRule(newPassword)"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';
const newPassword = ref('');

const { t } = useI18n();

const CustomRef = ref();

const deviceStore = useDeviceStore();

const enableSave = computed(() => {
	return passwordRule(newPassword.value).length == 0;
});

const passwordRule = (val: string) => {
	if (val.length < 10) return t('errors.at_least_10_digits_long');
	let rule = /^(?=.*[A-Z])(?=.*[a-z])[A-Za-z0-9@$!%*?&_.]{10,}$/;
	if (!rule.test(val)) {
		return t('errors.at_least_one_uppercase_and_lowercase_letter');
	}
	return '';
};

const updatePassword = async () => {
	if (newPassword.value.length === 0) {
		return;
	}

	CustomRef.value.onDialogOK(newPassword.value);
};
</script>
