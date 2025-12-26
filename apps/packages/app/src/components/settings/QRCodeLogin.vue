<template>
	<div class="content-bg column items-center justify-center">
		<slot name="mode" />
		<div
			class="row justify-center items-center text-body-1 q-mt-lg q-qr-code-wrapper"
		>
			<div
				class="login-dsc"
				:class="[
					loginQrCodeStatus !== QR_STATUS.EXPIRED &&
					loginQrCodeStatus !== QR_STATUS.SUCCESSFUL
						? 'login-dsc-wrapper'
						: ''
				]"
			>
				<span
					v-if="loginQrCodeStatus === QR_STATUS.EXPIRED"
					class="text-negative"
				>
					{{ t('login.qr_code_expired_refresh') }}
				</span>
				<span
					v-else-if="loginQrCodeStatus === QR_STATUS.SUCCESSFUL"
					class="text-positive"
				>
					{{ t('login.scan_successful') }}
				</span>
				<div
					v-else
					class="text-ink-2 q-px-xl"
					v-html="$t('login.Scan this code with the LarePass app to log in')"
				></div>
			</div>
			<q-img class="login-mask" src="settings/login_mask.png" />
		</div>
		<div class="cloud-qr-code q-pa-md q-mt-xl">
			<terminus-qr-code
				:url="loginUrl"
				:size="216"
				:status="loginQrCodeStatus"
				text-style="text-subtitle3"
				@on-refresh="resetDID"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusQrCode from 'src/components/settings/TerminusQrCode.vue';
import { useAdminStore } from 'src/stores/settings/admin';
import { useAccountStore } from 'src/stores/settings/account';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { Encoder, MessageTopic } from '@bytetrade/core';
import { QR_STATUS } from 'src/constant/index';
import { uid } from 'quasar';
import axios from 'axios';

export interface TokenData {
	userid: string;
	token: string;
	expired: number;
}

const { t } = useI18n();
const adminStore = useAdminStore();
const accountStore = useAccountStore();
const secret = ref<string>('');
const loginUrl = ref<string>('');
let user_info_interval: NodeJS.Timeout | null = null;

onMounted(async () => {
	resetDID();
	user_info_interval = setInterval(async () => {
		await updateInfo();
	}, 3 * 1000);
});

onBeforeUnmount(() => {
	if (user_info_interval) {
		clearInterval(user_info_interval);
	}
});

watch(
	() => adminStore.terminus.did,
	async () => {
		resetDID();
	}
);

const loginQrCodeStatus = ref(QR_STATUS.NORMAL);

async function updateInfo() {
	try {
		const result = await axios.post(
			accountStore.space.url + '/v2/user/activeLogin',
			{
				secret: secret.value
			}
		);

		const token: TokenData = result.data;
		if (!token.userid) {
			return;
		}

		if (user_info_interval) {
			clearInterval(user_info_interval);
		}
		// const saveData: SpaceSaveData = {
		// 	email: '',
		// 	token: token.token,
		// 	userid: token.userid,
		// 	expired: token.expired
		// };

		await accountStore.createAccount(adminStore.olaresId, {
			refresh_token: token.token,
			access_token: token.token,
			expires_in: 60 * 30 * 1000,
			expires_at: token.expired,
			userid: adminStore.terminus.did
		});

		await accountStore.listSecret();
		loginQrCodeStatus.value = QR_STATUS.SUCCESSFUL;
		setTimeout(() => {
			emit('success');
		}, 1000);
	} catch (e) {
		console.log(e);
	}
}

function resetDID() {
	//did.value = d;
	secret.value = uid().replace(/-/g, '');
	const time = new Date().getTime();
	loginUrl.value =
		'space://' +
		Encoder.stringToBase64Url(
			JSON.stringify({
				topic: MessageTopic.SIGN,
				event: 'login_cloud',
				message: {
					id: '1',
					data: {
						did: adminStore.terminus.did,
						secret: secret.value,
						time
					},
					sign: {
						callback_url: accountStore.space.url + '/v2/user/login',
						sign_body: {
							did: adminStore.terminus.did,
							secret: secret.value,
							time
						}
					}
				}
			})
		);
	loginQrCodeStatus.value = QR_STATUS.NORMAL;
	setTimeout(() => {
		loginQrCodeStatus.value = QR_STATUS.EXPIRED;
	}, 120 * 1000);
}

const emit = defineEmits(['success']);
</script>

<style scoped lang="scss">
.content-bg {
	width: 100%;
	height: 444px;
	border-radius: 12px;
	.q-qr-code-wrapper {
		background: #fff;
		// align-self: flex-start;
		z-index: 10;
		.login-mask {
			position: absolute;
			top: 105px;
			left: 40px;
			right: 0;
			display: none;
			width: 448px;
		}
		.login-dsc {
			position: absolute;
			left: 0;
			right: 0;
			text-align: center;
			// bottom: 32px;
			z-index: 1;
		}
		::v-deep(.login-dsc-wrapper:hover + .login-mask) {
			cursor: pointer;
			display: block;
		}
		::v-deep(.login-highlight) {
			color: $orange-default;
			cursor: pointer;
		}
	}
}

.cloud-qr-code {
	border: 1px solid $separator;
	border-radius: 10px;
	width: 240px;
	height: 240px;
	background-color: white;
	z-index: 0;
}
</style>
