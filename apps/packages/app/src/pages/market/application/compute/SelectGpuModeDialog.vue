<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('compute.select_mode_title')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:ok-disabled="!canSubmit"
		@onSubmit="submitMode"
	>
		<div class="text-body2 text-ink-2 q-mb-lg">
			{{ t('compute.select_mode_desc') }}
		</div>

		<div class="mode-select row items-center no-wrap">
			<div class="mode-select__value row items-center no-wrap">
				<span
					class="mode-select__label text-body1 ellipsis"
					:class="selectedOption ? 'text-ink-2' : 'text-ink-3'"
				>
					{{
						selectedOption
							? modeLabel(selectedOption.computeType)
							: t('compute.select_mode_placeholder')
					}}
				</span>
				<compute-resource-badge
					v-if="selectedOption && selectedOption.status !== 'installable'"
					:label="t(`compute.mode_status.${statusKey(selectedOption.status)}`)"
					tone="bad"
				/>
			</div>
			<q-icon name="sym_r_keyboard_arrow_down" size="24px" color="ink-3" />

			<q-menu
				ref="menuRef"
				anchor="bottom left"
				self="top left"
				fit
				class="mode-select__menu"
			>
				<div class="mode-select__menu-list column">
					<div
						v-for="opt in sortedOptions"
						:key="opt.computeType"
						class="mode-select__menu-item row items-center justify-between no-wrap"
						:class="{
							'mode-select__menu-item--active': selected === opt.computeType,
							'mode-select__menu-item--disabled': !isOptionEnabled(opt)
						}"
						@click="onPick(opt)"
					>
						<div class="mode-select__menu-label row items-center no-wrap">
							<span class="text-body1 ellipsis">
								{{ modeLabel(opt.computeType) }}
							</span>
							<compute-resource-badge
								v-if="opt.status !== 'installable'"
								:label="t(`compute.mode_status.${statusKey(opt.status)}`)"
								tone="bad"
							>
								<q-tooltip
									v-if="!isOptionEnabled(opt)"
									anchor="top middle"
									self="bottom middle"
									class="text-body3"
								>
									{{ t(`compute.mode_status_tip.${statusKey(opt.status)}`) }}
								</q-tooltip>
							</compute-resource-badge>
						</div>
						<q-icon
							v-if="selected === opt.computeType"
							name="sym_r_check_circle"
							size="24px"
							color="blue-default"
						/>
					</div>
				</div>
			</q-menu>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import ComputeResourceBadge from 'src/components/compute/ComputeResourceBadge.vue';
import { computed, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	ComputeModeOption,
	ComputeModeStatus,
	ComputeType,
	computeModeTitle,
	pickDefaultMode,
	sortModeOptions
} from 'src/constant/compute';

const props = defineProps({
	appTitle: {
		type: String,
		required: true
	},
	modes: {
		type: Array as PropType<ComputeModeOption[]>,
		required: true
	}
});

const { t } = useI18n();
const CustomRef = ref();
const menuRef = ref();

const sortedOptions = computed(() => sortModeOptions(props.modes));

const selected = ref<ComputeType | undefined>(pickDefaultMode(props.modes));

const selectedOption = computed(() =>
	sortedOptions.value.find((o) => o.computeType === selected.value)
);

const isOptionEnabled = (opt: ComputeModeOption) =>
	opt.status === 'installable';

const canSubmit = computed(() => {
	const opt = selectedOption.value;
	return !!opt && isOptionEnabled(opt);
});

const onPick = (opt: ComputeModeOption) => {
	if (!isOptionEnabled(opt)) return;
	selected.value = opt.computeType;
	menuRef.value?.hide();
};

/** Convert hyphenated `insufficient-resources` -> `insufficient_resources` for i18n key. */
const statusKey = (status: ComputeModeStatus) => status.replace(/-/g, '_');

const modeLabel = (type: ComputeType): string => computeModeTitle(type);

const submitMode = () => {
	if (!canSubmit.value || !selected.value) return;
	CustomRef.value.onDialogOK({
		selectedGpuType: selected.value
	});
};
</script>

<style scoped lang="scss">
.mode-select {
	gap: 4px;
	width: 100%;
	min-height: 40px;
	padding: 8px 8px 8px 12px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	cursor: pointer;

	&__value {
		flex: 1;
		min-width: 0;
		gap: 8px;
	}

	&__label {
		min-width: 0;
	}

	&__menu-list {
		padding: 8px;
		gap: 4px;
	}

	&__menu-item {
		padding: 8px;
		border-radius: 4px;
		gap: 8px;
		cursor: pointer;
		color: $ink-2;

		&:hover:not(&--disabled) {
			background: $background-hover;
		}

		&--active {
			color: $blue-default;
		}

		&--disabled {
			color: $ink-3;
			cursor: not-allowed;
		}
	}

	&__menu-label {
		min-width: 0;
		gap: 8px;
	}
}
</style>
