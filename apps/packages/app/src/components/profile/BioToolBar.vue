<template>
	<div class="tool-menu-bar">
		<button
			class="item flex flex-center justify-center cursor-pointer"
			:class="{ selected: selectedItem === 'appearance' }"
			@click="switchItem('appearance')"
		>
			<img
				id="appearance"
				:src="
					selectedItem === 'appearance'
						? '/svg/toolbar-appearance-selected.svg'
						: '/svg/toolbar-appearance.svg'
				"
			/>
		</button>
		<button
			class="item flex flex-center justify-center cursor-pointer"
			:class="{ selected: selectedItem === 'link' }"
			@click="switchItem('link')"
		>
			<img
				id="link"
				:src="
					selectedItem === 'link'
						? '/svg/toolbar-link-selected.svg'
						: '/svg/toolbar-link.svg'
				"
			/>
		</button>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
defineProps({
	selectedItem: String
});
const emits = defineEmits(['switchItem']);
const switchItem = (item: string) => {
	emits('switchItem', item);
};
</script>
<style lang="scss">
.tool-menu-bar {
	width: 5.6rem;
	height: 100vh;
	background: #ffffff;
	.item {
		width: 3.2rem;
		height: 3.2rem;
		margin: 2.4rem auto auto;
		display: flex;
		border: none;
		background: #ffffff;
		q-icon {
			width: 1.6rem;
			height: 1.6rem;
		}
		.img {
			width: 1.6rem;
			height: 1.6rem;
			cursor: pointer;
		}
	}
	.selected {
		background: rgba(70, 102, 246, 0.1) !important;
		border-radius: 4px;
	}
}
</style>
