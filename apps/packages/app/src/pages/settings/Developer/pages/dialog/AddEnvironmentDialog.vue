<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Add environment variable')"
		:skip="false"
		:ok="t('Save')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateEnable"
		@onSubmit="addEnv"
	>
		<terminus-edit
			prefix="OLARES_USER_"
			v-model="envKey"
			:label="t('Key')"
			:show-password-img="false"
			style="width: 100%"
		/>

		<terminus-edit
			v-model="envValue"
			:label="t('Value')"
			class="q-mt-lg"
			:show-password-img="false"
			style="width: 100%"
		/>

		<div class="text-body3 text-ink-3 q-mt-lg q-mb-xs">{{ t('Type') }}</div>

		<bt-select-v3 v-model="envType" :options="typeOptions" />

		<terminus-edit
			v-model="envDesc"
			:label="t('Description')"
			class="q-mt-lg"
			:show-password-img="false"
			style="width: 100%"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import BtSelectV3 from '../../../../../components/settings/base/BtSelectV3.vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDeviceStore } from 'src/stores/settings/device';
import { addUserEnv } from 'src/api/settings/env';
import { SelectorProps } from 'src/constant';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

const { t } = useI18n();
const CustomRef = ref();
const envKey = ref('');
const envValue = ref('');
const envType = ref('string');
const envDesc = ref('');
const deviceStore = useDeviceStore();
//int/bool/url/ip/domain/email/string/password
const typeOptions = ref<SelectorProps[]>([
	{
		label: 'string',
		value: 'string'
	},
	{
		label: 'boolean',
		value: 'bool'
	},
	{
		label: 'int',
		value: 'int'
	},
	{
		label: 'url',
		value: 'url'
	},
	{
		label: 'ip',
		value: 'ip'
	},
	{
		label: 'domain',
		value: 'domain'
	},
	{
		label: 'email',
		value: 'email'
	},
	{
		label: 'password',
		value: 'password'
	}
]);

const isUpdateEnable = computed(() => {
	if (!envKey.value || !envValue.value) {
		return false;
	}
	return true;
});

const addEnv = async () => {
	if (CustomRef.value) {
		addUserEnv({
			envName: 'OLARES_USER_' + envKey.value,
			value: envValue.value,
			type: envType.value,
			description: envDesc.value,
			required: false,
			editable: true
		})
			.then((env) => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('success')
				});
				console.log(env);
				CustomRef.value.onDialogOK(env);
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
