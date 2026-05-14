<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Add account')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateEnable"
		@onSubmit="addEnv"
	>
		<terminus-edit
			v-model="userName"
			:label="t('Username')"
			:show-password-img="false"
			style="width: 100%"
		/>

		<terminus-edit
			v-model="password"
			:label="t('Password')"
			class="q-mt-lg"
			:show-password-img="true"
			style="width: 100%"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDeviceStore } from 'src/stores/settings/device';
import { addSMBAccount } from 'src/api/settings/smb';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

const { t } = useI18n();
const CustomRef = ref();
const userName = ref('');
const password = ref('');
const deviceStore = useDeviceStore();

const isUpdateEnable = computed(() => {
	if (!userName.value || !password.value) {
		return false;
	}
	return true;
});

const addEnv = async () => {
	if (CustomRef.value) {
		addSMBAccount(userName.value, password.value)
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('success')
				});
				CustomRef.value.onDialogOK();
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: e.response.data.message || e.message
				});
			});
	}
};
</script>

<style scoped lang="scss"></style>
