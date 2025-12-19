<template>
	<div class="desktop-unlock-layout">
		<DesktopDefaultHeaderView />
		<div
			style="width: 100%; height: calc(100% - 76px)"
			class="row items-center justify-center"
		>
			<div class="desktop-unlock-root row justify-center items-center">
				<router-view class="desktop-unlock-page" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import DesktopDefaultHeaderView from '../../../components/DesktopDefaultHeaderView.vue';
</script>

<style scoped lang="scss">
.desktop-unlock-layout {
	height: 100vh;
	width: 100vw;
	background-color: $background-3;

	.desktop-unlock-root {
		padding: 52px;

		.desktop-unlock-page {
			height: 100%;
			border-radius: 12px;
			width: 480px;
		}
	}
}
</style>
