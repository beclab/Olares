<template>
	<div class="compute-device-row">
		<!-- Selection + metrics dim when the card is unselectable; the bound-apps
		     slot below stays fully interactive so apps can always be unbound. -->
		<div :class="{ 'compute-device-row--disabled': disabled }">
			<div class="row items-center justify-between no-wrap">
				<div class="row items-center no-wrap" style="gap: 8px; min-width: 0">
					<terminus-check-box
						v-if="selectionType === 'checkbox'"
						:model-value="selected"
						:size="18"
						:class="{
							'compute-device-row__select--locked': disabled && !selected
						}"
						@update:model-value="onSelect"
					/>
					<div
						v-else-if="selectionType === 'radio'"
						class="compute-device-row__radio row items-center justify-center"
						:class="{ 'compute-device-row__select--locked': disabled }"
						@click="onSelect"
					>
						<!-- Selected: filled blue dot radio. -->
						<div v-if="selected" class="compute-device-row__radio-on" />
						<terminus-circle-normal-checkbox v-else :size="16" />
					</div>

					<div class="column" style="min-width: 0; gap: 2px">
						<div class="row items-center no-wrap" style="gap: 8px">
							<span class="text-body2 text-ink-1 ellipsis">{{ name }}</span>
							<compute-support-type-badge
								v-if="supportType"
								:support-type="supportType"
							/>
						</div>
						<span v-if="capacityText" class="text-body3 text-ink-3 ellipsis">
							{{ capacityText }}
						</span>
					</div>

					<q-tooltip v-if="disabledReason" class="text-body3">
						{{ disabledReason }}
					</q-tooltip>
				</div>
				<div class="row items-center no-wrap">
					<slot name="meta" />
				</div>
			</div>

			<!-- selected-card content (inline quota allocation); renders nothing
			     until a card is selected, so no empty gap on idle rows. -->
			<slot />
		</div>

		<compute-validation-notice
			v-if="warning"
			class="q-mt-sm"
			:message="warning"
			:tone="warningTone"
		/>

		<!-- bound apps (always interactive, never dimmed) -->
		<div v-if="$slots.apps" class="q-mt-sm">
			<slot name="apps" />
		</div>
	</div>
</template>

<script setup lang="ts">
import ComputeSupportTypeBadge from './ComputeSupportTypeBadge.vue';
import ComputeValidationNotice from './ComputeValidationNotice.vue';
import TerminusCheckBox from 'src/components/common/TerminusCheckBox.vue';
import TerminusCircleNormalCheckbox from 'src/components/common/TerminusCircleNormalCheckbox.vue';
import { PropType } from 'vue';
import { ComputeTone } from 'src/constant/compute';

const props = defineProps({
	deviceId: {
		type: String,
		required: true
	},
	name: {
		type: String,
		required: true
	},
	/** Device support type (TimeSlice / MemorySlice / Exclusive) for the badge. */
	supportType: {
		type: String,
		default: ''
	},
	/** Capacity sub-label, e.g. "Dedicated VRAM: 23.9 Gi". */
	capacityText: {
		type: String,
		default: ''
	},
	selectionType: {
		type: String as PropType<'checkbox' | 'radio' | 'none'>,
		default: 'none'
	},
	selected: {
		type: Boolean,
		default: false
	},
	disabled: {
		type: Boolean,
		default: false
	},
	disabledReason: {
		type: String,
		default: ''
	},
	warning: {
		type: String,
		default: ''
	},
	warningTone: {
		type: String as PropType<ComputeTone>,
		default: 'bad'
	}
});

const emit = defineEmits<{ (e: 'select'): void }>();

const onSelect = () => {
	if (props.disabled && !props.selected) return;
	emit('select');
};
</script>

<style scoped lang="scss">
.compute-device-row {
	padding: 12px;
	border-radius: 12px;
	background: $background-6;

	&--disabled {
		opacity: 0.6;
	}

	&__radio {
		width: 20px;
		height: 20px;
		cursor: pointer;
		flex-shrink: 0;
	}

	// Selected radio: 16px solid blue circle with a centered white dot.
	&__radio-on {
		position: relative;
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: $blue-default;

		&::after {
			content: '';
			position: absolute;
			top: 50%;
			left: 50%;
			width: 6px;
			height: 6px;
			border-radius: 50%;
			background: #fff;
			transform: translate(-50%, -50%);
		}
	}

	&__select--locked {
		pointer-events: none;
		opacity: 0.5;
	}
}
</style>
