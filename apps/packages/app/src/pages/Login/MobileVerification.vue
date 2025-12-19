<template>
	<div class="login-box">
		<q-card class="login-card">
			<q-card-section class="column">
				<div
					class="avatar"
					:style="{
						width: `${avatarSize}px`,
						height: `${avatarSize}px`,
						borderRadius: `${avatarSize / 2}px`
					}"
				>
					<TerminusAvatar :info="tokenStore.user" :size="avatarSize" />
				</div>

				<p class="login-name">{{ t('mobile_veri') }}</p>
				<p class="login-hint" @click="toSecondFactor">
					{{ t('login_using_auth') }}
				</p>

				<div class="refresh row items-center justify-center" v-if="loading">
					<q-img
						src="../../assets/auth/progress_activity.svg"
						spinner-color="white"
						style="width: 18px; height: 18px"
					/>
				</div>
			</q-card-section>
		</q-card>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { DeviceType } from '@bytetrade/core';
import { useTokenStore } from '../../stores/token';
import { CurrentView } from '../../utils/constants';

const { t } = useI18n();
const tokenStore = useTokenStore();
const loading = ref(false);
const avatarSize = ref(
	tokenStore.deviceInfo.device === DeviceType.MOBILE ? 100 : 124
);

const toSecondFactor = () => {
	tokenStore.currentView = CurrentView.SECOND_FACTOR;
};

let interval: any = null;

let isInApiCheck = false;
async function checkApiState() {
	if (isInApiCheck) {
		return false;
	}

	if (!tokenStore.requestTermiPass) {
		return;
	}

	try {
		isInApiCheck = true;
		const res = await tokenStore.apiState();
		if (res) {
			loading.value = true;
			await tokenStore.replaceToDesktopUrl(tokenStore.target_url);
			loading.value = false;
		} else {
			isInApiCheck = false;
		}
	} catch (e) {
		console.log(e);
		isInApiCheck = false;
	}
}

onMounted(() => {
	interval = setInterval(async () => {
		await checkApiState();
	}, 500);
});

onUnmounted(() => {
	if (interval) {
		clearInterval(interval);
	}
});
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
	}
	.q-field__marginal {
		height: 32px;
	}
}
</style>
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
			.avatar {
				overflow: hidden;
			}
			.login-name {
				font-size: 24px;
				font-family: Roboto-Bold, Roboto;
				font-weight: bold;
				color: #ffffff;
				margin-top: 50px;
				margin-bottom: 4px;
			}
			.login-hint {
				font-size: 14px;
				font-family: Roboto-Regular, Roboto;
				font-weight: 400;
				color: rgba(255, 255, 255, 0.8);
				cursor: pointer;
				text-decoration: underline;
				text-underline-offset: 2px;
			}
			.login-btn {
				margin-top: 48px;
			}
		}
	}
	.item-margin {
		margin: 5px;
		width: 220px;
		height: 32px !important;
		background: rgba(255, 255, 255, 0.4);
		border-radius: 8px;
		backdrop-filter: blur(20px);
		font-size: 12px;
		padding-left: 12px;
		padding-right: 10px;
		.cursor-pointer {
			color: #ffffff;
		}
	}
}
:global(.q-field__control) {
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
	bottom: -10px;
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
</style>
