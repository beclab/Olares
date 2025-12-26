import {
	Logger,
	LogEvent,
	LoggerListOptions
} from '@didvault/sdk/src/core/logging';
import { Context } from '@didvault/sdk/src/core/server';
import { PostgresStorage } from '../storage/postgres';

export class PostgresLogger implements Logger {
	constructor(private _storage: PostgresStorage, public context?: Context) {}

	withContext(context: Context) {
		return new PostgresLogger(this._storage, context);
	}

	log(type: string, data?: any) {
		const event = new LogEvent(type, data, this.context);
		event.id = `${event.time.toISOString()}_${Math.floor(
			Math.random() * 1e6
		)}`;
		(async () => {
			try {
				this._storage.save(event);
			} catch (e) {
				//
			}
		})();
		return event;
	}

	list({ limit, offset, query: where, before, after }: LoggerListOptions) {
		where = {
			op: 'and',
			queries: where ? [where] : []
		};

		if (before) {
			where.queries.push({
				path: 'time',
				op: 'lt',
				value: before.toISOString()
			});
		}

		if (after) {
			where.queries.push({
				path: 'time',
				op: 'lt',
				value: after.toISOString()
			});
		}

		return this._storage.list(LogEvent, {
			limit,
			offset,
			query: where,
			orderBy: 'time',
			orderByDirection: 'desc'
		});
	}
}
