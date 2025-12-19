<template>
	<div class="selected-root row justify-center items-center" ref="rootRef">
		<div class="text-body1 selected-title" :class="inputClass">
			{{
				selected ? (selected.hideLabel ? selected.value : selected.label) : ''
			}}
		</div>
		<q-btn-dropdown
			class="selected-arrow"
			ref="dropdown"
			dropdown-icon="img:./settings/arrow.svg"
			size="10px"
			flat
			:menu-offset="[15, 10]"
			dense
		>
			<div class="select-card" :style="{ '--selectedWidth': selectedWidth }">
				<template v-for="(item, index) in options" :key="item.value">
					<div
						:class="
							item.disable
								? `select-item-disable ${inputClass ?? 'text-body2'}`
								: selected && item.value === selected.value
								? `select-item-selected ${inputClass ?? 'text-body2'}}`
								: `select-item-normal ${inputClass ?? 'text-body2'}}`
						"
						v-close-popup
						:style="{ marginTop: `${index === 0 ? '0' : '4px'}` }"
						@click="onItemClick(item)"
					>
						{{ item.hideLabel ? item.value : item.label }}
					</div>
				</template>
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
		require: true
	},
	inputClass: {
		type: String,
		require: false
	},
	isError: {
		type: Boolean,
		default: false,
		require: false
	},
	errorMessage: {
		type: String,
		default: '',
		require: false
	}
});

const selected = ref<SelectorProps>();
const rootRef = ref();
const dropdown = ref();
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const setFocused = inject('setFocused') as any;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const setBlured = inject('setBlured') as any;

onMounted(() => {
	if (setFocused) {
		setFocused(true);
	}
	if (setBlured) {
		setBlured(true);
	}
});

const selectedWidth = computed(() => {
	return rootRef.value ? rootRef.value.offsetWidth + 'px' : 0;
});

watch(
	() => [props.modelValue, props.options],
	() => {
		if (props.options && props.options.length > 0) {
			selected.value = props.options?.find((e) => e.value == props.modelValue);
		}
	},
	{
		immediate: true
	}
);

const emit = defineEmits(['update:modelValue']);

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
	height: 36px;
	width: 100%;
	color: $ink-2;

	.selected-title {
		width: calc(100% - 54px);
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
	background: #fff;
	color: $ink-2;

	.select-item-title {
		width: 100%;
		height: 34px;
		padding: 8px 0;
		border-radius: 4px;
		text-align: left;
		color: $ink-2;
	}

	.select-item-normal {
		@extend .select-item-title;
		cursor: pointer;
		text-decoration: none;

		&:hover {
			background: #f5f5f5 !important;
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
			background: #f5f5f5 !important;
		}
	}
}
</style>
