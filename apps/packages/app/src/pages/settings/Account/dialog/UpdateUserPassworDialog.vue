<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('change_password')"
		:skip="false"
		:ok="t('ok')"
		size="medium"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		:okDisabled="!enableSave"
		@onSubmit="updatePassword"
	>
		<terminus-edit
			v-model="oldPassword"
			:label="t('old_password')"
			style="width: 100%"
			:show-password-img="true"
			:is-error="
				oldPassword.length > 0 && oldPasswordRule(oldPassword).length > 0
			"
			:error-message="oldPasswordRule(oldPassword)"
		/>

		<terminus-edit
			v-model="newPassword"
			:label="t('new_password')"
			style="width: 100%"
			class="q-mt-lg"
			:show-password-img="true"
			:is-error="newPassword.length > 0 && passwordRule(newPassword).length > 0"
			:error-message="passwordRule(newPassword)"
		/>

		<terminus-edit
			v-model="repeatPassword"
			:label="t('repeat_password')"
			style="width: 100%"
			class="q-mt-lg"
			:show-password-img="true"
			:is-error="
				repeatPassword.length > 0 && repasswordRule(repeatPassword).length > 0
			"
			:error-message="repasswordRule(repeatPassword)"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useUserStore } from 'src/stores/settings/user';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { notifySuccess } from 'src/utils/settings/btNotify';
import { passwordAddSort } from 'src/utils/account';
import { useAdminStore } from 'src/stores/settings/admin';

const oldPassword = ref('');
const newPassword = ref('');
const repeatPassword = ref('');

const accountStore = useUserStore();
const quasar = useQuasar();
const { t } = useI18n();
const adminStore = useAdminStore();

const props = defineProps({
	name: {
		type: String,
		required: false,
		default: ''
	}
});

const enableSave = computed(() => {
	return (
		oldPasswordRule(oldPassword.value).length == 0 &&
		passwordRule(newPassword.value).length == 0 &&
		repasswordRule(repeatPassword.value).length == 0
	);
});

const oldPasswordRule = (val: string) => {
	if (val.length == 0) {
		return t('errors.old_password_is_empty');
	}
	return '';
};

const passwordRule = (val: string) => {
	if (val.length < 8) return t('errors.at_least_8_digits_long');
	let rule = /^(?=.*[A-Z])(?=.*[a-z])[A-Za-z0-9@$!%*?&_.]{8,}$/;
	if (!rule.test(val)) {
		return t('errors.at_least_one_uppercase_and_lowercase_letter');
	}
	return '';
};
const repasswordRule = (val: string) => {
	if (newPassword.value.length === 0) {
		return passwordRule(val);
	}
	if (newPassword.value !== val) {
		return t('errors.repassword_do_not_match_password');
	}
	return '';
};

const updatePassword = async () => {
	if (
		oldPassword.value.length === 0 ||
		newPassword.value.length === 0 ||
		repeatPassword.value.length === 0
	) {
		return;
	}
	if (newPassword.value !== repeatPassword.value) {
		return;
	}
	CustomRef.value.onDialogCancel();

	quasar
		.dialog({
			component: ReminderDialogComponent,
			componentProps: {
				title: t('change_password'),
				message: t(
					'Are you sure you want to change the password of the current account?'
				),
				useCancel: true,
				confirmText: t('confirm'),
				cancelText: t('cancel')
			}
		})
		.onOk(async () => {
			updatePasswordConfirm();
		});
};

const updatePasswordConfirm = async () => {
	try {
		await accountStore.reset_account_password({
			password: passwordAddSort(
				newPassword.value,
				adminStore.terminus.osVersion
			),
			current_password: passwordAddSort(
				oldPassword.value,
				adminStore.terminus.osVersion
			),
			username: props.name
		});
		notifySuccess(t('password_update_success'));
		onOKClick();
	} catch (error: any) {
		/* empty */
	}
};
const deviceStore = useDeviceStore();
const CustomRef = ref();

function onOKClick() {
	CustomRef.value.onDialogOK();
}
</script>
