import { createVCItem, submitPresentation } from '../../vc/vcutils';
import { UserItem, MnemonicItem } from '@didvault/sdk/src/core';
import { getDID, getPrivateJWK } from '../../../../did/did-key';
import { googleTestLogin } from '../../vc/provider/google_test';
import { TERMINUS_VC_TYPE } from '../../../../utils/constants';
import { getEthereumAddress } from '../../../../did/did-key';
import { facebookLogin } from '../../vc/provider/facebook';
import { twitterLogin } from '../../vc/provider/twitter';
import { googleLogin } from '../../vc/provider/google';
import { useUserStore } from '../../../../stores/user';
import { useSSIStore } from '../../../../stores/ssi';
import {
	BusinessAsyncCallback,
	BusinessCallback
} from '../../../../utils/Callback';
import { VCCardInfo } from 'src/utils/vc';
import { i18n } from '../../../../boot/i18n';
import { Submission } from '@bytetrade/core';
import { useMenuStore } from 'src/stores/menu';
import { signJWS } from 'src/layouts/dialog/sign';

export async function requestBindVC(
	request_type: TERMINUS_VC_TYPE,
	domain: string | null,
	callback: BusinessAsyncCallback
) {
	try {
		const userStore = useUserStore();
		const menuStore = useMenuStore();
		const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;

		if (!user) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}

		const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
			userStore.current_id!
		)!;

		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}

		const did = await getDID(mnemonic.mnemonic);
		const privateJWK = await getPrivateJWK(mnemonic.mnemonic);

		if (!did) {
			throw new Error(i18n.global.t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(i18n.global.t('errors.get_privatejwk_failure'));
		}

		let result: VCCardInfo | undefined;

		if (request_type == TERMINUS_VC_TYPE.FACEBOOK) {
			result = await facebookLogin(did, privateJWK);
		} else if (request_type == TERMINUS_VC_TYPE.TWITTER) {
			result = await twitterLogin(did, privateJWK);
		} else if (request_type == TERMINUS_VC_TYPE.GOOGLE) {
			if (process.env.IS_PC_TEST || menuStore.googleTest) {
				result = await googleTestLogin(did, privateJWK, domain);
			} else {
				if (!process.env.IS_BEX) {
					result = await googleLogin(did, privateJWK, domain);
				}
			}
		} else {
			throw new Error(i18n.global.t('errors.request_type_error'));
		}

		if (!result) {
			throw new Error(i18n.global.t('errors.vc_card_info_error'));
		}

		const item = await createVCItem(result);

		if (!item) {
			throw new Error(i18n.global.t('item_is_null'));
		}
		await callback.onSuccess(result);
	} catch (e) {
		let message = e.message;
		if (e.code) {
			message = i18n.global.t('code_message_info', {
				code: e.code,
				message: e.message
			});
		}
		callback.onFailure(message);
	}
}

export const readySubmitPresentation = async (
	chooseCard: VCCardInfo,
	domain: string | null,
	callback: BusinessCallback
) => {
	try {
		if (!chooseCard) {
			throw new Error(i18n.global.t('errors.not_choose_valid_vc_card'));
		}

		if (chooseCard.type && chooseCard.type.toLowerCase() === 'facebook') {
			throw new Error(
				i18n.global.t(
					'errors.facebook_does_not_support_binding_the_olares_id_choose_another_method'
				)
			);
		}

		const userStore = useUserStore();

		const user = userStore.users!.items.get(userStore.current_id!)!;
		if (!user) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}
		if (user.name) {
			throw new Error(i18n.global.t('errors.already_has_olares_id'));
		}
		const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
			userStore.current_id!
		)!;
		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}

		const did = await getDID(mnemonic.mnemonic);
		const privateJWK = await getPrivateJWK(mnemonic.mnemonic);
		const address = await getEthereumAddress(mnemonic.mnemonic);

		if (!did) {
			throw new Error(i18n.global.t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(i18n.global.t('errors.get_privatejwk_failure'));
		}
		if (!address) {
			throw new Error(i18n.global.t('errors.get_address_failure'));
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
			domain
		);
		if (submission && submission.status === 'approved') {
			user.name = submission.reason;
			userStore.users!.items.update(user);
			await userStore.save();
			callback.onSuccess('');
		} else {
			throw new Error(submission.reason);
		}
	} catch (e) {
		callback.onFailure(e.message);
	}
};

export const getBasicTerminusName = async (
	name: string,
	uid: string,
	version: string,
	agent: string,
	callback: BusinessCallback
) => {
	try {
		const ssiStore = useSSIStore();
		const userStore = useUserStore();

		const user = userStore.users!.items.get(userStore.current_id!)!;
		if (!user) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}
		if (user.name) {
			throw new Error(i18n.global.t('errors.already_has_olares_id'));
		}
		const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
			userStore.current_id!
		)!;
		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}

		const did = await getDID(mnemonic.mnemonic);
		const privateJWK = await getPrivateJWK(mnemonic.mnemonic);
		const address = await getEthereumAddress(mnemonic.mnemonic);

		if (!did) {
			throw new Error(i18n.global.t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(i18n.global.t('errors.get_privatejwk_failure'));
		}
		if (!address) {
			throw new Error(i18n.global.t('errors.get_address_failure'));
		}

		const body = {
			name,
			uid,
			version,
			agent,
			address
		};

		const jws = await signJWS(did, body, privateJWK);

		const obj: any = {
			jws: jws
		};

		const response = await ssiStore
			.vcInstance()
			.post('/get_basic_terminus_name/', obj);
		if (
			(response.status != 200 && response.status != 201) ||
			response.data.code != 0 ||
			!response.data.data ||
			!response.data.data.terminus_name
		) {
			throw Error(
				response.data.message
					? response.data.message
					: i18n.global.t('errors.get_response_result_failure')
			);
		}

		user.name = response.data.data.terminus_name;
		userStore.users!.items.update(user);
		await userStore.save();
		callback.onSuccess('');
	} catch (e) {
		callback.onFailure(e.message);
	}
};

export const precheckDidHadBindTerminusName = async () => {
	try {
		const userStore = useUserStore();

		const user = userStore.users!.items.get(userStore.current_id!)!;
		if (!user) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}
		if (user.name) {
			throw new Error(i18n.global.t('errors.already_has_olares_id'));
		}
		const mnemonic: MnemonicItem = userStore.users!.mnemonics.get(
			userStore.current_id!
		)!;
		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.get_user_failure'));
		}

		const ssiStore = useSSIStore();
		const did = await getDID(mnemonic.mnemonic);
		const terminusURL = await ssiStore.get_name_by_did(did);

		if (terminusURL === null || terminusURL === undefined) {
			return false;
		}

		if (terminusURL == '[local_user]') {
			return false;
		}
		const name = terminusURL.replace('.', '@');
		user.name = name;
		userStore.users!.items.update(user);
		await userStore.save();
		return true;
	} catch (error) {
		return false;
	}
};

export const basicTerminusNameMinLength = 8;

export const basicTerminusNameMaxLength = 63;

export const basicTerminusNameRule = '^(?=[a-z0-9]{8,64}$)[a-z0-9]+$';
