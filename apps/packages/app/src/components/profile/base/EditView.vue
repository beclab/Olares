<template>
	<div class="input-item" :style="{ '--height': height }">
		<div class="input-parent row justify-center items-center">
			<slot name="icon" />
			<q-input
				:type="height ? 'textarea' : ''"
				class="inner-input"
				standout
				:style="{
					'--width': slotIcon ? 'calc(100% - 32px)' : '100%',
					'--paddingLeft': slotIcon ? '0' : '12px'
				}"
				input-class="text-body1 text-ink-1"
				:input-style="{
					height: height,
					resize: 'none',
					padding: '8px !important'
				}"
				borderless
				dense
				name="search"
				debounce="500"
				:placeholder="placeholder"
				autocomplete="off"
				:readonly="isReadOnly"
				:disable="isReadOnly"
				@keyup.enter="onInput(modelValue)"
				:model-value="modelValue"
				@update:model-value="onInput"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useSlots } from 'vue';

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
	}
});

const emit = defineEmits(['input', 'update:modelValue']);

const slotIcon = !!useSlots().icon;

function onInput(value: string) {
	if (props.emitKey.length > 0) {
		emit('input', props.emitKey, value);
	} else {
		emit('update:modelValue', value);
	}
}
</script>

<style lang="scss" scoped>
.input-item {
	height: var(--height, 40px);
	width: 100%;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	overflow: hidden;

	.input-parent {
		height: 100%;
		overflow: hidden;

		.inner-input {
			width: var(--width);
			padding-left: (--paddingLeft);
			color: 1px solid $ink-2;
		}
	}
}
</style>
