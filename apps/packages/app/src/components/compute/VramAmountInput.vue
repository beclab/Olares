<template>
	<div class="vram-amount">
		<div class="row items-center no-wrap" style="gap: 8px">
			<div class="vram-amount__input">
				<terminus-edit
					:model-value="text"
					:is-number="false"
					inputmode="decimal"
					:show-password-img="false"
					:is-read-only="readonly"
					:is-error="!readonly && !!error"
					:error-message="''"
					:input-class="readonly ? 'text-ink-3' : 'text-ink-1'"
					@update:model-value="onInput"
				>
					<template v-if="!readonly" v-slot:right>
						<edit-number-right-slot
							:model-value="text"
							:max="maxInUnit"
							:min="minInUnit"
							:unit="step"
							@update:model-value="onStep"
						/>
					</template>
				</terminus-edit>
			</div>

			<!-- Read-only cards (Exclusive / Time slicing) reuse the same box; only
			     the unit becomes a static label instead of a switchable dropdown. -->
			<div
				class="vram-amount__unit row items-center justify-between no-wrap"
				:class="readonly ? 'vram-amount__unit--readonly' : 'cursor-pointer'"
			>
				<span
					class="text-body2"
					:class="readonly ? 'text-ink-3' : 'text-ink-2'"
					>{{ unitLabel }}</span
				>
				<template v-if="!readonly">
					<q-icon name="sym_r_expand_more" size="18px" class="text-ink-3" />
					<q-menu auto-close anchor="bottom right" self="top right">
						<q-list dense class="vram-amount__menu">
							<q-item
								v-for="opt in unitOptions"
								:key="opt.value"
								clickable
								:active="unit === opt.value"
								active-class="text-blue-default"
								@click="setUnit(opt.value)"
							>
								<q-item-section class="text-body2">{{
									opt.label
								}}</q-item-section>
							</q-item>
						</q-list>
					</q-menu>
				</template>
			</div>
		</div>

		<div v-if="!readonly && !hideHelp" class="text-body3 text-ink-3 q-mt-xs">
			{{ t('compute.vram_input_help', { max: maxLabel }) }}
		</div>
	</div>
</template>

<script setup lang="ts">
import EditNumberRightSlot from 'src/components/settings/EditNumberRightSlot.vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { formatComputeMemoryValue } from 'src/constant/compute';
import { computed, nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

type Unit = 'GB' | 'MB';

const GiB = 1024 ** 3;
const MiB = 1024 ** 2;
// Absorbs byte<->GB float drift before flooring/ceiling (see compute.ts), so an
// 8.06 GB allocation doesn't reappear as 8.05.
const GB_EPSILON = 1e-6;

const props = withDefaults(
	defineProps<{
		modelValue: number;
		minBytes: number;
		maxBytes: number;
		/** Render the same box but non-editable (Exclusive / Time slicing cards). */
		readonly?: boolean;
		/** Hide the bottom "Available: {max}" help line (inline quota row). */
		hideHelp?: boolean;
	}>(),
	{ readonly: false, hideHelp: false }
);

const emit = defineEmits<{
	(e: 'update:modelValue', bytes: number): void;
	(e: 'update:valid', valid: boolean): void;
	/** Current validation message ('' when valid) so the parent can place it. */
	(e: 'update:error', message: string): void;
}>();

const { t } = useI18n();

const unit = ref<Unit>('GB');
// Internal unit values stay GB/MB (the math uses GiB/MiB); the user-visible
// labels are Gi/Mi to match the global memory-unit wording (formatComputeMemory).
const UNIT_LABEL: Record<Unit, string> = { GB: 'Gi', MB: 'Mi' };
const unitLabel = computed(() => UNIT_LABEL[unit.value]);
const unitOptions: { label: string; value: Unit }[] = [
	{ label: UNIT_LABEL.GB, value: 'GB' },
	{ label: UNIT_LABEL.MB, value: 'MB' }
];

const unitBytes = computed(() => (unit.value === 'GB' ? GiB : MiB));
const step = computed(() => (unit.value === 'GB' ? 0.1 : 64));

// Floored so the displayed/steppable max never exceeds the real byte cap.
// GB keeps 2 decimals (matching the input limit); MB stays integer.
const maxInUnit = computed(() =>
	unit.value === 'GB'
		? Math.floor((props.maxBytes / GiB) * 100 + GB_EPSILON) / 100
		: Math.floor(props.maxBytes / MiB)
);
const minInUnit = computed(() =>
	unit.value === 'GB'
		? Math.ceil((props.minBytes / GiB) * 100 - GB_EPSILON) / 100
		: Math.ceil(props.minBytes / MiB)
);

const maxLabel = computed(() => `${maxInUnit.value} ${unitLabel.value}`);
const minLabel = computed(() => `${minInUnit.value} ${unitLabel.value}`);

const bytesToUnitText = (bytes: number): string => {
	// Floor so the shown value never exceeds the real bytes (GB keeps 2 decimals,
	// MB is integer-only).
	const v =
		unit.value === 'GB'
			? Math.floor((bytes / GiB) * 100 + GB_EPSILON) / 100
			: Math.floor(bytes / MiB);
	const clamped = Math.min(maxInUnit.value, Math.max(minInUnit.value, v));
	return String(clamped);
};

const clampBytes = (bytes: number): number =>
	Math.min(Math.max(Math.round(bytes), props.minBytes), props.maxBytes);

const toBytes = (raw: string): number => Number(raw) * unitBytes.value;

// Read-only cards just show their fixed capacity (GB), never clamped/edited.
const text = ref(
	props.readonly
		? formatComputeMemoryValue(props.modelValue)
		: bytesToUnitText(clampBytes(props.modelValue || props.minBytes))
);

// GB allows at most 2 decimals; MB is integer-only.
const numberPattern = computed(() =>
	unit.value === 'GB' ? /^\d+(\.\d{1,2})?$/ : /^\d+$/
);

/**
 * Clamp the raw text to what the current unit allows: MB drops the decimal part
 * entirely, GB keeps a single dot and caps the fraction at 2 digits. A lone
 * trailing dot (e.g. "1.") is preserved so the user can keep typing the
 * fraction.
 */
const sanitize = (raw: string): string => {
	const v = String(raw ?? '').replace(/[^\d.]/g, '');
	if (unit.value === 'MB') return v.split('.')[0];
	const dot = v.indexOf('.');
	if (dot < 0) return v;
	const intPart = v.slice(0, dot);
	const decPart = v
		.slice(dot + 1)
		.replace(/\./g, '')
		.slice(0, 2);
	return `${intPart}.${decPart}`;
};

const error = computed((): string => {
	const raw = text.value;
	if (raw === undefined || String(raw).trim() === '') {
		return t('compute.vram_required', { unit: unitLabel.value });
	}
	if (!numberPattern.value.test(String(raw))) {
		return t('compute.vram_not_number');
	}
	const bytes = toBytes(String(raw));
	if (bytes <= 0) return t('compute.vram_required', { unit: unitLabel.value });
	if (bytes < props.minBytes) {
		return t('compute.vram_below_required', { min: minLabel.value });
	}
	if (bytes > props.maxBytes) {
		return t('compute.vram_exceeds_remaining');
	}
	return '';
});

const onInput = (val: string) => {
	const raw = String(val ?? '');
	const clean = sanitize(raw);
	// When the input is rejected (e.g. a 3rd GB decimal) the sanitized value can
	// equal the current ref, so q-input wouldn't re-render and the dirty char
	// would linger in the DOM. Bounce through the raw value to force a redraw
	// back to the clean one.
	if (clean === text.value && raw !== clean) {
		text.value = raw;
		nextTick(() => {
			text.value = clean;
		});
		return;
	}
	text.value = clean;
};

const onStep = (val: number) => {
	let v = Math.min(maxInUnit.value, Math.max(minInUnit.value, Number(val)));
	// GB keeps at most 2 decimals; MB is integer-only.
	v = unit.value === 'GB' ? Math.floor(v * 100) / 100 : Math.floor(v);
	text.value = String(v);
};

const setUnit = (next: Unit) => {
	if (next === unit.value) return;
	const bytes = error.value
		? clampBytes(props.modelValue)
		: clampBytes(toBytes(text.value));
	unit.value = next;
	text.value = bytesToUnitText(bytes);
};

watch(
	// `error` is included so capacity changes (e.g. after refresh shrinks the
	// remaining VRAM) re-emit validity/message even when the text is untouched.
	[text, unit, error],
	() => {
		// Read-only cards carry a fixed allocation; never validate or emit edits.
		if (props.readonly) {
			emit('update:valid', true);
			emit('update:error', '');
			return;
		}
		const message = error.value;
		const invalid = !!message;
		emit('update:valid', !invalid);
		emit('update:error', message);
		if (!invalid) emit('update:modelValue', clampBytes(toBytes(text.value)));
	},
	{ immediate: true }
);
</script>

<style scoped lang="scss">
.vram-amount {
	&__input {
		flex: 1;
		min-width: 120px;
	}

	&__unit {
		margin-top: 4px;
		flex-shrink: 0;
		width: 64px;
		height: 36px;
		padding: 0 10px;
		gap: 2px;
		border: 1px solid $separator;
		border-radius: 8px;

		&:hover {
			border-color: $blue-default;
		}

		// Static unit for read-only cards: no hover affordance, disabled fill,
		// centered label (the dropdown caret is hidden).
		&--readonly {
			justify-content: center;
			background: $background-3;

			&:hover {
				border-color: $separator;
			}
		}
	}

	&__menu {
		min-width: 64px;
	}

	&__input {
		:deep(input[type='number']) {
			-moz-appearance: textfield;
			appearance: textfield;
		}

		:deep(input[type='number'])::-webkit-outer-spin-button,
		:deep(input[type='number'])::-webkit-inner-spin-button {
			-webkit-appearance: none;
			margin: 0;
		}
	}
}
</style>
