import { CacheRequest } from 'src/stores/market/CacheRequest';

export class CacheRequestBarrier {
	private requests = new Map<string, CacheRequest<any>>();
	private trackedKeys: string[];
	private updateCounts = new Map<string, number>();
	private readonly finalCallback: (
		results: Record<string, any>,
		isFirstLoad: boolean
	) => void;
	private firstLoadTriggered = false;
	private currentRound = 0;

	constructor(
		trackedKeys: string[],
		finalCallback: (results: Record<string, any>, isFirstLoad: boolean) => void
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
			await request.get();
			if (!this.firstLoadTriggered && this.allFirstCompleted()) {
				this.firstLoadTriggered = true;
				console.log('Barrier trackFirstCompletion');
				this.triggerFinalCallback(true);
			}
		} catch (err) {
			console.error(`Barrier request ${key} first load error:`, err);
		}
	}

	private trackUpdates<T>(key: string, request: CacheRequest<T>) {
		request.on('requestFinish', async (newData) => {
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
		return Array.from(this.requests.keys()).every((key) => {
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

	private triggerFinalCallback(isFirstLoad: boolean) {
		const results: Record<string, any> = {};
		this.requests.forEach((request, key) => {
			results[key] = request.getCacheData();
		});
		this.finalCallback(results, isFirstLoad);
	}
}
