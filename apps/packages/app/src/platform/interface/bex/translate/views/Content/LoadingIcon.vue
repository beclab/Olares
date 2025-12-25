<template>
	<div class="olares-loading-container">
		<div class="olares-loading-inner"></div>
	</div>
</template>

<script setup></script>

<style scoped>
.olares-loading-container {
	display: inline-flex;
	justify-content: center;
	align-items: flex-end;
}
.olares-loading-inner {
	width: 12px;
	height: 12px;
	border: 2px solid #fff;
	border-radius: 50%;
	border-top-color: grey;
	transform-origin: 50% 50%;
	animation: olares-rotation 1s linear infinite;
}
@keyframes olares-rotation {
	0% {
		transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
	}
}
</style>
