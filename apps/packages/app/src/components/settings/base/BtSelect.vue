<template>
	<div class="row justify-center items-center select-root">
		<q-btn-dropdown
			class="selected-arrow text-ink-2"
			ref="dropdown"
			dropdown-icon="sym_r_keyboard_arrow_down"
			:label="selected ? selected.label : ''"
			no-caps
			flat
			dense
			:menu-offset="[0, 5]"
			:class="{
				'text-body1': !textClass && !deviceStore.isMobile,
				'text-body3-m': !textClass && deviceStore.isMobile,
				textClass: !!textClass,
				'bg-background-1': !transparent
			}"
		>
			<q-list style="padding: 8px" class="select-items text-background-1">
				<q-item
					v-for="(item, index) in options"
					:key="index"
					:clickable="!item.disable"
					class="select-item-root item"
					:class="item.value === selected?.value ? 'bg-background-3' : ''"
					@click="onItemClick(item)"
					v-close-popup
				>
					<q-item-section
						class="text-body1"
						:class="{
							'text-blue-6': !item.disable && item.value === selected?.value,
							'text-ink-2': !item.disable && item.value !== selected?.value,
							'text-grey-4': item.disable
						}"
					>
						{{ item.label }}
					</q-item-section>
					<q-item-section side>
						<q-icon
							name="sym_r_check_circle"
							size="18px"
							class="text-blue-6"
							v-show="item.value === selected?.value"
						/>
					</q-item-section>
				</q-item>
			</q-list>
		</q-btn-dropdown>
	</div>
</template>

<script lang="ts" setup>
import { inject, onMounted, PropType, ref, watch } from 'vue';
import { useDeviceStore } from 'src/stores/settings/device';
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
	transparent: {
		type: Boolean,
		default: false
	},
	textClass: {
		type: String,
		default: ''
	}
});

const selected = ref<SelectorProps>();
const dropdown = ref();
const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
const deviceStore = useDeviceStore();

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
	border-radius: 8px;
}

.select-item-selected {
	color: $blue-6;
}

.select-item-root {
	height: 40px;
	min-height: 40px;
	border-radius: 4px;
	padding: 8px 8px 8px 12px;
}

.select-items .item:hover {
	background: $background-hover;
}
</style>
