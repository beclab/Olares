import { translate as $l } from '../util';
import {
	base32ToBytes,
	Serializable,
	AsSerializable,
	AsDate
} from './encoding';
import { totp } from './otp';
import { uuid } from './util';
import { AccountID } from './account';
import { AttachmentInfo } from './attachment';
import { openExternalUrl } from './platform';
import { add } from 'date-fns';
import cloneDeep from 'lodash/cloneDeep';

/** A tag that can be assigned to a [[VaultItem]] */
export type Tag = string;

export interface TagInfo {
	name: Tag;
	unlisted?: boolean;
	color?: string;
	count?: number;
	readonly?: number;
}

/** Unique identifier for [[VaultItem]]s */
export type VaultItemID = string;

export enum FieldType {
	Username = 'username',
	Password = 'password',
	ApiSecret = 'apiSecret',
	Mnemonic = 'mnemonic',
	Url = 'url',
	Email = 'email',
	Date = 'date',
	Month = 'month',
	Credit = 'credit',
	Phone = 'phone',
	Pin = 'pin',
	Totp = 'totp',
	Note = 'note',
	Text = 'text'
	// CryptoeAddress = 'cryptoaddress'
}

/**
 * Field definition containing meta data for a certain field type
 */
export interface FieldDef {
	/** content type */
	type: FieldType;
	/** regular expression describing pattern of field contents (used for validation) */
	pattern: RegExp;
	/** regular expression describing pattern of field contents (used for matching) */
	matchPattern: RegExp;
	/** whether the field should be masked when displayed */
	mask: boolean;
	/** whether the field value can have multiple lines */
	multiline: boolean;
	/** icon used for display */
	icon: string;
	/** display name */
	name: string;
	/** display formatting */
	format?: (value: string, masked: boolean) => string;
	/** for values that need to be prepared before being copied / filled */
	transform?: (value: string) => Promise<string>;
	actions?: {
		icon: string;
		label: string;
		action: (value: string) => void;
	}[];
}

/** Available field types and respective meta data (order matters for pattern matching) */
export const FIELD_DEFS: { [t in FieldType]: FieldDef } = {
	// [FieldType.CryptoeAddress]: {
	// 	type: FieldType.CryptoeAddress,
	// 	pattern: /.*/,
	// 	matchPattern: /.*/,
	// 	mask: true,
	// 	multiline: false,
	// 	icon: 'user',
	// 	get name() {
	// 		return $l('Address');
	// 	},
	// 	format(value, masked) {
	// 		return masked ? value.replace(/./g, '\u2022') : value;
	// 	}
	// },
	[FieldType.Username]: {
		type: FieldType.Username,
		pattern: /.*/,
		matchPattern: /.*/,
		mask: false,
		multiline: false,
		icon: 'sym_r_account_circle',
		get name() {
			return 'Username';
		}
	},
	[FieldType.Password]: {
		type: FieldType.Password,
		pattern: /.*/,
		matchPattern: /.*/,
		mask: true,
		multiline: true,
		icon: 'sym_r_key',
		get name() {
			return 'Password';
		},
		format(value, masked) {
			return masked ? value.replace(/./g, '\u2022') : value;
		}
	},
	[FieldType.ApiSecret]: {
		type: FieldType.ApiSecret,
		pattern: /.*/,
		matchPattern: /.*/,
		mask: true,
		multiline: true,
		icon: 'sym_r_key',
		get name() {
			return 'API Secret';
		},
		format(value, masked) {
			return masked ? value.replace(/./g, '\u2022') : value;
		}
	},
	[FieldType.Mnemonic]: {
		type: FieldType.Mnemonic,
		pattern: /.*/,
		matchPattern: /.*/,
		mask: true,
		multiline: true,
		icon: 'sym_r_key',
		get name() {
			return 'Mnemonic';
		},
		format(value, masked) {
			return masked ? value.replace(/./g, '\u2022') : value;
		}
	},
	[FieldType.Email]: {
		type: FieldType.Email,
		pattern: /(.*)@(.*)/,
		matchPattern: /^[\w-\.]+@([\w-]+\.)+[\w-]{2,8}$/,
		mask: false,
		multiline: false,
		icon: 'sym_r_mail',
		get name() {
			return 'Email Address';
		}
	},
	[FieldType.Url]: {
		type: FieldType.Url,
		pattern: /.*/,
		matchPattern:
			/[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,8}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)/i,
		mask: false,
		multiline: false,
		icon: 'sym_r_captive_portal',
		get name() {
			return 'URL';
		},
		actions: [
			{
				icon: 'webActive',
				label: $l('Open'),
				action: (value: string) =>
					openExternalUrl(value.startsWith('http') ? value : `https://${value}`)
			}
		]
	},
	[FieldType.Date]: {
		type: FieldType.Date,
		pattern: /^\d{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[0-1])$/,
		matchPattern: /^\d{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[0-1])$/,
		mask: false,
		multiline: false,
		icon: 'sym_r_calendar_month',
		get name() {
			return 'Date';
		},
		format(value) {
			return new Date(value).toLocaleDateString();
		}
	},
	[FieldType.Month]: {
		type: FieldType.Month,
		pattern: /^\d{4}-(0[1-9]|1[012])$/,
		matchPattern: /^\d{4}-(0[1-9]|1[012])$/,
		mask: false,
		multiline: false,
		icon: 'sym_r_calendar_today',
		get name() {
			return 'Month';
		}
	},
	[FieldType.Credit]: {
		type: FieldType.Credit,
		pattern: /.*/,
		matchPattern: /^\d{16}/,
		mask: true,
		multiline: false,
		icon: 'sym_r_credit_card',
		get name() {
			return 'Credit Card Number';
		},
		format(value, masked) {
			const parts: any = [];

			for (let i = 0; i < value.length; i += 4) {
				const part = value.slice(i, i + 4);
				parts.push(
					masked && i < value.length - 4 ? part.replace(/./g, '\u2022') : part
				);
			}

			return parts.join(' ');
		}
	},
	[FieldType.Phone]: {
		type: FieldType.Phone,
		pattern: /.*/,
		matchPattern: /\d+/,
		mask: false,
		multiline: false,
		icon: 'sym_r_call_log',
		get name() {
			return 'Phone Number';
		}
	},
	[FieldType.Pin]: {
		type: FieldType.Pin,
		pattern: /.*/,
		matchPattern: /\d+/,
		mask: true,
		multiline: false,
		icon: 'sym_r_fiber_pin',
		get name() {
			return 'PIN';
		},
		format(value, masked) {
			return masked ? value.replace(/./g, '\u2022') : value;
		}
	},
	[FieldType.Text]: {
		type: FieldType.Text,
		pattern: /.*/,
		matchPattern: /.*/,
		mask: false,
		multiline: true,
		icon: 'sym_r_sticky_note',
		get name() {
			return 'Plain Text';
		}
	},
	[FieldType.Note]: {
		type: FieldType.Note,
		pattern: /.*/,
		matchPattern: /(.*)(\n)?(.*)/,
		mask: false,
		multiline: true,
		icon: 'sym_r_assignment',
		get name() {
			return 'Richtext / Markdown';
		},
		format(value: string) {
			return value.split('\n')[0] || '';
		}
	},
	[FieldType.Totp]: {
		type: FieldType.Totp,
		pattern: /^([A-Z2-7=]{8})+$/i,
		matchPattern: /^([A-Z2-7=]{8})+$/i,
		mask: false,
		multiline: false,
		icon: 'sym_r_history',
		get name() {
			return 'One-Time Password';
		},
		async transform(value: string) {
			return await totp(base32ToBytes(value));
		}
	}
};

export class Field extends Serializable {
	constructor(vals: Partial<Field> = {}) {
		super();
		Object.assign(this, vals);
	}

	/**
	 * field type, determining meta data via the corresponding field definition
	 * in [[FIELD_DEFS]]
	 */
	type: FieldType = FieldType.Text;
	/** field name */
	name = '';
	/** field content */
	value = '';

	get def(): FieldDef {
		return FIELD_DEFS[this.type] || FIELD_DEFS[FieldType.Text];
	}

	get icon() {
		return this.def.icon;
	}

	async transform() {
		return this.def.transform
			? await this.def.transform(this.value)
			: this.value;
	}

	format(masked: boolean) {
		return this.def.format ? this.def.format(this.value, masked) : this.value;
	}

	protected _fromRaw(raw: any) {
		if (!raw.type) {
			raw.type = guessFieldType(raw);
		}
		return super._fromRaw(raw);
	}
}

/** Normalizes a tag value by removing invalid characters */
export function normalizeTag(tag: string): Tag {
	return tag.replace(',', '');
}

export enum AuditType {
	WeakPassword = 'weak_password',
	ReusedPassword = 'reused_password',
	CompromisedPassword = 'compromised_password',
	ExpiredItem = 'expired_item'
}

export interface AuditResult {
	type: AuditType;
	fieldIndex: number;
}

export class ItemHistoryEntry extends Serializable {
	constructor(item?: VaultItem) {
		super();
		if (item) {
			this.name = item.name;
			this.tags = item.tags;
			this.fields = item.fields;
			this.updated = item.updated;
			this.updatedBy = item.updatedBy;
		}
	}

	created: Date = new Date();

	updatedBy: AccountID = '';

	updated: Date = new Date();

	name = '';

	@AsSerializable(Field)
	fields: Field[] = [];

	tags: Tag[] = [];
}

export const ITEM_HISTORY_ENTRIES_LIMIT = 10;

export enum VaultType {
	Default = 0,
	VC = 1,
	PrivateKey = 2,
	TerminusTotp = 3,
    OlaresSSHPassword = 4
}

/** Represents an entry within a vault */
export class VaultItem extends Serializable {
	constructor(vals: Partial<VaultItem> = {}) {
		super();
		Object.assign(this, vals);
	}

	/** unique identfier */
	id: VaultItemID = '';

	/** item name */
	name = '';

	type: VaultType = VaultType.Default;

	/** icon to be displayed for this item */
	icon?: string = undefined;

	/** item fields */
	@AsSerializable(Field)
	fields: Field[] = [];

	/** array of tags assigned with this item */
	tags: Tag[] = [];

	/** Date and time of last update */
	@AsDate()
	updated: Date = new Date();

	/** [[Account]] the item was last updated by */
	updatedBy: AccountID = '';

	/**
	 * @DEPRECATED
	 * Accounts that have favorited this item
	 */
	favorited: AccountID[] = [];

	/** attachments associated with this item */
	@AsSerializable(AttachmentInfo)
	attachments: AttachmentInfo[] = [];

	auditResults: AuditResult[] = [];

	@AsDate()
	lastAudited?: Date = undefined;

	/** Number of days after which the item expires */
	expiresAfter?: number = undefined;

	/** Expiration date, calculated based on [[updated]] and [[expiresAfter]] properties */
	get expiresAt() {
		if (!this.expiresAfter) {
			return undefined;
		}

		return add(this.updated, { days: this.expiresAfter });
	}

	/** item history (first is the most recent change) */
	@AsSerializable(ItemHistoryEntry)
	history: ItemHistoryEntry[] = [];
}

/** Creates a new vault item */
export async function createVaultItem({
	name = 'Unnamed',
	fields = [],
	tags = [],
	icon,
	type = VaultType.Default,
	attachments = [],
	auditResults,
	lastAudited,
	expiresAfter,
    id
}: Partial<VaultItem>): Promise<VaultItem> {
	return new VaultItem({
		name,
		fields,
		tags,
		icon,
		type,
		attachments,
		auditResults,
		lastAudited,
		expiresAfter,
		id: id ? id : await uuid()
	});
}

/** Guesses the most appropriate field type based on field name and value */
export function guessFieldType({
	name,
	value = '',
	masked = false
}: {
	name: string;
	value?: string;
	masked?: boolean;
}): FieldType {
	if (masked) {
		return FieldType.Password;
	}

	const matchedTypeByName = Object.keys(FIELD_DEFS).filter((fieldType) =>
		new RegExp(fieldType, 'i').test(name)
	)[0] as FieldType;

	if (matchedTypeByName) {
		return matchedTypeByName;
	}

	// We skip some because they can match anything, and are only really valuable when matched by name
	const fieldTypesToSkipByValue = [FieldType.Username, FieldType.Password];

	const matchedTypeByValue = Object.keys(FIELD_DEFS)
		// @ts-ignore this is a string, deal with it, TypeScript (can't `as` as well)
		.filter((fieldType) => !fieldTypesToSkipByValue.includes(fieldType))
		.filter((fieldType) =>
			FIELD_DEFS[fieldType].matchPattern.test(value)
		)[0] as FieldType;

	if (value !== '' && matchedTypeByValue) {
		return matchedTypeByValue;
	}

	return FieldType.Text;
}

export interface ExchangeTemplate {
	name: string;
	icon: string;
	fields: { name: string; value?: string; type: FieldType }[];
}

export const EXCHANGE_TEMPLATES: ExchangeTemplate[] = [
	{
		name: 'Binance',
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Account';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			}
		]
	},
	{
		name: 'Deribit',
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Account';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			}
		]
	},
	{
		name: 'OKex',
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Account';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			},
			{
				get name() {
					return 'Passphrase';
				},
				type: FieldType.Password
			}
		]
	},
	{
		name: 'Huobi',
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Account';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			}
		]
	},
	{
		name: 'Bybit',
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Account';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			}
		]
	}
];

export interface VCTemplate {
	name: string;
	icon: string;
	fields: { name: string; value?: string; type: FieldType }[];
}

export const VC_TEMPLATES: VCTemplate[] = [
	{
		name: 'Facebook',
		icon: 'vault',
		fields: [
			{
				get name() {
					return $l('Name');
				},
				type: FieldType.Username
			},
			{
				get name() {
					return $l('Manifest');
				},
				type: FieldType.Note
			},
			{
				get name() {
					return $l('Credential');
				},
				type: FieldType.Note
			}
		]
	},
	{
		name: 'Twitter',
		icon: 'vault',
		fields: [
			{
				get name() {
					return $l('Name');
				},
				type: FieldType.Username
			},
			{
				get name() {
					return $l('Manifest');
				},
				type: FieldType.Note
			},
			{
				get name() {
					return $l('Credential');
				},
				type: FieldType.Note
			}
		]
	},
	{
		name: 'Google',
		icon: 'vault',
		fields: [
			{
				get name() {
					return $l('Name');
				},
				type: FieldType.Username
			},
			{
				get name() {
					return $l('Manifest');
				},
				type: FieldType.Note
			},
			{
				get name() {
					return $l('Credential');
				},
				type: FieldType.Note
			}
		]
	}
];

export interface CryptoTemplate {
	name: string;
	fields: { name: string; value?: string; type: FieldType }[];
	icon: string;
	tags: string[];
}

export interface ItemTemplate {
	id: string;
	name?: string;
	fields: { name: string; value?: string; type: FieldType }[];
	icon: string;
	iconSrc?: string;
	toString(): string;
	subTitle?: string;
	attachment?: boolean;
}

export const ITEM_TEMPLATES: ItemTemplate[] = [
	// {
	//     toString: () => $l("DID"),
	//     icon: "vault",
	//     fields: [
	//         {
	//             get name() {
	//                 return $l("Username");
	//             },
	//             type: FieldType.Username,
	//         },
	//         {
	//             get name() {
	//                 return $l("Password");
	//             },
	//             type: FieldType.Password,
	//         },
	//         {
	//             get name() {
	//                 return $l("URL");
	//             },
	//             type: FieldType.Url,
	//         },
	//     ],
	// },
	{
		id: 'vc',
		toString: () => $l('VC'),
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Username';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Password';
				},
				type: FieldType.Password
			},
			{
				get name() {
					return 'URL';
				},
				type: FieldType.Url
			}
		]
	},
	{
		id: 'exchange',
		toString: () => $l('Exchange'),
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Exchange';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Key';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Secret';
				},
				type: FieldType.Password
			},
			{
				get name() {
					return 'Passphrase';
				},
				type: FieldType.Password
			},
			{
				get name() {
					return 'Subaccount';
				},
				type: FieldType.Text
			}
		]
	},
	{
		id: 'crypto',
		toString: () => $l('Crypto Wallet'),
		icon: 'vault',
		fields: [
			{
				get name() {
					return 'Name';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Mnemonic';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Private Key';
				},
				type: FieldType.Password
			}
		],
		attachment: true
	},
	{
		id: 'web',
		toString: () => $l('Website / App'),
		icon: 'web',
		fields: [
			{
				get name() {
					return 'Username';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Password';
				},
				type: FieldType.Password
			},
			{
				get name() {
					return 'URL';
				},
				type: FieldType.Url
			}
		]
	},
	{
		id: 'computer',
		toString: () => $l('Computer'),
		icon: 'computer',
		fields: [
			{
				get name() {
					return 'Username';
				},
				type: FieldType.Username
			},
			{
				get name() {
					return 'Password';
				},
				type: FieldType.Password
			}
		]
	},
	{
		id: 'credit_card',
		toString: () => $l('Credit Card'),
		icon: 'creditCard',
		fields: [
			{
				get name() {
					return 'Password';

				},
				type: FieldType.Credit
			},
			{
				get name() {
					return 'Card Owner';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Valid Until';
				},
				type: FieldType.Month
			},
			{
				get name() {
					return 'CVC';
				},
				type: FieldType.Pin
			},
			{
				get name() {
					return 'PIN';
				},
				type: FieldType.Pin
			}
		]
	},
	{
		id: 'bank',
		toString: () => $l('Bank Account'),
		icon: 'bank',
		fields: [
			{
				get name() {
					return 'Account Owner';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'IBAN';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'BIC';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Card PIN';
				},
				type: FieldType.Pin
			}
		]
	},
	{
		id: 'wifi',
		toString: () => $l('WIFI Password'),
		icon: 'wifi',
		fields: [
			{
				get name() {
					return 'Name';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Password';
				},
				type: FieldType.Password
			}
		]
	},
	{
		id: 'passport',
		toString: () => $l('Passport'),
		icon: 'passport',
		fields: [
			{
				get name() {
					return 'Full Name';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Passport Number';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Country';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Birthdate';
				},
				type: FieldType.Date
			},
			{
				get name() {
					return 'Birthplace';
				},
				type: FieldType.Text
			},
			{
				get name() {
					return 'Issued On';
				},
				type: FieldType.Date
			},
			{
				get name() {
					return 'Expires';
				},
				type: FieldType.Date
			}
		]
	},
	// {
	//     id: 'note',
	//     toString: () => $l("Note"),
	//     icon: "note",
	//     fields: [
	//         {
	//             get name() {
	//                 return $l("Note");
	//             },
	//             type: FieldType.Note,
	//         },
	//     ],
	// },
	{
		id: 'authenticator',
		toString: () => $l('Authenticator'),
		icon: 'authenticator',
		fields: [
			{
				get name() {
					return 'One-Time Password';
				},
				type: FieldType.Totp
			}
		]
	},
	{
		id: 'document',
		toString: () => $l('Document'),
		icon: 'document',
		fields: [],
		attachment: true
	},
	{
		id: 'custom',
		toString: () => $l('Custom'),
		icon: 'custom',
		fields: []
	}
];
export const cloneItemTemplates = () => {
    return cloneDeep(ITEM_TEMPLATES) as ItemTemplate[];
}