const STORAGE_KEY = 'olares.market.debug.flags';

export const MARKET_DEBUG_ENABLED = !!process.env.DEV;

export interface MarketDebugFlags {
	bypassV2Install: boolean;
	bypassV2Upgrade: boolean;
}

const DEFAULT_FLAGS: MarketDebugFlags = {
	bypassV2Install: false,
	bypassV2Upgrade: false
};

export function readMarketDebugFlags(): MarketDebugFlags {
	if (!MARKET_DEBUG_ENABLED) return { ...DEFAULT_FLAGS };
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return { ...DEFAULT_FLAGS };
		const parsed = JSON.parse(raw);
		if (!parsed || typeof parsed !== 'object') return { ...DEFAULT_FLAGS };
		return { ...DEFAULT_FLAGS, ...parsed };
	} catch {
		return { ...DEFAULT_FLAGS };
	}
}

export function writeMarketDebugFlags(value: MarketDebugFlags) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
	} catch {
		// debug-only flags: never block the user on storage errors.
	}
}
