<template>
	<div class="login-box">
		<q-card class="login-card">
			<q-card-section class="column">
				<OneTimePasswordMethod
					ref="onetimeRef"
					:passwordErr="passwordErr"
					:digits="optConfig.digits"
					:period="optConfig.period"
					@handleOnComplete="handleOnComplete"
				/>

				<div class="refresh row items-center justify-center" v-if="loading">
					<q-img
						src="../../../assets/auth/progress_activity.svg"
						spinner-color="white"
						style="width: 18px; height: 18px"
					/>
				</div>
			</q-card-section>
		</q-card>
	</div>
</template>
<script lang="ts">
import {
	defineComponent,
	ref,
	onUnmounted,
	nextTick,
	getCurrentInstance,
	ComponentInternalInstance
} from 'vue';
import { Token } from '@bytetrade/core';
import { useTokenStore } from '../../../stores/token';
import OneTimePasswordMethod from './OneTimePasswordMethod.vue';

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
		const oneTimePasswordMethod = ref();
		const passwordErr = ref(false);
		const instance = getCurrentInstance() as ComponentInternalInstance | null;
		const proxy = instance?.proxy as any & { clearInput: () => void };
		const loading = ref(false);

		const handleOnComplete = (value: any) => {
			oneTimePasswordMethod.value = value;
			onLogin();
		};

		const onLogin = async () => {
			loading.value = true;
			try {
				const data: Token = await tokenStore.secondFactor(
					oneTimePasswordMethod.value
				);

				await tokenStore.replaceToDesktopUrl(data.redirect);
			} catch (err) {
				passwordErr.value = true;
				handleClearInput();
				setTimeout(() => {
					passwordErr.value = false;
				}, 2000);
			} finally {
				loading.value = false;
			}
		};

		const handleClearInput = () => {
			nextTick(() => {
				oneTimePasswordMethod.value = null;
				proxy.$refs['onetimeRef'].clearInput();
			});
		};

		onUnmounted(() => {
			handleClearInput();
		});

		return {
			userInfo,
			optConfig,
			passwordErr,
			loading,
			tokenStore,
			onLogin,
			handleOnComplete,
			handleClearInput
		};
	}
});
</script>
<style lang="scss" scoped>
.login-box {
	width: 100%;
	height: 100%;
	background-position: center;
	background-repeat: no-repeat;
	background-size: 100% 100%;
	position: fixed;
	top: 0px;
	left: 0;
	display: flex;
	align-items: center;

	.login-card {
		height: 300px;
		margin: 0 auto 10vh;
		background-color: transparent;
		box-shadow: none;
		.column {
			display: flex;
			justify-content: center;
			align-items: center;

			.refresh {
				width: 32px;
				height: 32px;
				border-radius: 16px;
				background: rgba(31, 24, 20, 0.3);
				margin: 0 auto;
				animation: rotate 1s linear infinite;
				position: absolute;
				left: 0;
				right: 0;
				bottom: -40px;
				margin: auto;
				overflow: hidden;
			}

			@keyframes rotate {
				0% {
					transform: rotate(0deg);
				}
				100% {
					transform: rotate(360deg);
				}
			}
		}
	}
}
</style>
