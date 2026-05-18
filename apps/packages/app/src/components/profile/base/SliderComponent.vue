<template>
	<div class="slider-root column full-width justify-start q-mt-lg">
		<div class="text-subtitle1 text-ink-1">{{ label }}</div>

		<div class="full-width row justify-start items-center q-mt-xs">
			<q-slider
				:model-value="modelValue"
				@update:model-value="onSliderUpdate"
				thumb-path="M 15,10
A 5,5 0 1,0 5,10
A 5,5 0 1,0 15,10"
				:min="min"
				:max="max"
				thumb-color="light-green-default"
				track-size="2px"
				class="slider-scroll"
				track-color="background-4"
				color="light-green-default"
			/>

			<q-input
				borderless
				input-class="text-body1 text-ink-1"
				class="slider-edit q-ml-lg"
				:model-value="inputFocused ? inputDraft : displayModel"
				@focus="onInputFocus"
				@blur="onInputBlur"
				@keyup.enter.prevent="onInputEnter"
				@update:model-value="onInputDraftUpdate"
			>
				<template v-slot:append>
					<div
						class="text-body1 text-ink-1 full-height"
						style="text-align: center"
					>
						{{ unit }}
					</div>
				</template>
			</q-input>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue';

const props = defineProps({
	label: {
		type: String,
		require: true
	},
	modelValue: {
		type: Number,
		require: true
	},
	min: {
		type: Number,
		default: 0
	},
	max: {
		type: Number,
		default: 100
	},
	unit: {
		type: String,
		default: '%'
	}
});

const emit = defineEmits(['update:modelValue']);

const inputFocused = ref(false);
const inputDraft = ref('');

const displayModel = computed(() => String(props.modelValue));

function clampInt(n: number): number {
	const rounded = Math.round(n);
	return Math.min(props.max, Math.max(props.min, rounded));
}

function onSliderUpdate(val: number | string | null) {
	const num = typeof val === 'number' ? val : Number(val);
	if (Number.isNaN(num)) {
		return;
	}
	const next = clampInt(num);
	emit('update:modelValue', next);
}

function onInputFocus() {
	inputFocused.value = true;
	inputDraft.value = displayModel.value;
}

function onInputDraftUpdate(val: string | number | null) {
	if (!inputFocused.value) {
		return;
	}
	inputDraft.value = String(val ?? '');
}

function commitInput(fromBlur: boolean) {
	if (!inputFocused.value) {
		return;
	}
	const n = parseFloat(inputDraft.value.trim());
	if (Number.isNaN(n)) {
		inputDraft.value = displayModel.value;
		if (fromBlur) {
			inputFocused.value = false;
		}
		return;
	}
	emit('update:modelValue', clampInt(n));
	if (fromBlur) {
		inputFocused.value = false;
	} else {
		nextTick(() => {
			inputDraft.value = displayModel.value;
		});
	}
}

function onInputBlur() {
	commitInput(true);
}

function onInputEnter() {
	commitInput(false);
}

watch(
	() => props.modelValue,
	() => {
		if (!inputFocused.value) {
			inputDraft.value = displayModel.value;
		}
	}
);
</script>

<style scoped lang="scss">
.slider-root {
	height: auto;

	.slider-scroll {
		width: calc(100% - 100px);
	}

	.slider-edit {
		padding: 2px 12px;
		height: 40px;
		border-radius: 8px;
		border: 1px solid $input-stroke;
		width: 80px;
	}
}
</style>
