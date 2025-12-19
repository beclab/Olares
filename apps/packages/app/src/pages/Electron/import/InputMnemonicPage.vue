<template>
	<div class="input-mnemonic-page">
		<PadInputContentScrollView>
			<terminus-page-title
				:center="true"
				class="page-title"
				:label="t('Import an account')"
				:desc="t('Enter the 12-word mnemonic phrase to import your Olares ID')"
			/>
			<terminus-mnemonics-component
				ref="mnemonicRef"
				:lager-title="false"
				:row-capacity="4"
				class="input-mnemonic-page__mnemonic"
				@on-mnemonic-change="mnemonicUpdate"
				@keyup.enter="onConfirm"
			/>
		</PadInputContentScrollView>

		<q-btn
			icon="sym_r_arrow_back"
			class="input-mnemonic-page__back btn-no-text btn-no-border btn-size-sm"
			flat
			dense
			@click="onReturn"
			v-if="userStore.current_user"
		/>

		<confirm-button
			class="input-mnemonic-page__button"
			:btn-title="t('next')"
			@onConfirm="onConfirm"
			:btn-status="btnStatusRef"
		/>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import TerminusMnemonicsComponent from '../../../components/common/TerminusMnemonicsComponent.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { useI18n } from 'vue-i18n';
import { parsingMnemonics } from '../../Mobile/login/account/ImportUserBusiness';
import TerminusPageTitle from '../../../components/common/TerminusPageTitle.vue';
import { importUser } from '../../../utils/BindTerminusBusiness';
import { TerminusDefaultDomain, TerminusInfo } from '@bytetrade/core';
import PadInputContentScrollView from '../../../components/ios/PadInputContentScrollView.vue';
import {
	notifyWarning,
	notifyFailed
} from '../../../utils/notifyRedefinedUtil';
import { useUserStore } from '../../../stores/user';
import { getOlaresInfo as getBaseOlaresInfo } from '../../../utils/account';

const $q = useQuasar();
const router = useRouter();
const mnemonic = ref<string>('');
const userStore = useUserStore();
const { t } = useI18n();

async function onConfirm() {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	btnStatusRef.value = ConfirmButtonStatus.disable;
	$q.loading.show();
	await parsingMnemonics(mnemonic.value, {
		async onSuccess(data: any) {
			if (data) {
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
								'Unable to connect with Olares. It could be a network issue or your Olares needs to be activated.'
							)
						);
					}
				} catch (e) {
					$q.loading.hide();
					notifyFailed(e.message);
				}
			} else {
				notifyWarning(
					t(
						'The corresponding DID for this mnemonic phrase does not have a corresponding Olares ID, so it cannot be imported. Please try importing it using a mobile device and apply for an Olares ID according to the prompts.'
					)
				);
			}
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
	btnStatusRef.value = ConfirmButtonStatus.normal;
	$q.loading.hide();
}

const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);

const mnemonicUpdate = (value: string) => {
	mnemonic.value = value;
	const masterPasswordArray = mnemonic.value.split(' ');
	if (
		masterPasswordArray.length !== 12 ||
		masterPasswordArray.find((e) => e.length === 0) !== undefined
	) {
		btnStatusRef.value = ConfirmButtonStatus.disable;
		return;
	}
	btnStatusRef.value = ConfirmButtonStatus.normal;
};

const onReturn = () => {
	router.go(-1);
};

async function getOlaresInfo(baseUrl: string): Promise<TerminusInfo | null> {
	try {
		return getBaseOlaresInfo(baseUrl);
	} catch (e) {
		return null;
	}
}
</script>

<style lang="scss" scoped>
.input-mnemonic-page {
	width: 100%;
	height: 100%;
	background: $background-2;
	padding-top: 20px;
	padding-left: 32px;
	padding-right: 32px;
	position: relative;

	&__mnemonic {
		margin-top: 32px;
	}

	&__button {
		position: absolute;
		bottom: 52px;
		width: calc(100% - 64px);
		left: 32px;
		right: 32px;
	}

	&__back {
		position: absolute;
		top: 20px;
		left: 20px;
	}
}
</style>
