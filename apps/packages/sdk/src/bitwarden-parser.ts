import { translate as $l } from './util';

export enum BitwardenItemFieldType {
	Text = 0,
	Hidden = 1,
	Boolean = 2,
	Linked = 3
}

export type BitwardenItemField = {
	name: string;
	value: string;
	type: BitwardenItemFieldType;
};

export type BitwardenItemLoginUri = {
	match?: string;
	uri: string;
};

export type BitwardenFolder = {
	id: string;
	name: string;
};

export enum BitwardenItemType {
	Login = 1,
	SecureNote = 2,
	Card = 3,
	Identity = 4
}

export enum BitwardenItemRepromptType {
	None = 0,
	Password = 1
}

export enum BitwardenItemSecureNoteType {
	Generic = 0
}

export type BitwardenItem = {
	id: string;
	organizationId?: string;
	folderId?: string;
	type: BitwardenItemType;
	reprompt?: BitwardenItemRepromptType;
	name: string;
	notes: string;
	favorite: boolean;
	fields?: BitwardenItemField[];
	login?: {
		uris?: BitwardenItemLoginUri[];
		username?: string;
		password?: string;
		totp?: string;
	};
	secureNote?: {
		type?: BitwardenItemSecureNoteType;
	};
	card?: {
		cardholderName?: string;
		brand?: string;
		number?: string;
		expMonth?: string;
		expYear?: string;
		code?: string;
	};
	identity?: {
		title?: string;
		firstName?: string;
		middleName?: string;
		lastName?: string;
		address1?: string;
		address2?: string;
		address3?: string;
		city?: string;
		state?: string;
		postalCode?: string;
		country?: string;
		company?: string;
		email?: string;
		phone?: string;
		ssn?: string;
		username?: string;
		passportNumber?: string;
		licenseNumber?: string;
	};
	collectionIds?: (string | null)[];
};

export type BitwardenExport = {
	folders?: BitwardenFolder[];
	items: BitwardenItem[];
};

export async function parseBitwardenFile(
	dataContent: string
): Promise<BitwardenExport> {
	try {
		const data = JSON.parse(dataContent);

		return data as BitwardenExport;
	} catch (error) {
		console.error('Failed to parse Bitwarden .json file');
		throw error;
	}
}

type RowData = {
	name: string;
	tags: string;
	url: string;
	username: string;
	password: string;
	notes: string;
	extraFields: ExtraField[];
};

type ExtraFieldType =
	| 'username'
	| 'password'
	| 'url'
	| 'email'
	| 'date'
	| 'month'
	| 'credit'
	| 'phone'
	| 'totp'
	| 'text'
	| 'pin';

type ExtraField = { name: string; value: string; type: ExtraFieldType };

function parseFieldTypeToExtraFieldType(
	field: BitwardenItemField
): ExtraFieldType {
	if (field.type === BitwardenItemFieldType.Linked) {
		return 'url';
	} else if (field.type === BitwardenItemFieldType.Hidden) {
		return 'password';
	}

	return 'text';
}

export function parseToRowData(
	item: BitwardenItem,
	defaultTags?: string[],
	folders?: BitwardenFolder[]
): RowData | undefined {
	if (!item) {
		return;
	}

	const folderName = item.folderId
		? folders?.find((folder) => folder.id === item.folderId)?.name
		: '';

	const rowData: RowData = {
		name: item.name,
		tags: [...(defaultTags || []), ...(folderName ? [folderName] : [])].join(
			','
		),
		url: item.login?.uris ? item.login?.uris[0]?.uri : '',
		username: item.login?.username || '',
		password: item.login?.password || '',
		notes: item.notes || '',
		extraFields: []
	};

	// Extract card into extraFields
	if (item.type === BitwardenItemType.Card && item.card) {
		if (item.card.number) {
			rowData.extraFields.push({
				name: $l('Card Number'),
				type: 'credit',
				value: item.card.number
			});
		}

		if (item.card.cardholderName) {
			rowData.extraFields.push({
				name: $l('Card Owner'),
				type: 'text',
				value: item.card.cardholderName
			});
		}

		if (item.card.expYear) {
			rowData.extraFields.push({
				name: $l('Valid Until'),
				type: 'month',
				value: `${item.card.expYear}-${item.card?.expMonth || '01'}`
			});
		}

		if (item.card.code) {
			rowData.extraFields.push({
				name: $l('CVC'),
				type: 'pin',
				value: item.card.code
			});
		}
	}

	// Extract identity into extraFields
	if (item.type === BitwardenItemType.Identity && item.identity) {
		Object.keys(item.identity).forEach((identityField) => {
			rowData.extraFields.push({
				name: identityField,
				type: 'text',
				value: item.identity![identityField]
			});
		});
	}

	// Extract some more extraFields
	(item.fields || []).forEach((field) => {
		rowData.extraFields.push({
			name: field.name,
			value: field.value,
			type: parseFieldTypeToExtraFieldType(field)
		});
	});

	return rowData;
}
