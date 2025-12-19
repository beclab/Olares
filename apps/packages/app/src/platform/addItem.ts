import {
	Field,
	FieldType,
	cloneItemTemplates,
	VaultItem
} from '@didvault/sdk/src/core';
import { createVaultItem } from '@didvault/sdk/src/core/item';

import { app } from '../globals';
import { Router } from 'vue-router';
import { useMenuStore } from '../stores/menu';
import { bexVaultUpdate } from 'src/utils/bexFront';
import { useUserStore } from 'src/stores/user';
import { useVaultStore } from 'src/stores/vault';
import protobuf from 'protobufjs';
import base32 from 'hi-base32';

export async function updateUIToAddWeb(
	identify: string,
	router: Router,
	username = '',
	password = '',
	direct = false
) {
	console.log('approvalTypeRef-2');
	const userStore = useUserStore();
	if (!(await userStore.unlockFirst())) {
		return;
	}
	const meunStore = useMenuStore();
	const selectedTemplate = cloneItemTemplates().find((i) => i.id == 'web');
	if (!selectedTemplate) {
		return;
	}
	meunStore.isEdit = !direct;

	const field = selectedTemplate.fields.find((t) => t.type === FieldType.Url);
	if (field) {
		if (identify.startsWith('http')) {
			try {
				const urlObj = new URL(identify);
				const baseUrl = urlObj.origin;
				field.value = baseUrl;
			} catch (error) {
				field.value = identify;
			}
		} else {
			field.value = identify;
		}
	}
	if (username) {
		const nameField = selectedTemplate.fields.find(
			(t) => t.type === FieldType.Username
		);
		if (nameField) {
			nameField.value = username;
		}
	}
	if (password) {
		const passwordField = selectedTemplate.fields.find(
			(t) => t.type === FieldType.Password
		);
		if (passwordField) {
			passwordField.value = password;
		}
	}

	// const item: any = await addNewItem(
	// 	// selectedTemplate.toString() || '',
	// 	username ? username : '',
	// 	selectedTemplate.icon,
	// 	selectedTemplate.fields,
	// 	[]
	// );
	const name = !direct ? '' : username || '';
	const editing_item = await addItem(
		name,
		selectedTemplate.icon,
		selectedTemplate.fields,
		[]
	);

	// menuStore.isEdit = true;

	if (!editing_item) {
		return;
	}

	// if (editing_item) {
	// 	const id = editing_item.id;
	// 	// emits('toolabClick', id);
	// }
	if (!direct && editing_item) {
		const vaultStore = useVaultStore();
		vaultStore.editing_item = editing_item;
	}

	if (editing_item && router && !direct) {
		router.push({
			path: '/items/' + editing_item.id
		});
	}
}

export async function addItem(
	name: string,
	icon: string,
	fields: any,
	tags: string[],
	vault = app.mainVault,
	auditResults = [],
	lastAudited = undefined,
	expiresAfter = undefined,
	attachments = [],
	isNew?: boolean,
	id?: string
): Promise<VaultItem | undefined> {
	// const vault = app.mainVault;
	if (!vault) {
		return;
	}

	const item: VaultItem = await app.createItem({
		name,
		vault,
		icon,
		fields: isNew
			? fields
			: fields.map((f: Field) => new Field({ ...f, value: f.value || '' })),
		tags,
		auditResults,
		lastAudited,
		expiresAfter: expiresAfter && expiresAfter > 0 ? expiresAfter : undefined,
		attachments,
		id
	});

	console.log('addItem item', item);

	bexVaultUpdate();
	return item;
}

export async function addNewItem(
	name: string,
	icon: string,
	fields: any,
	tags: string[],
	vault = app.mainVault
): Promise<{ item: VaultItem; vault: any } | undefined> {
	console.log('addNewItem fields', fields);
	const item: VaultItem = await createVaultItem({
		name,
		fields: fields.map((f: Field) => new Field({ ...f, value: f.value || '' })),
		tags,
		icon
	});

	console.log('addItem item', item);

	bexVaultUpdate();
	return {
		item,
		vault
	};
}

const PROTO_DEFINITION = `
syntax = "proto2";
message OtpParameters {
  required bytes secret = 1;
  optional string name = 2 [default=""];
  optional string issuer = 3 [default=""];
  optional int32 algorithm = 4 [default=1];
  optional int32 digits = 5 [default=6];
  optional int32 type = 6 [default=2];
}
message MigrationPayload {
  repeated OtpParameters otpParameters = 1;
}
`;

export const decodeAuthenticatorMigrationUrl = (migrationUrl: string) => {
	const urlObj = new URL(migrationUrl);
	if (
		urlObj.protocol !== 'otpauth-migration:' ||
		!urlObj.searchParams.has('data')
	) {
		throw new Error('Invalid migration URL');
	}

	const root = protobuf.parse(PROTO_DEFINITION).root;
	const MigrationPayload = root.lookupType('MigrationPayload');

	const data = urlObj.searchParams.get('data');
	if (!data) {
		return [];
	}

	const base64Data = decodeURIComponent(data)
		.replace(/-/g, '+')
		.replace(/_/g, '/');

	const buffer = Buffer.from(base64Data, 'base64');
	if (buffer.length < 10) throw new Error('Data too short');

	try {
		const payload = MigrationPayload.decode(buffer);
		return (payload as any).otpParameters.map((param) => {
			const secret = base32.encode(param.secret).replace(/=/g, '');
			const type = param.type === 1 ? 'hotp' : 'totp';
			const params = new URLSearchParams({
				secret: secret,
				issuer: param.issuer,
				algorithm: ['SHA1', 'SHA256', 'SHA512'][param.algorithm - 1] || 'SHA1',
				digits: param.digits.toString(),
				period: '30'
			});
			return (
				`otpauth://${type}/${encodeURIComponent(param.issuer)}:` +
				`${encodeURIComponent(param.name)}?${params}`
			);
		});
	} catch (e) {
		console.error('Decode Error:', {
			offset: e.offset,
			buffer: buffer
				.slice(Math.max(0, e.offset - 16), e.offset + 16)
				.toString('hex')
		});
		throw new Error(`Protocol Buffers decode failed: ${e.message}`);
	}
};
