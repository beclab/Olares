<template>
	<div class="terminus-mnemonic">
		<div
			class="terminus-mnemonic__bg row wrap justify-between items-center"
			:class="
				isError
					? 'terminus-mnemonic__error_bg'
					: isBackup && inputValue.length > 0
					? 'terminus-mnemonic__backup_bg'
					: 'terminus-mnemonic__bg'
			"
		>
			<div class="terminus-mnemonic__bg__input__index text-body2 q-ml-xs">
				{{ index + 1 }}.
			</div>
			<q-input
				ref="inputRef"
				:name="inputName(index)"
				autocomplete="off"
				v-model="inputValue"
				type="text"
				dense
				class="terminus-mnemonic__bg__input text-body2"
				borderless
				:input-style="{ textAlign: 'left' }"
				input-class="text-ink-1"
				:readonly="isReadOnly"
				@update:model-value="onTextChange"
				@keydown="onkeydown"
				@blur="inputBlur"
				@focus="inputFocus"
				text-align="center"
			/>

			<q-icon
				v-if="isBackup && !backupReadonly && inputText.length > 0"
				@click="backupItemRemove"
				name="sym_r_cancel"
				size="20px"
				style="
					position: absolute;
					right: -8px;
					top: -8px;
					cursor: pointer;
					border-radius: 12px;
				"
				color="grey-4"
			>
			</q-icon>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import NativeInputBlurMonitor from 'src/utils/nativeInputBlur';
import { useQuasar } from 'quasar';

const props = defineProps({
	inputText: {
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
	index: {
		type: Number,
		default: 0,
		require: true
	},
	isBackup: {
		type: Boolean,
		required: false,
		default: false
	},
	backupReadonly: {
		type: Boolean,
		required: false,
		default: true
	}
});

const inputRef = ref();

const inputValue = ref(props.inputText);
const emit = defineEmits([
	'onTextChange',
	'onFinishedEdit',
	'onBackupDeleteItem',
	'onUpdateError'
]);

let isUpdating = false;

const $q = useQuasar();

function onTextChange(value: string) {
	if (isUpdating) {
		return;
	}
	isUpdating = true;
	if (props.isError) {
		emit('onUpdateError', props.index, false);
	}
	if (
		inputValue.value.endsWith(' ') ||
		inputValue.value.split(' ').length > 1
	) {
		inputBlur();
	}
	isUpdating = false;
}

function onkeydown(event) {
	console.log(event.key);
}

const isFocus = ref(false);

const inputFocus = () => {
	isFocus.value = true;
	if (nativeBlurMonitor) {
		nativeBlurMonitor.isFocus = true;
	}
};

const inputBlur = () => {
	isFocus.value = false;
	console.log('inputBlur ===>', props.index, inputValue.value);
	let formatValue = inputValue.value;
	let next = false;
	if (formatValue.endsWith(' ')) {
		formatValue = formatValue.trim();
		next = true;
	}
	formatValue = formatValue.toLowerCase();

	emit(
		'onTextChange',
		props.index,
		formatValue,
		next,
		inputName(props.index + 1)
	);

	if (formatValue.split(' ').length == 1) {
		setInputText(formatValue);
		emit('onFinishedEdit', props.index, formatValue);
	}
};

function setInputText(text: string) {
	inputValue.value = text;
}

const inputName = (index: number) => {
	return 'mnemonic_name_index_' + index;
};

const backupItemRemove = () => {
	emit('onBackupDeleteItem', props.index);
};

defineExpose({ setInputText });

let nativeBlurMonitor: NativeInputBlurMonitor | undefined = undefined;
onMounted(() => {
	if ($q.platform.is.nativeMobile) {
		nativeBlurMonitor = new NativeInputBlurMonitor();
		nativeBlurMonitor.onStart();
		nativeBlurMonitor.blur = () => {
			inputRef.value.blur();
		};
	}
});

onBeforeUnmount(() => {
	if ($q.platform.is.nativeMobile && nativeBlurMonitor) {
		nativeBlurMonitor.onEnd();
		nativeBlurMonitor = undefined;
	}
});
</script>

<style lang="scss" scoped>
.terminus-mnemonic {
	width: auto;
	min-height: 36px;

	&__bg {
		min-height: 100%;
		width: 100%;
		padding: 0;
		border-radius: 8px;
		height: 36px;
		border: 1px solid $separator;

		&__input {
			width: calc(100% - 25px);
			margin-top: -3px;
		}

		&__input__index {
			margin-top: -5px;
			text-align: left;
			width: 18px;
			height: 16px;
			color: $ink-2;
		}
	}

	&__error_bg {
		min-height: 100%;
		width: 100%;
		padding: 0;
		border-radius: 8px;
		border: 1px solid $red;
	}

	&__backup_bg {
		min-height: 100%;
		width: 100%;
		padding: 0;
		border-radius: 8px;
		border: 1px solid $yellow;
	}
}
</style>
