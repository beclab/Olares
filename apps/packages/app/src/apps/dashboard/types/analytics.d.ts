export interface PageviewsParams {
	startAt: number;
	endAt: number;
	unit: string;
	timezone?: string;
}

export interface MetricsParams {
	type: string;
	startAt?: number;
	endAt?: number;
	url?: string;
	referrer?: string;
	os?: string;
	browser?: string;
	device?: string;
	country?: string;
	region?: string;
	city?: string;
}

export interface StatsParams {
	startAt: number;
	endAt: number;
}
interface StatsResponseCell {
	value: number;
	change: number;
}
export interface StatsResponse {
	pageviews: StatsResponseCell;
	uniques: StatsResponseCell;
	bounces: StatsResponseCell;
	totaltime: StatsResponseCell;
}

export interface WebsiteItem {
	createdAt: string;
	deletedAt: number | null;
	domain: string;
	id: string;
	name: string;
	resetAt: number | null;
	shareId: number | null;
	updatedAt: string;
	userId: string;
}
