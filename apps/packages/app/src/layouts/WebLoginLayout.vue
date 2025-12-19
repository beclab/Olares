<template>
	<div class="web-login-layout row justify-center items-center">
		<div class="web-login-root row justify-center items-center">
			<router-view class="web-login-page" />
		</div>
	</div>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
.web-login-layout {
	height: 100vh;
	width: 100vw;
	background: $background-1;

	.web-login-root {
		height: 530px;
		width: 360px;

		.web-login-page {
			height: 100%;
			border-radius: 12px;
			width: 400px;
		}
	}
}
</style>
