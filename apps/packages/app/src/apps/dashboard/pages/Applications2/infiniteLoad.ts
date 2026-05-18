export type Applications2GridInfiniteConfig = {
	fixedGridRows: number;
	gridColWidth: number;
	gridGap: number;
	wideColThreshold: number;
	wideFirstVisibleRows: number;
	appendMaxNarrow: number;
	appendMaxCol4: number;
	appendMaxCol5Plus: number;
	skeletonMaxNarrow: number;
	skeletonMaxWide: number;
	offsetBaseRatio: number;
	offsetMinPx: number;
	offsetRatioBump: number;
	offsetRatioBumpMinCols: number;
};

export const APPLICATIONS2_GRID_INFINITE_DEFAULTS: Applications2GridInfiniteConfig =
	{
		fixedGridRows: 8,
		gridColWidth: 510,
		gridGap: 32,
		wideColThreshold: 3,
		wideFirstVisibleRows: 4,
		appendMaxNarrow: 20,
		appendMaxCol4: 14,
		appendMaxCol5Plus: 10,
		skeletonMaxNarrow: 20,
		skeletonMaxWide: 14,
		offsetBaseRatio: 1.5,
		offsetMinPx: 800,
		offsetRatioBump: 0.135,
		offsetRatioBumpMinCols: 2
	};

export function computeColumnCount(
	containerWidth: number,
	cfg: Applications2GridInfiniteConfig
): number {
	const { gridColWidth, gridGap } = cfg;
	return Math.max(
		1,
		Math.floor((containerWidth - gridGap) / (gridColWidth + gridGap))
	);
}

export function itemsPerPageFromGrid(rows: number, cols: number): number {
	return rows * cols;
}

export function targetVisibleAfterReset(
	itemsPerPage: number,
	cols: number,
	cfg: Applications2GridInfiniteConfig
): number {
	if (cols <= cfg.wideColThreshold) return itemsPerPage;
	return Math.min(itemsPerPage, cols * cfg.wideFirstVisibleRows);
}

export function maxAppendStepCap(
	cols: number,
	cfg: Applications2GridInfiniteConfig
): number {
	if (cols <= cfg.wideColThreshold) return cfg.appendMaxNarrow;
	if (cols === cfg.wideColThreshold + 1) return cfg.appendMaxCol4;
	return cfg.appendMaxCol5Plus;
}

export function skeletonCountCap(
	cols: number,
	cfg: Applications2GridInfiniteConfig
): number {
	return cols > cfg.wideColThreshold
		? cfg.skeletonMaxWide
		: cfg.skeletonMaxNarrow;
}

export function clampVisibleLimit(visible: number, total: number): number {
	if (total <= 0) return visible;
	return Math.min(visible, total);
}

export function computeAppendStep(args: {
	visible: number;
	total: number;
	itemsPerPage: number;
	cols: number;
	cfg: Applications2GridInfiniteConfig;
}): number {
	const { visible, total, itemsPerPage, cols, cfg } = args;
	const remaining = total - visible;
	const cap = maxAppendStepCap(cols, cfg);
	return Math.min(remaining, itemsPerPage, cap);
}

export function computeInfiniteScrollOffsetPx(
	viewportHeight: number,
	cols: number,
	cfg: Applications2GridInfiniteConfig
): number {
	const ratioBump = cols > cfg.offsetRatioBumpMinCols ? cfg.offsetRatioBump : 0;
	return Math.max(
		cfg.offsetMinPx,
		Math.round(viewportHeight * (cfg.offsetBaseRatio + ratioBump))
	);
}

export function isVerticallyScrollable(el: HTMLElement): boolean {
	const st = window.getComputedStyle(el);
	const oy = st.overflowY;
	if (oy !== 'auto' && oy !== 'scroll' && oy !== 'overlay') return false;
	return el.scrollHeight > el.clientHeight + 2;
}

export function getScrollViewportHeight(fromEl: HTMLElement | null): number {
	if (typeof window === 'undefined') {
		return 400;
	}
	if (!fromEl) {
		return window.innerHeight;
	}
	let el: HTMLElement | null = fromEl.parentElement;
	while (el && el !== document.documentElement) {
		if (isVerticallyScrollable(el)) {
			return Math.max(200, el.clientHeight);
		}
		el = el.parentElement;
	}
	return window.innerHeight;
}
