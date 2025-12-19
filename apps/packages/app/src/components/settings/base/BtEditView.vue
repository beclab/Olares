<template>
	<div class="input-item" :style="{ '--height': height }">
		<div class="input-parent row justify-center items-center">
			<slot name="icon" />
			<q-input
				:type="inputType"
				class="text-body2 inner-input"
				standout
				:style="{
					'--width': slotIcon ? 'calc(100% - 32px)' : '100%',
					'--paddingLeft': slotIcon ? '0' : '12px'
				}"
				:input-class="`${height ? 'text-resize' : ''} ${
					right ? 'text-right' : ''
				} custom-placeholder`"
				borderless
				dense
				name="search"
				debounce="500"
				reverse-fill-mask
				:placeholder="placeholder"
				autocomplete="off"
				@focus="onFocus"
				@blur="onBlur"
				:readonly="isReadOnly"
				:disable="isReadOnly"
				@keyup.enter="onInput(modelValue)"
				:model-value="modelValue"
				@update:model-value="onInput"
				placeholder-color="red"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { inject, ref, useSlots } from 'vue';

const props = defineProps({
	modelValue: {
		type: String,
		require: true
	},
	height: {
		type: String,
		default: '',
		required: false
	},
	right: {
		type: Boolean,
		default: false,
		require: false
	},
	isReadOnly: {
		type: Boolean,
		default: false,
		require: false
	},
	placeholder: {
		type: String,
		default: '',
		required: false
	},
	emitKey: {
		type: String,
		default: '',
		require: false
	},
	isPassword: {
		type: Boolean,
		default: false,
		required: false
	}
});

const inputType = ref<'textarea' | 'password' | undefined>(
	props.height ? 'textarea' : props.isPassword ? 'password' : undefined
);

const emit = defineEmits(['input', 'update:modelValue']);
const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
let focused = false;
let blured = false;

const onFocus = () => {
	if (!focused && setFocused) {
		setFocused(true);
		focused = true;
	}
};
const onBlur = () => {
	if (!blured && setBlured) {
		setBlured(true);
		blured = true;
	}
};

const slotIcon = !!useSlots().icon;

function onInput(value: string) {
	if (setBlured) {
		setBlured(false);
		blured = false;
	}
	if (props.emitKey.length > 0) {
		emit('input', props.emitKey, value);
	} else {
		emit('update:modelValue', value);
	}
}
</script>

<style lang="scss" scoped>
.input-item {
	height: var(--height, 32px);
	width: 100%;
	overflow: hidden;

	.input-parent {
		margin-top: -3px;
		overflow: hidden;

		.inner-input {
			width: var(--width);
			padding-left: (--paddingLeft);
			color: 1px solid $separator;
		}

		.text-resize {
			resize: none;
		}

		.text-right {
			text-align: end !important;
		}
	}
}
</style>
