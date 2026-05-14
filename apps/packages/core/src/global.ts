enum SUPPORT_ENVIRONMENT {
	EN = 'en',
	CN = 'cn'
}

const DID_GATE_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/did',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/did'
};

const OLARES_SPACE_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/space',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/space'
};

const FIREBASE_PUSH_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/push',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/push'
};

const FRP_LIST_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/frp',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/frp'
};

const TAILSCALE_CONTROLPLANE_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/controlplane',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/controlplane'
};

const DOWNLOAD_CDN_URL = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://cdn.olares.com',
	[SUPPORT_ENVIRONMENT.CN]: 'https://cdn.olares.cn'
};

const VC = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/vc',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/vc'
};

const MARKET_PROVIDER = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/market',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/market'
};

const OLARES_UPGRADE = {
	[SUPPORT_ENVIRONMENT.EN]: 'https://api.olares.com/upgrade',
	[SUPPORT_ENVIRONMENT.CN]: 'https://api.olares.cn/upgrade'
};

const userNameToEnvironment = (name: string) => {
	if (name.endsWith('olares.cn')) {
		return SUPPORT_ENVIRONMENT.CN;
	}
	return SUPPORT_ENVIRONMENT.EN;
};

export const GolbalHost = {
	VC,
	DOWNLOAD_CDN_URL,
	TAILSCALE_CONTROLPLANE_URL,
	FRP_LIST_URL,
	FIREBASE_PUSH_URL,
	OLARES_SPACE_URL,
	DID_GATE_URL,
	MARKET_PROVIDER,
	OLARES_UPGRADE,
	SUPPORT_ENVIRONMENT,
	userNameToEnvironment
};
