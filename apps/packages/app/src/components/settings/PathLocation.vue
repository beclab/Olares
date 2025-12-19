<template>
	<div class="path-location-root row items-center">
		<div class="path-location-root__start row justify-start items-center">
			<slot name="start" />
		</div>

		<div class="path-location-root__midden row justify-end items-center">
			<slot name="midden" />
		</div>

		<div class="path-location-root__end row justify-end items-center">
			<slot name="end" />
		</div>
	</div>
</template>

<script lang="ts" setup></script>

<style scoped lang="scss">
.path-location-root {
	width: 100%;
	height: auto;

	&__start {
		width: 50%;
		margin-right: 20px;
	}

	&__midden {
		margin-right: 20px;
		width: 16%;
	}

	&__end {
		width: 26%;
	}
}
</style>
