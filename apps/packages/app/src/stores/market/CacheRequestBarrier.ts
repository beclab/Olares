import { CacheRequest } from 'src/stores/market/CacheRequest';

export class CacheRequestBarrier {
	private requests = new Map<string, CacheRequest<any>>();
	private trackedKeys: string[];
	private updateCounts = new Map<string, number>();
	private readonly finalCallback: (
		results: Record<string, any>,
		isFromCache: boolean
	) => void;
	private firstLoadTriggered = false;
	private currentRound = 0;
	private hasColdStart = false;

	constructor(
		trackedKeys: string[],
		finalCallback: (results: Record<string, any>, isFromCache: boolean) => void
	) {
		this.trackedKeys = trackedKeys;
		this.finalCallback = finalCallback;
		trackedKeys.forEach((key) => this.updateCounts.set(key, 0));
	}

	addRequest<T>(key: string, request: CacheRequest<T>) {
		this.requests.set(key, request);

		this.trackFirstCompletion(key, request);

		if (this.trackedKeys.includes(key)) {
			this.trackUpdates(key, request);
		}
	}

	private async trackFirstCompletion<T>(key: string, request: CacheRequest<T>) {
		try {
			const preCache = request.getCacheData();
			if (preCache === null || preCache === undefined) {
				this.hasColdStart = true;
			}

			await request.get();
			if (!this.firstLoadTriggered && this.allFirstCompleted()) {
				this.firstLoadTriggered = true;
				console.log('Barrier trackFirstCompletion');
				const isFromCache = !this.hasColdStart;
				this.triggerFinalCallback(isFromCache);
			}
		} catch (err) {
			console.error(`Barrier request ${key} first load error:`, err);
		}
	}

	private trackUpdates<T>(key: string, request: CacheRequest<T>) {
		request.on('requestFinish', async (newData) => {
			if (!this.firstLoadTriggered) {
				return;
			}

			const currentCount = this.updateCounts.get(key) || 0;
			this.updateCounts.set(key, currentCount + 1);

			console.log(
				`Barrier trackUpdates ${key} ${this.updateCounts.get(key)}`,
				newData
			);

			if (this.allTrackedUpdated()) {
				this.triggerFinalCallback(false);
				this.currentRound++;
				console.log('Barrier allTrackedUpdated', this.currentRound);
			}
		});
	}

	private allFirstCompleted(): boolean {
		return this.trackedKeys.every((key) => {
			const request = this.requests.get(key);
			return request?.getCacheData() !== null;
		});
	}

	private allTrackedUpdated(): boolean {
		const counts = this.trackedKeys.map(
			(key) => this.updateCounts.get(key) || 0
		);

		const targetCount = this.currentRound + 1;
		return counts.every((count) => count === targetCount);
	}

	private triggerFinalCallback(isFromCache: boolean) {
		const results: Record<string, any> = {};
		this.requests.forEach((request, key) => {
			results[key] = request.getCacheData();
		});
		this.finalCallback(results, isFromCache);
	}
}
