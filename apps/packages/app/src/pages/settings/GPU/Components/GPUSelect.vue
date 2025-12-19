<template>
	<div
		class="row items-center justify-between select-root selected-arrow text-body2 text-ink-1"
		:class="border ? `items-border ${classes}` : classes"
		:style="{ '--height': height + 'px' }"
	>
		<div class="row items-center">
			<div class="q-mr-sm">
				<q-img
					src="settings/imgs/root/gpu.svg"
					style="border-radius: 8px"
					:width="iconSize + 'px'"
					:height="iconSize + 'px'"
				/>
			</div>
			<div>{{ selected ? selected.label : '' }}</div>
		</div>

		<q-icon
			v-if="options && options.length > 0"
			:name="menuShow ? 'sym_r_keyboard_arrow_up' : 'sym_r_keyboard_arrow_down'"
			size="24px"
		/>
		<q-menu
			v-if="options && options.length > 0"
			class="bg-background-2"
			fit
			:offset="offset"
			v-model="menuShow"
			:class="menuClasses"
		>
			<q-list>
				<q-item
					v-for="(item, index) in options"
					:key="index"
					:clickable="!item.disable"
					class="select-item-root"
					:class="item.value === selected?.value ? 'bg-background-3' : ''"
					:style="{ '--menuItemHeight': menuItemHeight + 'px' }"
					@click="onItemClick(item)"
					v-close-popup
				>
					<q-item-section class="item-margin-left">
						<div class="row items-center">
							<div style="position: relative" :style="'alpha: 0.5'">
								<q-img
									src="settings/imgs/root/gpu.svg"
									style="border-radius: 8px"
									:width="iconSize + 'px'"
									:height="iconSize + 'px'"
								/>
							</div>
							<div
								class="application-name text-body1"
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
							</div>
						</div>
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
import { inject, onMounted, PropType, ref, watch } from 'vue';
import { SelectorProps } from 'src/constant';

const props = defineProps({
	modelValue: {
		type: [String],
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
	height: {
		type: Number,
		default: 64,
		required: false
	},
	iconSize: {
		type: Number,
		default: 24,
		required: false
	},
	classes: {
		type: String,
		required: false,
		default: 'q-px-lg'
	},
	menuClasses: {
		type: String,
		required: false,
		default: 'q-pa-md'
	},
	menuItemHeight: {
		type: Number,
		default: 48,
		required: false
	}
});

const selected = ref<SelectorProps>();
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

const onItemClick = (item: SelectorProps) => {
	if (!item.disable) {
		emit('update:modelValue', item.value);
	}
};

const menuShow = ref(false);
</script>

<style scoped lang="scss">
.selected-arrow {
	height: var(--height, 40px);

	border-radius: 12px;
	background: $background-1;
	&:hover {
		background: $background-3;
	}
}

.items-border {
	border: solid 1px $separator;
}

.select-item-root {
	height: var(--menuItemHeight, 48px);
	border-radius: 4px;
	padding: 8px 8px 8px 12px;

	.application-logo {
		border-radius: 8px;
	}
	.application-name {
		margin-left: 8px;
	}
}
</style>
