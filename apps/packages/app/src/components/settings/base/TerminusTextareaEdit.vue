<template>
	<div class="terminus-textarea-edit-root">
		<div class="text-body3 text-ink-3 q-mb-xs" v-if="label">
			{{ label }}
		</div>
		<div class="common-border" :class="editBorderClass">
			<q-input
				:model-value="modelValue"
				borderless
				dense
				type="textarea"
				input-class="custom-placeholder text-body3"
				:input-style="`padding: 8px 32px 8px 16px !important; height: ${normalHeight}px`"
				@focus="onFocus"
				@blur="onBlur"
				@update:model-value="onTextChange"
			/>
			<div
				class="clear_btn row items-center"
				v-if="modelValue.length > 0"
				@click="onTextChange('')"
			>
				<q-icon name="sym_r_cancel" color="ink-3" size="18px" />
			</div>
		</div>
		<div
			class="text-overline terminus-textarea-edit-root__error"
			v-if="errorMessage.length > 0 && isError"
		>
			{{ errorMessage }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
const props = defineProps({
	label: {
		type: String,
		default: '',
		require: false
	},
	modelValue: {
		type: String,
		default: '',
		require: false
	},
	normalHeight: {
		type: Number,
		default: 116,
		required: false
	},
	errorMessage: {
		type: String,
		default: '',
		require: false
	},
	isError: {
		type: Boolean,
		default: false,
		require: false
	},
	isReadOnly: {
		type: Boolean,
		default: false,
		require: false
	},
	emitKey: {
		type: String,
		default: '',
		require: false
	}
});

let focused = ref(false);

const onFocus = () => {
	focused.value = true;
};
const onBlur = () => {
	focused.value = false;
};

const editBorderClass = computed(() => {
	if (props.errorMessage.length > 0 && props.isError) {
		return 'terminus_input_border_edit_error';
	}
	if (props.isReadOnly) {
		return 'terminus_input_border_edt_read_only';
	}
	if (focused.value) {
		return 'terminus_input_border_editing';
	}
	return 'terminus_input_border_edit_normal';
});

const emit = defineEmits(['onTextChange', 'update:modelValue', 'submit']);

function onTextChange(value: any) {
	if (props.emitKey?.length > 0) {
		emit('onTextChange', props.emitKey, value);
		return;
	}
	emit('update:modelValue', value);
}
</script>

<style scoped lang="scss">
.terminus-textarea-edit-root {
	.common-border {
		border-radius: 8px;
		position: relative;

		.clear_btn {
			position: absolute;
			height: 24px;
			width: 24px;
			right: 0px;
			top: 4px;
		}
	}

	&__error {
		width: 100%;
		margin-top: 6px;
		color: $red;
	}
}
</style>
