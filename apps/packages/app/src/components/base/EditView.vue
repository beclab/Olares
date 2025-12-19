<template>
	<div class="input-item">
		<div style="margin-top: -5px" class="row justify-center items-center">
			<slot name="icon" />
			<q-input
				class="inner-input"
				standout
				:style="{
					'--width': slotIcon ? 'calc(100% - 32px)' : '100%',
					'--paddingLeft': slotIcon ? '0' : '12px'
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
	height: 32px;
	width: 100%;
	border: 1px solid #dddddd;
	border-radius: 8px;

	.inner-input {
		width: var(--width);
		font-family: Roboto;
		font-size: 12px;
		padding-left: (--paddingLeft);
		padding-right: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
		color: $sub-title;
	}
}
</style>
