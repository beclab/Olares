<template>
	<q-item
		dense
		clickable
		@mouseenter="setHover(true)"
		@mouseleave="setHover(false)"
		:active="hoverRef"
		:active-class="activeSoft"
		:style="{ '--active-hover': softRef }"
		class="bt-menu-item full-width row justify-start items-center cursor-pointer"
		@click="emit('onItemClick')"
	>
		<div class="full-width row justify-between items-center">
			<div
				:class="selected ? `${activeText}` : 'text-ink-2'"
				class="row justify-start items-center"
			>
				<q-icon v-if="icon" class="q-mr-sm" size="20px" :name="icon" />
				<div class="text-body3">
					{{ title }}
				</div>
				<slot name="after" :hover="hoverRef" />
			</div>
			<q-icon
				v-if="selected && selectedIcon"
				:class="activeText"
				size="16px"
				name="sym_r_check_circle"
			/>
			<bt-hot-key-icon
				v-else-if="hotkey"
				:hotkey="hotkey"
				:show-board="false"
			/>
		</div>
	</q-item>
</template>

<script lang="ts" setup>
import { PropType, ref } from 'vue';
import { useColor } from '@bytetrade/ui';
import BtHotKeyIcon from 'src/components/base/BtHotKeyIcon.vue';

const props = defineProps({
	title: {
		type: String,
		require: true
	},
	icon: {
		type: String,
		default: ''
	},
	selected: {
		type: Boolean,
		default: false
	},
	selectedIcon: {
		type: Boolean,
		default: true
	},
	hotkey: {
		type: String,
		default: ''
	},
	activeSoft: {
		type: String,
		default: 'orange-soft'
	},
	activeText: {
		type: String,
		default: 'text-orange-default'
	}
});

const hoverRef = ref();
const setHover = (hover: boolean) => {
	hoverRef.value = hover;
};
const emit = defineEmits(['onItemClick']);
const { color: softRef } = useColor(props.activeSoft);
</script>

<style lang="scss" scoped>
.q-list--dense > .q-item,
.q-item--dense {
	min-height: 36px;
	padding: 8px !important;
}

.q-item--active {
	color: transparent !important;
}

.q-item {
	padding: 8px !important;
}

.bt-menu-item {
	border-radius: 4px;

	&:hover {
		background: var(--active-hover);
	}
}
</style>
