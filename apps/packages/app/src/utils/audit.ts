import { HashParams } from '@didvault/sdk/src/core';
import { bytesToHex, stringToBytes } from '@didvault/sdk/src/core';
import {
	AuditResult,
	AuditType,
	FieldType,
	VaultItem
} from '@didvault/sdk/src/core';
import { getCryptoProvider } from '@didvault/sdk/src/core';
import { Vault } from '@didvault/sdk/src/core';
import { sub } from 'date-fns';
import { app } from '../globals';
import { passwordStrength } from '@didvault/sdk/src/util';

export interface ListItem {
	item: VaultItem;
	vault: Vault;
	section?: string;
	firstInSection?: boolean;
	lastInSection?: boolean;
	warning?: boolean;
}

async function sha1(password: string) {
	const hashedPasswordData = await getCryptoProvider().hash(
		stringToBytes(password),
		new HashParams({ algorithm: 'SHA-1' })
	);
	const hashedPassword = bytesToHex(hashedPasswordData);
	return hashedPassword;
}

async function isPasswordWeak(password: string) {
	// If the password is longer that 100 characters, calculating the
	// entropy becomes too expensive, so we assume it's probably
	// strong enough.
	if (password.length > 100) {
		return false;
	}

	const { score } = await passwordStrength(password);

	return score < 2;
}

async function hasPasswordBeenCompromised(passwordHash: string) {
	const hashPrefix = passwordHash.slice(0, 5);

	const response = await fetch(
		`https://api.pwnedpasswords.com/range/${hashPrefix}`
	);

	const result = await response.text();

	const matchingHashSuffixes = result.split('\r\n');

	for (const matchingHashSuffix of matchingHashSuffixes) {
		const fullLowercaseHash = `${hashPrefix}${
			matchingHashSuffix.toLowerCase().split(':')[0]
		}`;

		if (fullLowercaseHash === passwordHash) {
			return true;
		}
	}

	return false;
}

export async function auditVaults(
	vaults: Vault[] = app.vaults,
	{
		updateOnlyItemWithId,
		updateOnlyIfOutdated
	}: { updateOnlyItemWithId?: string; updateOnlyIfOutdated?: boolean } = {}
) {
	const reusedPasswords: ListItem[] = [];
	const weakPasswords: ListItem[] = [];
	const compromisedPasswords: ListItem[] = [];
	const expiredItems: ListItem[] = [];

	// Don't try to run if the app has locked
	if (app.state.locked) {
		return {
			reusedPasswords,
			weakPasswords,
			compromisedPasswords,
			expiredItems
		};
	}

	const usedPasswordHashCounts = new Map<string, number>();
	const reusedPasswordItemIds: Set<string> = new Set();
	const weakPasswordItemIds: Set<string> = new Set();
	const compromisedPasswordItemIds: Set<string> = new Set();

	// We need to do a run once for all the password hashes, to calculate reused afterwards, otherwise order can become a problem
	for (const vault of vaults) {
		// const feature = vault.org
		//     ? app.getOrgFeatures(vault.org).securityReport
		//     : app.getAccountFeatures().securityReport;
		// if (feature.disabled) {
		//     continue;
		// }
		for (const item of vault.items) {
			const passwordFields = item.fields
				.map((field, fieldIndex) => ({ field, fieldIndex }))
				.filter((field) => field.field.type === FieldType.Password)
				.filter((field) => Boolean(field.field.value));

			for (const passwordField of passwordFields) {
				const passwordHash = await sha1(passwordField.field.value);

				const currentPasswordHashCount =
					usedPasswordHashCounts.get(passwordHash) || 0;

				usedPasswordHashCounts.set(passwordHash, currentPasswordHashCount + 1);
			}
		}
	}

	const now = new Date();
	const oneWeekAgo = sub(new Date(), { weeks: 1 });

	let resultsFound = false;

	for (const vault of vaults) {
		// const feature = vault.org
		//     ? app.getOrgFeatures(vault.org).securityReport
		//     : app.getAccountFeatures().securityReport;

		// if (
		//     feature.disabled ||
		//     !app.hasWritePermissions(vault) ||
		//     !vault.accessors.some((a) => a.id === app.account!.id)
		// ) {
		//     continue;
		// }

		let vaultResultsFound = false;

		for (const item of vault.items) {
			if (updateOnlyItemWithId) {
				if (item.id !== updateOnlyItemWithId) {
					continue;
				}
			}

			if (
				updateOnlyIfOutdated &&
				item.lastAudited &&
				item.lastAudited >= oneWeekAgo
			) {
				continue;
			}

			const passwordFields = item.fields
				.map((field, fieldIndex) => ({ field, fieldIndex }))
				.filter((field) => field.field.type === FieldType.Password)
				.filter((field) => Boolean(field.field.value));

			// If an item had password fields that failed audits and were since removed, we need to run the audit again to clear and update it
			const itemHasFailedAudits = (item.auditResults || []).length > 0;

			if (passwordFields.length === 0 && !itemHasFailedAudits) {
				continue;
			}

			const auditResults: AuditResult[] = [];

			for (const passwordField of passwordFields) {
				const passwordHash = await sha1(passwordField.field.value);

				// Reused audit can't be skipped as it's interdependent
				if (usedPasswordHashCounts.get(passwordHash)! > 1) {
					// Don't add the same item twice to the list, if there are more than one reused password fields in it
					if (!reusedPasswordItemIds.has(item.id)) {
						reusedPasswords.push({ item, vault });
						reusedPasswordItemIds.add(item.id);
					}

					if (app.account?.settings.securityReport.reusedPasswords) {
						auditResults.push({
							type: AuditType.ReusedPassword,
							fieldIndex: passwordField.fieldIndex
						});
					}

					vaultResultsFound = true;
				}

				if (app.account?.settings.securityReport.weakPasswords) {
					const isThisPasswordWeak = await isPasswordWeak(
						passwordField.field.value
					);
					if (isThisPasswordWeak) {
						// Don't add the same item twice to the list, if there are more than one weak password fields in it
						if (!weakPasswordItemIds.has(item.id)) {
							weakPasswords.push({ item, vault });
							weakPasswordItemIds.add(item.id);
						}

						auditResults.push({
							type: AuditType.WeakPassword,
							fieldIndex: passwordField.fieldIndex
						});

						vaultResultsFound = true;
					}
				}

				if (app.account?.settings.securityReport.compromisedPaswords) {
					const isPasswordCompromised = await hasPasswordBeenCompromised(
						passwordHash
					);
					if (isPasswordCompromised) {
						// Don't add the same item twice to the list, if there are more than one compromised password fields in it
						if (!compromisedPasswordItemIds.has(item.id)) {
							compromisedPasswords.push({ item, vault });
							compromisedPasswordItemIds.add(item.id);
						}

						auditResults.push({
							type: AuditType.CompromisedPassword,
							fieldIndex: passwordField.fieldIndex
						});

						vaultResultsFound = true;
					}
				}
			}

			if (app.account?.settings.securityReport.expiredItems && item.expiresAt) {
				const isThisItemExpired = item.expiresAt < now;

				if (isThisItemExpired) {
					auditResults.push({
						type: AuditType.ExpiredItem,
						fieldIndex: -1
					});

					vaultResultsFound = true;
				}
			}

			item.auditResults = auditResults;
			item.lastAudited = new Date();
			vault.items.update(item);
		}

		if (!vaultResultsFound) {
			await app.saveVault(vault);
		}

		resultsFound = resultsFound || vaultResultsFound;
	}

	if (resultsFound) {
		await app.save();
	}

	return {
		reusedPasswords,
		weakPasswords,
		compromisedPasswords,
		expiredItems
	};
}
