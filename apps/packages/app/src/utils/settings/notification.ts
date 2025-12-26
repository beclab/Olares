/* eslint-disable @typescript-eslint/no-unused-vars */
export const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export type NotificationCredential = Record<string, string>;
export type NotificationRecipientData = Record<string, string>;

export interface NotificationCredentialValidator {
	params: CredentialProperty[];

	validate(data: NotificationCredential): boolean;
}

export class NotificationApplicationCredentialValidator
	implements NotificationCredentialValidator
{
	params = [
		{ name: 'appId', required: true, type: 'string' },
		{ name: 'appName', required: true, type: 'string' },
		{ name: 'group', required: true, type: 'string' },
		{ name: 'dataType', required: true, type: 'string' },
		{ name: 'op', required: true, type: 'string' },
		{ name: 'url', required: true, type: 'string' }
	];

	validate(credential: NotificationCredential) {
		if (!credential) {
			return false;
		}
		if (!credential.group || !credential.dataType || !credential.op) {
			return false;
		}
		return true;
	}
}

export class NotificationSmtpCredentialValidator
	implements NotificationCredentialValidator
{
	params = [
		{ name: 'host', required: true, type: 'string' },
		{ name: 'port', required: true, type: 'number', default: 25 },
		{ name: 'email', required: true, type: 'string' },
		{ name: 'username', required: false, type: 'string' },
		{ name: 'password', required: false, type: 'string' },
		{ name: 'secure', required: false, type: 'boolean', default: false }
	];

	validate(_data: NotificationCredential) {
		return true;
	}
}
export class NotificationSMSCredentialValidator
	implements NotificationCredentialValidator
{
	params = [
		{ name: 'apiKey', required: false, type: 'string' },
		{ name: 'authToken', required: false, type: 'string' }
	];

	validate(_data: NotificationCredential) {
		return true;
	}
}

export class NotificationWebhookCredentialValidator
	implements NotificationCredentialValidator
{
	params = [
		{ name: 'username', required: false, type: 'string' },
		{ name: 'password', required: false, type: 'string' },
		{ name: 'token', required: false, type: 'string' }
	];

	validate(_data: NotificationCredential) {
		return true;
	}
}

export interface CredentialProperty {
	name: string;
	required: boolean;
	default?: any;
	type?: string;
}

export interface TerminusNotificationSenderTemplate {
	icon: string;
	type: string;
	name: string;
	recipientType: string;
	description: string;
	credential: NotificationCredentialValidator;
}

export const senderTemplates: TerminusNotificationSenderTemplate[] = [
	{
		icon: '',
		type: 'Application',
		name: 'Application',
		recipientType: 'Firebase',
		description: 'Send notification use Olares Application',
		credential: new NotificationApplicationCredentialValidator()
	},
	{
		icon: '',
		type: 'SMTP',
		name: 'Email SMTP',
		recipientType: 'Email',
		description: 'Send notification use SMTP through email',
		credential: new NotificationSmtpCredentialValidator()
	},
	{
		icon: '',
		type: 'SMS',
		name: 'SMS',
		recipientType: 'SMS',
		description: 'Send notification use SMS',
		credential: new NotificationSMSCredentialValidator()
	},
	{
		icon: '',
		type: 'Webhook',
		name: 'Webhook',
		recipientType: 'Webhook',
		description: 'Send notification use webhook',
		credential: new NotificationWebhookCredentialValidator()
	},
	{
		icon: '',
		type: 'Lark',
		name: 'Lark',
		recipientType: 'NoNeed',
		description: 'Send notification use Lark',
		credential: new NotificationWebhookCredentialValidator()
	},
	{
		icon: '',
		type: 'Slack',
		name: 'Slack',
		recipientType: 'NoNeed',
		description: 'Send notification use Slack',
		credential: new NotificationWebhookCredentialValidator()
	}
];

export class NotificationRecipient {
	data: NotificationRecipientData;
	constructor(data: NotificationRecipientData) {
		this.data = data;
	}
}

export interface NotificationRecipientValidator {
	//type: RecipientType;
	params: string[];
	validate(data: NotificationRecipientData): boolean;
}

export class NoNeedRecipient extends NotificationRecipient {
	constructor(data: NotificationRecipientData) {
		super(data);
	}
}

export class NoNeedRecipientValidator
	implements NotificationRecipientValidator
{
	params: string[] = [];

	validate(_data: NotificationRecipientData): boolean {
		return true;
	}
}

export class FirebaseRecipient extends NotificationRecipient {
	token: string;

	constructor(data: NotificationRecipientData) {
		super(data);
		this.token = data.token;
	}
}

export class FirebaseRecipientValidator
	implements NotificationRecipientValidator
{
	params: string[] = ['token'];

	validate(data: NotificationRecipientData): boolean {
		if (data.token) {
			return true;
		}
		return false;
	}
}

export class EmailRecipient extends NotificationRecipient {
	email: string;

	constructor(data: NotificationRecipientData) {
		super(data);
		this.data = data;
		this.email = data.email;
	}
}

export class EmailRecipientValidator implements NotificationRecipientValidator {
	params: string[] = ['email'];

	validate(data: NotificationRecipientData): boolean {
		if (data.email) {
			return true;
		}
		return false;
	}
}

export class PhoneRecipient extends NotificationRecipient {
	countryCode: string;
	phoneNumber: string;

	constructor(data: NotificationRecipientData) {
		super(data);
		this.data = data;
		this.countryCode = data.countryCode;
		this.phoneNumber = data.phoneNumber;
	}
}

export class PhoneRecipientValidator implements NotificationRecipientValidator {
	// type = RecipientType.Phone;
	params: string[] = ['countryCode', 'phoneNumber'];

	validate(data: NotificationRecipientData): boolean {
		if (data.countryCode && data.phoneNumber) {
			return true;
		}
		return false;
	}
}

export class WebhookRecipient extends NotificationRecipient {
	url: string;

	constructor(data: NotificationRecipientData) {
		super(data);
		this.data = data;
		this.url = data.url;
	}
}

export class WebhookRecipientValidator
	implements NotificationRecipientValidator
{
	params: string[] = ['url'];

	validate(data: NotificationRecipientData): boolean {
		if (data.url) {
			return true;
		}
		return false;
	}
}

export interface TerminusNotificationRecipientTemplate {
	icon: string;
	type: string;
	name: string;
	validator: NotificationRecipientValidator;
}

export const recipientTemplates: TerminusNotificationRecipientTemplate[] = [
	{
		icon: '',
		type: 'NoNeed',
		name: 'NoNeed',
		validator: new NoNeedRecipientValidator()
	},
	{
		icon: '',
		type: 'Firebase',
		name: 'Firebase',
		validator: new FirebaseRecipientValidator()
	},
	{
		icon: '',
		type: 'Email',
		name: 'Email',
		validator: new EmailRecipientValidator()
	},
	{
		icon: '',
		type: 'Phone',
		name: 'Phone',
		validator: new PhoneRecipientValidator()
	},
	{
		icon: '',
		type: 'Webhook',
		name: 'Webhook',
		validator: new WebhookRecipientValidator()
	}
];
