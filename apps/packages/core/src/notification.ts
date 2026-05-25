import axios, { AxiosResponse } from 'axios';

export type FireBaseToken = string;

// export enum PushOpType {
// 	Create,
// 	Query
// }

export enum NotificationResultCode {
	Success = 0,
	NetworkError = 1,
	PermissionError = 2,
	Waiting = 3,
	NotFound = 4,
	Error = 1000
}

export interface NotificationResult {
	code: NotificationResultCode;
	message?: string;
	data?: number;
}

export enum MessageTopic {
	PONG = 'pong',
	SIGN = 'sign',
	CANCEL_SIGN = 'cancel_sign',
	Data = 'data',
	Notification = 'notification'
}

export interface SignMessageData {
	callback_url: string;
	sign_body: {
		[key: string]: any;
	};
	sign_eth?: {
		domain: {
			[key: string]: any;
		};
		types: {
			[key: string]: any;
		};
		data: {
			[key: string]: any;
		};
		primaryType: string;
	};
	sign_vc?: {
		type: string;
		name: string;
		request_path: string;
		data: {
			[key: string]: any;
		};
	};
}

export interface AppInfo {
	id: string;
	icon: string;
	title: string;
}

export interface MessageData {
	id: string;
	template_id?: string;
	vars?: {
		[key: string]: any;
	};
	data?: {
		[key: string]: any;
	};
	sign?: SignMessageData;
}

export interface MessageBody {
	topic: MessageTopic;
	event: string;
	terminusName: string;
	collapse_key?: string;
	notification?: {
		body: string;
		title: string;
	};
	app?: AppInfo;
	message?: MessageData;
	android?: {
		priority?: string;
	};
	apns?: {
		headers?: {
			'apns-priority': string;
		};
	};
	webpush?: {
		headers?: {
			Urgency: string;
		};
	};
}

export interface NotificationMessage {
	id: number;
	recipient?: { token?: string };
	message: MessageBody;
}

export interface SystemMessage {
	receiver: {
		kind: string;
		apiVersion: string;
		metadata: {
			name: string;
			creationTimestamp: string | null;
			labels: {
				app: string;
				type: string;
			};
		};
		spec: {
			webhook: {
				enabled: boolean;
				url: string;
			};
		};
		status: any;
	};
	alert: {
		alerts: [
			{
				status: string;
				labels: {
					payload: string;
					type: string;
					version: string;
					namespace: string;
				};
				annotations: {
					message: string;
				};
				startsAt: string;
				endsAt: string;
				generatorURL?: string;
				fingerprint?: string;
			}
		];
	};
}

export function createSystemMessage(
	username: string,
	payload: string,
	message: string
): SystemMessage {
	return {
		receiver: {
			kind: 'Receiver',
			apiVersion: 'notification.kubesphere.io/v2beta2',
			metadata: {
				name: 'sys-receiver',
				creationTimestamp: new Date().toISOString(),
				labels: {
					app: 'notification-manager',
					type: 'global'
				}
			},
			spec: {
				webhook: {
					enabled: true,
					url:
						'http://notifications-server.user-space-' +
						username +
						'/notification/system/push'
				}
			},
			status: {}
		},
		alert: {
			alerts: [
				{
					status: '',
					labels: {
						namespace: 'user-system-' + username,
						payload: payload,
						type: 'notification',
						version: 'v1'
					},
					annotations: {
						message: message
					},
					startsAt: new Date().toISOString(),
					endsAt: new Date().toISOString(),
					generatorURL: '',
					fingerprint: ''
				}
			]
		}
	};
}

export async function postSystemNotification(
	message: SystemMessage
): Promise<boolean> {
	try {
		const data: AxiosResponse = await axios.post(
			'http://notification-manager-svc.kubesphere-monitoring-system:19093/api/v2/notifications',
			message
		);
		console.log(data.data);
		if (data.data == 'notification center response body') {
			return true;
		}
		return true;
	} catch (e) {
		console.log(e);
		return false;
	}
}
