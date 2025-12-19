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
					v-html="$t('login.Scan this code with the LarePass app')"
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
import TerminusQrCode from './TerminusQrCode.vue';
import { useAdminStore } from '../../stores/settings/admin';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { Encoder, MessageTopic } from '@bytetrade/core';
import { QR_STATUS } from '../../constant/index';
import { uid } from 'quasar';
import { useTokenStore } from '../../stores/settings/token';
import { bus } from 'src/utils/bus';

const props = defineProps({
	command: {
		type: String,
		required: true
	},
	title: {
		type: String,
		required: true
	},
	body: {
		type: String,
		required: true
	},
	data: {
		type: Object,
		required: false,
		default: {} as any
	}
});

export interface TokenData {
	userid: string;
	token: string;
	expired: number;
}

const { t } = useI18n();
const adminStore = useAdminStore();
const secret = ref<string>('');
const loginUrl = ref<string>('');
const tokenStore = useTokenStore();

onMounted(async () => {
	resetDID();
	bus.on('olaresStatusUpdate', olaresStatusUpdate);
});

onBeforeUnmount(() => {
	bus.off('olaresStatusUpdate', olaresStatusUpdate);
});

const loginQrCodeStatus = ref(QR_STATUS.NORMAL);

function resetDID() {
	//did.value = d;
	secret.value = uid().replace(/-/g, '');
	const time = new Date().getTime();
	const domain =
		process.env.NODE_ENV == 'development'
			? process.env.SETTINGS_URL
			: tokenStore.url;
	const callback_url =
		(process.env.NODE_ENV == 'development'
			? process.env.SETTINGS_URL
			: tokenStore.url) + `/api/command/${props.command}`;
	loginUrl.value =
		'space://' +
		Encoder.stringToBase64Url(
			JSON.stringify({
				topic: MessageTopic.SIGN,
				event: 'olaresd_command',
				notification: {
					body: props.body,
					title: props.title
				},
				message: {
					id: '1',
					data: {
						did: adminStore.terminus.did,
						secret: secret.value,
						time
					},
					sign: {
						callback_url: callback_url,
						sign_body: {
							did: adminStore.terminus.did,
							name: adminStore.olaresId,
							time: `${time}`,
							domain: domain,
							challenge: 'challenge',
							body: {
								...props.data
							}
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

const olaresStatusUpdate = (data: { command: string }) => {
	if (data.command == props.command) {
		emit('success');
	}
};
</script>

<style scoped lang="scss">
.content-bg {
	width: 100%;
	height: 444px;
	border-radius: 12px;
	.q-qr-code-wrapper {
		background: #fff;
		align-self: flex-start;
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
			color: $orange-default;
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
