<template>
	<div class="factor-box">
		<q-img
			fit="fill"
			class="desktop-bg"
			:src="`auth/${tokenStore.user.loginBackground}`"
		/>
		<q-card class="factor-card row items-center justify-center">
			<OneTimePasswordMethod
				ref="onetimeRef"
				:digits="optConfig.digits"
				:period="optConfig.period"
				@handleOnComplete="handleOnComplete"
			/>

			<div
				class="login-btn"
				:class="passwordErr ? 'errShock' : ''"
				@click="onLogin"
			>
				Submit
			</div>
		</q-card>
	</div>
</template>
<script lang="ts">
import {
	defineComponent,
	ref,
	onMounted,
	onUnmounted,
	nextTick,
	getCurrentInstance,
	ComponentInternalInstance
} from 'vue';
import { Notify, Loading } from 'quasar';
import { useTokenStore } from '../../../stores/token';
import OneTimePasswordMethod from './OneTimePasswordMethod.vue';
import { Token } from '@bytetrade/core';

export default defineComponent({
	components: {
		OneTimePasswordMethod
	},
	setup() {
		const tokenStore = useTokenStore();
		const userInfo = ref();
		const optConfig = ref({
			digits: 6,
			period: 30
		});
		const passwordErr = ref(false);
		const oneTimePasswordMethod = ref();
		const instance = getCurrentInstance() as ComponentInternalInstance | null;
		const proxy = instance?.proxy as any & {
			clearInput: () => void;
		};

		const handleOnComplete = (value: any) => {
			oneTimePasswordMethod.value = value;
		};

		const onLogin = async () => {
			Loading.show();

			try {
				const data: Token = await tokenStore.secondFactor(
					oneTimePasswordMethod.value
				);
				if (data.redirect) {
					await tokenStore.replaceToDesktopUrl(data.redirect);
					handleClearInput();
				}
			} catch (err) {
				passwordErr.value = true;
				handleClearInput();
				setTimeout(() => {
					passwordErr.value = false;
				}, 2000);
				Notify.create({
					type: 'negative',
					message: (err as Error).message
				});
			} finally {
				Loading.hide();
			}
		};

		const handleClearInput = () => {
			nextTick(() => {
				oneTimePasswordMethod.value = null;
				proxy.$refs['onetimeRef'].clearInput();
			});
		};

		const keydownEnter = (event: any) => {
			if (event.keyCode !== 13) return false;
			if (
				!oneTimePasswordMethod.value ||
				(oneTimePasswordMethod.value && oneTimePasswordMethod.value.length < 6)
			) {
				return Notify.create({
					type: 'warning',
					message: 'Please enter the Verification Code.',
					timeout: 800
				});
			}
			onLogin();
		};

		onMounted(() => {
			window.addEventListener('keydown', keydownEnter);
		});

		onUnmounted(() => {
			window.removeEventListener('keydown', keydownEnter);
		});

		return {
			userInfo,
			optConfig,
			passwordErr,
			tokenStore,
			onLogin,
			handleOnComplete,
			handleClearInput
		};
	}
});
</script>
<style lang="scss">
.factor-box {
	width: 100vw;
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	background-position: center;
	background-repeat: no-repeat;
	background-size: 100% 100%;

	.desktop-bg {
		width: 100%;
		height: 100%;
		position: absolute;
		left: 0;
		right: 0;
		top: 0;
		bottom: 0;
	}
	.factor-card {
		width: 480px;
		padding: 20px;
		box-shadow: none;
		background: none;
		.login-btn {
			width: 120px;
			height: 48px;
			line-height: 48px;
			text-align: center;
			font-size: 16px;
			font-weight: 500;
			color: #242424;
			background: #ffffff;
			box-shadow: 0px 2px 12px 0px rgba(0, 0, 0, 0.2);
			opacity: 1;
			border-radius: 8px;
		}
	}
}
</style>
