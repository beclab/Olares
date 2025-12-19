import { App, AppState } from '@didvault/sdk/src/core/app';
import { getPlatform } from '@didvault/sdk/src/core';
import { CredentialManifest, GetSchemaResponse } from '@bytetrade/core';
import { PlatformAjaxSender } from './platform/platformAjaxSender';
import { busEmit } from './utils/bus';

let url = process.env.PL_SERVER_URL || global.location.origin + '/server';
if (process.env.NODE_ENV === 'development') {
	url = '/server';
}

const sender = new PlatformAjaxSender(url);
export let app = (global.app = new App(sender));

global.getPlatform = getPlatform;

export async function clearSenderUrl() {
	sender.url = '';
}

export function setSenderUrl({ url }: { url: string | undefined }) {
	if (url) {
		sender.url = url;
	} else {
		clearSenderUrl();
	}
}

export function getSenderUrl() {
	return sender.url;
}

const appSubscribeBlock = (state: AppState) => {
	busEmit('appSubscribe', state);
	busEmit('orgSubscribe', state.orgs);
};

export function resetAPP() {
	if (app) {
		app.unsubscribe(appSubscribeBlock);
	}
	app = global.app = new App(sender);
	app.subscribe(appSubscribeBlock);
}

resetAPP();

export enum ContactSource {
	Terminus = 1
}

export enum ContactLogType {
	Add = 0,
	Remove = 1,
	Friend_Request = 2,
	Friend_Accept = 3,
	Friend_Reject = 4,
	Friend_Request_Timeout = 5
}

export class Contact {
	id: number;
	name: string;
	did: string; //"did:cmt:"address
	source: ContactSource;
	create_at: number;
	update_at: number;
	delete_at: number;

	constructor(props?: Partial<Contact>) {
		props && Object.assign(this, props);
	}
}

export enum FreindRequestStatus {
	Wait_Friend_Operate = 0,
	Wait_Me_Operate = 1,
	Friend_Accepted = 2,
	Me_Accepted = 3,
	Friend_Rejected = 4,
	Me_Rejected = 5,
	Friend_Timeout = 6,
	Me_Timeout = 7
}

export interface FreindRequestLog {
	time: number;
	message?: string;
}

export interface FreindRequest {
	did: string;
	name: string;
	status: FreindRequestStatus;
	time: number;
	logs: FreindRequestLog[];
}

export interface ContactFreindRequestLog {
	id: number;
	type: ContactLogType;
	from: string;
	from_did: string;
	to: string;
	to_did: string;
	time: number;
}

export interface ContactRequestLog {
	id: number;
	type: ContactLogType;
	from: string;
	from_did: string;
	to: string;
	to_did: string;
	time: number;
}

export type ContactLog = ContactFreindRequestLog | ContactRequestLog | Contact;

export enum MessageStatus {
	Sent = 0,
	Send = 1,
	Unsent = 2,
	Resent = 3
}

export class Message {
	id: number;
	from: string;
	from_did: string;
	to: string;
	to_did: string;
	time: number;
	content: string;
	create_at: number;
	update_at: number;
	status: MessageStatus = 1;

	constructor(props?: Partial<Message>) {
		props && Object.assign(this, props);
	}
}

export interface MessageLastRead {
	did: string;
	time: number;
}

export class MessageThread {
	did: string;
	name: string;
	time: number;
	content: string;
	unread: number;

	constructor(props?: Partial<MessageThread>) {
		props && Object.assign(this, props);
	}
}

export interface ClientSchema {
	manifest: CredentialManifest;
	application_verifiable_credential: GetSchemaResponse;
}

export enum OrgMenu {
	DASHBOARD = 'Dashboard',
	MEMBERS = 'Members',
	VAULTES = 'Vaults',
	SETTINGS = 'Settings',
	INVITES = 'Invites'
}
