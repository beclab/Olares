import { RouteLocationRaw } from 'vue-router';

export interface Approval {
	id: number;
	taskId: number | null;
	data: {
		params?: any;
		origin?: string;
		requestDefer?: Promise<any>;
		approvalType?: string;
		routerPath?: string;
		redirectRoute?: RouteLocationRaw;
		session: any;
		sessionId?: number;
	};
	winProps: any;
	resolve?(params?: any): void;
	reject?(err: Error): void;
}

export interface Controller {
	getApproval: () => Promise<Approval | null>;
	resolveApproval: (
		data?: any,
		forceReject?: boolean
	) => Promise<Approval | null>;
	clearAllApproval: () => Promise<void>;
	rejectApproval: (err?: string, isInternal?: boolean) => Promise<void>;

	sendUnlocked: (data: string) => void;

	sendLocked: () => void;

	sendAppState: (
		appState: string,
		accountsData: string,
		mnemonicsData: string,
		currentAccountId: string
	) => Promise<void>;

	requestPassword: () => string | undefined;

	getConnectedSite: () => Promise<any>;

	removeConnectedSite: (origin: string) => Promise<void>;

	removeConnectedSites: () => Promise<void>;

	changeAccount: (didKey: string) => void;

	setAutofillBadgeEnable: (enable: boolean) => Promise<any>;

	getAutofillBadgeEnable: () => Promise<boolean>;

	setRssBadgeEnable: (enable: boolean) => Promise<any>;

	getRssBadgeEnable: (enable: boolean) => Promise<any>;

	setApprovalBadgeEnable: (enable: boolean) => Promise<any>;

	getApprovalBadgeEnable: () => Promise<boolean>;

	storeGetItem: (key: string) => Promise<any>;

	storeSetItem: (key: string, value: any) => Promise<void>;

	storeRemoveItem: (key: string) => Promise<void>;

	getCurrentTab: () => Promise<any>;

	toggleBexDisplay: () => Promise<void>;

	autofillById: (id: string) => Promise<void>;
}
