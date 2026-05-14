<template>
	<q-dialog class="card-dialog" ref="dialogRef">
		<q-card
			class="card-container no-shadow"
			:style="{ width: platform === 'web' ? '400px' : '320px' }"
		>
			<dialog-bar
				:title="title"
				:icon="icon"
				:titAlign="titAlign"
				:platform="platform"
				@close="onCancel"
			/>

			<div class="dialog-content">
				<div v-if="prompt">
					<div class="prompt-name q-mb-xs" v-if="prompt.name">
						{{ prompt.name }}
					</div>
					<q-input
						class="prompt-input"
						v-model="promptModel"
						borderless
						input-class="text-ink-2"
						dense
						:placeholder="prompt.placeholder"
					/>
				</div>
				<div v-else class="message text-ink-2 text-body2">
					{{ message }}
					<slot />
				</div>
			</div>

			<dialog-footer
				:ok="ok"
				:cancel="cancel"
				:okStyle="okStyle"
				:loading="okLoading"
				:platform="platform"
				@onCancel="onCancel"
				@onSubmit="submit"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useDialogPluginComponent } from 'quasar';
import { ref, defineProps } from 'vue';

import DialogBar from './../../custom-dialog/src/DialogBar.vue';
import DialogFooter from './../../custom-dialog/src/DialogFooter.vue';

const props = defineProps({
	platform: {
		type: String,
		default: 'web',
		required: false
	},
	message: {
		type: String,
		default: '',
		required: false
	},
	title: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	titAlign: {
		type: String,
		default: 'left',
		required: false
	},
	ok: {
		type: [String, Boolean],
		default: true,
		required: false
	},
	okStyle: {
		type: Object as () => any,
		required: false
	},
	cancel: {
		type: [String, Boolean],
		default: false,
		required: false
	},
	okLoading: {
		type: [String, Boolean],
		default: false,
		required: false
	},
	prompt: {
		type: Object as () => any,
		required: false
	}
});

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const loading = ref(false);
const promptModel = ref<string>();

if (props.prompt && props.prompt.model) {
	promptModel.value = props.prompt.model as string;
}

const submit = async () => {
	loading.value = true;
	if (props.prompt) {
		if (
			props.prompt.isValid &&
			typeof props.prompt.isValid === 'function' &&
			!props.prompt.isValid(promptModel.value ?? '')
		) {
			loading.value = false;
			return;
		}
		if (promptModel.value) {
			onDialogOK(promptModel.value);
		} else {
			onDialogCancel();
		}
	} else {
		onDialogOK(true);
	}
};

const onCancel = () => {
	onDialogCancel();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-container {
		border-radius: 12px;
		padding: 20px;

		.dialog-content {
			padding: 20px 0 32px;
		}

		.prompt-name {
			color: $ink-3;
			font-size: 12px;
			line-height: 16px;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.prompt-input {
			border: 1px solid $input-stroke;
			border-radius: 8px;
			padding: 0 10px;
		}

		.message {
			white-space: normal;
			word-break: normal;
			overflow-wrap: break-word;
		}
	}
}
</style>
