import { MessageBody, PrivateJwk } from '@bytetrade/core';

export interface NativeScanQRProtocol {
	protocol: string;
	method: (result: string) => Promise<boolean>;
	canResponseQRContent: (result: string) => Promise<boolean>;
	success?: () => Promise<void>;
}

export const commonResponseQRContent = (result: string, protocol: string) => {
	return result.startsWith(protocol + '://');
};

export const commonGetRealQRConent = (result: string) => {
	return result.split('://')[1];
};

export interface NativeSignProtocol {
	protocol: string;
	precheck: (message: MessageBody) => boolean;
	signAction?: (
		message: MessageBody,
		params: {
			did: string;
			privateJWK: PrivateJwk;
		}
	) => Promise<{
		callback_url: string;
		postData: any;
	}>;
	afterSign?: (message: MessageBody) => Promise<void>;
}
