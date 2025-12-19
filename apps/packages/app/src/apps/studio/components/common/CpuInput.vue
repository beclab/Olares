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

// 解析初始值
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

// 初始化
parseInitialValue(props.modelValue);

// CPU 验证规则
const cpuRules = computed(() => [
	(val: string) => (val && val.length > 0) || t('cpu_rule'),
	(val: string) => {
		if (!val) return true;

		const numValue = parseFloat(val);
		if (isNaN(numValue)) {
			return t('cpu_rule_invalid') || 'CPU 值无效';
		}

		// 检查是否为 0 或负数
		if (numValue <= 0) {
			return t('cpu_rule_not_zero') || 'CPU 不能为 0';
		}

		// 根据单位转换为 millicores
		let cpuInMillicores = 0;
		if (cpuUnit.value === 'm') {
			cpuInMillicores = numValue;
		} else {
			// 'core' 单位
			cpuInMillicores = numValue * 1000;
		}

		// 检查是否超过最大值
		if (cpuInMillicores > maxCpu) {
			return (
				t('cpu_rule_max', { max: maxCpu / 1000 }) ||
				`CPU 不能超过 ${maxCpu / 1000} cores`
			);
		}

		return true;
	}
]);

// 组合值并发出更新
const emitValue = () => {
	if (!cpuValue.value) {
		emit('update:modelValue', '');
		return;
	}

	const fullValue =
		cpuUnit.value === 'm' ? `${cpuValue.value}m` : `${cpuValue.value}`;
	emit('update:modelValue', fullValue);
};

// 监听数值变化
watch(cpuValue, () => {
	emitValue();
});

// 处理单位变化
const handleUnitChange = () => {
	emitValue();
	// 触发验证
	if (inputRef.value) {
		inputRef.value.validate();
	}
};

// 监听外部 modelValue 变化
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

// 暴露验证方法
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
