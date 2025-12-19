<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Configure Environment Variables')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:okDisabled="!isSetEnable"
		@onSubmit="submitEnvironmentVariable"
	>
		<env-var-input
			v-for="item in missingValues"
			:key="item.envName"
			:env="item"
			@update:env="updateMissingValues"
		/>

		<env-var-input
			v-for="item in invalidValues"
			:key="item.envName"
			:first-error="true"
			:env="item"
			@update:env="updateInvalidValues"
		/>

		<env-var-input
			v-for="item in missingRefs"
			:key="item.envName"
			:env="item"
			@update:env="updateMissingRefs"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import EnvVarInput from 'src/components/EnvVarInput.vue';
import { AppEnvResponse, BaseEnv } from 'src/constant';
import { ref, onMounted, computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	data: {
		type: Object as PropType<AppEnvResponse>,
		required: true
	}
});

const { t } = useI18n();
const CustomRef = ref();
const missingValues = ref<BaseEnv[]>([]);
const invalidValues = ref<BaseEnv[]>([]);
const missingRefs = ref<BaseEnv[]>([]);

onMounted(() => {
	if (props.data.missingValues && props.data.missingValues.length > 0) {
		missingValues.value.push(...props.data.missingValues);
	}
	if (props.data.invalidValues && props.data.invalidValues.length > 0) {
		invalidValues.value.push(...props.data.invalidValues);
	}
	if (props.data.missingRefs && props.data?.missingRefs.length > 0) {
		missingRefs.value.push(...props.data.missingRefs);
	}
});

const updateMissingValues = (newEnv: BaseEnv) => {
	const index = missingValues.value.findIndex(
		(item) => item.envName === newEnv.envName
	);
	if (index !== -1) {
		missingValues.value[index] = { ...newEnv };
	}
};

const updateInvalidValues = (newEnv: BaseEnv) => {
	const index = invalidValues.value.findIndex(
		(item) => item.envName === newEnv.envName
	);
	if (index !== -1) {
		invalidValues.value[index] = { ...newEnv };
	}
};

const updateMissingRefs = (newEnv: BaseEnv) => {
	const index = missingRefs.value.findIndex(
		(item) => item.envName === newEnv.envName
	);
	if (index !== -1) {
		missingRefs.value[index] = { ...newEnv };
	}
};

const isSetEnable = computed(() => {
	const list = missingValues.value
		.concat(invalidValues.value)
		.concat(missingRefs.value);
	for (let i = 0; i < list.length; i++) {
		if (!list[i].value) {
			return false;
		}
		if (list[i].regex && !new RegExp(list[i].regex).test(list[i].value)) {
			return false;
		}
	}
	return true;
});

const submitEnvironmentVariable = async () => {
	const list = missingValues.value
		.concat(invalidValues.value)
		.concat(missingRefs.value);
	const updateEnvs = list.map((item) => {
		return {
			envName: item.envName,
			value: item.value
		};
	});
	console.log(updateEnvs);
	CustomRef.value.onDialogOK({
		envs: updateEnvs
	});
};
</script>

<style scoped lang="scss"></style>
