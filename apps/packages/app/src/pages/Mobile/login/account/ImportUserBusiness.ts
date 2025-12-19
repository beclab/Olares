import { getDID } from '../../../../did/did-key';
import { useUserStore } from '../../../../stores/user';
import { useSSIStore } from '../../../../stores/ssi';
import { BusinessAsyncCallback } from '../../../../utils/Callback';
import { i18n } from '../../../../boot/i18n';

export async function parsingMnemonics(
	mnemonic: string,
	callback: BusinessAsyncCallback
) {
	try {
		if (!mnemonic) {
			throw new Error(i18n.global.t('errors.please_enter_the_mnemonic_phrase'));
		}

		const array = mnemonic.split(' ');
		if (array.length != 12) {
			throw new Error(
				i18n.global.t(
					'errors.the_format_of_the_mnemonic_phrase_does_not_meet_the_requirements'
				)
			);
		}

		const did = await getDID(mnemonic);
		const ssiStore = useSSIStore();
		const terminusURL = await ssiStore.get_name_by_did(did);
		if (terminusURL === null) {
			throw new Error(
				i18n.global.t(
					'errors.currently_the_mnemonic_phrase_corresponding_to_the_did_is_not_bound_to_any_olares_id_on_the_blockchain'
				)
			);
		}

		if (terminusURL === undefined) {
			throw new Error(
				i18n.global.t(
					'errors.therea_is_a_network_issue_with_the_service_providing_did_query_please_try_importing_again_later'
				)
			);
		}

		const name =
			terminusURL == '[local_user]' ? '' : terminusURL.replace('.', '@');

		// if (name.length == 0) {
		// 	throw new Error(i18n.global.t('errors.olares_info_not_found'));
		// }

		const userStore = useUserStore();
		const user = userStore.users!.items.get(did);
		if (user) {
			throw new Error(
				i18n.global.t(
					'errors.the_did_corresponding_to_the_current_mnemonic_phrase_already_exists_in_larePass'
				)
			);
		}
		userStore.temp_import_data.mnemonic = mnemonic;
		userStore.temp_import_data.terminusName = name;

		await callback.onSuccess(name);
	} catch (error) {
		console.error(error);
		callback.onFailure(error.message);
	}
}
