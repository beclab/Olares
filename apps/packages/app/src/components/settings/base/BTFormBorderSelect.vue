<template>
	<div>
		<div class="text-body3 text-ink-3 q-mb-xs" v-if="title">
			{{ title }}
		</div>
		<div
			class="row items-center justify-between select-root selected-arrow text-body1 text-ink-2"
			:class="border ? 'items-border' : ''"
		>
			<div>{{ selected ? selected.label : '' }}</div>
			<q-icon name="sym_r_keyboard_arrow_down" size="20px" />
			<q-menu class="bg-background-2 q-pa-sm" fit :offset="offset">
				<q-list>
					<q-item
						v-for="(item, index) in options"
						:key="index"
						:clickable="item.enable"
						class="select-item-root"
						:class="item.value === selected?.value ? 'bg-background-3' : ''"
						@click="onItemClick(item)"
						v-close-popup
					>
						<q-item-section
							class="text-body2"
							:class="
								item.enable
									? item.value === selected?.value
										? color
										: 'text-ink-2'
									: 'text-grey-4'
							"
						>
							{{ item.label }}
						</q-item-section>
						<q-item-section side>
							<q-icon
								name="sym_r_check_circle"
								size="18px"
								:class="color"
								v-show="item.value === selected?.value"
							/>
						</q-item-section>
					</q-item>
				</q-list>
			</q-menu>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { inject, onMounted, PropType, ref, watch } from 'vue';

interface SelectData {
	label: string;
	value: string | number;
	enable: boolean;
}

const props = defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	modelValue: {
		type: [String, Number],
		require: true
	},
	options: {
		type: Object as PropType<SelectData[]>,
		require: true
	},
	border: {
		type: Boolean,
		default: false,
		required: false
	},
	offset: {
		type: Array,
		default: () => [0, 0],
		required: false
	},
	color: {
		type: String,
		default: 'text-blue-6'
	}
});

const selected = ref<SelectData>();
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

watch(
	() => props.modelValue,
	() => {
		selected.value = props.options?.find((e) => e.value == props.modelValue);
	},
	{
		immediate: true
	}
);

const emit = defineEmits(['update:modelValue']);

const onItemClick = (item: SelectData) => {
	if (item.enable) {
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
	height: 40px;
	padding-left: 20px;
	padding-right: 20px;
	border-radius: 8px;
	background: $background-1;
	&:hover {
		background: $background-3;
	}
}

.items-border {
	border: solid 1px $separator;
}

.select-item-root {
	height: 40px;
	min-height: 40px;
	border-radius: 4px;
	padding: 8px 8px 8px 12px;
}
</style>
