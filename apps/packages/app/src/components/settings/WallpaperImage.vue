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
				:fit="imageFit"
				:noSpinner="true"
				:ratio="deviceStore.isMobile ? 6 / 11 : 16 / 9"
				style="border-radius: 4px"
			/>
			<div
				v-if="style === IMG_CONTENT_MODE.Tile"
				:style="{
					width: '300px',
					height: '200px',
					border: '1px solid #eee',
					backgroundImage: `url(${src})`,
					backgroundRepeat: 'repeat',
					backgroundSize: '80px 80px'
				}"
			/>
		</div>
		<div class="row items-center justify-center">
			<slot name="legend" />
		</div>
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
import { useDeviceStore } from 'src/stores/settings/device';
import { IMG_CONTENT_MODE } from 'src/constant';
const deviceStore = useDeviceStore();
import { computed, ref } from 'vue';

const props = defineProps({
	width: Number,
	src: String,
	style: String,
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

const emits = defineEmits(['deleteI']);
const deleteAction = () => {
	emits('deleteI');
};
const imageFit = computed(() => {
	if (props.style === IMG_CONTENT_MODE.Stretch) {
		return 'fill';
	} else if (props.style === IMG_CONTENT_MODE.Fill) {
		return 'cover';
	}
	return 'fill';
});
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
