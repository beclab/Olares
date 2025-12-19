<template>
	<template v-if="tokenStore.deviceInfo.device === DeviceType.MOBILE">
		<div
			class="login row items-center justify-center"
			v-if="!loading && password"
			@click="onLogin"
		>
			<q-img
				src="../../assets/auth/login-icon.svg"
				style="width: 18px; height: 18px"
			/>
		</div>

		<div class="refresh-mobile row items-center justify-center" v-if="loading">
			<q-img
				src="../../assets/auth/login-loading.svg"
				spinner-color="white"
				style="width: 26px; height: 26px"
			/>
		</div>
	</template>

	<template v-else>
		<div class="refresh-pc row items-center justify-center" v-if="loading">
			<q-img
				src="../../assets/auth/progress_activity.svg"
				spinner-color="white"
				style="width: 18px; height: 18px"
			/>
		</div>
	</template>
</template>
<script lang="ts" setup>
import { useTokenStore } from '../../stores/token';
import { DeviceType } from '@bytetrade/core';

defineProps({
	password: {
		type: String,
		required: false,
		default: ''
	},
	loading: {
		type: Boolean,
		required: false,
		default: false
	}
});

const emits = defineEmits(['onLogin']);

const tokenStore = useTokenStore();

const onLogin = async () => {
	emits('onLogin');
};
</script>

<style lang="scss">
.q_field_password {
	.q-field__native {
		font-size: 14px;
		font-family: Roboto-Medium, Roboto;
		font-weight: 500;
		color: #ffffff;
	}
	.q-field__control {
		height: 100% !important;
		padding-left: 24px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.q-field__marginal {
		height: 32px;
	}
}
</style>
<style lang="scss" scoped>
.login-box {
	width: 100%;
	height: 100vh;
	background-position: center;
	background-repeat: no-repeat;
	background-size: 100% 100%;
	position: fixed;
	top: 0px;
	left: 0;
	display: flex;
	align-items: center;
	overflow: hidden;

	.login-card {
		width: 500px;
		margin: 0 auto 50px;
		background-color: transparent;
		box-shadow: none;
		padding-bottom: 56px;
		.column {
			display: flex;
			justify-content: center;
			align-items: center;
			.avatar {
				width: 100px;
				height: 100px;
				border-radius: 62px;
				overflow: hidden;
			}
			.login-name {
				font-size: 24px;
				font-family: Roboto-Bold, Roboto;
				font-weight: bold;
				color: #ffffff;
				margin-top: 30px;
				margin-bottom: 4px;
				text-underline-position: from-font;
				text-decoration-skip-ink: none;
			}
			.login-hint {
				font-size: 14px;
				font-family: Roboto-Regular, Roboto;
				font-weight: 400;
				color: #ffffff;
			}
			.login-btn {
				margin-top: 48px;
			}
		}
	}
}

:global(.text-white) {
	font-size: 16px;
	font-family: Roboto-Medium, Roboto;
	font-weight: 500;
	color: #242424;
}

:global(.q-btn--rectangle) {
	border-radius: 8px;
	width: 120px;
	height: 48px;
}
:global(.q-field--standard .q-field__control:before) {
	border: none;
}
:global(.q-field--standard .q-field__control:after) {
	height: 0px;
}

.login {
	width: 56px;
	height: 56px;
	border-radius: 28px;
	background: rgba(255, 255, 255, 0.98);
	position: absolute;
	left: 0;
	right: 0;
	bottom: -70px;
	margin: auto;
	overflow: hidden;
}

.refresh-mobile {
	width: 56px;
	height: 56px;
	border-radius: 28px;
	background: rgba(31, 24, 20, 0.3);
	animation: rotate-mobile 1s linear infinite;
	position: absolute;
	left: 0;
	right: 0;
	bottom: -70px;
	margin: auto;
	overflow: hidden;
}

@keyframes rotate-mobile {
	0% {
		transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
	}
}

.refresh-pc {
	width: 32px;
	height: 32px;
	border-radius: 16px;
	background: rgba(31, 24, 20, 0.3);
	margin: 0 auto;
	animation: rotate-pc 1s linear infinite;
	position: absolute;
	left: 0;
	right: 0;
	bottom: -40px;
	margin: auto;
	overflow: hidden;
}

@keyframes rotate-pc {
	0% {
		transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
	}
}
</style>
