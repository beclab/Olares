<template>
	<div class="row justify-center items-center select-root">
		<q-btn-dropdown
			class="selected-arrow text-ink-2 bg-background-1"
			ref="dropdown"
			dropdown-icon="sym_r_keyboard_arrow_down"
			no-caps
			flat
			dense
			:menu-offset="[0, 5]"
			:class="{
				'text-body1': !deviceStore.isMobile,
				'text-body3-m': deviceStore.isMobile
			}"
		>
			<template v-slot:label>
				<div class="row justify-center items-center text-body1 text-ink-2">
					<q-img
						class="application-logo"
						no-spinner
						:src="modelValue ? modelValue.app.icon : ''"
					/>
					{{ modelValue ? modelValue.label : '' }}
				</div>
			</template>
			<q-list style="padding: 8px" class="select-items text-background-1">
				<q-item
					v-for="(item, index) in options"
					:key="index"
					:clickable="!item.disable"
					class="select-item-root item"
					:class="item.value === modelValue?.value ? 'bg-background-3' : ''"
					@click="onItemClick(item)"
					v-close-popup
				>
					<q-item-section
						class="text-body1"
						:class="{
							'text-blue-6': !item.disable && item.value === modelValue?.value,
							'text-ink-2': !item.disable && item.value !== modelValue?.value,
							'text-grey-4': item.disable
						}"
					>
						<q-img class="application-logo" no-spinner :src="item.app.icon" />
						{{ item.label }}
					</q-item-section>
				</q-item>
			</q-list>
		</q-btn-dropdown>
	</div>
</template>

<script lang="ts" setup>
import { inject, onMounted, PropType, ref, watch } from 'vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { ApplicationSelectorState } from 'src/constant';

const props = defineProps({
	modelValue: {
		type: Object as PropType<ApplicationSelectorState>,
		require: true
	},
	options: {
		type: Object as PropType<ApplicationSelectorState[]>,
		require: true
	}
});

const selected = ref<ApplicationSelectorState>();
const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
const deviceStore = useDeviceStore();
const dropdown = ref();

onMounted(() => {
	if (setFocused) {
		setFocused(true);
	}
	if (setBlured) {
		setBlured(true);
	}
});

const emit = defineEmits(['update:modelValue']);

const onItemClick = (item: ApplicationSelectorState) => {
	if (!item.disable) {
		emit('update:modelValue', item);
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

.application-logo {
	width: 24px;
	height: 24px;
	margin-right: 4px;
	border-radius: 8px;
}
</style>
