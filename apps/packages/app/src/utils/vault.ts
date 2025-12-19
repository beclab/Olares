import * as imp from '@didvault/sdk/src/import';
import { VaultItem } from '@didvault/sdk/src/core';
import { QVueGlobals } from 'quasar';

export const parseData = async (
	file: File,
	formatSelect: imp.ImportFormat,
	itemColumns: imp.ImportCSVColumn[] = [],
	quasar: QVueGlobals,
	resetCSVColumns = false,
	hasColumnsOnFirstRow = false
) => {
	let items: VaultItem[] = [];
	let returnItemColumns: imp.ImportCSVColumn[] = [];
	switch (formatSelect.value) {
		case imp.PADLOCK_LEGACY.value:
			// #warning
			break;
		case imp.LASTPASS.value:
			items = await imp.asLastPass(file);
			break;
		case imp.CSV.value:
			{
				const result = await imp.asCSV(
					file,
					resetCSVColumns ? [] : itemColumns,
					hasColumnsOnFirstRow
				);
				items = result.items;
				returnItemColumns = result.itemColumns;
			}
			break;
		case imp.ONEPUX.value:
			items = await imp.as1Pux(file);
			break;
		case imp.BITWARDEN.value:
			items = await imp.asBitwarden(file);
			break;
		case imp.DASHLANE.value:
			items = await imp.asDashlane(file);
			break;
		case imp.KEEPASS.value:
			items = await imp.asKeePass(file);
			break;
		case imp.NORDPASS.value:
			items = await imp.asNordPass(file);
			break;
		case imp.ICLOUD.value:
			items = await imp.asICloud(file);
			break;
		case imp.CHROME.value:
			items = await imp.asChrome(file);
			break;
		case imp.FIREFOX.value:
			items = await imp.asFirefox(file);
			break;
		case imp.PBES2.value:
			{
				const VaultImportPbes2PasswordDialog = (
					await import(
						'../components/setting/VaultImportPbes2PasswordDialog.vue'
					)
				).default;
				const pItems = await new Promise<VaultItem[]>((resolve) =>
					quasar
						.dialog({
							component: VaultImportPbes2PasswordDialog,
							componentProps: {
								file: file
							}
						})
						.onOk((items: VaultItem[]) => {
							resolve(items);
						})
						.onCancel(() => {
							resolve([]);
						})
				);
				items = pItems;
			}

			break;
	}
	return {
		items,
		returnItemColumns
	};
};
