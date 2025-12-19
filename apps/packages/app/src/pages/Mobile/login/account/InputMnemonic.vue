<template>
	<div class="input-mnemonic-root column justify-start items-center">
		<terminus-title-bar v-if="userStore.current_user" />
		<terminus-scroll-area
			class="input-mnemonic-scroll"
			:class="
				keyboardOpen
					? userStore.current_user
						? 'scroll-area-conf-header-open'
						: 'scroll-area-conf-open'
					: userStore.current_user
					? 'scroll-area-conf-header-close'
					: 'scroll-area-conf-close'
			"
		>
			<template v-slot:content>
				<div class="input-mnemonic-page column justify-start items-center">
					<q-img
						class="input-mnemonic-page__image"
						:src="getRequireImage('login/import_terminus_mnemonic.svg')"
					/>
					<span class="input-mnemonic-page__desc login-sub-title">{{
						t('Enter the 12-word mnemonic phrase to import your Olares ID')
					}}</span>
					<q-checkbox
						v-if="isLocalTest"
						v-model="use_local"
						@update:model-value="checkboxChange"
						>local server</q-checkbox
					>
					<terminus-mnemonics-component
						ref="mnemonicRef"
						class="input-mnemonic-page__mnemonic q-mb-lg"
						@on-mnemonic-change="mnemonicUpdate"
					/>
				</div> </template
		></terminus-scroll-area>
		<confirm-button
			class="input-mnemonic-root-button"
			:btn-title="btnText"
			@onConfirm="onConfirm"
			:btn-status="btnStatusRef"
		/>
		<q-inner-loading :showing="loading" dark color="white" size="64px">
		</q-inner-loading>
	</div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import TerminusMnemonicsComponent from '../../../../components/common/TerminusMnemonicsComponent.vue';
import ConfirmButton from '../../../../components/common/ConfirmButton.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import { useI18n } from 'vue-i18n';
import { parsingMnemonics } from './ImportUserBusiness';
import { getRequireImage } from '../../../../utils/imageUtils';
import MonitorKeyboard from '../../../../utils/monitorKeyboard';
import { importUser } from '../../../../utils/BindTerminusBusiness';
import { useUserStore } from '../../../../stores/user';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import {
	notifyFailed,
	notifyWarning
} from '../../../../utils/notifyRedefinedUtil';
import { onUnmounted } from 'vue';
import { TerminusDefaultDomain, TerminusInfo } from '@bytetrade/core';
import { getOlaresInfo as getBaseOlaresInfo } from '../../../../utils/account';

const $q = useQuasar();
const router = useRouter();
const mnemonic = ref<string>('');
const { t } = useI18n();
const keyboardOpen = ref(false);
let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const use_local = ref(false);
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);
const userStore = useUserStore();
const isLocalTest = computed(() => {
	return process.env.IS_PC_TEST;
});
const btnText = ref(
	isLocalTest.value && use_local.value ? t('skip') : t('next')
);
const loading = ref(false);

const checkboxChange = () => {
	btnText.value = isLocalTest.value && use_local.value ? t('skip') : t('next');
};

onMounted(() => {
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

async function onConfirm() {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	loading.value = true;
	btnStatusRef.value = ConfirmButtonStatus.disable;
	await parsingMnemonics(mnemonic.value, {
		async onSuccess(data: any) {
			if (data) {
				if (process.env.IS_BEX) {
					try {
						const array: string[] = data.split('@');
						let baseUrl = '';
						if (array.length == 2) {
							baseUrl = 'https://auth.' + array[0] + '.' + array[1] + '/';
						} else {
							baseUrl =
								'https://auth.' + array[0] + '.' + TerminusDefaultDomain + '/';
						}
						const info: TerminusInfo | null = await getOlaresInfo(baseUrl);
						if (info && info.wizardStatus == 'completed') {
							await importUser(data, mnemonic.value);
							router.push({ path: '/connectLoading' });
						} else {
							throw Error(
								t(
									'Unable to connect with Olares. It could be a network issue or your Olares needs to be activated. '
								)
							);
						}
					} catch (e) {
						// $q.loading.hide();
						loading.value = true;
						btnStatusRef.value = ConfirmButtonStatus.normal;
						notifyFailed(e.message);
					}
					return;
				}
				try {
					await importUser(data, mnemonic.value);
					router.replace({ path: '/connectLoading' });
				} catch (e) {
					btnStatusRef.value = ConfirmButtonStatus.normal;
					loading.value = false;
					notifyFailed(e.message);
				}
			} else {
				if (process.env.IS_BEX) {
					btnStatusRef.value = ConfirmButtonStatus.normal;
					loading.value = false;
					notifyWarning(t('bex_import_mnemonics_not_active_reminder'));
					return;
				}

				try {
					await importUser(null, mnemonic.value);
					loading.value = true;
					btnStatusRef.value = ConfirmButtonStatus.normal;
					router.replace({ path: '/BindTerminusName' });
				} catch (error) {
					notifyFailed(error.message);
				}
			}
		},
		onFailure(message: string) {
			notifyFailed(message);
			loading.value = false;
			btnStatusRef.value = ConfirmButtonStatus.normal;
		}
	});
	loading.value = false;
}

const mnemonicUpdate = (value: string) => {
	mnemonic.value = value;
	const masterPasswordArray = mnemonic.value.split(' ');
	if (
		masterPasswordArray.length !== 12 ||
		masterPasswordArray.findIndex((e) => e.length === 0) >= 0
	) {
		btnStatusRef.value = ConfirmButtonStatus.disable;
		return;
	}
	btnStatusRef.value = ConfirmButtonStatus.normal;
};

async function getOlaresInfo(baseUrl: string): Promise<TerminusInfo | null> {
	try {
		return await getBaseOlaresInfo(baseUrl);
	} catch (e) {
		return null;
	}
}
</script>

<style lang="scss" scoped>
.input-mnemonic-root {
	width: 100%;
	height: 100%;

	.input-mnemonic-scroll {
		width: 100%;

		.input-mnemonic-page {
			width: 100%;
			height: 100%;
			padding-left: 20px;
			padding-right: 20px;

			&__image {
				width: 160px;
				height: 160px;
				margin-top: 32px;
			}

			&__desc {
				margin-top: 20px;
			}

			&__mnemonic {
				margin-top: 32px;
			}
		}
	}

	.input-mnemonic-root-button {
		width: calc(100% - 40px);
	}
}
</style>
