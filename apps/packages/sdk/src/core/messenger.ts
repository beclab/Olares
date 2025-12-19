export type MessageData = { [param: string]: string };

/**
 * A message to be sent to a user
 */
export abstract class Message<T extends MessageData> {
	/** Message title */
	abstract get title(): string;

	/** Template name */
	abstract readonly template: string;

	constructor(public readonly data: T) {}
}

export class EmailAuthMessage extends Message<{
	code: string;
	requestId: string;
}> {
	template = 'email-auth';

	get title() {
		const appName = process.env.PL_APP_NAME;
		return `${
			appName ? appName + ' ' : ''
		}Email Verification (Request ID: ${this.data.requestId})`;
	}
}

abstract class OrgInviteMessage extends Message<{
	orgName: string;
	invitedBy: string;
	acceptInviteUrl: string;
}> {}

export class JoinOrgInviteMessage extends OrgInviteMessage {
	template = 'join-org-invite';

	get title() {
		return `${this.data.invitedBy} wants you to join the "${this.data.orgName}" org on ${process.env.PL_APP_NAME}!`;
	}
}

export class ConfirmMembershipInviteMessage extends OrgInviteMessage {
	template = 'confirm-org-member-invite';

	get title() {
		return `Confirm your membership for the "${this.data.orgName}" org on ${process.env.PL_APP_NAME}!`;
	}
}

export class JoinOrgInviteAcceptedMessage extends Message<{
	orgName: string;
	invitee: string;
	confirmMemberUrl: string;
}> {
	template = 'join-org-invite-accepted';

	get title() {
		return `${this.data.invitee} has accepted your invite!`;
	}
}

export class JoinOrgInviteCompletedMessage extends Message<{
	orgName: string;
	openAppUrl: string;
}> {
	template = 'join-org-invite-completed';

	get title() {
		return `You have successfully joined ${this.data.orgName} on ${process.env.PL_APP_NAME}!`;
	}
}

export class FailedLoginAttemptMessage extends Message<{ location: string }> {
	template = 'failed-login-attempt';

	get title() {
		const appName = process.env.PL_APP_NAME;
		return `${appName ? appName + ' ' : ''}Failed Login Attempt from ${
			this.data.location
		})`;
	}
}

export class NewLoginMessage extends Message<{ location: string }> {
	template = 'new-login';

	get title() {
		const appName = process.env.PL_APP_NAME;
		return `${appName ? appName + ' ' : ''}New Login from ${
			this.data.location
		})`;
	}
}

export class PlainMessage extends Message<{ message: string }> {
	template = 'plain';

	get title() {
		return 'Padloc Error Notification';
	}
}

/**
 * Generic interface for sending messages to users
 */
export interface Messenger {
	/**
	 * Sends a message to a given address
	 */
	send<T extends MessageData>(addr: string, msg: Message<T>): Promise<void>;
}

/**
 * Stub implementation of the [[Messenger]] interface.  Simply stores messages
 * in memory. Useful for testing purposes.
 */
export class StubMessenger implements Messenger {
	/**
	 * An array of messages passed to the [[send]] method. Sorted from
	 * most recent to oldest.
	 */
	messages: { recipient: string; message: Message<any> }[] = [];

	async send<T extends MessageData>(recipient: string, message: Message<T>) {
		this.messages.unshift({ recipient, message });
	}

	/**
	 * Returns the most recent message sent to `addr`.
	 */
	lastMessage(addr: string): Message<any> | null {
		const msg = this.messages.find(({ recipient }) => recipient === addr);
		return msg ? msg.message : null;
	}
}

export class ConsoleMessenger implements Messenger {
	async send(recipient: string, message: Message<any>) {
		console.log(
			`Message sent to ${recipient}: ${JSON.stringify(message.data)}`
		);
	}
}
