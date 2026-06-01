<template>
	<div class="row inline dot-flashing-container">
		<div class="dot-flashing"></div>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
$dotWidth: 4px;
$dotHeight: 4px;
$dotTime: 0.6s;
.dot-flashing-container {
	width: 16px;
	padding-left: 6px;
}
.dot-flashing {
	position: relative;
	width: $dotWidth;
	height: $dotHeight;
	border-radius: 50%;
	background-color: #fff;
	color: #fff;
	animation: dot-flashing $dotTime infinite linear alternate;
	animation-delay: 0.3s;
}
.dot-flashing::before,
.dot-flashing::after {
	content: '';
	display: inline-block;
	position: absolute;
	top: 0;
}
.dot-flashing::before {
	left: -6px;
	width: $dotWidth;
	height: $dotHeight;
	border-radius: 50%;
	background-color: #fff;
	color: #fff;
	animation: dot-flashing $dotTime infinite alternate;
	animation-delay: 0s;
}
.dot-flashing::after {
	left: 6px;
	width: $dotWidth;
	height: $dotHeight;
	border-radius: 50%;
	background-color: #fff;
	color: #fff;
	animation: dot-flashing $dotTime infinite alternate;
	animation-delay: $dotTime;
}

@keyframes dot-flashing {
	0% {
		background-color: #fff;
	}
	50%,
	100% {
		background-color: $background-alpha;
	}
}
</style>
