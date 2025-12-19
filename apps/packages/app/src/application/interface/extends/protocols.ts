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
