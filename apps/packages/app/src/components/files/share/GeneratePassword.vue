<template>
	<div class="row items-center justify-between">
		<terminus-edit
			:inputHeight="40"
			class="q-mr-md"
			style="flex: 1"
			:modelValue="modelValue"
			@update:modelValue="$emit('update:modelValue', $event)"
			:show-password-img="false"
		>
			<template v-slot:right v-if="copy">
				<div
					class="row justify-center items-center"
					v-if="modelValue && modelValue.length > 0"
					@click="copyPassword"
				>
					<q-icon
						class="q-mt-sm"
						name="sym_r_content_copy"
						color="light-blue-default"
						size="24px"
					/>
				</div>
			</template>
		</terminus-edit>
		<div
			class="generate-password text-body3 row items-center text-light-blue-default"
			@click="generatePassword"
		>
			{{ t('regenerate') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/common/TerminusEdit.vue';
import { useI18n } from 'vue-i18n';
import { generatePasword } from 'src/utils/format';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { getApplication } from 'src/application/base';

const props = defineProps({
	modelValue: {
		type: String,
		require: true
	},
	passwordLength: {
		type: Number,
		required: false,
		default: 6
	},
	copy: {
		type: Boolean,
		required: false,
		default: false
	}
});

const { t } = useI18n();

const emit = defineEmits(['update:modelValue']);

const generatePassword = () => {
	emit('update:modelValue', generatePasword(props.passwordLength));
};

const copyPassword = () => {
	if (!props.modelValue) {
		return;
	}
	getApplication()
		.copyToClipboard(props.modelValue)
		.then(() => {
			notifySuccess(t('copy_successfully'));
		});
};
</script>

<style scoped lang="scss">
.generate-password {
	border: 1px solid $light-blue-default;
	padding: 0px 12px;
	border-radius: 8px;
	height: 40px;
	cursor: pointer;

	-webkit-tap-highlight-color: transparent;
	background-color: transparent;
	tap-highlight-color: transparent;

	&:hover {
		background-color: $background-3;
	}

	&:active {
		background-color: $background-3;
	}

	&:focus {
		background-color: transparent;
		outline: none;
	}
}
</style>
