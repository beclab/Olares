<template>
	<q-card class="instance-container" flat>
		<q-card-section class="text-h6 text-ink-1">
			{{ t('image_create') }}
		</q-card-section>

		<q-card-section class="q-py-none">
			<card-form-item
				:name="t('containers_dev_env')"
				:required="true"
				tip="CPU"
			>
				<div class="env-container">
					<q-input
						ref="envRef"
						dense
						borderless
						no-error-icon
						v-model.trim="imageConfig.devEnv"
						class="env-input"
						input-class="env-input-in text-ink-2"
						:rules="ruleConfig.env.rules"
						:placeholder="ruleConfig.env.placeholder"
						:input-style="{ textIndent: '10px' }"
					>
					</q-input>

					<q-select
						dense
						options-dense
						emit-value
						map-options
						borderless
						no-error-icon
						:options="envOptions"
						input-debounce="0"
						hide-dropdown-icon
						class="env-select"
						input-class="env-select-in"
						color="ink-3"
						popup-content-class="options_selected_content"
						@update:model-value="updateEnv"
					>
						<template v-slot:append>
							<div class="select-down">
								<q-icon name="sym_r_keyboard_arrow_down" />
							</div>
						</template>
						<template v-slot:option="{ itemProps, opt }">
							<q-item dense v-bind="itemProps" class="select-item">
								<q-item-section class="select-section text-ink-2">
									<q-item-label>{{ opt.label }}</q-item-label>
								</q-item-section>
							</q-item>
						</template>
					</q-select>
				</div>
			</card-form-item>

			<card-form-item name="CPU" :required="true" tip="CPU">
				<cpu-input
					ref="cpuRef"
					v-model="imageConfig.requiredCpu"
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
					v-model="imageConfig.requiredMemory"
					:placeholder="ruleConfig.memory.placeholder"
					:min-memory="minMemory"
				/>
			</card-form-item>

			<card-form-item
				:name="t('docker.volume_size')"
				:required="true"
				:tip="t('docker.volume_size')"
			>
				<q-input
					ref="volumeRef"
					dense
					borderless
					no-error-icon
					v-model.trim="imageConfig.requiredDisk"
					class="form-item-input"
					input-class="text-ink-2"
					:rules="ruleConfig.volume.rules"
					:placeholder="ruleConfig.volume.placeholder"
				>
					<template v-slot:append>
						<q-select
							dense
							borderless
							v-model="requiredDiskUnit"
							:options="diskOptions"
							dropdown-icon="sym_r_keyboard_arrow_down"
							style="width: 50px"
							@update:model-value="updateDiskUnit"
						/>
					</template>
				</q-input>
			</card-form-item>

			<card-form-item
				:name="t('docker.expose_ports')"
				:required="false"
				:tip="t('docker.expose_ports')"
			>
				<q-input
					ref="portsRef"
					dense
					borderless
					no-error-icon
					v-model.trim="imageConfig.ports"
					class="form-item-input"
					input-class="text-ink-2"
					:rules="ruleConfig.ports.rules"
					:placeholder="ruleConfig.ports.placeholder"
				>
				</q-input>
			</card-form-item>

			<card-form-item name="GPU" :required="false" tip="GPU">
				<q-toggle color="teal-6" v-model="imageConfig.requiredGpu" />
			</card-form-item>

			<card-form-item
				v-if="imageConfig.requiredGpu"
				:name="t('docker.manufacturer')"
				:required="true"
				:tip="t('docker.manufacturer')"
			>
				<div class="q-pa-md q-gutter-sm">
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="imageConfig.gpuVendor"
						:val="VENDOR.NVIDIA"
						label="NVIDIA"
						color="teal-default"
					/>
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="imageConfig.gpuVendor"
						:val="VENDOR.AMD"
						label="AMD"
						color="teal-default"
					/>
					<q-radio
						class="q-mr-lg text-ink-1 text-body1"
						dense
						v-model="imageConfig.gpuVendor"
						:val="VENDOR.INTEL"
						label="Intel"
						color="teal-default"
					/>
				</div>
			</card-form-item>
		</q-card-section>
	</q-card>
</template>

<script lang="ts" setup>
import { reactive, watch, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ruleConfig } from '@apps/studio/src/types/config';
import { envOptions, diskOptions } from '@apps/studio/src/types/constants';

import CardFormItem from './../common/CardFormItem.vue';
import CpuInput from './../common/CpuInput.vue';
import MemoryInput from './../common/MemoryInput.vue';
import { VENDOR } from '../../types/core';

interface Props {
	minMemory?: number;
}

const props = withDefaults(defineProps<Props>(), {
	minMemory: 128
});

const emits = defineEmits(['updateImage']);

const { t } = useI18n();
const envRef = ref();
const memoryRef = ref();
const cpuRef = ref();
const volumeRef = ref();
const portsRef = ref();

const requiredDiskUnit = ref('Mi');
const imageConfig = reactive({
	devEnv: '',
	requiredCpu: '',
	requiredGpu: false,
	gpuVendor: VENDOR.NVIDIA,
	requiredMemory: '',
	requiredDisk: '',
	ports: ''
});

const updateEnv = (value) => {
	imageConfig.devEnv = value;
};

const updateImage = () => {
	const data = {
		...imageConfig,
		requiredDisk: `${imageConfig.requiredDisk}${requiredDiskUnit.value}`
	};
	console.log('create-container', data);
	emits('updateImage', data);
};

const updateDiskUnit = () => {
	updateImage();
};

watch(
	() => imageConfig,
	() => {
		updateImage();
	},
	{
		deep: true
	}
);

const validate = () => {
	envRef.value.validate();
	memoryRef.value.validate();
	cpuRef.value.validate();
	volumeRef.value.validate();
	portsRef.value.validate();

	if (
		envRef.value.hasError ||
		volumeRef.value.hasError ||
		memoryRef.value.hasError ||
		cpuRef.value.hasError ||
		portsRef.value.hasError
	) {
		return false;
	}

	return true;
};

defineExpose({
	validate
});
</script>

<style lang="scss" scoped>
.env-container {
	width: 100%;
	height: 42px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	padding: 0 10px;
	position: relative;
	.env-input {
		width: calc(100% - 40px);
		position: absolute;
		top: 0;
		left: 0;
		z-index: 2;
	}

	.env-select {
		width: 100%;
		position: absolute;
		top: 0;
		left: 0;
		z-index: 1;

		.select-down {
			width: 40px;
			height: 100%;
			background-color: $background-6;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 0 8px 8px 0;
			overflow: hidden;
		}

		::v-deep(.q-field__append) {
			padding: 0 !important;
		}
	}
}

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
