import { VaultItem } from './core';
import { PBES2Container } from './core';
import { marshal, stringToBytes } from './core';
import { loadPapa, ImportFormat, CSV, PBES2 } from './import';

export const supportedFormats: ImportFormat[] = [CSV, PBES2];
export { CSV, PBES2 } from './import';

function itemsToTable(items: VaultItem[]) {
	// Array of column names
	let cols = ['name', 'tags'];
	// Column indizes associated with field/column names
	let colInds = {};
	// Two dimensional array, starting with column names
	let table = [cols];

	// Fill up columns array with distinct field names
	for (let item of items) {
		for (let field of item.fields) {
			if (!colInds[field.name]) {
				colInds[field.name] = cols.length;
				cols.push(field.name);
			}
		}
	}

	// Creates an array of empty strings with the length of the `cols` array
	function emptyRow() {
		var l = cols.length;
		var row: string[] = [];
		while (l--) {
			row.push('');
		}
		return row;
	}

	// Add a row for each item
	items.forEach(function (item) {
		// Create an empty row to be filled with item name, category and field values
		var row = emptyRow();
		// VaultItem name and category are always the first and second column respectively
		row[0] = item.name;
		row[1] = item.tags.join(',');

		// Fill up columns with corrensponding field values if the fields exist on the item. All
		// other columns remain empty
		item.fields.forEach(function (item) {
			row[colInds[item.name]] = item.value;
		});

		// Add row to table
		table.push(row);
	});

	return table;
}

export async function asCSV(items: VaultItem[]): Promise<string> {
	const papa = await loadPapa();
	return papa.unparse(itemsToTable(items));
}

export async function asPBES2Container(
	items: VaultItem[],
	password: string
): Promise<string> {
	for (const item of items) {
		item.attachments = [];
	}

	const container = new PBES2Container();
	await container.unlock(password);
	await container.setData(
		stringToBytes(marshal({ items: items.map((item) => item.toRaw()) }))
	);
	return container.toJSON();
}
