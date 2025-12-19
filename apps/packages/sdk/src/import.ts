import { unmarshal, bytesToString } from './core';
import { PBES2Container } from './core';
//import { validateLegacyContainer, parseLegacyContainer } from './core';
import {
	VaultItem,
	Field,
	createVaultItem,
	FieldType,
	guessFieldType
} from './core';
import { Err, ErrorCode } from './core';
import { uuid, capitalize } from './core';
import { translate } from './util';
import { readFileAsText, readFileAsArrayBuffer } from './core';

import { OnePuxItem } from './1pux-parser';
import { BitwardenExport, BitwardenItem } from './bitwarden-parser';

export interface ImportFormat {
	value:
		| 'csv'
		| 'padlock-legacy'
		| 'lastpass'
		| 'padloc'
		| '1pux'
		| 'bitwarden'
		| 'dashlane'
		| 'keepass'
		| 'nordpass'
		| 'icloud'
		| 'chrome'
		| 'firefox';
	label: string;
}

export interface ImportCSVColumn {
	name: string;
	displayName: string;
	type: FieldType | 'name' | 'tags' | 'skip';
	values: string[];
}

export const CSV: ImportFormat = {
	value: 'csv',
	label: 'CSV'
};

export const PADLOCK_LEGACY: ImportFormat = {
	value: 'padlock-legacy',
	label: 'Padlock (v2)'
};

export const LASTPASS: ImportFormat = {
	value: 'lastpass',
	label: 'LastPass'
};

export const PBES2: ImportFormat = {
	value: 'padloc',
	label: 'Encrypted Container'
};

export const ONEPUX: ImportFormat = {
	value: '1pux',
	label: '1Password (1pux)'
};

export const BITWARDEN: ImportFormat = {
	value: 'bitwarden',
	label: 'Bitwarden (JSON)'
};

export const DASHLANE: ImportFormat = {
	value: 'dashlane',
	label: 'Dashlane (CSV)'
};

export const KEEPASS: ImportFormat = {
	value: 'keepass',
	label: 'KeePass (CSV)'
};

export const NORDPASS: ImportFormat = {
	value: 'nordpass',
	label: 'NordPass (CSV)'
};

export const ICLOUD: ImportFormat = {
	value: 'icloud',
	label: 'iCloud (CSV)'
};

export const CHROME: ImportFormat = {
	value: 'chrome',
	label: 'Chrome (CSV)'
};

export const FIREFOX: ImportFormat = {
	value: 'firefox',
	label: 'Firefox (CSV)'
};

export const supportedFormats: ImportFormat[] = [
	CSV,
	PADLOCK_LEGACY,
	LASTPASS,
	PBES2,
	ONEPUX,
	BITWARDEN,
	DASHLANE,
	KEEPASS,
	NORDPASS,
	ICLOUD,
	CHROME,
	FIREFOX
];

export const csvSupportedFormats: ImportFormat[] = [
	CSV,
	LASTPASS,
	DASHLANE,
	KEEPASS,
	NORDPASS,
	ICLOUD,
	CHROME,
	FIREFOX
];

export function loadPapa(): Promise<any> {
	return import(/* webpackChunkName: "papaparse" */ 'papaparse');
}

/**
 * Takes a data table (represented by a two-dimensional array) and converts it
 * into an array of items
 * @param  Array    data            Two-dimensional array containing tabular item data; The first 'row'
 *                                  might contain field names. All other rows represent items, containing
 *                                  the item name, field values and optionally a list of tags.
 * @param  Array    columnTypes     Array containing the type of field per column.
 * @param  Boolean  columnsOnFirstRow  Boolean, representing if there are columnms on the first row.
 */
async function fromTable(
	data: string[][],
	columnTypes: ImportCSVColumn[],
	columnsOnFirstRow: boolean
): Promise<VaultItem[]> {
	let nameColumnIndex = columnTypes.findIndex(
		(columnType) => columnType.type === 'name'
	);
	const tagsColumnIndex = columnTypes.findIndex(
		(columnType) => columnType.type === 'tags'
	);

	if (nameColumnIndex === -1) {
		nameColumnIndex = 0;
	}

	const dataRows = columnsOnFirstRow ? data.slice(1) : data;

	// All subsequent rows should contain values
	const items = dataRows
		.filter((row) => {
			// Skip empty rows
			if (row.length === 1 && row[0] === '') {
				return false;
			}

			return true;
		})
		.map((row) => {
			// Construct an array of field object from column names and values
			const fields: Field[] = [];
			for (let columnIndex = 0; columnIndex < row.length; ++columnIndex) {
				if (columnTypes[columnIndex]?.type === 'skip') {
					continue;
				}

				// Skip name column, category column (if any) and empty fields
				if (
					columnIndex !== nameColumnIndex &&
					columnIndex !== tagsColumnIndex &&
					row[columnIndex]
				) {
					const name = columnTypes[columnIndex]?.displayName || '';
					const value = row[columnIndex];
					const type = columnTypes[columnIndex]?.type || undefined;
					fields.push(
						new Field().fromRaw({
							name,
							value,
							type
						})
					);
				}
			}

			const name = row[nameColumnIndex!];
			const tags = row[tagsColumnIndex!];
			return createVaultItem({
				name,
				fields,
				tags: (tags && tags.split(',')) || []
			});
		});

	return Promise.all(items);
}

export async function isCSV(data: string): Promise<boolean> {
	const papa = await loadPapa();
	return papa.parse(data).errors.length === 0;
}

export async function asCSV(
	file: File,
	mappedItemColumns: ImportCSVColumn[],
	columnsOnFirstRow: boolean
): Promise<{ items: VaultItem[]; itemColumns: ImportCSVColumn[] }> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const parsed = papa.parse(data);
	if (parsed.errors.length) {
		throw new Err(ErrorCode.INVALID_CSV, 'Failed to parse .csv file.');
	}
	const rows = parsed.data as string[][];

	if (rows.length === 0) {
		throw new Err(ErrorCode.INVALID_CSV, 'No rows found in .csv file.');
	}

	const columnNames = columnsOnFirstRow
		? rows[0].map((column) => column.toLowerCase())
		: rows[0].map((_value, index) =>
				translate('Column {column}', {
					column: index.toString()
				})
		  );

	let hasNameColumn = false;
	let hasTagsColumn = false;

	const itemColumns =
		mappedItemColumns.length > 0
			? mappedItemColumns
			: columnNames.map((columnName, columnIndex) => {
					const values = (columnsOnFirstRow ? rows.slice(1) : rows).map(
						(row) => row[columnIndex] || ''
					);

					// Guess field type based on first non-empty value
					// TODO: Sample all values for more reliable results?
					let type = guessFieldType({
						name: columnsOnFirstRow ? columnName : '',
						value: values.find((value) => Boolean(value))
					}) as ImportCSVColumn['type'];

					// If we're not given field names by the first row, base the name on the type
					const name = columnsOnFirstRow
						? columnName.toLocaleLowerCase()
						: type;

					if (!hasNameColumn && name === 'name') {
						type = 'name';
						hasNameColumn = true;
					}

					if (!hasTagsColumn && ['tags', 'category'].includes(name)) {
						type = 'tags';
						hasTagsColumn = true;
					}

					return {
						name,
						displayName: capitalize(name),
						type,
						values
					};
			  });

	// Ensure there's at least one nameColumn
	if (mappedItemColumns.length === 0 && !hasNameColumn) {
		itemColumns[0].type = 'name';
	}

	const items = await fromTable(rows, itemColumns, columnsOnFirstRow);

	return { items, itemColumns };
}
// export async function isPadlockV1(file: File): Promise<boolean> {
// 	try {
// 		const data = await readFileAsText(file);
// 		return validateLegacyContainer(unmarshal(data));
// 	} catch (e) {
// 		return false;
// 	}
// }

// export async function asPadlockLegacy(
// 	file: File,
// 	password: string
// ): Promise<VaultItem[]> {
// 	const data = await readFileAsText(file);
// 	const container = parseLegacyContainer(unmarshal(data));
// 	await container.unlock(password);
// 	return importLegacyContainer(container);
// }

export async function importLegacyContainer(container: PBES2Container) {
	const records = unmarshal(bytesToString(await container.getData())) as any[];
	const items = records
		.filter(({ removed }) => !removed)
		.map(async ({ name = 'Unnamed', fields = [], tags, category, updated }) => {
			return new VaultItem().fromRaw({
				id: await uuid(),
				name,
				fields,
				tags: tags || [category],
				updated,
				updatedBy: '',
				attachments: []
			});
		});

	return Promise.all(items);
}

export async function isPBES2Container(file: File) {
	try {
		const data = await readFileAsText(file);
		new PBES2Container().fromRaw(unmarshal(data));
		return true;
	} catch (error) {
		return false;
	}
}

export async function asPBES2Container(
	file: File,
	password: string
): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const container = new PBES2Container().fromRaw(unmarshal(data));
	await container.unlock(password);

	const raw = unmarshal(bytesToString(await container.getData())) as any;

	const items = raw.items.map((item: any) => {
		// Due to a bug in < v1.3.4 items were not serialized properly, so we may
		// need this additional step
		if (typeof item === 'string') {
			try {
				item = unmarshal(item);
			} catch (e) {
				console.log(e);
			}
		}
		return new VaultItem().fromRaw(item);
	});

	return items;
}

/*
 * Lastpass secure notes are exported by putting non-standard fields into the 'extra' column. Every line
 * represents a field in the following format:
 *
 *     field_name:data
 *
 * We're parsing that information to retrieve the individual fields
 */
function lpParseNotes(str: string): Field[] {
	const lines = str.split('\n');
	const fields = lines
		.filter((line) => !!line)
		.map((line) => {
			const split = line.indexOf(':');
			return new Field({
				name: line.substring(0, split),
				value: line.substring(split + 1),
				type: FieldType.Text
			});
		});
	return fields;
}

/*
 * Parses a single row in a LastPass CSV file. Apart from extracting the default fields, we also parse
 * the 'extra' column for 'special notes' and remove any special fields that are not needed outside of
 * LastPass
 */
async function lpParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 4;
	const categoryIndex = 5;
	const urlIndex = 0;
	const usernameIndex = 1;
	const passwordIndex = 2;
	const notesIndex = 3;
	const totpIndex = 7;

	let fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];
	const notes = row[notesIndex];

	if (row[urlIndex] === 'http://sn') {
		// The 'http://sn' url indicates that this line represents a 'secure note', which means
		// we'll have to parse the 'extra' column to retrieve the individual fields
		fields.push(...lpParseNotes(notes));
		// In case of 'secure notes' we don't want the url and NoteType field
		fields = fields.filter((f) => f.name != 'url' && f.name != 'NoteType');
	} else {
		// We've got a regular 'site' item, so the 'extra' column simply contains notes
		fields.push(
			new Field({
				name: translate('Notes'),
				value: notes,
				type: FieldType.Note
			})
		);
	}

	if (row[totpIndex]) {
		fields.push(
			new Field({
				name: translate('One-Time Password'),
				value: row[totpIndex],
				type: FieldType.Totp
			})
		);
	}

	const dir = row[categoryIndex];
	// Create a basic item using the standard fields
	return createVaultItem({
		name: row[nameIndex],
		fields,
		tags: dir ? [dir] : []
	});
}

export async function asLastPass(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		.map(lpParseRow);

	return Promise.all(items);
}

export async function isLastPass(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return (
			headerRow === 'url,username,password,extra,name,grouping,fav' ||
			headerRow === 'url,username,password,extra,name,grouping,fav,totp'
		);
	} catch (error) {
		return false;
	}
}

async function parse1PuxItem(
	accountName: string,
	vaultName: string,
	item: OnePuxItem['item']
): Promise<VaultItem | undefined> {
	if (!item) {
		return;
	}

	const { parseToRowData } = await import('./1pux-parser');

	const rowData = parseToRowData(item, [accountName, vaultName]);

	if (!rowData) {
		return;
	}

	const itemName = rowData.name;
	const tags = rowData.tags.split(',');

	if (item.trashed) {
		tags.push('trashed');
	}

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: rowData.username,
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: rowData.password,
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: rowData.url,
			type: FieldType.Url
		})
	];

	if (rowData.notes) {
		fields.push(
			new Field({
				name: translate('Notes'),
				value: rowData.notes,
				type: FieldType.Note
			})
		);
	}

	for (const extraField of rowData.extraFields) {
		if (extraField.type === 'totp') {
			// Extract just the secret
			try {
				const secret = new URL(extraField.value).searchParams.get('secret');
				if (secret) {
					fields.push(
						new Field({
							name: extraField.name,
							value: secret,
							type: FieldType.Totp
						})
					);
				}
			} catch (error) {
				// Do nothing
			}
		} else {
			fields.push(
				new Field({
					name: extraField.name,
					value: extraField.value,
					type: extraField.type as FieldType
				})
			);
		}
	}

	return createVaultItem({ name: itemName, fields, tags });
}

export async function as1Pux(file: File): Promise<VaultItem[]> {
	try {
		const { parse1PuxFile } = await import('./1pux-parser');
		const data = await readFileAsArrayBuffer(file);
		const dataExport = await parse1PuxFile(data);

		const items: VaultItem[] = [];

		for (const account of dataExport.data.accounts) {
			for (const vault of account.vaults) {
				for (const vaultItem of vault.items) {
					if (vaultItem.item) {
						const parsedItem = await parse1PuxItem(
							account.attrs.name,
							vault.attrs.name,
							vaultItem.item
						);
						if (parsedItem) {
							items.push(parsedItem);
						}
					}
				}
			}
		}

		return items;
	} catch (error) {
		throw new Err(ErrorCode.INVALID_1PUX, 'Failed to parse .1pux file.');
	}
}

/**
 * Checks if a given file name ends with .1pux to avoid trying to parse unnecessarily
 */
export function is1Pux(file: File): boolean {
	return file.name.endsWith('.1pux');
}

async function parseBitwardenItem(
	vaultName: string,
	item: BitwardenItem,
	folders: BitwardenExport['folders']
): Promise<VaultItem | undefined> {
	if (!item) {
		return;
	}

	const { parseToRowData } = await import('./bitwarden-parser');

	const rowData = parseToRowData(item, [vaultName], folders);

	if (!rowData) {
		return;
	}

	const itemName = rowData.name;
	const tags = rowData.tags.split(',');

	const fields: Field[] = [];

	if (rowData.username) {
		fields.push(
			new Field({
				name: 'Username',
				value: rowData.username,
				type: FieldType.Username
			})
		);
	}

	if (rowData.password) {
		fields.push(
			new Field({
				name: 'Password',
				value: rowData.password,
				type: FieldType.Password
			})
		);
	}

	if (rowData.url) {
		fields.push(
			new Field({
				name: 'URL',
				value: rowData.url,
				type: FieldType.Url
			})
		);
	}

	if (rowData.notes) {
		fields.push(
			new Field({
				name: 'Notes',
				value: rowData.notes,
				type: FieldType.Note
			})
		);
	}

	for (const extraField of rowData.extraFields) {
		if (extraField.type === 'totp') {
			// Extract just the secret
			try {
				const secret = new URL(extraField.value).searchParams.get('secret');
				if (secret) {
					fields.push(
						new Field({
							name: extraField.name,
							value: secret,
							type: FieldType.Totp
						})
					);
				}
			} catch (error) {
				// Do nothing
			}
		} else {
			fields.push(
				new Field({
					name: extraField.name,
					value: extraField.value,
					type: extraField.type as FieldType
				})
			);
		}
	}

	return createVaultItem({ name: itemName, fields, tags });
}

export async function asBitwarden(file: File): Promise<VaultItem[]> {
	try {
		const { parseBitwardenFile } = await import('./bitwarden-parser');
		const contents = await readFileAsText(file);
		const dataExport = await parseBitwardenFile(contents);

		const items: VaultItem[] = [];

		for (const vaultItem of dataExport.items) {
			if (vaultItem) {
				const parsedItem = await parseBitwardenItem(
					file.name.replace('.json', ''),
					vaultItem,
					dataExport.folders
				);
				if (parsedItem) {
					items.push(parsedItem);
				}
			}
		}

		return items;
	} catch (error) {
		throw new Err(
			ErrorCode.INVALID_BITWARDEN,
			'Failed to parse Bitwarden .json file.'
		);
	}
}

export async function isBitwarden(file: File): Promise<boolean> {
	try {
		const content = await readFileAsText(file);
		const data = JSON.parse(content);
		return Object.prototype.hasOwnProperty.call(data, 'items');
	} catch (error) {
		return false;
	}
}

async function dashlaneParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 3;
	const categoryIndex = 7;
	const urlIndex = 6;
	const usernameIndex = 0;
	const otherUsernameIndexes = [1, 2];
	const passwordIndex = 4;
	const notesIndex = 5;
	const totpIndex = 8;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	for (const otherUsernameIndex of otherUsernameIndexes) {
		if (row[otherUsernameIndex]) {
			fields.push(
				new Field({
					name: translate('Other Username'),
					value: row[otherUsernameIndex],
					type: FieldType.Username
				})
			);
		}
	}

	if (row[notesIndex]) {
		fields.push(
			new Field({
				name: translate('Notes'),
				value: row[notesIndex],
				type: FieldType.Note
			})
		);
	}

	if (row[totpIndex]) {
		fields.push(
			new Field({
				name: translate('One-Time Password'),
				value: row[totpIndex],
				type: FieldType.Totp
			})
		);
	}

	const tags = row[categoryIndex] ? [row[categoryIndex]] : [];

	return createVaultItem({ name: row[nameIndex], fields, tags });
}

export async function asDashlane(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		.map(dashlaneParseRow);

	return Promise.all(items);
}

export async function isDashlane(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return (
			headerRow ===
			'username,username2,username3,title,password,note,url,category,otpSecret'
		);
	} catch (error) {
		return false;
	}
}

async function keePassParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 0;
	const urlIndex = 3;
	const usernameIndex = 1;
	const passwordIndex = 2;
	const notesIndex = 4;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	if (row[notesIndex]) {
		fields.push(
			new Field({
				name: translate('Notes'),
				value: row[notesIndex],
				type: FieldType.Note
			})
		);
	}

	return createVaultItem({ name: row[nameIndex], fields, tags: [] });
}

export async function asKeePass(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		.map(keePassParseRow);

	return Promise.all(items);
}

export async function isKeePass(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return (
			headerRow === '"Account","Login Name","Password","Web Site","Comments"'
		);
	} catch (error) {
		return false;
	}
}

async function nordPassParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 0;
	const urlIndex = 1;
	const usernameIndex = 2;
	const passwordIndex = 3;
	const notesIndex = 4;
	const cardHolderNameIndex = 5;
	const cardNumberIndex = 6;
	const cvcIndex = 7;
	const expiryDateIndex = 8;
	const zipCodeIndex = 9;
	const folderIndex = 10;
	const fullNameIndex = 11;
	const phoneIndex = 12;
	const emailIndex = 13;
	const address1Index = 14;
	const address2Index = 15;
	const cityIndex = 16;
	const countryIndex = 17;
	const stateIndex = 18;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	if (row[notesIndex]) {
		fields.push(
			new Field({
				name: translate('Notes'),
				value: row[notesIndex],
				type: FieldType.Note
			})
		);
	}

	if (row[cardHolderNameIndex]) {
		fields.push(
			new Field({
				name: translate('Cardholder Name'),
				value: row[cardHolderNameIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[cardNumberIndex]) {
		fields.push(
			new Field({
				name: translate('Card Number'),
				value: row[cardNumberIndex],
				type: FieldType.Credit
			})
		);
	}

	if (row[cvcIndex]) {
		fields.push(
			new Field({
				name: translate('CVC'),
				value: row[cvcIndex],
				type: FieldType.Pin
			})
		);
	}

	if (row[expiryDateIndex]) {
		fields.push(
			new Field({
				name: translate('Expiry Date'),
				value: row[expiryDateIndex],
				type: FieldType.Date
			})
		);
	}

	if (row[zipCodeIndex]) {
		fields.push(
			new Field({
				name: translate('Zip Code'),
				value: row[zipCodeIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[fullNameIndex]) {
		fields.push(
			new Field({
				name: translate('Full Name'),
				value: row[fullNameIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[phoneIndex]) {
		fields.push(
			new Field({
				name: translate('Phone Number'),
				value: row[phoneIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[emailIndex]) {
		fields.push(
			new Field({
				name: translate('Email'),
				value: row[emailIndex],
				type: FieldType.Email
			})
		);
	}

	if (row[address1Index]) {
		fields.push(
			new Field({
				name: translate('Address 1'),
				value: row[address1Index],
				type: FieldType.Text
			})
		);
	}

	if (row[address2Index]) {
		fields.push(
			new Field({
				name: translate('Address 2'),
				value: row[address2Index],
				type: FieldType.Text
			})
		);
	}

	if (row[cityIndex]) {
		fields.push(
			new Field({
				name: translate('City'),
				value: row[cityIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[countryIndex]) {
		fields.push(
			new Field({
				name: translate('Country'),
				value: row[countryIndex],
				type: FieldType.Text
			})
		);
	}

	if (row[stateIndex]) {
		fields.push(
			new Field({
				name: translate('State'),
				value: row[stateIndex],
				type: FieldType.Text
			})
		);
	}

	const tags = row[folderIndex] ? [row[folderIndex]] : [];

	return createVaultItem({ name: row[nameIndex], fields, tags });
}

export async function asNordPass(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const typeIndex = 19;
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		// Filter out folders
		.filter((row: string[]) => row[typeIndex] !== 'folder')
		.map(nordPassParseRow);

	return Promise.all(items);
}

export async function isNordPass(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return (
			headerRow ===
			'name,url,username,password,note,cardholdername,cardnumber,cvc,expirydate,zipcode,folder,full_name,phone_number,email,address1,address2,city,country,state,type'
		);
	} catch (error) {
		return false;
	}
}

async function iCloudParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 0;
	const urlIndex = 1;
	const usernameIndex = 2;
	const passwordIndex = 3;
	const notesIndex = 4;
	const totpIndex = 5;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	if (row[notesIndex]) {
		fields.push(
			new Field({
				name: translate('Notes'),
				value: row[notesIndex],
				type: FieldType.Note
			})
		);
	}

	if (row[totpIndex]) {
		// Extract just the secret
		try {
			const secret = new URL(row[totpIndex]).searchParams.get('secret');
			if (secret) {
				fields.push(
					new Field({
						name: translate('One-Time Password'),
						value: secret,
						type: FieldType.Totp
					})
				);
			}
		} catch (error) {
			// Do nothing
		}
	}

	return createVaultItem({ name: row[nameIndex], fields, tags: [] });
}

export async function asICloud(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		.map(iCloudParseRow);

	return Promise.all(items);
}

export async function isICloud(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return headerRow === 'Title,URL,Username,Password,Notes,OTPAuth';
	} catch (error) {
		return false;
	}
}

async function chromeParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 0;
	const urlIndex = 1;
	const usernameIndex = 2;
	const passwordIndex = 3;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	return createVaultItem({ name: row[nameIndex], fields, tags: [] });
}

export async function asChrome(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		.map(chromeParseRow);

	return Promise.all(items);
}

export async function isChrome(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return headerRow === 'name,url,username,password';
	} catch (error) {
		return false;
	}
}

async function firefoxParseRow(row: string[]): Promise<VaultItem> {
	const nameIndex = 0;
	const urlIndex = 0;
	const usernameIndex = 1;
	const passwordIndex = 2;

	const fields: Field[] = [
		new Field({
			name: translate('Username'),
			value: row[usernameIndex],
			type: FieldType.Username
		}),
		new Field({
			name: translate('Password'),
			value: row[passwordIndex],
			type: FieldType.Password
		}),
		new Field({
			name: translate('URL'),
			value: row[urlIndex],
			type: FieldType.Url
		})
	];

	return createVaultItem({
		name: row[nameIndex].replace('https://', '').replace('http://', ''),
		fields,
		tags: []
	});
}

export async function asFirefox(file: File): Promise<VaultItem[]> {
	const data = await readFileAsText(file);
	const papa = await loadPapa();
	const urlIndex = 0;
	const items = papa
		.parse(data)
		.data // Remove first row as it only contains field names
		.slice(1)
		// Filter out empty rows
		.filter((row: string[]) => row.length > 1)
		// Filter out Firefox Sync identity
		.filter((row: string[]) => row[urlIndex] !== 'chrome://FirefoxAccounts')
		.map(firefoxParseRow);

	return Promise.all(items);
}

export async function isFirefox(file: File): Promise<boolean> {
	try {
		const data = await readFileAsText(file);
		const headerRow = data.split('\n')[0].trim();
		return (
			headerRow ===
			'"url","username","password","httpRealm","formActionOrigin","guid","timeCreated","timeLastUsed","timePasswordChanged"'
		);
	} catch (error) {
		return false;
	}
}

export async function guessFormat(file: File): Promise<ImportFormat> {
	// Try to guess 1pux first (won't need parsing)
	if (is1Pux(file)) {
		return ONEPUX;
	}
	if (await isBitwarden(file)) {
		return BITWARDEN;
	}
	if (await isPBES2Container(file)) {
		return PBES2;
	}
	// if (await isPadlockV1(file)) {
	// 	return PADLOCK_LEGACY;
	// }
	if (await isLastPass(file)) {
		return LASTPASS;
	}
	if (await isDashlane(file)) {
		return DASHLANE;
	}
	if (await isKeePass(file)) {
		return KEEPASS;
	}
	if (await isNordPass(file)) {
		return NORDPASS;
	}
	if (await isICloud(file)) {
		return ICLOUD;
	}
	if (await isChrome(file)) {
		return CHROME;
	}
	if (await isFirefox(file)) {
		return FIREFOX;
	}

	return CSV;
}
