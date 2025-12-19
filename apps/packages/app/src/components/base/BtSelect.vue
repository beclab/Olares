<template>
	<div
		class="row items-center justify-between selected-arrow text-body2 text-ink-1"
		:class="{
			'items-border': border,
			'select-disable': disable
		}"
	>
		<div>{{ selected ? selected.label : '' }}</div>
		<q-icon name="sym_r_keyboard_arrow_down" size="20px" />
		<q-menu
			class="bg-background-2 q-pa-sm"
			fit
			:offset="offset"
			v-model="menuDisplay"
			v-if="!disable"
		>
			<q-list>
				<q-item
					v-for="(item, index) in options"
					:key="index"
					:clickable="!item.disable"
					class="select-item-root"
					:class="item.value === selected?.value ? 'bg-background-3' : ''"
					@click="onItemClick(item)"
					v-close-popup
				>
					<q-item-section
						class="text-body2"
						:class="
							!item.disable
								? item.value === selected?.value
									? color
									: item.titleClass
									? item.titleClass
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
</template>

<script lang="ts" setup>
import { computed, inject, onMounted, PropType, ref, watch } from 'vue';
import { SelectorProps } from 'src/constant';

const props = defineProps({
	modelValue: {
		type: [String, Number],
		require: true
	},
	options: {
		type: Object as PropType<SelectorProps[]>,
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
	},
	disable: {
		type: Boolean,
		required: false,
		default: false
	}
});

const selected = ref<SelectorProps>();
const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
const menuStatus = ref(false);

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

watch(
	() => props.options,
	() => {
		selected.value = props.options?.find((e) => e.value == props.modelValue);
	},
	{
		immediate: true
	}
);

const menuDisplay = computed({
	get: () => menuStatus.value,
	set: (value) => {
		menuStatus.value = value;
		emit('update:menu', value);
	}
});

const emit = defineEmits(['update:modelValue', 'update:menu']);

const onItemClick = (item: SelectorProps) => {
	if (!item.disable) {
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
	padding-left: 10px;
	padding-right: 10px;
	border-radius: 8px;
	background: $background-1;
	&:hover {
		background: $background-3;
	}
}

.select-disable {
	background-color: $background-3;
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
