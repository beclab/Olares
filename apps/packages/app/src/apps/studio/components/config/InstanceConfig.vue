<template>
	<q-card class="instance-container" flat>
		<q-card-section class="text-h6 text-ink-1">
			{{ t('docker.instance_specifications') }}
		</q-card-section>

		<q-card-section class="q-py-none">
			<card-form-item name="CPU" :required="true" tip="CPU">
				<cpu-input
					ref="cpuRef"
					v-model="instanceConfig.requiredCpu"
					:placeholder="ruleConfig.cpu.placeholder"
				/>
			</card-form-item>

			<card-form-item
				:name="t('docker.memory')"
				:required="true"
				:tip="t('docker.memory')"
			>
				<memory-input
					ref="memoryRef"
					v-model="instanceConfig.requiredMemory"
					:placeholder="ruleConfig.memory.placeholder"
				/>
			</card-form-item>

			<card-form-item name="GPU" :required="false" tip="GPU">
				<q-toggle color="teal-6" v-model="instanceConfig.requiredGpu" />
			</card-form-item>

			<card-form-item
				v-if="instanceConfig.requiredGpu"
				:name="t('docker.manufacturer')"
				:required="true"
				:tip="t('docker.manufacturer')"
			>
				<div class="q-pa-md q-gutter-sm">
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="instanceConfig.gpuVendor"
						:val="VENDOR.NVIDIA"
						label="NVIDIA"
						color="teal-default"
					/>
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="instanceConfig.gpuVendor"
						:val="VENDOR.AMD"
						label="AMD"
						color="teal-default"
					/>
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="instanceConfig.gpuVendor"
						:val="VENDOR.INTEL"
						label="Intel"
						color="teal-default"
					/>
				</div>
			</card-form-item>

			<card-form-item name="Postgres" :required="false" tip="Postgres">
				<q-toggle color="teal-6" v-model="instanceConfig.needPg" />
			</card-form-item>

			<card-form-item name="Redis" :required="false" tip="Redis">
				<q-toggle color="teal-6" v-model="instanceConfig.needRedis" />
			</card-form-item>
		</q-card-section>
	</q-card>
</template>

<script lang="ts" setup>
import { reactive, watch, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { VENDOR } from '@apps/studio/src/types/core';
import { ruleConfig } from '@apps/studio/src/types/config';

import CardFormItem from './../common/CardFormItem.vue';
import CpuInput from './../common/CpuInput.vue';
import MemoryInput from './../common/MemoryInput.vue';

interface Props {
	defaultValues?: {
		requiredCpu?: string;
		requiredMemory?: string;
	};
}

const props = withDefaults(defineProps<Props>(), {
	defaultValues: () => ({})
});

const emits = defineEmits(['updateInstance']);

const { t } = useI18n();
const memoryRef = ref();
const cpuRef = ref();

const instanceConfig = reactive({
	requiredCpu: '',
	requiredMemory: '',
	requiredGpu: false,
	needPg: false,
	needRedis: false,
	gpuVendor: VENDOR.NVIDIA
});

onMounted(() => {
	if (props.defaultValues?.requiredCpu) {
		instanceConfig.requiredCpu = props.defaultValues.requiredCpu;
	}
	if (props.defaultValues?.requiredMemory) {
		instanceConfig.requiredMemory = props.defaultValues.requiredMemory;
	}
});

const updateInstance = () => {
	emits('updateInstance', instanceConfig);
};

watch(
	() => instanceConfig,
	() => {
		updateInstance();
	},
	{
		deep: true
	}
);

const validate = () => {
	memoryRef.value.validate();
	cpuRef.value.validate();

	if (memoryRef.value.hasError || cpuRef.value.hasError) {
		return false;
	}

	return true;
};

defineExpose({
	validate
});
</script>

<style lang="scss" scoped>
.instance-container {
	margin: 20px 20px 0 20px;
	padding: 4px;
	border-radius: 12px;
	background-color: $background-1;
}
::v-deep(.q-toggle__track) {
	height: 0.6em !important;
	width: 1.1em !important;
	border-radius: 0.3em;
	position: absolute;
	top: 0.21em;
	left: 0.16em;
	background: #dbdbdb;
	opacity: 1;
}

::v-deep(.q-toggle__thumb::before) {
	background: none !important;
}
::v-deep(.q-toggle__thumb:after) {
	box-shadow: none;
}
</style>
