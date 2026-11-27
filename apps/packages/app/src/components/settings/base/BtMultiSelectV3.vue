<template>
	<div
		class="multi-selected-root"
		ref="rootRef"
		:style="{ '--width': width, '--height': height }"
	>
		<div
			class="multi-selected-control row items-center no-wrap"
			:class="{ 'multi-selected-control--error': isError }"
		>
			<div class="multi-selected-chips row items-center">
				<template v-if="selectedOptions.length > 0">
					<div
						v-for="opt in selectedOptions"
						:key="String(opt.value)"
						class="multi-selected-chip text-body3"
					>
						<span class="multi-selected-chip-label">{{ opt.label }}</span>
						<q-icon
							name="sym_r_close"
							size="14px"
							class="multi-selected-chip-remove"
							@click.stop="removeValue(opt.value)"
						/>
					</div>
				</template>
				<div v-else class="multi-selected-placeholder text-body3 text-ink-3">
					{{ placeholder }}
				</div>
			</div>
			<q-icon
				class="multi-selected-arrow text-ink-2"
				name="sym_r_keyboard_arrow_down"
				size="20px"
			/>
			<q-menu fit anchor="bottom left" self="top left" :offset="[0, 5]">
				<div class="multi-select-card">
					<template v-for="item in options" :key="item.value">
						<div
							:class="
								item.disable
									? `multi-select-item-disable ${inputClass ?? 'text-body2'}`
									: isSelected(item.value)
									? `multi-select-item-selected ${inputClass ?? 'text-body2'}`
									: `multi-select-item-normal ${inputClass ?? 'text-body2'}`
							"
							class="multi-select-item-row"
							@click="onItemClick(item)"
						>
							<span class="multi-select-item-label">{{
								item.hideLabel ? item.value : item.label
							}}</span>
							<q-icon
								v-if="isSelected(item.value)"
								name="sym_r_check_circle"
								class="multi-select-item-check"
								size="18px"
							/>
						</div>
					</template>
				</div>
			</q-menu>
		</div>
	</div>
	<div
		class="text-overline q-mt-xs text-red"
		v-if="errorMessage.length > 0 && isError"
	>
		{{ errorMessage }}
	</div>
</template>

<script lang="ts" setup>
import { computed, inject, onMounted, PropType, ref } from 'vue';
import { SelectorProps } from 'src/constant';

const props = defineProps({
	modelValue: {
		type: Array as PropType<string[]>,
		default: () => []
	},
	options: {
		type: Object as PropType<SelectorProps[]>,
		required: true
	},
	inputClass: {
		type: String,
		required: false
	},
	isError: {
		type: Boolean,
		default: false,
		required: false
	},
	errorMessage: {
		type: String,
		default: '',
		required: false
	},
	width: {
		type: String,
		default: '100%',
		required: false
	},
	height: {
		type: String,
		default: '36px',
		required: false
	},
	placeholder: {
		type: String,
		default: '',
		required: false
	}
});

const emit = defineEmits(['update:modelValue']);

const rootRef = ref();
const setFocused = inject<(focused: boolean) => void>('setFocused');
const setBlured = inject<(blured: boolean) => void>('setBlured');

onMounted(() => {
	if (rootRef.value) {
		rootRef.value.addEventListener('focus', () => setFocused?.(true));
		rootRef.value.addEventListener('blur', () => setBlured?.(true));
	}
});

const selectedOptions = computed<SelectorProps[]>(() =>
	props.modelValue.map((v) => {
		const found = props.options?.find((o) => o.value === v);
		return found ?? { label: v, value: v };
	})
);

function isSelected(value: SelectorProps['value']) {
	return props.modelValue.includes(String(value));
}

const onItemClick = (item: SelectorProps) => {
	if (item.disable) {
		return;
	}
	const value = String(item.value);
	const next = props.modelValue.includes(value)
		? props.modelValue.filter((v) => v !== value)
		: [...props.modelValue, value];
	emit('update:modelValue', next);
};

const removeValue = (value: SelectorProps['value']) => {
	const target = String(value);
	emit(
		'update:modelValue',
		props.modelValue.filter((v) => v !== target)
	);
};
</script>

<style scoped lang="scss">
.multi-selected-root {
	width: var(--width);
}

.multi-selected-control {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	min-height: var(--height);
	width: 100%;
	color: $ink-2;
	padding: 2px 4px 2px 8px;
	box-sizing: border-box;
	cursor: pointer;

	&--error {
		border-color: $negative;
	}
}

.multi-selected-chips {
	flex: 1 1 auto;
	min-width: 0;
	gap: 4px;
	flex-wrap: wrap;
	padding: 2px 0;
}

.multi-selected-placeholder {
	line-height: 28px;
}

.multi-selected-chip {
	display: flex;
	align-items: center;
	gap: 4px;
	max-width: 100%;
	padding: 2px 6px;
	border-radius: 4px;
	background: $background-3;
	color: $ink-2;

	.multi-selected-chip-label {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.multi-selected-chip-remove {
		flex-shrink: 0;
		cursor: pointer;
		color: $ink-3;

		&:hover {
			color: $ink-1;
		}
	}
}

.multi-selected-arrow {
	flex-shrink: 0;
	width: 20px;
	height: 20px;
}

.multi-select-card {
	width: 100%;
	display: flex;
	padding: 12px;
	flex-direction: column;
	align-items: flex-start;
	gap: 4px;
	background: $background-2;
	color: $ink-2;

	.multi-select-item-title {
		width: 100%;
		height: 34px;
		padding: 8px 0;
		border-radius: 4px;
		text-align: left;
		color: $ink-2;
	}

	.multi-select-item-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		gap: 8px;
		height: 34px;
		padding: 8px 0;
		border-radius: 4px;

		.multi-select-item-label {
			flex: 1;
			min-width: 0;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.multi-select-item-check {
			flex-shrink: 0;
			color: $blue-default;
		}
	}

	.multi-select-item-normal {
		cursor: pointer;
		text-decoration: none;

		&:hover {
			background: $background-hover !important;
		}
	}

	.multi-select-item-disable {
		background: darkgray;
		color: grey;
	}

	.multi-select-item-selected {
		cursor: pointer;
		text-decoration: none;
		color: $orange-default;

		&:hover {
			background: $background-hover !important;
		}
	}
}
</style>
