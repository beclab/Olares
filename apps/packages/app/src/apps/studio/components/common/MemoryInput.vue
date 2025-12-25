<template>
	<q-input
		ref="inputRef"
		dense
		borderless
		no-error-icon
		v-model.trim="memoryValue"
		class="form-item-input"
		input-class="text-ink-2"
		:rules="memoryRules"
		:placeholder="placeholder"
	>
		<template v-slot:append>
			<q-select
				dense
				borderless
				v-model="memoryUnit"
				:options="memoryUnitOptions"
				dropdown-icon="sym_r_keyboard_arrow_down"
				@update:model-value="handleUnitChange"
			/>
		</template>
	</q-input>
</template>

<script lang="ts" setup>
import { ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { maxMemory } from '@apps/studio/src/types/constants';

interface Props {
	modelValue?: string;
	placeholder?: string;
	minMemory?: number; // Mi
}

const props = withDefaults(defineProps<Props>(), {
	modelValue: '',
	placeholder: '',
	minMemory: 128
});

const emit = defineEmits<{
	(e: 'update:modelValue', value: string): void;
}>();

const { t } = useI18n();
const inputRef = ref();
const memoryValue = ref('');
const memoryUnit = ref<string>('Mi');
const memoryUnitOptions = ['Mi', 'Gi'];

// 解析初始值
const parseInitialValue = (value: string) => {
	if (!value) return;

	const trimmedValue = value.trim();
	if (trimmedValue.endsWith('Gi')) {
		memoryValue.value = trimmedValue.slice(0, -2);
		memoryUnit.value = 'Gi';
	} else if (trimmedValue.endsWith('Mi')) {
		memoryValue.value = trimmedValue.slice(0, -2);
		memoryUnit.value = 'Mi';
	} else if (trimmedValue.match(/^\d+(\.\d+)?$/)) {
		// 纯数字，默认为 Mi
		memoryValue.value = trimmedValue;
		memoryUnit.value = 'Mi';
	}
};

parseInitialValue(props.modelValue);

const memoryRules = computed(() => [
	(val: string) => (val && val.length > 0) || t('memory_rule'),
	(val: string) => {
		if (!val) return true;

		const numValue = parseFloat(val);
		if (isNaN(numValue)) {
			return t('memory_rule_invalid') || 'Memory 值无效';
		}

		if (numValue <= 0) {
			return t('memory_rule_not_zero') || 'Memory 不能为 0';
		}

		let memoryInMi = 0;
		if (memoryUnit.value === 'Mi') {
			memoryInMi = numValue;
		} else {
			memoryInMi = numValue * 1024;
		}

		if (memoryInMi < props.minMemory) {
			return (
				t('memory_rule_min', { min: props.minMemory }) ||
				`Memory 不能低于 ${props.minMemory}Mi`
			);
		}

		if (memoryInMi > maxMemory) {
			return (
				t('memory_rule_max', { max: maxMemory / 1024 }) ||
				`Memory 不能超过 ${maxMemory / 1024} Gi`
			);
		}

		return true;
	}
]);

const emitValue = () => {
	if (!memoryValue.value) {
		emit('update:modelValue', '');
		return;
	}

	const fullValue = `${memoryValue.value}${memoryUnit.value}`;
	emit('update:modelValue', fullValue);
};

watch(memoryValue, () => {
	emitValue();
});

const handleUnitChange = () => {
	emitValue();
	if (inputRef.value) {
		inputRef.value.validate();
	}
};

watch(
	() => props.modelValue,
	(newVal) => {
		const currentValue = `${memoryValue.value}${memoryUnit.value}`;
		if (newVal !== currentValue) {
			parseInitialValue(newVal);
		}
	}
);

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
