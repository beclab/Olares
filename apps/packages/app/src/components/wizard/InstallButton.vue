<template>
	<div class="row install_btn_bg">
		<div
			class="percent"
			:style="{
				'--percent': `${progress}%`
			}"
		></div>
		<div
			class="install-button row items-center justify-center text-subtitle2 text-white"
			:class="progress <= 40 ? 'text-light-blue-default' : ''"
		>
			{{ status }}
		</div>
	</div>
</template>
<script setup lang="ts">
defineProps({
	progress: {
		type: Number,
		default: 42,
		required: false
	},
	status: {
		type: String,
		default: '',
		required: false
	}
});
</script>

<style lang="scss" scoped>
.install_btn_bg {
	width: 100%;
	height: 32px;
	border-radius: 8px;
	background-color: $light-blue-soft;
	position: relative;

	.pre-percent {
		// background-color: $light-blue-default;
		height: 100%;
		position: absolute;
		left: 0;
		top: 0;
		width: var(--percent);
		border-top-left-radius: 8px;
		border-bottom-left-radius: 8px;
	}

	.percent {
		background-color: $light-blue-default;
		height: 100%;
		position: absolute;
		left: 0;
		top: 0;
		width: var(--percent);
		border-top-left-radius: 8px;
		border-bottom-left-radius: 8px;
	}

	.install-button {
		width: 100%;
		height: 100%;
		position: absolute;
		// mix-blend-mode: difference;
	}
}
</style>
