let worker: Worker | undefined;
let isLeader = false;
let channel: BroadcastChannel | null = null;
const requestMap = new Map();
let knownTabs: string[] = [];
const subscriptionMaps = new Map<string, Subscription>();
let leaderElectionResolve: (() => void) | null = null;
let leaderElectionPromise = new Promise<void>((resolve) => {
	leaderElectionResolve = resolve;
});

type MessageType = 'init' | 'close' | 'query' | 'execute' | 'transaction';

const tabId = generateId('tab');

export function workerInit() {
	channel = new BroadcastChannel('db_channel');
	channel.postMessage({ type: 'tab_opened', tabId });

	const electionTimeout = setTimeout(() => {
		if (isLeader) {
			return;
		}
		if (!isLeader) {
			console.log('tabId is leader !! ', tabId);
			isLeader = true;
			knownTabs = [];
			channel?.postMessage({ type: 'leader_elected', tabId });
			startWorker(false);
			leaderElectionResolve?.();
		}
	}, 100);

	channel.onmessage = (event) => {
		console.log('tab id', tabId);
		console.log('tab event', event);
		const message = event.data;

		if (!message || typeof message.type !== 'string') {
			console.warn('Invalid message received:', message);
			return;
		}

		switch (message.type) {
			case 'leader_elected':
				clearTimeout(electionTimeout);
				console.log(`new leader: ${message.tabId}`);
				isLeader = message.tabId === tabId;
				if (isLeader) {
					console.log(`I am the new leader: ${tabId}`);
					startWorker(true);
					channel?.postMessage({ type: 'send_tab_id', leaderTabId: tabId });
				}
				markElectionComplete();
				break;

			case 'send_tab_id':
				if (!isLeader) {
					channel?.postMessage({ type: 'receive_tab_id', tabId });
				}
				break;
			case 'receive_tab_id':
				if (isLeader) {
					if (!knownTabs.includes(message.tabId)) {
						knownTabs.push(message.tabId);
					}
				}
				break;
			case 'tab_opened':
				if (isLeader) {
					if (!knownTabs.includes(message.tabId)) {
						knownTabs.push(message.tabId);
					}
					channel?.postMessage({ type: 'leader_elected', tabId });
				}
				break;

			case 'tab_closed':
				if (isLeader) {
					knownTabs = knownTabs.filter((id) => id !== message.tabId);
				}
				break;

			case 'db_request':
				if (isLeader && worker) {
					worker.postMessage(message.data);
				}
				break;

			case 'db_response':
				if (message.tabId === tabId) {
					handleMessage(message.data);
				}
				break;

			case 'restore_subscription':
				subscriptionMaps.forEach((subscription) => {
					if (subscription) {
						handleSubscription(subscription);
					}
				});
				break;

			default:
				console.warn('Unknown message type:', message.type);
		}
	};

	window.addEventListener('beforeunload', () => {
		if (isLeader) {
			electNewLeader();
		} else {
			channel?.postMessage({
				type: 'tab_closed',
				tabId
			});
		}

		if (isLeader && worker) {
			worker.terminate();
		}
	});
}

function markElectionComplete() {
	leaderElectionResolve?.();
	leaderElectionResolve = null;
	leaderElectionPromise = Promise.resolve();
}

function electNewLeader() {
	const sortedTabs = knownTabs.sort((a, b) => a.localeCompare(b));
	if (sortedTabs.length === 0) {
		return;
	}
	const newLeaderTabId = sortedTabs[0];
	channel?.postMessage({ type: 'leader_elected', tabId: newLeaderTabId });
}

function startWorker(restoreSubscription: boolean) {
	worker = new Worker(
		new URL('./worker/messageHandler.worker.ts', import.meta.url)
	);

	worker.onmessage = (event) => {
		if (event.data.tabId === tabId) {
			handleMessage(event.data);
		} else {
			channel?.postMessage({
				tabId: event.data.tabId,
				type: 'db_response',
				data: event.data
			});
		}
	};

	sendMessageToWorker('init');

	if (restoreSubscription) {
		subscriptionMaps.forEach((subscription) => {
			if (subscription) {
				handleSubscription(subscription);
			}
		});
		channel?.postMessage({
			type: 'restore_subscription'
		});
	}
}

async function handleSubscription(subscription: Subscription) {
	const data = {
		type: 'subscribe',
		tabId,
		id: subscription.subscriptionId,
		payload: {
			sql: subscription.sql,
			params: subscription.params,
			tables: subscription.tables
		}
	};
	await leaderElectionPromise;
	if (isLeader) {
		worker?.postMessage(data);
	} else {
		channel?.postMessage({
			tabId,
			type: 'db_request',
			data
		});
	}
}

function handleMessage(data: any) {
	const { id, type, result, error } = data;
	if (subscriptionMaps.has(id)) {
		const subscription = subscriptionMaps.get(id);
		if (subscription) {
			if (type === 'subscribe' && subscription.success) {
				subscription.success(result, id);
			} else if (type === 'error' && subscription.failure) {
				subscription.failure(error);
			} else {
				console.warn(`Unhandled subscription message for id: ${id}`);
			}
		}
	} else if (requestMap.has(id)) {
		const { resolve, reject } = requestMap.get(id);
		requestMap.delete(id);

		if (error) {
			console.error(`message error id: ${id}`, error);
			reject(error);
		} else {
			resolve(result);
		}
	} else {
		console.warn(`Unhandled message with id: ${id} and type: ${type}`);
	}
}

export async function sendMessageToWorker(
	type: MessageType,
	payload = {},
	id = generateId('msg')
) {
	await leaderElectionPromise;
	return new Promise((resolve, reject) => {
		const timeoutId = setTimeout(() => {
			requestMap.delete(id);
			reject(new Error(`Request ${id} timed out after 30s`));
		}, 30000);
		requestMap.set(id, {
			resolve: (result) => {
				clearTimeout(timeoutId);
				resolve(result);
			},
			reject: (error) => {
				clearTimeout(timeoutId);
				reject(error);
			}
		});

		const data = { id, tabId, type, payload };

		if (isLeader) {
			worker?.postMessage(data);
		} else {
			channel?.postMessage({
				tabId,
				type: 'db_request',
				data
			});
		}
	});
}

function generateId(prefix: string): string {
	return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export class Subscription {
	subscriptionId: string;
	private timeoutHandle: any = null;
	sql: string;
	params?: any[] = [];
	tables?: string[] = [];
	success?: (result: any, id?: string) => void | undefined = undefined;
	failure?: (message: string) => void | undefined = undefined;

	constructor(
		subscriptionId: string,
		sql: string,
		params: any[] = [],
		tables: string[] = [],
		private timeout?: number
	) {
		this.sql = sql;
		this.params = params;
		this.tables = tables;
		this.subscriptionId = subscriptionId;

		if (this.timeout) {
			this.startTimeout();
		}
	}

	subscribe(callback: (result: any, id?: string) => void): this {
		this.success = callback;
		subscriptionMaps.set(this.subscriptionId, this);
		return this;
	}

	catch(callback: (message: string) => void): this {
		this.failure = callback;
		subscriptionMaps.set(this.subscriptionId, this);
		return this;
	}

	unsubscribe(): void {
		this.clearTimeout();
		this.success = undefined;
		this.failure = undefined;
		subscriptionMaps.delete(this.subscriptionId);
		worker?.postMessage({ id: this.subscriptionId, type: 'unsubscribe' });
	}

	private startTimeout(): void {
		this.clearTimeout();
		this.timeoutHandle = setTimeout(() => {
			console.warn(`Subscription ${this.subscriptionId} timed out.`);
			this.unsubscribe();
		}, this.timeout);
	}

	private clearTimeout(): void {
		if (this.timeoutHandle) {
			clearTimeout(this.timeoutHandle);
			this.timeoutHandle = null;
		}
	}
}

export function liveQuery(
	subscriptionId: string,
	sql: string,
	params?: any[],
	tables?: string[],
	options?: { timeout?: number }
): Subscription {
	const subscription = new Subscription(
		subscriptionId,
		sql,
		params,
		tables,
		options?.timeout
	);
	handleSubscription(subscription);
	return subscription;
}
