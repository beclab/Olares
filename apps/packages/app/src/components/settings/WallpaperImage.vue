<template>
	<div
		class="background-select-bg column justify-start"
		:style="{
			width: `${width + 2 * padding + 2}px`
		}"
	>
		<div
			:style="{
				'--padding': `${padding}px`,
				'--borderWidth': `${borderWidth}px`,
				'--borderRadius': `${borderRadius}px`
			}"
			:class="
				selected ? 'background-mode-item-select' : 'background-mode-item-normal'
			"
			class="row justify-center item-center"
		>
			<q-img
				:src="src"
				height="auto"
				:width="`${width}px`"
				:fit="imgContentModeRef"
				:noSpinner="true"
				:ratio="deviceStore.isMobile ? 6 / 11 : 16 / 9"
				style="border-radius: 4px"
			/>
		</div>
		<div class="row items-center justify-center">
			<slot name="legend" />
		</div>
		{{ deleteEnable }}
		<div
			v-if="deleteEnable"
			class="delete row items-center justify-center"
			@click.stop="deleteAction"
			:style="{
				'--padding': `${padding}px`,
				'--borderWidth': `${borderWidth}px`
			}"
		>
			<q-icon name="sym_r_close" size="16px" color="negative" class="icon" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { imgContentModes } from 'src/constant/index';
import { useDeviceStore } from 'src/stores/settings/device';
const deviceStore = useDeviceStore();

defineProps({
	width: Number,
	src: String,
	padding: {
		type: Number,
		default: 4
	},
	selected: {
		type: Boolean,
		default: false
	},
	borderWidth: {
		type: Number,
		required: false,
		default: 2
	},
	borderRadius: {
		type: Number,
		required: false,
		default: 4
	},
	deleteEnable: {
		type: Boolean,
		required: false,
		default: false
	}
});

const imgContentModeRef = ref(imgContentModes[0]);
const emits = defineEmits(['deleteI']);
const deleteAction = () => {
	emits('deleteI');
};
</script>

<style scoped lang="scss">
.background-select-bg {
	height: auto;
	cursor: pointer;
	text-decoration: none;
	position: relative;

	.background-mode-item-normal {
		border-radius: var(--borderRadius);
		padding: var(--padding);
		border: var(--borderWidth) solid transparent;
	}

	.background-mode-item-select {
		border-radius: var(--borderRadius);
		padding: var(--padding);
		border: var(--borderWidth) solid $blue;
	}

	.delete {
		position: absolute;
		right: -10px;
		top: -10px;
		height: 20px;
		width: 20px;
		opacity: 0;
		visibility: hidden;
		border-radius: 10px;
		overflow: hidden;
		background-color: $background-3;
	}
}

.background-select-bg:hover .delete {
	opacity: 1;
	visibility: visible;
}
</style>
