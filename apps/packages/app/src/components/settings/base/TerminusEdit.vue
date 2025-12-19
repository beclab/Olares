<template>
	<div class="terminus-edit">
		<div class="row justify-start items-center">
			<div class="text-body3 terminus-edit__label" v-show="label">
				{{ label }}
			</div>
			<q-icon
				v-if="descriptions && descriptions.length > 0"
				size="20px"
				name="sym_r_info"
				class="text-ink-3"
			>
				<q-tooltip
					self="top left"
					class="text-body3"
					:offset="[0, 0]"
					style="width: 284px"
				>
					<template v-for="item in descriptions" :key="item">
						<div>{{ item }}</div>
					</template>
				</q-tooltip>
			</q-icon>
		</div>
		<div
			class="terminus-edit__bg row wrap justify-between items-center terminus_background_edit_base"
			:class="editBorderClass"
			:style="
				(transaction ? 'background : transparent;' : '') +
				(isTextarea ? '' : 'height: 36px;')
			"
		>
			<img
				:src="inputImageRef"
				style="margin-left: 12px"
				v-if="inputImg.length > 0"
			/>
			<q-input
				:model-value="modelValue"
				:type="inputTypeRef"
				class="text-body3 terminus-edit__bg__input"
				:style="isTextarea ? '' : 'margin-top: -6px;'"
				bg-color="transparent"
				:autogrow="isTextarea"
				:placeholder="hintText"
				borderless
				:inputClass="
					inputClass +
					(inputTypeRef == 'password' ? ' custom-input-padding' : '')
				"
				:input-style="{
					fontWeight: `${
						inputTypeRef === 'password' && !$q.platform.is.ios ? '900' : '400'
					}`,
					fontSize: `${
						inputTypeRef === 'password' && !$q.platform.is.ios
							? '28px'
							: inputTypeRef === 'password' && $q.platform.is.ios
							? '9px'
							: '12px'
					}`,
					maxHeight: isTextarea ? '100px' : ''
				}"
				:readonly="isReadOnly"
				@update:model-value="onTextChange"
				dense
				@keyup.enter="submit"
				@focus="onFocus"
				@blur="onBlur"
			>
				<template v-slot:append>
					<q-icon
						size="16px"
						v-if="showPasswordImg"
						:name="
							inputTypeRef === 'password'
								? 'sym_r_visibility_off'
								: 'sym_r_visibility'
						"
						@click="changeInputType"
						color="text-ink-2"
					/>
					<div v-else>
						<slot name="right"> </slot>
					</div>
				</template>
			</q-input>

			<!-- <div style="position: absolute; right: 0;height: 100%; width: 100px; background: green"/> -->
		</div>
		<div
			class="text-overline terminus-edit__error"
			v-if="errorMessage.length > 0 && isError"
		>
			{{ errorMessage }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from 'src/utils/settings/helper';
import { computed, inject, PropType, ref } from 'vue';
import { useQuasar } from 'quasar';

const props = defineProps({
	label: {
		type: String,
		default: '',
		require: false
	},
	descriptions: {
		type: Object as PropType<string[]>,
		default: [] as string[],
		require: false
	},
	modelValue: {
		type: String,
		default: '',
		require: false
	},
	isError: {
		type: Boolean,
		default: false,
		require: false
	},
	inputImg: {
		type: String,
		default: '',
		require: false
	},
	hintText: {
		type: String,
		default: '',
		required: false
	},
	showPasswordImg: {
		type: Boolean,
		default: false,
		required: false
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
	},
	inputClass: {
		type: String,
		default: 'text-ink-1',
		require: false
	},
	errorMessage: {
		type: String,
		default: '',
		require: false
	},
	transaction: {
		type: Boolean,
		default: false,
		require: false
	},
	isNumber: {
		type: Boolean,
		default: false,
		required: false
	},
	isTextarea: {
		type: Boolean,
		default: false,
		required: false
	}
});

const inputImageRef = ref();
const $q = useQuasar();
const inputTypeRef = ref<
	| 'text'
	| 'password'
	| 'textarea'
	| 'email'
	| 'search'
	| 'tel'
	| 'file'
	| 'number'
	| 'url'
	| 'time'
	| 'date'
	| undefined
>('text');
if (props.inputImg) {
	inputImageRef.value = getRequireImage(`${props.inputImg}`);
}
if (props.showPasswordImg) {
	inputTypeRef.value = 'password';
}

if (props.isNumber) {
	inputTypeRef.value = 'number';
}

// if (props.isTextarea) {
// 	inputTypeRef.value = 'textarea';
// }

const emit = defineEmits(['onTextChange', 'update:modelValue', 'submit']);

function onTextChange(value: any) {
	if (setBlured) {
		setBlured(false);
	}
	if (props.emitKey?.length > 0) {
		emit('onTextChange', props.emitKey, value);
		return;
	}
	emit('update:modelValue', value);
}

function changeInputType() {
	if (inputTypeRef.value === 'text') {
		inputTypeRef.value = 'password';
	} else {
		inputTypeRef.value = 'text';
	}
}
const submit = () => {
	emit('submit');
};

const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
let focused = ref(false);

const onFocus = () => {
	focused.value = true;
	if (setFocused) {
		setFocused(true);
	}
};
const onBlur = () => {
	focused.value = false;
	if (setBlured) {
		setBlured(true);
	}
};
const getErrorStatus = () => {
	return props.isError;
};

defineExpose({ getErrorStatus });

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
</script>

<style lang="scss" scoped>
.terminus_background_edit_base {
	backdrop-filter: blur(6.07811px);
	border-radius: 8px;
	background: transparent;
}

.terminus_input_border_edt_read_only {
	background: linear-gradient(0deg, $background-3, $background-3);
	border: 1px solid $separator;
}

.terminus-edit {
	width: auto;

	&__label {
		color: $ink-3;
	}

	&__bg {
		width: 100%;
		margin-top: 4px;
		position: relative;

		&__input {
			height: 100%;
			width: calc(100% - 30px);
			margin-left: 16px;
			color: $ink-2;
		}
	}

	&__input__less_width {
		width: calc(100% - 64px);
	}

	&__error {
		width: 100%;
		margin-top: 4px;
		color: $red;
	}
}

.terminus_input_border_edit_normal {
	border: 1px solid $separator;
}

.terminus_input_border_editing {
	border: 1px solid $blue-default;
}

.terminus_input_border_edit_error {
	border: 1px solid $red;
}
</style>
