import { DBManager } from './dbManager';
import { SubscriptionManager } from './subscriptionManager';
import { MessageQueue } from './messageQueue';

const dbManager = new DBManager();
const subscriptionManager = new SubscriptionManager();
const messageQueue = new MessageQueue();

async function handleMessage(message: {
	id: string;
	type: string;
	tabId: string;
	payload: any;
}): Promise<void> {
	const { id, tabId, type, payload } = message;

	try {
		switch (type) {
			case 'init': {
				const result = await dbManager.initDB();
				self.postMessage({ id, tabId, type: 'init', result });
				break;
			}

			case 'close': {
				await dbManager.closeDB();
				self.postMessage({
					id,
					tabId,
					type: 'close',
					result: 'Database closed successfully.'
				});
				break;
			}

			case 'execute': {
				dbManager.execute(payload.sql, payload.params);
				self.postMessage({ id, tabId, type: 'execute', result: 'Success' });
				const affectedTables = subscriptionManager.getAffectedTables(
					payload.sql
				);
				subscriptionManager.notifySubscribers(
					affectedTables,
					(id, sql, params) => {
						const results = dbManager.query(sql, params);
						self.postMessage({ id, tabId, type: 'subscribe', result: results });
					}
				);
				break;
			}

			case 'transaction': {
				dbManager.executeTransaction(payload.sql, payload.params);
				self.postMessage({ id, tabId, type: 'transaction', result: 'Success' });
				const affectedTables = subscriptionManager.getAffectedTables(
					payload.sql
				);
				subscriptionManager.notifySubscribers(
					affectedTables,
					(id, sql, params) => {
						const results = dbManager.query(sql, params);
						self.postMessage({ id, tabId, type: 'subscribe', result: results });
					}
				);
				break;
			}

			case 'query': {
				const results = dbManager.query(payload.sql, payload.params);
				self.postMessage({ id, tabId, type: 'query', result: results });
				break;
			}

			case 'subscribe': {
				let tables = payload.tables;
				if (tables.length === 0) {
					tables = dbManager.getTablesFromQuery(payload.sql);
				}
				subscriptionManager.addSubscription(
					id,
					payload.sql,
					payload.params,
					tables
				);
				const results = dbManager.query(payload.sql, payload.params);
				self.postMessage({
					id,
					tabId,
					type: 'subscribe',
					result: results
				});
				break;
			}

			case 'unsubscribe': {
				subscriptionManager.removeSubscription(id);
				self.postMessage({
					id,
					tabId,
					type: 'unsubscribe',
					result: 'Unsubscribed successfully.'
				});
				break;
			}

			default: {
				self.postMessage({
					id,
					tabId,
					type: 'error',
					error: `Unknown message type: ${type}`
				});
			}
		}
	} catch (error) {
		console.error(`Error processing message ${id}:`, error);
		self.postMessage({
			id,
			tabId,
			type: 'error',
			error: error.message || 'Unexpected error occurred.'
		});
	}
}

self.onmessage = (event: any) => {
	messageQueue.addMessage(event.data);

	if (event.data.type === 'init') {
		handleMessage(event.data);
		return;
	}

	if (!dbManager.inited) {
		console.log('Database is not ready, message added to queue.');
		return;
	}

	messageQueue.processQueue(handleMessage);
};

dbManager.on('init', () => {
	console.log('Database initialized, processing queued messages.');
	messageQueue.processQueue(handleMessage);
});
