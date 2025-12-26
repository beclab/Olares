<template>
	<div class="row justify-center items-center">
		<div v-if="selected" class="text-body2 selected-title">
			{{ selected.label }}
		</div>
		<q-btn-dropdown
			class="selected-arrow"
			ref="dropdown"
			dropdown-icon="img:/settings/arrow.svg"
			size="10px"
			flat
			dense
		>
			<q-list style="padding: 8px">
				<template v-for="(item, index) in options" :key="item">
					<div
						class="text-body2"
						:class="
							!item.disable
								? item.value === selected.value
									? 'select-item-selected'
									: 'select-item-normal'
								: 'select-item-disable'
						"
						v-close-popup
						:style="{ marginTop: `${index === 0 ? '0' : '4px'}` }"
						@click="onItemClick(item)"
					>
						{{ item.label }}
					</div>
				</template>
			</q-list>
		</q-btn-dropdown>
	</div>
</template>

<script lang="ts" setup>
import { inject, onMounted, PropType, ref, watch } from 'vue';
import { SelectorProps } from '../../../constant';

const props = defineProps({
	modelValue: {
		type: String,
		require: true
	},
	options: {
		type: Object as PropType<SelectorProps[]>,
		require: true
	}
});

const selected = ref<SelectorProps>();
const dropdown = ref();
const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;

onMounted(() => {
	if (setFocused) {
		setFocused(true);
	}
	if (setBlured) {
		setBlured(true);
	}
});

const emit = defineEmits(['update:modelValue']);

watch(
	() => props.modelValue,
	() => {
		selected.value = props.options?.find((e) => e.value == props.modelValue);
	},
	{
		immediate: true
	}
);

const onItemClick = (item: SelectorProps) => {
	if (!item.disable) {
		selected.value = item;
		emit('update:modelValue', item.value);
	}
};
</script>

<style scoped lang="scss">
.selected-title {
	margin-right: 8px;
	text-align: right;
	color: $ink-1;
}

.selected-arrow {
	width: 20px;
	height: 20px;
}

.select-item-title {
	width: 134px;
	height: 24px;
	padding: 2px 8px;
	border-radius: 4px;
}

.select-item-normal {
	color: $ink-1;
	@extend .select-item-title;
	cursor: pointer;
	text-decoration: none;

	&:hover {
		background: $background-3;
	}
}

.select-item-disable {
	background: $separator;
	color: $ink-2;
	@extend .select-item-title;
}

.select-item-selected {
	background: $blue;
	@extend .select-item-title;
	cursor: pointer;
	text-decoration: none;
}
</style>
