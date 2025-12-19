<template>
	<div class="select-terminus-name">
		<back-title-bar :title="t('select_terminus_name')" />
		<q-scroll-area class="select-terminus-name__scroll">
			<div class="select-terminus-name__layout">
				<terminus-item
					v-for="item in vcsRef"
					:key="item.id"
					@click="readySubmitPresentation(item)"
					class="select-terminus-name__item"
					:whole-picture-size="36"
					:image-path="getTerminusLogo(item)"
				>
					<template v-slot:title>{{ getName(item) }}</template>
				</terminus-item>
			</div>
		</q-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import BackTitleBar from '../../../components/common/BackTitleBar.vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { submitPresentation } from '../vc/vcutils';
import { Submission } from '@bytetrade/core';
import { onMounted, ref } from 'vue';
import {
	getDID,
	getEthereumAddress,
	getPrivateJWK
} from '../../../did/did-key';
import { useUserStore } from '../../../stores/user';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import { getTerminusName, getVCCardItemList } from '../items/VCItemUtil';
import { useI18n } from 'vue-i18n';
import {
	notifyFailed,
	notifySuccess
} from '../../../utils/notifyRedefinedUtil';
import { VCCardInfo, VCCardItem } from '../../../utils/vc';

const userStore = useUserStore();
const $q = useQuasar();
const vcsRef = ref();
const router = useRouter();
const { t } = useI18n();

onMounted(async () => {
	vcsRef.value = getVCCardItemList();
});

function getTerminusLogo(chooseCard: VCCardInfo) {
	const path = 'secret/';
	switch (chooseCard.type) {
		case 'Google':
			return path + 'google_logo.svg';
		case 'Twitter':
			return path + 'twitter_logo.svg';
		case 'Facebook':
			return path + 'facebook_logo.svg';
		default:
			return '';
	}
}

function getName(card: VCCardItem) {
	let name = getTerminusName(card);
	if (!name) {
		name = t('not_support_bind');
	}
	return name;
}

async function readySubmitPresentation(chooseCard: VCCardInfo) {
	try {
		$q.loading.show();

		if (!chooseCard) {
			throw new Error(t('errors.not_choose_valid_vc_card'));
		}

		if (chooseCard.type && chooseCard.type.toLowerCase() === 'facebook') {
			throw new Error(
				t(
					'errors.facebook_does_not_support_binding_the_olares_id_choose_another_method'
				)
			);
		}

		const user = userStore.users!.items.get(userStore.current_id!)!;
		if (!user) {
			throw new Error(t('errors.get_user_failure'));
		}
		if (user.name) {
			throw new Error(t('errors.already_has_olares_id'));
		}
		const did = await getDID(userStore.current_mnemonic?.mnemonic);
		const privateJWK = await getPrivateJWK(
			userStore.current_mnemonic?.mnemonic
		);
		const address = await getEthereumAddress(
			userStore.current_mnemonic?.mnemonic
		);
		if (!did) {
			throw new Error(t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(t('errors.get_privatejwk_failure'));
		}
		if (!address) {
			throw new Error(t('errors.get_address_failure'));
		}

		const submission: Submission = await submitPresentation(
			chooseCard.type,
			did,
			privateJWK,
			chooseCard.verifiable_credential.substring(
				0,
				chooseCard.verifiable_credential.length
			),
			address,
			null
		);
		if (submission && submission.status === 'approved') {
			user.name = submission.reason;
			await userStore.users!.items.update(user);
			await userStore.save();
			$q.loading.hide();
			notifySuccess(t('bind_success'));
			router.back();
		} else {
			throw new Error(submission.reason);
		}
	} catch (e) {
		$q.loading.hide();
		notifyFailed(e.message);
	}
}
</script>

<style scoped lang="scss">
.select-terminus-name {
	height: 100%;
	width: 100%;

	&__scroll {
		height: 100%;
		width: 100%;
		padding-left: 24px;
		padding-right: 24px;
	}

	&__item {
		margin-top: 16px;
	}

	&__layout {
		width: 100%;
		height: 100%;
	}
}
</style>
