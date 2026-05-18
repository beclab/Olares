<template>
	<q-input
		ref="inputRef"
		dense
		borderless
		no-error-icon
		v-model.trim="cpuValue"
		class="form-item-input"
		input-class="text-ink-2"
		:rules="cpuRules"
		:placeholder="placeholder"
	>
		<template v-slot:append>
			<q-select
				dense
				borderless
				v-model="cpuUnit"
				:options="cpuUnitOptions"
				dropdown-icon="sym_r_keyboard_arrow_down"
				@update:model-value="handleUnitChange"
			/>
		</template>
	</q-input>
</template>

<script lang="ts" setup>
import { ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { cpuUnitOptions, maxCpu } from '@apps/studio/src/types/constants';

interface Props {
	modelValue?: string;
	placeholder?: string;
}

const props = withDefaults(defineProps<Props>(), {
	modelValue: '',
	placeholder: ''
});

const emit = defineEmits<{
	(e: 'update:modelValue', value: string): void;
}>();

const { t } = useI18n();
const inputRef = ref();
const cpuValue = ref('');
const cpuUnit = ref<string>(cpuUnitOptions[0]);

// Parse initial value.
const parseInitialValue = (value: string) => {
	if (!value) return;

	const trimmedValue = value.trim();
	if (trimmedValue.endsWith('m')) {
		cpuValue.value = trimmedValue.slice(0, -1);
		cpuUnit.value = 'm';
	} else if (trimmedValue.endsWith('core')) {
		cpuValue.value = trimmedValue.slice(0, -4);
		cpuUnit.value = 'core';
	} else if (trimmedValue.match(/^\d+(\.\d+)?$/)) {
		cpuValue.value = trimmedValue;
		cpuUnit.value = 'core';
	}
};

// Initialize.
parseInitialValue(props.modelValue);

// CPU validation rules.
const cpuRules = computed(() => [
	(val: string) => (val && val.length > 0) || t('cpu_rule'),
	(val: string) => {
		if (!val) return true;

		const numValue = parseFloat(val);
		if (isNaN(numValue)) {
			return t('cpu_rule_invalid') || 'CPU 值无效';
		}

		// Validate against zero or negative values.
		if (numValue <= 0) {
			return t('cpu_rule_not_zero') || 'CPU 不能为 0';
		}

		// Convert to millicores based on unit.
		let cpuInMillicores = 0;
		if (cpuUnit.value === 'm') {
			cpuInMillicores = numValue;
		} else {
			// 'core' unit
			cpuInMillicores = numValue * 1000;
		}

		// Check upper bound.
		if (cpuInMillicores > maxCpu) {
			return (
				t('cpu_rule_max', { max: maxCpu / 1000 }) ||
				`CPU 不能超过 ${maxCpu / 1000} cores`
			);
		}

		return true;
	}
]);

// Compose value and emit update.
const emitValue = () => {
	if (!cpuValue.value) {
		emit('update:modelValue', '');
		return;
	}

	const fullValue =
		cpuUnit.value === 'm' ? `${cpuValue.value}m` : `${cpuValue.value}`;
	emit('update:modelValue', fullValue);
};

// Watch numeric value changes.
watch(cpuValue, () => {
	emitValue();
});

// Handle unit changes.
const handleUnitChange = () => {
	emitValue();
	// Trigger validation.
	if (inputRef.value) {
		inputRef.value.validate();
	}
};

// Watch external modelValue changes.
watch(
	() => props.modelValue,
	(newVal) => {
		const currentValue =
			cpuUnit.value === 'm' ? `${cpuValue.value}m` : `${cpuValue.value}`;
		if (newVal !== currentValue) {
			parseInitialValue(newVal);
		}
	}
);

// Expose validation method.
const validate = () => {
	return inputRef.value?.validate();
};

const hasError = computed(() => {
	return inputRef.value?.hasError ?? false;
});

defineExpose({
	validate,
	hasError
});
</script>
