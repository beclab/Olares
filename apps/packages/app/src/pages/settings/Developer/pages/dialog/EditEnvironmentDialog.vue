<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Edit environment variable')"
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
			v-if="envRef"
			:label="t('Value')"
			:env="envRef"
			:enable-env-ref-picker="enableEnvRefPicker"
			@update:env="updateNewEnv"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import EnvVarInput from 'src/components/EnvVarInput.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { ref, onMounted, computed, PropType } from 'vue';
import { BaseEnv } from 'src/constant';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	data: {
		type: Object as PropType<BaseEnv>,
		required: true
	},
	enableEnvRefPicker: {
		type: Boolean,
		default: true
	}
});

const deviceStore = useDeviceStore();
const { t } = useI18n();
const CustomRef = ref();
const key = ref('');
const envRef = ref();

onMounted(() => {
	if (props.data) {
		key.value = props.data.envName || '';
		envRef.value = { ...props.data };
	}
});

const updateNewEnv = (newEnv: any) => {
	envRef.value = newEnv;
};

function resolveDisplayValue(env: BaseEnv | undefined) {
	if (!env) {
		return '';
	}
	const v = env.value;
	if (!v || String(v).trim() === '') {
		return !!env.default ? String(env.default).trim() : '';
	}
	return String(v).trim();
}

function valueFromSignature(vf: BaseEnv['valueFrom'] | undefined) {
	if (!vf) {
		return '';
	}
	return `${vf.envName}\0${vf.status || ''}`;
}

const isUpdateEnable = computed(() => {
	if (!props.data || !envRef.value) {
		return false;
	}

	const cur = envRef.value;
	const orig = props.data;
	const inputValue = resolveDisplayValue(cur);

	if (orig.required && !inputValue) {
		return false;
	}

	const unchanged =
		resolveDisplayValue(cur) === resolveDisplayValue(orig) &&
		valueFromSignature(cur.valueFrom) === valueFromSignature(orig.valueFrom) &&
		(cur.applyOnChange ?? false) === (orig.applyOnChange ?? false);

	if (unchanged) {
		return false;
	}

	if (cur.right === false) {
		return false;
	}

	const { regex } = cur;
	if (regex && typeof regex === 'string' && regex.trim()) {
		try {
			const reg = new RegExp(regex);
			if (!reg.test(inputValue)) {
				return false;
			}
		} catch (e) {
			console.error('Invalid regex pattern:', regex, e);
			return false;
		}
	}

	return true;
});

const updateEnv = async () => {
	if (CustomRef.value && envRef.value) {
		CustomRef.value.onDialogOK({
			key: key.value,
			value: envRef.value.value,
			applyOnChange: envRef.value.applyOnChange,
			valueFrom: envRef.value.valueFrom ?? null
		});
	}
};
</script>

<style scoped lang="scss"></style>
