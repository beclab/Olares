<template>
	<div class="row justify-start items-center q-mt-xs">
		<div class="text-body3 text-ink-3" v-show="label || env.envName">
			{{ label ? label : env.envName }}
		</div>
		<q-icon
			v-if="env.description || hasBoundEnvRef"
			size="16px"
			name="sym_r_help"
			class="text-ink-3 q-ml-xs"
		>
			<q-tooltip self="top left" class="text-body3" :offset="[0, 0]">
				<div style="max-width: 284px">
					<div v-if="env.description">{{ env.description }}</div>
					<div v-if="hasBoundEnvRef">
						{{
							t('This value is set by a system environment variable', {
								envName: env.valueFrom!.envName
							})
						}}
					</div>
				</div>
			</q-tooltip>
		</q-icon>
	</div>
	<div class="column env-var-input-body q-mt-xs">
		<div class="row env-var-input-row items-start no-wrap">
			<div ref="refPickerMainEl" class="env-var-input-main col">
				<bt-select-v3
					v-if="options.length > 0 && !hasBoundEnvRef && !isMissingRefRepairUi"
					:model-value="String(inputValue ?? '')"
					input-class="text-body3 text-ink-1"
					class="full-width"
					width="100%"
					:options="options"
					:is-error="inlineFieldError"
					:error-message="inlineErrorMessage"
					@update:modelValue="onSelectInput"
				/>
				<div
					v-else-if="options.length > 0 && hasBoundEnvRef"
					class="ref-bound-display ref-bound-display--linked text-body3"
					:class="{
						'ref-bound-display--error': showValidationError
					}"
				>
					{{ refBoundLabel }}
				</div>
				<terminus-edit
					v-else-if="!hasBoundEnvRef"
					:model-value="inputValueStr"
					:show-password-img="env.type === 'password'"
					:is-error="inlineFieldError"
					:error-message="inlineErrorMessage"
					:is-read-only="false"
					class="full-width"
					@update:modelValue="onValueInput"
				/>
				<div
					v-else
					class="ref-bound-display ref-bound-display--linked text-body3"
					:class="{
						'ref-bound-display--error': showValidationError
					}"
				>
					{{ refBoundLabel }}
				</div>
			</div>

			<div v-if="showNotFoundRefPicker" class="env-var-input-actions col-auto">
				<q-btn
					round
					dense
					class="text-ink-2"
					padding="6px"
					icon="sym_r_drag_click"
					:title="t('env_ref_picker_tooltip')"
					@click.stop="onLinkMenuToggle"
				/>
				<q-menu
					v-model="linkMenuOpen"
					:target="refPickerMainEl ?? false"
					anchor="bottom left"
					self="top left"
					:offset="[0, 6]"
					class="env-ref-q-menu"
					:style="refMenuPaneStyle"
					@show="onLinkMenuShown"
				>
					<div class="env-ref-select-card">
						<div class="env-ref-list-scroll">
							<div
								v-if="refListLoadPending"
								class="env-ref-item-row env-ref-item text-body3 text-ink-3 q-pa-sm"
							>
								{{ t('loading') }}
							</div>
							<template v-else>
								<div
									v-for="name in refNames"
									:key="name"
									class="env-ref-item-row env-ref-item text-body2 text-ink-2"
									v-close-popup
									@click="pickSystemEnvRef(name)"
								>
									<span class="env-ref-item-label">{{ name }}</span>
								</div>
								<div
									v-if="refNames.length === 0"
									class="env-ref-item-row env-ref-item text-body3 text-ink-3 q-pa-sm"
								>
									{{ t('No available environment variable configurations') }}
								</div>
							</template>
						</div>
					</div>
				</q-menu>
			</div>
		</div>
		<div
			v-if="hasBoundEnvRef && showValidationError"
			class="text-caption text-negative q-mt-xs"
		>
			{{ firstErrorRef ? itemError : env.error }}
		</div>
	</div>
</template>

<script setup lang="ts">
import BtSelectV3 from './settings/base/BtSelectV3.vue';
import TerminusEdit from './settings/base/TerminusEdit.vue';
import { BaseEnv, EnvOption, SelectorProps } from '../constant';
import { computed, onMounted, PropType, ref, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	getSystemEnvList,
	getUserEnvList,
	remoteOptionsProxy
} from 'src/api/settings/env';

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
	},
	enableEnvRefPicker: {
		type: Boolean,
		default: true
	},
	isFromMissingRefs: {
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
const refNames = ref<string[]>([]);
const refValueByName = ref<Record<string, string>>({});
const refPickerMainEl = ref<HTMLElement | null>(null);
const refPickerMenuWidth = ref('100%');
const linkMenuOpen = ref(false);
const refListLoadPending = ref(false);

const refMenuPaneStyle = computed(() => {
	const w = refPickerMenuWidth.value;
	return {
		width: w,
		minWidth: w,
		maxWidth: `min(calc(100vw - 24px), ${w})`
	};
});

function syncRefMenuWidth() {
	const el = refPickerMainEl.value;
	refPickerMenuWidth.value = el?.offsetWidth ? `${el.offsetWidth}px` : '100%';
}

function effectiveEnvLiteral(e: BaseEnv): string {
	const v = e.value;
	if (v !== undefined && v !== null && String(v).trim() !== '') {
		return String(v);
	}
	return String(e.default ?? '');
}

async function loadRefEnvLists() {
	try {
		const [systemList, userList] = await Promise.all([
			getSystemEnvList(),
			getUserEnvList()
		]);
		let sys;
		let usr;

		if (process.env.APPLICATION === 'MARKET') {
			sys = Array.isArray(systemList.data) ? systemList.data : [];
			usr = Array.isArray(userList.data) ? userList.data : [];
		} else if (process.env.APPLICATION === 'SETTINGS') {
			sys = Array.isArray(systemList) ? systemList : [];
			usr = Array.isArray(userList) ? userList : [];
		} else {
			new Error('Channel must be supported');
		}
		const valueMap: Record<string, string> = {};
		for (const e of sys) {
			if (e?.envName) {
				valueMap[e.envName] = effectiveEnvLiteral(e);
			}
		}
		for (const e of usr) {
			if (e?.envName) {
				valueMap[e.envName] = effectiveEnvLiteral(e);
			}
		}
		refValueByName.value = valueMap;
		const names = new Set<string>();
		sys.forEach((e) => e?.envName && names.add(e.envName));
		usr.forEach((e) => e?.envName && names.add(e.envName));
		refNames.value = Array.from(names).sort((a, b) =>
			a.localeCompare(b, undefined, { sensitivity: 'base' })
		);
	} catch (e) {
		console.error(e);
		refNames.value = [];
		refValueByName.value = {};
	}
}

function onLinkMenuToggle() {
	if (linkMenuOpen.value) {
		linkMenuOpen.value = false;
		return;
	}
	syncRefMenuWidth();
	linkMenuOpen.value = true;
}

async function onLinkMenuShown() {
	nextTick(() => {
		requestAnimationFrame(() => {
			syncRefMenuWidth();
		});
	});
	const showLoading = refNames.value.length === 0;
	if (showLoading) {
		refListLoadPending.value = true;
	}
	try {
		await loadRefEnvLists();
	} finally {
		refListLoadPending.value = false;
	}
}

const hasBoundEnvRef = computed(() => {
	const vf = props.env.valueFrom;
	return !!(vf?.envName?.trim() && vf.status === 'synced');
});

const isMissingRefRepairUi = computed(() => {
	const vf = props.env.valueFrom;
	if (!vf?.envName?.trim()) {
		return false;
	}
	if (vf.status === 'notfound') {
		return true;
	}
	return props.isFromMissingRefs;
});

const showNotFoundRefPicker = computed(
	() => props.enableEnvRefPicker && isMissingRefRepairUi.value
);

const showValidationError = computed(() => {
	if (firstErrorRef.value) {
		return !itemRight.value;
	}
	return props.env.right === false;
});

const missingRefErrorMessage = computed(() => {
	if (!isMissingRefRepairUi.value || !props.env.valueFrom?.envName) {
		return '';
	}
	return t('env_ref_system_var_missing', {
		envName: props.env.valueFrom.envName
	});
});

const inlineErrorMessage = computed(() => {
	if (missingRefErrorMessage.value) {
		return missingRefErrorMessage.value;
	}
	return firstErrorRef.value ? itemError.value : props.env.error ?? '';
});

const inlineFieldError = computed(
	() => !!missingRefErrorMessage.value || showValidationError.value
);

watch(isMissingRefRepairUi, (active) => {
	if (!active) {
		linkMenuOpen.value = false;
	}
});

const refBoundLabel = computed(() => {
	const vf = props.env.valueFrom;
	if (hasBoundEnvRef.value && vf?.envName) {
		return vf.envName;
	}
	return String(props.env.value ?? '').trim();
});

const inputValue = computed(() => {
	if (props.env.value === undefined) {
		return props.env.default;
	} else {
		return props.env.value;
	}
});

const inputValueStr = computed(() => {
	const v = inputValue.value;
	if (v === undefined || v === null) {
		return '';
	}
	return String(v);
});

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
	if (props.env.type === 'bool') {
		options.value = [
			{
				label: 'true',
				value: 'true'
			},
			{
				label: 'false',
				value: 'false'
			}
		];
	} else {
		options.value = await envOptionToSelector(
			props.env.options ?? [],
			props.env.remoteOptions ?? ''
		);
	}

	await loadRefEnvLists();
});

function isRefBindingValue(item: BaseEnv) {
	const vf = item.valueFrom;
	return !!(vf?.envName?.trim() && vf.status === 'synced');
}

function validateValueAgainstSchema(item: BaseEnv, valueToCheck: string) {
	if (!valueToCheck && !!item.required) {
		return { right: false, error: t('cannot be empty') };
	}

	const trimmed = String(valueToCheck ?? '').trim();
	if (item.options?.length && trimmed !== '') {
		const ok = item.options.some((o) => String(o.value).trim() === trimmed);
		if (!ok) {
			return {
				right: false,
				error: t('Application variable validation failed', {
					type: item.type ?? 'value'
				})
			};
		}
	}

	if (item.regex) {
		try {
			const reg = new RegExp(item.regex);
			if (!reg.test(valueToCheck)) {
				return {
					right: false,
					error: t(
						'does not meet format requirements (regex validation failed)',
						{ regex: item.regex }
					)
				};
			}
		} catch {
			return {
				right: false,
				error: t(
					'does not meet format requirements (regex validation failed)',
					{ regex: String(item.regex) }
				)
			};
		}
	}

	switch (item.type) {
		case 'number':
		case 'int':
			if (isNaN(Number(valueToCheck))) {
				return { right: false, error: t('must be a number') };
			}
			break;
		case 'email':
			if (!emailReg.test(valueToCheck)) {
				return { right: false, error: t('invalid email format') };
			}
			break;
		case 'bool':
			{
				const v = String(valueToCheck).trim().toLowerCase();
				if (v !== 'true' && v !== 'false') {
					return {
						right: false,
						error: t('Application variable validation failed', {
							type: item.type
						})
					};
				}
			}
			break;
		default:
			break;
	}

	return { right: true, error: '' };
}

function literalForRefValidation(refName: string, value: string): string {
	const fromMap = refValueByName.value[refName];
	if (fromMap !== undefined) {
		return String(fromMap ?? '');
	}
	if (value === refName) {
		return '';
	}
	return String(value ?? '');
}

function validateItem(item: BaseEnv, value: string) {
	if (isRefBindingValue(item)) {
		const refName = item.valueFrom!.envName;
		return validateValueAgainstSchema(
			item,
			literalForRefValidation(refName, value)
		);
	}

	if (!value && !!item.required) {
		return { right: false, error: t('cannot be empty') };
	}

	return validateValueAgainstSchema(item, value);
}

function pickSystemEnvRef(name: string) {
	const preserveApplyOnChange =
		props.env.valueFrom?.status === 'notfound' || props.isFromMissingRefs;
	const resolved = refValueByName.value[name] ?? '';
	const next: BaseEnv = {
		...props.env,
		value: resolved,
		valueFrom: {
			envName: name,
			status: 'synced'
		},
		applyOnChange: preserveApplyOnChange ? props.env.applyOnChange : true
	};
	const { right, error } = validateItem(next, resolved);
	firstErrorRef.value = false;
	itemRight.value = true;
	itemError.value = '';
	emit('update:env', { ...next, right, error });
	linkMenuOpen.value = false;
}

function onSelectInput(result: string) {
	const next: BaseEnv = {
		...props.env,
		value: result,
		valueFrom: undefined,
		applyOnChange: props.env.applyOnChange
	};
	const { right, error } = validateItem(next, result);
	firstErrorRef.value = false;
	itemRight.value = true;
	itemError.value = '';
	emit('update:env', { ...next, right, error });
}

function onValueInput(result: string) {
	const next: BaseEnv = {
		...props.env,
		value: result,
		valueFrom: undefined,
		applyOnChange: props.env.applyOnChange
	};

	const { right, error } = validateItem(next, result);
	firstErrorRef.value = false;
	itemRight.value = true;
	itemError.value = '';
	emit('update:env', { ...next, right, error });
}
</script>

<style scoped lang="scss">
.env-var-input-body {
	width: 100%;
}

.env-var-input-row {
	width: 100%;
	gap: 8px;
	align-items: center;
}

.env-var-input-main {
	min-width: 0;
	flex: 1 1 auto;
}

.env-var-input-actions {
	flex-shrink: 0;
}

.full-width {
	width: 100%;
}

.ref-bound-display {
	display: flex;
	align-items: center;
	min-height: 40px;
	padding: 8px 12px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	box-sizing: border-box;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.ref-bound-display--linked {
	background: $background-2;
	color: $ink-3;
	cursor: default;
	user-select: text;
}

.ref-bound-display--error {
	border-color: $negative;
}
</style>

<style lang="scss">
.env-ref-q-menu {
	box-sizing: border-box;
}

.env-ref-q-menu .env-ref-select-card {
	width: 100%;
	display: flex;
	padding: 0 0 12px;
	flex-direction: column;
	align-items: stretch;
	gap: 0;
	background: $background-2;
	color: $ink-2;
	box-sizing: border-box;
}

.env-ref-q-menu .env-ref-list-scroll {
	max-height: 220px;
	overflow: auto;
	padding: 0 12px 0;
}

.env-ref-q-menu .env-ref-item-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	width: 100%;
	gap: 8px;
	min-height: 34px;
	padding: 8px 0;
	border-radius: 4px;
	box-sizing: border-box;
}

.env-ref-q-menu .env-ref-item-label {
	flex: 1;
	min-width: 0;
	overflow: hidden;
	text-overflow: ellipsis;
}

.env-ref-q-menu .env-ref-item {
	cursor: pointer;
	text-decoration: none;

	&:hover {
		background: $background-hover !important;
	}
}
</style>
