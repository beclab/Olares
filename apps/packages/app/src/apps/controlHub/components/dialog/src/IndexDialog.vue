<template>
	<MyQDialog2
		class="card-dialog"
		v-model="visible"
		ref="dialogRef"
		:title="title"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
		:okLoading="loading"
	>
		<div>
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
			<template v-else>
				{{ message }}
				<q-form @reset="onReset">
					<slot />
				</q-form>
			</template>
		</div>

		<!-- <slot name="dialogFooter" />   -->
	</MyQDialog2>
</template>

<script lang="ts" setup>
import { useDialogPluginComponent } from 'quasar';
import { ref, defineProps } from 'vue';
import MyQDialog2 from '@apps/control-panel-common/src/components/MyQDialog2.vue';

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
		default: 'Alert',
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
	loading: {
		type: [String, Boolean],
		default: false,
		required: false
	},
	prompt: {
		type: Object as () => any,
		required: false
	},
	submit: Object as () => any,
	reset: Object as () => any
});

const visible = ref(false);

const emits = defineEmits(['submit', 'reset']);
const onSubmit = () => {
	emits('submit');
};

const onReset = () => {
	emits('reset');
};
const show = () => {
	visible.value = true;
};
const close = () => {
	visible.value = false;
};

const onCancel = () => {
	onReset();
	close();
};

defineExpose({
	show,
	close
});
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		border-radius: 12px;
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
	}
}
</style>
