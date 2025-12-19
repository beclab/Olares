<template>
	<div class="row justify-start items-center q-mt-xs">
		<div class="text-body3 text-ink-3" v-show="label || env.envName">
			{{ label ? label : env.envName }}
		</div>
		<q-icon
			v-if="!!env.valueFrom || env.description"
			size="16px"
			name="sym_r_help"
			class="text-ink-3 q-ml-xs"
		>
			<q-tooltip
				self="top left"
				class="text-body3"
				:offset="[0, 0]"
				style="width: 284px"
			>
				<div v-if="env.description">{{ env.description }}</div>
				<div v-if="!!env.valueFrom">
					{{
						t('This value is set by a system environment variable', {
							envName: env.valueFrom.envName
						})
					}}
				</div>
			</q-tooltip>
		</q-icon>
	</div>
	<bt-select-v3
		:model-value="env.value || env.default"
		v-if="(env?.options && env.options.length > 0) || !!env.remoteOptions"
		input-class="text-body3 text-ink-1"
		class="q-mt-xs"
		:options="options"
		:is-error="firstErrorRef ? !itemRight : !env.right"
		:error-message="firstErrorRef ? itemError : env.error"
		@update:modelValue="
			(result) => {
				const { right, error } = validateItem(env, result);
				firstErrorRef = false;
				itemRight = true;
				itemError = '';
				emit('update:env', {
					...env,
					value: result,
					right,
					error
				});
			}
		"
	>
	</bt-select-v3>
	<terminus-edit
		v-else
		:model-value="env.value || env.default"
		:show-password-img="env.type === 'password'"
		:is-error="firstErrorRef ? !itemRight : !env.right"
		:error-message="firstErrorRef ? itemError : env.error"
		:is-read-only="!!env.valueFrom"
		@update:modelValue="
			(result) => {
				const { right, error } = validateItem(env, result);
				firstErrorRef = false;
				itemRight = true;
				itemError = '';
				emit('update:env', {
					...env,
					value: result,
					right,
					error
				});
			}
		"
		style="width: 100%"
	/>
</template>

<script setup lang="ts">
import BtSelectV3 from './settings/base/BtSelectV3.vue';
import TerminusEdit from './settings/base/TerminusEdit.vue';
import { BaseEnv, EnvOption, SelectorProps } from '../constant';
import { remoteOptionsProxy } from 'src/api/settings/env';
import { onMounted, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	env: {
		type: Object as PropType<BaseEnv>,
		required: true
	},
	label: {
		type: String,
		default: ''
	},
	firstError: {
		type: Boolean,
		default: false
	}
});
const emit = defineEmits(['update:env']);
const { t } = useI18n();
const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const options = ref<SelectorProps[]>([]);
const itemRight = ref(false);
const itemError = ref(
	t('Application variable validation failed', { type: props.env.type })
);
const firstErrorRef = ref(props.firstError);

async function envOptionToSelector(
	localOptions: EnvOption[],
	remoteUrl: string
): Promise<SelectorProps[]> {
	if (localOptions?.length) {
		return localOptions.map((item) => ({
			label: item.title,
			value: item.value
		}));
	}

	if (!remoteUrl) return [];

	try {
		const res = await remoteOptionsProxy(remoteUrl);
		if (Array.isArray(res.data)) {
			return res.data
				.filter(
					(item) =>
						typeof item === 'object' &&
						item?.title != null &&
						item?.value != null
				)
				.map((item) => ({
					label: item.title,
					value: item.value
				}));
		}
	} catch (e) {
		console.error(e.message);
	}

	return [];
}

onMounted(async () => {
	options.value = await envOptionToSelector(
		props.env.options,
		props.env.remoteOptions
	);
});

function validateItem(item: BaseEnv, value: string) {
	console.log(item);
	if (!value) {
		return { right: false, error: t('cannot be empty') };
	}

	if (item.regex) {
		const reg = new RegExp(item.regex);
		if (!reg.test(value)) {
			return {
				right: false,
				error: t(
					'does not meet format requirements (regex validation failed)',
					{ regex: item.regex }
				)
			};
		}
	}

	switch (item.type) {
		case 'number':
			if (isNaN(Number(value))) {
				return { right: false, error: t('must be a number') };
			}
			break;
		case 'email':
			if (!emailReg.test(value)) {
				return { right: false, error: t('invalid email format') };
			}
			break;
		default:
			break;
	}

	return { right: true, error: '' };
}
</script>

<style scoped lang="scss"></style>
