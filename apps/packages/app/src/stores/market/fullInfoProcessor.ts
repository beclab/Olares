import { AppFullInfoLatest, MarketAppRequest } from 'src/constant/constants';
import { getMarketApps } from 'src/api/market/centerApi';

export interface AppQueueProcessorConfig {
	normalBatchSize?: number;
	priorityBatchSize?: number;
}

export interface AppHandler {
	handleAppInfo(key: string, appInfo: any): void;
	handleNotFoundApp?(key: string): void;
}

export class FullInfoProcessor {
	public normalBatchSize = 20;
	public priorityBatchSize = 10;

	private normalQueue: Set<string> = new Set();
	private priorityQueue: Set<string> = new Set();
	private notFoundQueue: Set<string> = new Set();
	private isProcessingQueue = false;

	private readonly handler: AppHandler;

	constructor(handler: AppHandler, config: AppQueueProcessorConfig = {}) {
		this.handler = handler;

		if (config.normalBatchSize !== undefined) {
			this.normalBatchSize = config.normalBatchSize;
		}
		if (config.priorityBatchSize !== undefined) {
			this.priorityBatchSize = config.priorityBatchSize;
		}
	}

	public getQueueStatus() {
		return {
			normalQueueSize: this.normalQueue.size,
			priorityQueueSize: this.priorityQueue.size,
			notFoundQueueSize: this.notFoundQueue.size,
			isProcessing: this.isProcessingQueue,
			normalBatchSize: this.normalBatchSize,
			priorityBatchSize: this.priorityBatchSize
		};
	}

	public updateConfig(config: AppQueueProcessorConfig) {
		if (config.normalBatchSize !== undefined) {
			this.normalBatchSize = config.normalBatchSize;
		}
		if (config.priorityBatchSize !== undefined) {
			this.priorityBatchSize = config.priorityBatchSize;
		}
	}

	public addToNormalQueue(key: string, forceRequest = false) {
		if (!forceRequest && this.notFoundQueue.has(key)) {
			return;
		}
		if (!this.normalQueue.has(key) && !this.priorityQueue.has(key)) {
			this.normalQueue.add(key);
			this.tryProcessQueue();
		}
	}

	public addToPriorityQueue(key: string, forceRequest = false) {
		if (!forceRequest && this.notFoundQueue.has(key)) {
			return;
		}
		this.normalQueue.delete(key);
		if (!this.priorityQueue.has(key)) {
			this.priorityQueue.add(key);
			this.tryProcessQueue();
		}
	}

	public isNotFound(key: string): boolean {
		return this.notFoundQueue.has(key);
	}

	public clearNotFoundQueue() {
		this.notFoundQueue.clear();
	}

	private tryProcessQueue() {
		if (!this.isProcessingQueue) {
			this.processTempAppQueue();
		}
	}

	public async processTempAppQueue() {
		if (this.isProcessingQueue) {
			return;
		}

		try {
			this.isProcessingQueue = true;
			await this.processQueue(this.priorityQueue, this.priorityBatchSize);
			await this.processQueue(this.normalQueue, this.normalBatchSize);
		} finally {
			this.isProcessingQueue = false;
			if (this.priorityQueue.size > 0 || this.normalQueue.size > 0) {
				this.processTempAppQueue();
			}
		}
	}

	private async processQueue(queue: Set<string>, batchSize: number) {
		if (queue.size === 0) {
			return;
		}

		const appIds = Array.from(queue);
		queue.clear();

		const batches = Math.ceil(appIds.length / batchSize);

		for (let i = 0; i < batches; i++) {
			const start = i * batchSize;
			const end = Math.min(start + batchSize, appIds.length);
			const batchIds = appIds.slice(start, end);
			await this.processBatch(batchIds);
		}
	}

	private async processBatch(keys: string[]) {
		const appsBySource: { [sourceId: string]: string[] } = {};

		keys.forEach((key) => {
			const [sourceId] = key.split('_');
			if (!appsBySource[sourceId]) {
				appsBySource[sourceId] = [];
			}
			appsBySource[sourceId].push(key);
		});

		for (const [sourceId, keys] of Object.entries(appsBySource)) {
			try {
				const appIds = keys.map((key) => key.split('_')[1]);
				const apps: MarketAppRequest[] = appIds.map((appId) => ({
					appid: appId,
					sourceDataName: sourceId
				}));

				const data: any = await getMarketApps(apps);

				const marketApps: AppFullInfoLatest[] = data.apps || [];
				const notFoundApps: { sourceDataName: string; appid: string }[] =
					data.not_found || [];

				marketApps.forEach((app, index) => {
					this.handler.handleAppInfo(keys[index], app);
				});

				notFoundApps.forEach((notFound) => {
					const key = `${notFound.sourceDataName}_${notFound.appid}`;
					this.notFoundQueue.add(key);
					if (this.handler.handleNotFoundApp) {
						this.handler.handleNotFoundApp(key);
					}
					this.normalQueue.delete(key);
					this.priorityQueue.delete(key);
				});
			} catch (error) {
				console.error(`deal app failure:`, error);
			}
		}
	}
}
