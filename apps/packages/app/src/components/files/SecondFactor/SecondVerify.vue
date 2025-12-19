<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('login')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="small"
		@onSubmit="onLogin"
	>
		<OneTimePasswordMethod
			ref="onetimeRef"
			:digits="infitotp.digits"
			:period="infitotp.period"
			@handleOnChange="handleOnChange"
		/>
	</bt-custom-dialog>
</template>
<script lang="ts" setup>
import { ref, onMounted, onUnmounted, getCurrentInstance } from 'vue';
import { useRouter } from 'vue-router';
import { Loading } from 'quasar';
import OneTimePasswordMethod from './OneTimePasswordMethod.vue';
import { useSecondVerifyStore } from '../../../stores/second';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	request: {
		type: Object,
		required: true
	}
});

const CustomRef = ref();

const router = useRouter();
const secondVerifyStore = useSecondVerifyStore();
const infitotp = ref({
	digits: 6,
	period: 30
});
const oneTimePasswordMethod = ref();
const passwordErr = ref(false);
const { proxy } = getCurrentInstance() as any;

const { t } = useI18n();

const handleOnChange = (value: any) => {
	oneTimePasswordMethod.value = value;
};

const onLogin = async () => {
	if (
		!oneTimePasswordMethod.value ||
		(oneTimePasswordMethod.value && oneTimePasswordMethod.value?.length < 6)
	) {
		return false;
	}

	Loading.show();
	const responseURL = props.request.responseURL;
	try {
		const res: any = await secondVerifyStore.cert_secondfactor_totp(
			oneTimePasswordMethod.value,
			responseURL
		);

		let redirect = res?.data?.data?.redirect;
		if (redirect) {
			let path = '/files' + redirect.slice(redirect.indexOf('resources') + 9);
			router.push({ path });
			CustomRef.value.onDialogOK({ redirect });
		}
	} catch (err) {
		passwordErr.value = true;
		setTimeout(() => {
			passwordErr.value = false;
		}, 2000);
		notifyFailed((err as Error).message);
	} finally {
		Loading.hide();
		await handleClearInput();
	}
};

const handleClearInput = () => {
	oneTimePasswordMethod.value = null;
	proxy.$refs['onetimeRef'].clearInput();
};

const keydownEnter = (event: any) => {
	if (event.keyCode !== 13) return false;
	onLogin();
};

onMounted(() => {
	window.addEventListener('keydown', keydownEnter);
});

onUnmounted(() => {
	window.removeEventListener('keydown', keydownEnter);
});
</script>
<style lang="scss">
.factor-box {
	display: flex;
	align-items: center;
	justify-content: center;

	.factor-card {
		width: 480px;
		padding: 20px;
		box-shadow: none;
		background-color: $background-1;
		border-radius: 20px;

		.login-btn {
			width: 120px;
			height: 48px;
			line-height: 48px;
			text-align: center;
			color: $ink-1;
			background: $white;
			box-shadow: 0px 2px 12px 0px $grey-2;
			opacity: 1;
			border-radius: 8px;
		}

		.errShock {
			animation-delay: 0s;
			animation-name: shock;
			animation-duration: 0.1s;
			animation-iteration-count: 3;
			animation-direction: normal;
			animation-timing-function: linear;
		}

		@keyframes shock {
			0% {
				margin-left: 0px;
				margin-right: 5px;
			}
			100% {
				margin-left: 5px;
				margin-right: 0px;
			}
		}
	}
}
</style>
