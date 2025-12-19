<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Edit Environment Variable')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateEnable"
		@onSubmit="updateEnv"
	>
		<terminus-edit
			v-model="key"
			:label="t('Key')"
			is-read-only
			:show-password-img="false"
			style="width: 100%"
		/>

		<env-var-input
			:label="t('Value')"
			:env="envRef"
			@update:env="updateNewEnv"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import EnvVarInput from 'src/components/EnvVarInput.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { ref, onMounted, computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { BaseEnv } from 'src/constant';

const props = defineProps({
	data: {
		type: Object as PropType<BaseEnv>,
		required: true
	}
});

const deviceStore = useDeviceStore();
const { t } = useI18n();
const CustomRef = ref();
const key = ref('');
const envRef = ref();

onMounted(() => {
	key.value = props.data?.envName;
	envRef.value = props.data;
});

const updateNewEnv = (newEnv) => {
	envRef.value = newEnv;
};

const isUpdateEnable = computed(() => {
	if (
		(props.data?.required && !envRef.value.value) ||
		(key.value === props.data.envName &&
			envRef.value.value === props.data.value)
	) {
		return false;
	}
	if (
		envRef.value &&
		envRef.value.regex &&
		!new RegExp(envRef.value.regex).test(envRef.value.value)
	) {
		return false;
	}
	return true;
});

const updateEnv = async () => {
	CustomRef.value.onDialogOK({
		key: key.value,
		value: envRef.value.value
	});
};
</script>

<style scoped lang="scss"></style>
