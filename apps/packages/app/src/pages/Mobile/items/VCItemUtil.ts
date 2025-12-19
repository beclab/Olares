import { getTerminusNameFromVC } from '@bytetrade/core';
import { base64ToString } from '@didvault/sdk/src/core';
import { app } from '../../../globals';
import { VCCardItem, convertVault2CardItem } from 'src/utils/vc';

export function getVCCardItemList(): VCCardItem[] {
	const res: VCCardItem[] = [];
	for (const vault of app.state.vaults) {
		for (const item of vault.items) {
			const card = convertVault2CardItem(item);
			if (card) {
				res.push(card);
			}
		}
	}
	return res;
}

export function getTerminusName(card: VCCardItem) {
	let name = '';

	const vc = JSON.parse(
		base64ToString(card.verifiable_credential.split('.')[1])
	).vc;

	if (card.type == 'Google') {
		const tName = vc.credentialSubject['email'];
		name = getTerminusNameFromVC('Google', tName);
	} else if (card.type == 'Twitter') {
		const tName = vc.credentialSubject['username'];
		name = getTerminusNameFromVC('Twitter', tName);
	}
	return name ? name : '';
}
