import { loadAsync } from 'jszip';

export type OnePuxItemDetailsLoginField = {
	value: string;
	id: string;
	name: string;
	fieldType: 'A' | 'B' | 'C' | 'E' | 'I' | 'N' | 'P' | 'R' | 'S' | 'T' | 'U';
	designation?: 'username' | 'password';
};

export type OnePuxItemDetailsSection = {
	title: string;
	name: string;
	fields: [
		{
			title: string;
			id: string;
			value: {
				concealed?: string;
				reference?: string;
				string?: string;
				email?: string;
				phone?: string;
				url?: string;
				totp?: string;
				gender?: string;
				creditCardType?: string;
				creditCardNumber?: string;
				monthYear?: number;
				date?: number;
			};
			indexAtSource: number;
			guarded: boolean;
			multiline: boolean;
			dontGenerate: boolean;
			inputTraits: {
				keyboard: string;
				correction: string;
				capitalization: string;
			};
		}
	];
};

export type OnePuxItemDetailsPasswordHistory = {
	value: string;
	time: number;
};

export type OnePuxItemOverviewUrl = {
	label: string;
	url: string;
};

export type OnePuxItem = {
	item?: {
		uuid: string;
		favIndex: number;
		createdAt: number;
		updatedAt: number;
		trashed: boolean;
		categoryUuid: string;
		details: {
			loginFields: OnePuxItemDetailsLoginField[];
			notesPlain?: string;
			sections: OnePuxItemDetailsSection[];
			passwordHistory: OnePuxItemDetailsPasswordHistory[];
			documentAttributes?: {
				fileName: string;
				documentId: string;
				decryptedSize: number;
			};
		};
		overview: {
			subtitle: string;
			urls?: OnePuxItemOverviewUrl[];
			title: string;
			url: string;
			ps?: number;
			pbe?: number;
			pgrng?: boolean;
			tags?: string[];
		};
	};
	file?: {
		attrs: {
			uuid: string;
			name: string;
			type: string;
		};
		path: string;
	};
};

export type OnePuxVault = {
	attrs: {
		uuid: string;
		desc: string;
		avatar: string;
		name: string;
		type: 'P' | 'E' | 'U';
	};
	items: OnePuxItem[];
};

export type OnePuxAccount = {
	attrs: {
		accountName: string;
		name: string;
		avatar: string;
		email: string;
		uuid: string;
		domain: string;
	};
	vaults: OnePuxVault[];
};

export type OnePuxData = {
	accounts: OnePuxAccount[];
};

export type OnePuxAttributes = {
	version: number;
	description: string;
	createdAt: number;
};

export type OnePuxExport = {
	attributes: OnePuxAttributes;
	data: OnePuxData;
};

export async function parse1PuxFile(
	fileContents: ArrayBuffer
): Promise<OnePuxExport> {
	try {
		const zip = await loadAsync(fileContents);

		const attributesContent = await zip
			.file('export.attributes')!
			.async('string');
		const attributes = JSON.parse(attributesContent);
		const dataContent = await zip.file('export.data')!.async('string');
		const data = JSON.parse(dataContent);

		return {
			attributes,
			data
		} as OnePuxExport;
	} catch (error) {
		console.error('Failed to parse .1pux file');
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
	| 'text';

type ExtraField = { name: string; value: string; type: ExtraFieldType };

function parseFieldTypeToExtraFieldType(
	field: OnePuxItemDetailsLoginField
): ExtraFieldType {
	if (field.designation === 'username') {
		return 'username';
	} else if (field.designation === 'password') {
		return 'password';
	} else if (field.fieldType === 'E') {
		return 'email';
	} else if (field.fieldType === 'U') {
		return 'url';
	}
	return 'text';
}

export function parseToRowData(
	item: OnePuxItem['item'],
	defaultTags?: string[]
): RowData | undefined {
	if (!item) {
		return;
	}

	const rowData: RowData = {
		name: item.overview.title,
		tags: [...(defaultTags || []), ...(item.overview.tags || [])].join(','),
		url: item.overview.url || '',
		username: '',
		password: '',
		notes: item.details.notesPlain || '',
		extraFields: []
	};

	// Skip documents
	if (
		item.details.documentAttributes &&
		item.details.loginFields.length === 0
	) {
		return;
	}

	// Extract username, password, and some extraFields
	item.details.loginFields.forEach((field) => {
		if (field.designation === 'username') {
			rowData.username = field.value;
		} else if (field.designation === 'password') {
			rowData.password = field.value;
		} else if (
			field.fieldType === 'I' ||
			field.fieldType === 'C' ||
			field.id.includes(';opid=__') ||
			field.value === ''
		) {
			// Skip these noisy form-fields
			return;
		} else {
			rowData.extraFields.push({
				name: field.name || field.id,
				value: field.value,
				type: parseFieldTypeToExtraFieldType(field)
			});
		}
	});

	// Extract some more extraFields
	item.details.sections.forEach((section) => {
		section.fields.forEach((field) => {
			let value = '';
			let type: ExtraFieldType = 'text';

			if (field.value.concealed) {
				value = field.value.concealed || '';
			} else if (field.value.reference) {
				value = field.value.reference || '';
			} else if (field.value.string) {
				value = field.value.string || '';
			} else if (field.value.email) {
				value = field.value.email || '';
				type = 'email';
			} else if (field.value.phone) {
				value = field.value.phone || '';
				type = 'phone';
			} else if (field.value.url) {
				value = field.value.url || '';
				type = 'url';
			} else if (field.value.totp) {
				value = field.value.totp || '';
				type = 'totp';
			} else if (field.value.gender) {
				value = field.value.gender || '';
			} else if (field.value.creditCardType) {
				value = field.value.creditCardType || '';
			} else if (field.value.creditCardNumber) {
				value = field.value.creditCardNumber || '';
				type = 'credit';
			} else if (field.value.monthYear) {
				value =
					(field.value.monthYear && field.value.monthYear.toString()) || '';
				type = 'month';
			} else if (field.value.date) {
				value = (field.value.date && field.value.date.toString()) || '';
				type = 'date';
			} else {
				// Default, so no data is lost when something new comes up
				value = JSON.stringify(field.value);
			}

			rowData.extraFields.push({
				name: field.title || field.id,
				value,
				type
			});
		});
	});

	return rowData;
}
