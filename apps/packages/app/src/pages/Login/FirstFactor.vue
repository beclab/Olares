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
				<p class="login-name">{{ username }}</p>
				<p class="login-hint">
					{{ t('login_title') }}
				</p>
				<div class="item-margin" :class="passwordErr ? 'shake' : ''">
					<div class="input-wrap">
						<input
							class="text-white"
							:class="loading ? 'disable' : ''"
							type="password"
							id="password"
							v-model="password"
							ref="loginRef"
							:disabled="loading"
							@keydown="onkeydown"
							:style="{ fontSize: inputFontSize }"
						/>
					</div>

					<label v-if="!password" class="placeholder-label">{{
						t('login_hint_password')
					}}</label>
					<q-icon
						v-if="
							password && tokenStore.deviceInfo.device !== DeviceType.MOBILE
						"
						class="cursor-pointer animated fadeIn"
						name="sym_r_arrow_circle_right"
						@click="onLogin"
						size="20px"
					/>
				</div>

				<login-action
					:password="password"
					:loading="loading"
					@on-login="onLogin"
				/>
			</q-card-section>
		</q-card>
	</div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useTokenStore } from '../../stores/token';
import { CurrentView } from '../../utils/constants';
import { DeviceType } from '@bytetrade/core';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import LoginAction from './LoginAction.vue';

const { t } = useI18n();
const tokenStore = useTokenStore();
const password = ref();
const loginRef = ref();
const loading = ref(false);
const passwordErr = ref(false);
const username =
	tokenStore.olaresId &&
	tokenStore.olaresId.slice(0, tokenStore.olaresId.indexOf('@'));
const avatarSize = ref(
	tokenStore.deviceInfo.device === DeviceType.MOBILE ? 100 : 124
);
const inputFontSize = ref(
	tokenStore.deviceInfo.device == DeviceType.DESKTOP ? '22px' : '18px'
);

const onLogin = async () => {
	loading.value = true;
	try {
		const res = await tokenStore.login(
			tokenStore.olaresId.split('@')[0],
			password.value
		);

		console.log('res ===>', res);

		if (res.fa2) {
			tokenStore.currentView = CurrentView.MOBILE_VERIFICATION;
		} else {
			await tokenStore.replaceToDesktopUrl(res.redirect);
		}
	} catch (err) {
		console.log('error--->', err);

		if (err.response) {
			BtNotify.show({
				type: NotifyDefinedType.MESSAGE,
				message:
					typeof err.response.data == 'string'
						? err.response.data
						: err.response.data?.message
						? err.response.data.message
						: ''
			});
		} else {
			BtNotify.show({
				type: NotifyDefinedType.MESSAGE,
				message: err.message
			});
		}

		shakeInput();
		handleClearInput();
	} finally {
		loading.value = false;
	}
};

const handleClearInput = () => {
	loading.value = false;
	setTimeout(() => {
		loginRef.value.focus();
		loginRef.value.select();
	}, 1200);
};

const shakeInput = () => {
	passwordErr.value = true;
	setTimeout(() => {
		passwordErr.value = false;
	}, 1000);
};

const onkeydown = async (e: any) => {
	if (e.keyCode !== 13) return false;
	if (e.keyCode === 13 && !password.value) {
		shakeInput();
		handleClearInput();
	}

	onLogin();
};
</script>

<style lang="scss" scoped>
input[type='password']::-ms-reveal {
	display: none;
}
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
				margin-top: 16px;
				margin-bottom: 4px;
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
	.item-margin {
		margin: 5px;
		width: 220px;
		height: 32px !important;
		background: rgba(255, 255, 255, 0.4);
		border-radius: 8px;
		backdrop-filter: blur(20px);
		padding-left: 12px;
		padding-right: 10px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		box-sizing: border-box;
		position: relative;

		.input-wrap {
			width: 100%;
			height: 18px;
			overflow: hidden;
			display: flex;
			align-items: center;
			justify-content: flex-start;
		}

		.cursor-pointer {
			color: #ffffff;
		}

		.placeholder-label {
			position: absolute;
			left: 14px;
			top: 50%;
			transform: translateY(-50%);
			color: rgba(255, 255, 255, 0.5);
			pointer-events: none;
		}

		input {
			width: 180px;
			background-color: transparent;
			border: none;
			padding: 0;
			margin: 0;
			caret-color: #ffffff;
			letter-spacing: 1px;
			height: 28px;
			line-height: 28px;
			font-weight: 900;

			&.disable {
				pointer-events: none;
			}
			&:focus {
				outline: none;
				box-shadow: none;
			}
			&::placeholder {
				font-size: 12px;
				line-height: 20px;
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
</style>
