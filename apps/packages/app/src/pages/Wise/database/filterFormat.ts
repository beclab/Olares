import {
	FilterInfo,
	ORDER_TYPE,
	SORT_TYPE,
	SPLIT_TYPE
} from 'src/utils/rss-types';
import { TabType } from 'src/utils/rss-menu';

function getTimeStamp(timeFormat: string): number {
	if (
		timeFormat.includes('day') ||
		timeFormat.includes('week') ||
		timeFormat.includes('month') ||
		timeFormat.includes('year')
	) {
		const currentDate = new Date();
		const daysCheckAgo = new Date(currentDate);
		const datePart: string[] = timeFormat.split(' ');
		if (timeFormat.includes('day')) {
			daysCheckAgo.setDate(currentDate.getDate() - Number(datePart[0]));
		}
		if (timeFormat.includes('week')) {
			daysCheckAgo.setDate(currentDate.getDate() - 7 * Number(datePart[0]));
		}
		if (timeFormat.includes('month')) {
			daysCheckAgo.setDate(currentDate.getDate() - 30 * Number(datePart[0]));
		}
		if (timeFormat.includes('year')) {
			daysCheckAgo.setDate(currentDate.getDate() - 365 * Number(datePart[0]));
		}
		return daysCheckAgo.getTime() / 1000;
	} else {
		const date = new Date(timeFormat);
		return date.getTime() / 1000;
	}
}

function getTimeFormatStr(timeFormat: string): string {
	if (
		timeFormat.includes('day') ||
		timeFormat.includes('week') ||
		timeFormat.includes('month') ||
		timeFormat.includes('year')
	) {
		const currentDate = new Date();
		const daysCheckAgo = new Date(currentDate);
		const datePart: string[] = timeFormat.trim().split(' ');
		if (timeFormat.includes('day')) {
			daysCheckAgo.setDate(currentDate.getDate() - Number(datePart[0]));
		}
		if (timeFormat.includes('week')) {
			daysCheckAgo.setDate(currentDate.getDate() - 7 * Number(datePart[0]));
		}
		if (timeFormat.includes('month')) {
			daysCheckAgo.setDate(currentDate.getDate() - 30 * Number(datePart[0]));
		}
		if (timeFormat.includes('year')) {
			daysCheckAgo.setDate(currentDate.getDate() - 365 * Number(datePart[0]));
		}
		return daysCheckAgo.toISOString();
	} else {
		return timeFormat;
	}
}

export class FilterFormat {
	rawQuery: string;
	rawSortField: SORT_TYPE;
	rawOrderDirection: ORDER_TYPE;
	rawSplitCondition: SPLIT_TYPE;
	rawTabType?: TabType;

	query: string;
	sortField: string;
	orderDirection: string;

	offset?: number = 0;
	limit?: number = 100;
	lastTime?: number = 0;

	constructor(data: any) {
		this.rawQuery = data.query || '';
		this.rawSortField = data.sortField || SORT_TYPE.PUBLISHED;
		this.rawOrderDirection = data.orderDirection || ORDER_TYPE.DESC;
		this.rawSplitCondition = data.splitCondition || SPLIT_TYPE.NONE;
		this.rawTabType = data.tabType;

		this.query = this.cleanConflicts(this.rawQuery);
		this.sortField = this.getSortField(this.rawSortField);
		this.orderDirection = this.getOrderDirection(this.rawOrderDirection);

		this.offset = data.offset || 0;
		this.limit = data.limit;
		this.lastTime = data.lastTime ? data.lastTime / 1000 : 0;
	}

	private cleanConflicts(rawQuery: string): string {
		let userQuery = rawQuery.trim();

		const sortRegex = /sort:\w+/i;
		const orderRegex = /order:(asc|desc)/i;
		const splitRegex = /split:\w+/i;

		userQuery = userQuery
			.replace(sortRegex, '')
			.replace(orderRegex, '')
			.replace(splitRegex, '')
			.trim();

		return userQuery;
	}

	private getSortField(rawSortField: SORT_TYPE): string {
		switch (rawSortField) {
			case SORT_TYPE.CREATED:
				return 'entries.createdAt';
			case SORT_TYPE.PUBLISHED:
				return 'entries.published_at';
			default:
				return 'entries.updatedAt';
		}
	}

	private getOrderDirection(rawOrderDirection: ORDER_TYPE): string {
		return rawOrderDirection === ORDER_TYPE.ASC ? 'ASC' : 'DESC';
	}

	static fromFilterInfo(
		filterInfo: FilterInfo,
		tabType?: TabType | string
	): FilterFormat {
		return new FilterFormat({
			query: filterInfo.query,
			sortField: filterInfo.sortby,
			orderDirection: filterInfo.orderby,
			splitCondition: filterInfo.splitview,
			tabType: tabType
		});
	}

	private parseConditions(
		query: string,
		lastTime: number
	): {
		tableMap: Map<string, string>;
		logicExpression: string;
	} {
		const tableMap = new Map<string, string>();

		tableMap.set(
			'entries',
			`entries.updatedAt >= DATETIME(${lastTime}, 'unixepoch')`
		);

		const conditionHandlers: Record<
			string,
			(key: string, value: string) => string
		> = {
			has: (_, value) => {
				//TODO enclosures table
				// if (value.includes('attachment')) {
				// 	return 'EXISTS (SELECT 1 FROM enclosures WHERE enclosures.entry_id = entries.id)';
				// }
				if (value.includes('note')) {
					return 'EXISTS (SELECT 1 FROM notes WHERE notes.entry_id = entries.id)';
				}
				if (value.includes('tag')) {
					return 'EXISTS (SELECT 1 FROM labels, json_each(labels.entries) WHERE json_each.value = entries.id)';
				}
				return '';
			},
			islibrary: (_, value) =>
				value === 'true'
					? `json_extract(entries.sources, '$') LIKE '%"library"%'`
					: `NOT json_extract(entries.sources, '$') LIKE '%"library"%'`,
			isfeed: (_, value) =>
				value === 'true'
					? `json_extract(entries.sources, '$') LIKE '%"wise"%'`
					: `NOT json_extract(entries.sources, '$') LIKE '%"wise"%'`,
			seen: (_, value) =>
				value === 'true' ? `entries.unread = false` : `entries.unread = true`,
			location: (_, value) => {
				if (value.includes('all')) {
					return `json_extract(entries.sources, '$') LIKE '%"library"%'`;
				}
				if (value.includes('readlater')) {
					return `json_extract(entries.sources, '$') LIKE '%"library"%' AND entries.readlater = true`;
				}
				if (value.includes('inbox')) {
					return `json_extract(entries.sources, '$') LIKE '%"library"%' AND entries.readlater = false`;
				}
				return '';
			},
			tag: (_, value) =>
				`EXISTS (SELECT 1 FROM (SELECT id, entries FROM labels WHERE labels.name = '${value}') AS filtered_labels CROSS JOIN json_each(filtered_labels.entries) WHERE json_each.value = entries.id)`,
			tag_id: (_, value) =>
				// `EXISTS (SELECT 1 FROM labels CROSS JOIN json_each(labels.entries) WHERE json_each.value = entries.id AND labels.id = '${value}')`
				`EXISTS (SELECT 1 FROM (SELECT id, entries FROM labels WHERE labels.id = '${value}') AS filtered_labels CROSS JOIN json_each(filtered_labels.entries) WHERE json_each.value = entries.id)`,
			generic: (key, value) => {
				const keysToCheck = ['feed_id', 'file_type', 'author'];
				if (keysToCheck.includes(key) && !value) {
					throw new Error(
						`Invalid value for key: ${key}. Value cannot be empty.`
					);
				}
				return `${key} = '${value}'`;
			}
		};

		conditionHandlers.feed_id = conditionHandlers.generic;
		conditionHandlers.file_type = conditionHandlers.generic;
		conditionHandlers.author = conditionHandlers.generic;

		const timeConditions = (
			key: string,
			operator: string,
			value: string
		): string => {
			if (key.startsWith('published_at')) {
				return `entries.published_at ${convertOperator(
					operator
				)} ${getTimeStamp(value)}`;
			}
			if (key.startsWith('created_at')) {
				return `entries.createdAt ${convertOperator(
					operator
				)} '${getTimeFormatStr(value)}'`;
			}
			if (key.startsWith('updated_at')) {
				return `entries.updatedAt ${convertOperator(
					operator
				)} '${getTimeFormatStr(value)}'`;
			}
			if (key.startsWith('last_opened')) {
				return `entries.last_opened ${convertOperator(operator)} ${getTimeStamp(
					value
				)}`;
			}
			return '';
		};

		const convertOperator = (op: string): string => {
			const operatorMap: { [key: string]: string } = {
				gt: '>',
				gte: '>=',
				lt: '<',
				lte: '<='
			};

			return operatorMap[op] || op;
		};

		const stack: string[] = [];
		const tokens =
			query.match(/(\(|\)|AND|OR|[^\s()]+:"[^"]+"|[^\s()]+:[^\s()]+)/g) || [];

		for (const token of tokens) {
			if (token === '(' || token === ')') {
				stack.push(token);
			} else if (token === 'AND' || token === 'OR') {
				stack.push(token);
			} else {
				const [key, value] = token.split(':').map((x) => x.trim());

				// Handle time-based conditions
				if (key.match(/(published_at|created_at|last_opened|updated_at)__/)) {
					const [timeKey, operator] = key.split('__');
					stack.push(timeConditions(timeKey, operator, value));
				}
				// Handle quoted values as a single unit
				else if (value.startsWith('"') && value.endsWith('"')) {
					const unquotedValue = value.slice(1, -1);
					stack.push(
						conditionHandlers[key]?.(key, unquotedValue) ||
							`${key} = '${unquotedValue}'`
					);
				} else if (conditionHandlers[key]) {
					stack.push(conditionHandlers[key]?.(key, value));
				}
			}
		}

		console.log(stack);

		const logicExpression = stack.length > 0 ? `(${stack.join(' ')})` : '';

		return { tableMap, logicExpression };
	}

	private handleUserSetting(parseInfos: {
		tableMap: Map<string, string>;
		logicExpression: string;
	}): string {
		const conditions: string[] = [];

		conditions.push(...Array.from(parseInfos.tableMap.values()));

		if (parseInfos.logicExpression) {
			conditions.push(parseInfos.logicExpression);
		}

		if (this.rawTabType) {
			if (this.rawSplitCondition === SPLIT_TYPE.LOCATION) {
				conditions.push(
					`json_extract(entries.sources, '$') LIKE '%"library"%'`
				);
				if (this.rawTabType === TabType.Inbox) {
					conditions.push(`entries.readlater = false`);
				} else if (this.rawTabType === TabType.ReadLater) {
					conditions.push(`entries.readlater = true`);
				}
			} else if (this.rawSplitCondition === SPLIT_TYPE.SEEN) {
				if (this.rawTabType === TabType.Seen) {
					conditions.push(`entries.unread = false`);
				} else if (this.rawTabType === TabType.UnSeen) {
					conditions.push(`entries.unread = true`);
				}
			}
		}

		const whereString = conditions.filter(Boolean).join(' AND ');

		const tableString = Array.from(parseInfos.tableMap.keys()).join(', ');

		return `FROM ${tableString} WHERE ${whereString}`;
	}

	public buildQuery(columnSpecifier = 'entries.*'): string {
		try {
			this.logDetails();
			const parseInfos = this.parseConditions(this.query, this.lastTime || 0);
			const whereClause = this.handleUserSetting(parseInfos);
			let sql = `SELECT ${columnSpecifier} ${whereClause} ORDER BY ${this.sortField} ${this.orderDirection}`;
			console.log('===> sql ', sql);
			if (this.limit && this.limit > 0) {
				sql += ` LIMIT ${this.limit} OFFSET ${this.offset}`;
			}
			return sql;
		} catch (e) {
			console.error(e);
			return '';
		}
	}

	public logDetails(): void {
		console.log('=======>FilterFormat Details:');
		console.log(`Raw Query: ${this.rawQuery}`);
		console.log(`Processed Query: ${this.query}`);
		console.log(`Sort Field: ${this.sortField}`);
		console.log(`Order Direction: ${this.orderDirection}`);
		console.log(`rawSplitCondition: ${this.rawSplitCondition}`);
		console.log(`Tab Type: ${this.rawTabType}`);
		console.log(`Offset: ${this.offset}`);
		console.log(`Limit: ${this.limit}`);
		console.log(`<=======Last Time: ${this.lastTime}`);
	}
}
