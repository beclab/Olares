<template>
	<div class="opacity-input-background row">
		<q-input
			:disable="inputDisabled"
			class="opacity-input"
			borderless
			@focus="focusChange(true)"
			@blur="focusChange(false)"
			v-model="percentageInput"
			@update:model-value="onInputChange"
			:rules="[validatePercentage]"
		/>
		<div
			style="position: absolute; right: 0; top: 0"
			class="column justify-between items-center"
		>
			<div class="icon-parent" @click="increasePercentage" title="increase">
				<q-icon size="8px" name="sym_r_expand_less" />
			</div>
			<div class="icon-parent" @click="decreasePercentage" title="decrease">
				<q-icon
					size="8px"
					style="position: absolute; top: 0"
					name="sym_r_expand_more"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref } from 'vue';
import { colorsRgba } from 'quasar/dist/types/utils/colors';
import hexToRgb = colors.hexToRgb;
import { colors } from 'quasar';

const props = defineProps({
	hexColor: String,
	modelValue: Number,
	inputDisabled: Boolean
});
const emit = defineEmits(['onUpdate', 'update:modelValue']);

const percentageSign = ref(true);

const focusChange = (focus: boolean) => {
	percentageSign.value = !focus;
};

const percentageInput = computed({
	get: () => `${props.modelValue}${percentageSign.value ? '%' : ''}`,
	set: (value) => {
		const parsedValue = parseFloat(value);
		console.log(parsedValue);
		emit(
			'update:modelValue',
			isNaN(parsedValue) ? 0 : Math.min(Math.max(parsedValue, 0), 100)
		);
	}
});

const onInputChange = (value: string) => {
	console.log(value);
	// 处理直接修改百分比输入
	const valueWithoutPercentageSign = value.replace('%', '');
	emit('onUpdate', valueWithoutPercentageSign);
};

const increasePercentage = () => {
	// 处理上调
	if (props.modelValue < 100) {
		const value = (props.modelValue + 1).toFixed(0);
		percentageInput.value = value;
		emit('onUpdate', value);
	}
};

const decreasePercentage = () => {
	// 处理下调
	if (props.modelValue > 0) {
		const value = (props.modelValue - 1).toFixed(0);
		percentageInput.value = value;
		emit('onUpdate', value);
	}
};

const validatePercentage = (value) => {
	// 验证规则
	let isValid;
	if (percentageSign.value) {
		isValid = /^(?:100|[1-9]?\d|0)%$/.test(value);
	} else {
		isValid = /^(?:100|[1-9]?\d|0)$/.test(value);
	}
	return isValid || 'Please enter a valid percentage';
};

onMounted(() => {
	setHexColor(props.hexColor);
});

const setHexColor = (hexColor: string) => {
	console.log(hexColor);
	if (hexColor && (hexColor.length === 7 || hexColor.length === 9)) {
		//#00000000 表示黑色完全透明， rgba为(0,0,0,0) rgba.a = 0  不透明度为 0%
		//#ffffffff 表示白色完全不透明， rgba为(255,255,255,0) rgba.a = 100  不透明度为 100%
		const rgba: colorsRgba = hexToRgb(hexColor);
		let a;
		if (hexColor.length === 7) {
			a = 100;
		} else {
			a = rgba.a;
		}
		console.log(a);
		percentageInput.value = `${a.toFixed(0)}%`;
	}
};

defineExpose({ setHexColor });
</script>

<style lang="scss">
.opacity-input-background {
	width: 164px;
	height: 32px;
	border-radius: 8px;
	border: 1px solid $input-stroke;
	position: relative;

	.opacity-input {
		width: calc(100% - 20px);
		margin-top: -13px;
		padding-left: 10px;
		padding-right: 10px;
		position: absolute;
	}

	.icon-parent {
		width: 20px;
		height: 16px;
		position: relative;
		cursor: pointer;
		text-decoration: none;
	}
}
</style>
