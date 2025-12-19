export class SubscriptionManager {
	private subscriptions = new Map<
		string,
		{ sql: string; params: any[]; tables: string[] }
	>();

	addSubscription(
		id: string,
		sql: string,
		params: any[],
		tables: string[]
	): void {
		this.subscriptions.set(id, { sql, params, tables });
	}

	removeSubscription(id: string): void {
		this.subscriptions.delete(id);
	}

	notifySubscribers(
		affectedTables: string[],
		callback: (id: string, sql: string, params: any[]) => void
	): void {
		for (const [id, { sql, params, tables }] of this.subscriptions.entries()) {
			if (tables.some((table) => affectedTables.includes(table))) {
				console.log(`Notifying subscriber ${id}`);
				callback(id, sql, params);
			}
		}
	}

	getAffectedTables(sql: string): string[] {
		const tables: string[] = [];

		const insertMatch = sql.match(/INSERT\s+INTO\s+(\w+)/i);
		const updateMatch = sql.match(/UPDATE\s+(\w+)/i);
		const deleteMatch = sql.match(/DELETE\s+FROM\s+(\w+)/i);

		if (insertMatch) tables.push(insertMatch[1]);
		if (updateMatch) tables.push(updateMatch[1]);
		if (deleteMatch) tables.push(deleteMatch[1]);

		console.log('=====> affected tables ', tables);
		return tables;
	}
}
