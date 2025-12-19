import { PrivateJwk, Submission } from '@bytetrade/core';
import {
	VaultItem,
	VaultType,
	Field,
	cloneItemTemplates,
	ItemTemplate
} from '@didvault/sdk/src/core';
import { useSSIStore } from '../../../stores/ssi';
import { useCloudStore } from '../../../stores/cloud';
import { app } from '../../../globals';
import { i18n } from '../../../boot/i18n';
import { getPresentationJWS, VCCardInfo } from 'src/utils/vc';

export async function createVCItem(result: VCCardInfo): Promise<VaultItem> {
	const vault = app.mainVault;
	if (!vault) {
		throw new Error(i18n.global.t('errors.main_vault_is_null'));
	}

	const template: ItemTemplate | undefined = cloneItemTemplates().find(
		(template) => template.id === 'vc'
	);
	if (!template) {
		throw new Error(i18n.global.t('errors.template_is_null'));
	}
	template.fields[0].value = result.type;
	template.fields[1].value = result.manifest;
	template.fields[2].value = result.verifiable_credential;

	return await app.createItem({
		name: result.type,
		vault,
		fields: template?.fields.map(
			(f) => new Field({ ...f, value: f.value || '' })
		),
		tags: [],
		icon: template?.icon,
		type: VaultType.VC
	});
}

export async function submitPresentation(
	request_type: string,
	holder: string,
	holderPrivateKey: PrivateJwk,
	vc: string,
	address: string,
	domain: string | null
): Promise<Submission> {
	const ssiStore = useSSIStore();
	const cloudStore = useCloudStore();

	const definition = await ssiStore.get_presentation_definition(request_type);
	if (!definition) {
		throw new Error(i18n.global.t('errors.no_presentation_definition'));
	}

	const jws = await getPresentationJWS(
		holder,
		holderPrivateKey,
		definition,
		vc
	);

	if (request_type === 'Domain') {
		if (!domain) {
			throw new Error(i18n.global.t('need_domain'));
		}
		const submission = await cloudStore.verifyDomainPresentation(
			jws,
			domain // domain
		);
		if (!submission) {
			throw new Error(i18n.global.t('no_submission_result'));
		}
		return submission;
	} else {
		const submission = await ssiStore.submit_presentation(jws, address, domain);
		if (!submission) {
			throw new Error(i18n.global.t('no_submission_result'));
		}
		return submission;
	}
}
