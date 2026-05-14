<template>
	<div
		class="selected-root row justify-center items-center"
		ref="rootRef"
		:style="{ '--width': width, '--height': height }"
	>
		<div class="text-body1 selected-title" :class="inputClass">
			{{
				selected
					? selected.hideLabel
						? selected.value
						: selected.label
					: fallbackLabel ?? ''
			}}
		</div>
		<q-btn-dropdown
			class="selected-arrow text-ink-2"
			ref="dropdown"
			dropdown-icon="sym_r_keyboard_arrow_down"
			size="10px"
			flat
			:menu-offset="[15, 10]"
			dense
		>
			<div
				class="select-card"
				:class="{ 'select-card--show-selected': showSelectedIcon }"
				:style="{ '--selectedWidth': selectedWidth }"
			>
				<div
					v-if="$slots.header"
					:class="[
						headerIsSelected
							? `select-item-selected ${inputClass ?? 'text-body2'}`
							: `select-item-normal ${inputClass ?? 'text-body2'}`,
						'select-item-title',
						'select-item-row'
					]"
					v-close-popup
					@click="onHeaderClick"
				>
					<span class="select-item-label">
						<slot name="header" />
					</span>
					<q-icon
						v-if="showSelectedIcon && headerIsSelected"
						name="sym_r_check_circle"
						class="select-item-check"
						size="18px"
					/>
				</div>
				<template v-for="(item, index) in options" :key="item.value">
					<div
						:class="
							item.disable
								? `select-item-disable ${inputClass ?? 'text-body2'}`
								: selected && item.value === selected.value
								? `select-item-selected ${inputClass ?? 'text-body2'}`
								: `select-item-normal ${inputClass ?? 'text-body2'}`
						"
						class="select-item-row"
						v-close-popup
						:style="{
							marginTop: index === 0 && !$slots.header ? '0' : '4px'
						}"
						@click="onItemClick(item)"
					>
						<span class="select-item-label">{{
							item.hideLabel ? item.value : item.label
						}}</span>
						<q-icon
							v-if="
								showSelectedIcon && selected && item.value === selected.value
							"
							name="sym_r_check_circle"
							class="select-item-check"
							size="18px"
						/>
					</div>
				</template>
				<div
					v-if="$slots.footer"
					class="select-item-normal select-item-title"
					:class="inputClass ?? 'text-body2'"
					style="margin-top: 4px"
					v-close-popup
					@click="onFooterClick"
				>
					<slot name="footer" />
				</div>
			</div>
		</q-btn-dropdown>
	</div>
	<div
		class="text-overline q-mt-xs text-red"
		v-if="errorMessage.length > 0 && isError"
	>
		{{ errorMessage }}
	</div>
</template>

<script lang="ts" setup>
import { computed, inject, onMounted, PropType, ref, watch } from 'vue';
import { SelectorProps } from 'src/constant';

const props = defineProps({
	modelValue: {
		type: String,
		require: true
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
	fallbackLabel: {
		type: String,
		default: '',
		required: false
	},
	showSelectedIcon: {
		type: Boolean,
		default: false,
		required: false
	},
	headerValue: {
		type: String,
		default: '',
		required: false
	}
});

const selected = ref<SelectorProps>();
const rootRef = ref();
const dropdown = ref();
const setFocused = inject<(focused: boolean) => void>('setFocused');
const setBlured = inject<(blured: boolean) => void>('setBlured');

onMounted(() => {
	if (rootRef.value) {
		rootRef.value.addEventListener('focus', () => setFocused?.(true));
		rootRef.value.addEventListener('blur', () => setBlured?.(true));
	}
});

const selectedWidth = computed(() => {
	return rootRef.value ? `${rootRef.value.offsetWidth}px` : '0px';
});

const headerIsSelected = computed(
	() =>
		!!(
			props.showSelectedIcon &&
			props.headerValue &&
			props.modelValue === props.headerValue
		)
);

watch(
	() => [props.modelValue, props.options],
	() => {
		if (props.options && props.options.length > 0) {
			selected.value = props.options?.find((e) => e.value === props.modelValue);
		}
	},
	{
		immediate: true
	}
);

const emit = defineEmits(['update:modelValue', 'header-click', 'footer-click']);

const onHeaderClick = () => {
	emit('header-click');
};

const onFooterClick = () => {
	emit('footer-click');
};

const onItemClick = (item: SelectorProps) => {
	if (!item.disable) {
		selected.value = item;
		emit('update:modelValue', item.value);
	}
};
</script>

<style scoped lang="scss">
.selected-root {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	height: var(--height);
	width: var(--width);
	color: $ink-2;

	.selected-title {
		width: calc(var(--width) - 54px);
		margin-right: 8px;
		overflow: hidden;
		color: $ink-2;
		text-overflow: ellipsis;
	}

	.selected-arrow {
		width: 20px;
		height: 20px;
	}
}

.select-card {
	width: var(--selectedWidth);
	display: flex;
	padding: 12px;
	flex-direction: column;
	align-items: flex-start;
	gap: 4px;
	background: $background-2;
	color: $ink-2;

	.select-item-title {
		width: 100%;
		height: 34px;
		padding: 8px 0;
		border-radius: 4px;
		text-align: left;
		color: $ink-2;
	}

	.select-item-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		gap: 8px;

		.select-item-label {
			flex: 1;
			min-width: 0;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.select-item-check {
			flex-shrink: 0;
			color: $orange-default;
		}
	}

	.select-item-normal {
		@extend .select-item-title;
		cursor: pointer;
		text-decoration: none;

		&:hover {
			background: $background-hover !important;
		}
	}

	.select-item-disable {
		background: darkgray;
		color: grey;
		@extend .select-item-title;
	}

	.select-item-selected {
		@extend .select-item-title;
		cursor: pointer;
		text-decoration: none;

		&:hover {
			background: $background-hover !important;
		}
	}

	&--show-selected .select-item-selected {
		color: $orange-default;
	}
}
</style>
