<template>
	<div class="second-verification-page column justify-between items-center">
		<one-time-password-method
			class="second-verification-page__fa"
			ref="onetimeRef"
			@handleOnComplete="handleOnComplete"
			@handle-on-change="handleOnComplete"
			@keyup.enter="onConfirm"
		/>
		<confirm-button
			class="second-verification-page__button"
			:btn-title="t('confirm')"
			@onConfirm="onConfirm"
			:btn-status="btnStatusRef"
		/>
	</div>
</template>

<script setup lang="ts">
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { getCurrentInstance, nextTick, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { useQuasar } from 'quasar';
import { importUserCheckFa } from '../../../utils/BindTerminusBusiness';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../../stores/user';
import OneTimePasswordMethod from '../../../components/common/OneTimePasswordMethod.vue';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

const oneTimePasswordMethodRef = ref();
const { proxy } = getCurrentInstance() as any;
const { t } = useI18n();
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const $q = useQuasar();
const userStore = useUserStore();
const router = useRouter();

onMounted(() => {
	const terminusName = userStore.temp_import_data.terminusName;
	const mnemonic = userStore.temp_import_data.mnemonic;
	const token = userStore.temp_import_data.token;
	const osName = userStore.temp_import_data.osName;
	if (!mnemonic || !terminusName || !token || !osName) {
		router.replace({ path: '/import_mnemonic' });
		return;
	}
});

const handleOnComplete = (value: any) => {
	oneTimePasswordMethodRef.value = value;
	btnStatusRef.value = value
		? ConfirmButtonStatus.normal
		: ConfirmButtonStatus.disable;
};

const onConfirm = async () => {
	const terminusName = userStore.temp_import_data.terminusName;
	const mnemonic = userStore.temp_import_data.mnemonic;
	const token = userStore.temp_import_data.token;
	const osName = userStore.temp_import_data.osName;
	const localServer = userStore.temp_import_data.localServer;
	if (!mnemonic || !terminusName || !token || !osName) {
		router.replace({ path: '/import_mnemonic' });
		return;
	}

	$q.loading.show();

	await importUserCheckFa(
		terminusName,
		token,
		oneTimePasswordMethodRef.value,
		mnemonic,
		localServer,
		{
			async onSuccess() {
				router.push({ path: '/home' });
			},
			onFailure(message: string) {
				handleClearInput();
				notifyFailed(message);
			}
		}
	);

	$q.loading.hide();
};

const handleClearInput = () => {
	nextTick(() => {
		oneTimePasswordMethodRef.value = null;
		proxy.$refs['onetimeRef'].clearInput();
	});
};
</script>

<style scoped lang="scss">
.second-verification-page {
	width: 100%;
	height: 100%;

	&__fa {
		width: 100%;
		margin-top: 20px;
	}

	&__button {
		width: calc(100% - 64px);
		margin-bottom: 52px;
	}
}
</style>
