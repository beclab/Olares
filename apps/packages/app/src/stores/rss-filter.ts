import { defineStore } from 'pinia';
import {
	addFilter,
	deleteFilter,
	getFilterList,
	updateFilter
} from '../api/wise';
import {
	FilterInfo,
	ORDER_TYPE,
	SORT_TYPE,
	SPLIT_TYPE
} from '../utils/rss-types';
import cloneDeep from 'lodash/cloneDeep';
import { TabType } from '../utils/rss-menu';
import { useConfigStore } from './rss-config';
import { liveQuery, Subscription } from '../pages/Wise/database/sqliteService';
import { FilterFormat } from '../pages/Wise/database/filterFormat';
import {
	addOrUpdateFilters,
	getAllFilters,
	removeFilterById
} from '../pages/Wise/database/tables/view';

export type FilterState = {
	inited: boolean;
	filterList: FilterInfo[];
	unseenMap: Record<string, number>;
	subscriptionMap: Record<string, Subscription>;
	feedMap: Map<string, Set<FilterInfo>>;
	labelMap: Map<string, Set<FilterInfo>>;
};
//sym_r_library_books
//sym_r_video_library
//sym_r_music_cast
//sym_r_picture_as_pdf
//sym_r_book_2
//sym_r_filter

export const useFilterStore = defineStore('filter', {
	state: () => {
		return {
			inited: false,
			filterList: [],
			unseenMap: {},
			subscriptionMap: {},
			feedMap: new Map<string, Set<FilterInfo>>(),
			labelMap: new Map<string, Set<FilterInfo>>()
		} as FilterState;
	},
	getters: {
		customList(state: FilterState) {
			return state.filterList.filter((item) => !item.system);
		}
	},
	actions: {
		async init() {
			console.log('init');
			const list = await getAllFilters();
			for (let i = 0; i < list.length; i++) {
				this.updateMenuBySplit(list[i].id, list[i].splitview);
				this.subscribeFilter(list[i]);
				this.parseFilter(list[i]);
			}
			this.filterList = list;
			this.inited = true;
			this.syncWithRemote();
		},
		async syncWithRemote() {
			try {
				const remoteFilters = await getFilterList();
				const localFilters = this.filterList;

				const localMap = new Map(localFilters.map((f) => [f.id, f]));
				const remoteMap = new Map(remoteFilters.map((f) => [f.id, f]));

				for (const remoteFilter of remoteFilters) {
					const localFilter = localMap.get(remoteFilter.id);

					if (!localFilter) {
						this.updateMenuBySplit(remoteFilter.id, remoteFilter.splitview);
						await addOrUpdateFilters([remoteFilter]);
						this.filterList.push(remoteFilter);
						this.subscribeFilter(remoteFilter);
						this.parseFilter(remoteFilter);
					}
				}

				this.filterList = this.filterList.filter((localFilter) => {
					if (!remoteMap.has(localFilter.id)) {
						this.deleteFilter(localFilter.id);
						return false;
					}
					return true;
				});
			} catch (error) {
				console.error('sync filter:', error);
			}
		},
		async addFilter(name: string, description: string, query: string) {
			const data = await addFilter(
				name ? name : 'view' + (this.customList.length + 1),
				description,
				query,
				SORT_TYPE.PUBLISHED,
				ORDER_TYPE.DESC,
				SPLIT_TYPE.NONE
			);
			if (data) {
				await addOrUpdateFilters([data]);
				this.filterList.push(data);
				this.subscribeFilter(data);
				this.parseFilter(data);
			}
			return data;
		},
		subscribeFilter(filter: FilterInfo) {
			const subscription = this.subscriptionMap[filter.id];
			if (subscription) {
				subscription.unsubscribe();
			}
			if (!filter.showbadge) {
				return;
			}
			this.subscriptionMap[filter.id] = liveQuery(
				filter.id + '_badge',
				FilterFormat.fromFilterInfo(
					{ ...filter, splitview: SPLIT_TYPE.SEEN },
					TabType.UnSeen
				).buildQuery('COUNT(*)')
			).subscribe((data: any) => {
				if (data && data.length > 0) {
					this.unseenMap[filter.id] = data[0]['COUNT(*)'];
				} else {
					this.unseenMap[filter.id] = 0;
				}
			});
		},
		getFilterListById(id: string) {
			return this.filterList.find((item) => item.id === id);
		},
		updateMenuBySplit(
			filterId: string,
			splitView: SPLIT_TYPE | string | undefined,
			changeTab = false
		) {
			const configStore = useConfigStore();
			switch (splitView) {
				case SPLIT_TYPE.LOCATION:
					configStore.setUserTabs(filterId, [TabType.Inbox, TabType.ReadLater]);
					break;
				case SPLIT_TYPE.SEEN:
					configStore.setUserTabs(filterId, [TabType.UnSeen, TabType.Seen]);
					break;
				case SPLIT_TYPE.NONE:
					configStore.setUserTabs(filterId, []);
					break;
			}
			console.log('userTabs ', configStore.userTabs);
			if (changeTab) {
				const tabs = configStore.userTabs.get(filterId);
				if (tabs) {
					configStore.menuChoice.tab = tabs[0];
				} else {
					configStore.menuChoice.tab = TabType.Empty;
				}
			}
		},
		async deleteFilter(id: string) {
			const index = this.filterList.findIndex((item) => item.id === id);
			const item = cloneDeep(this.filterList[index]);
			this.filterList.splice(index, 1);
			await deleteFilter(id)
				.then(() => {})
				.catch(() => {
					this.filterList.splice(index, 0, item);
				});
			removeFilterById(id);
			const subscription = this.subscriptionMap[id];
			if (subscription) {
				subscription.unsubscribe();
			}
			this.cleanupFilter(id);
		},
		async addToQuery(
			filter: FilterInfo,
			type: 'feed_id' | 'tag_id',
			id: string
		) {
			let updatedQuery = `${type}:${id}`;
			if (filter.query) {
				updatedQuery = `(${filter.query}) OR ` + updatedQuery;
			}
			await this.modifyFilter({ ...filter, query: updatedQuery });
		},
		async removeFromQuery(
			filter: FilterInfo,
			type: 'feed_id' | 'tag_id',
			id: string
		) {
			const target = `${type}:${id}`;

			const tokens =
				filter.query.match(/(\(|\)|AND|OR|[^\s()]+:[^\s()]+)/g) || [];

			function parse(tokens: string[]): string[] {
				const stack: string[] = [];

				for (const token of tokens) {
					if (token === ')') {
						const insideBracket: string[] = [];
						while (stack.length && stack[stack.length - 1] !== '(') {
							insideBracket.unshift(stack.pop()!);
						}
						stack.pop();

						const processed = cleanUp(insideBracket);
						if (processed.length === 1) {
							stack.push(processed[0]);
						} else {
							stack.push(`(${processed.join(' ')})`);
						}
					} else {
						stack.push(token);
					}
				}

				return cleanUp(stack);
			}

			function cleanUp(tokens: string[]): string[] {
				const cleaned: string[] = [];

				for (let i = 0; i < tokens.length; i++) {
					if (tokens[i] === target) {
						continue;
					}
					if (
						tokens[i] === 'OR' &&
						(cleaned.length === 0 || cleaned[cleaned.length - 1] === 'OR')
					) {
						continue;
					}
					cleaned.push(tokens[i]);
				}

				if (cleaned[0] === 'OR') cleaned.shift();
				if (cleaned[cleaned.length - 1] === 'OR') cleaned.pop();

				return cleaned;
			}

			const updatedQuery = parse(tokens).join(' ');

			if (updatedQuery !== filter.query) {
				await this.modifyFilter({ ...filter, query: updatedQuery });
			}
		},
		parseFilter(filter: FilterInfo) {
			try {
				const query = filter.query;

				if (!query) {
					console.warn(`Filter ${filter.id} has an empty query.`);
					this.cleanupFilter(filter.id);
					return;
				}

				const feedRegex = /\bfeed_id:([\w-]+)\b/g;
				const labelRegex = /\btag_id:([\w-]+)\b/g;

				let match;

				this.cleanupFilter(filter.id);

				while ((match = feedRegex.exec(query)) !== null) {
					const feedId = match[1];
					if (!this.feedMap.has(feedId)) {
						this.feedMap.set(feedId, new Set<FilterInfo>());
					}
					this.feedMap.get(feedId)!.add(filter);
				}

				while ((match = labelRegex.exec(query)) !== null) {
					const labelId = match[1];
					if (!this.labelMap.has(labelId)) {
						this.labelMap.set(labelId, new Set<FilterInfo>());
					}
					this.labelMap.get(labelId)!.add(filter);
				}
			} catch (error) {
				console.error(`Error parsing filter ${filter.id}:`, error);
			}
		},
		cleanupFilter(filterId: string) {
			for (const [feedId, filters] of this.feedMap) {
				for (const filter of filters) {
					if (filter.id === filterId) {
						filters.delete(filter);
						break;
					}
				}
				if (filters.size === 0) {
					this.feedMap.delete(feedId);
				}
			}

			for (const [labelId, filters2] of this.labelMap) {
				for (const filter2 of filters2) {
					if (filter2.id === filterId) {
						filters2.delete(filter2);
						break;
					}
				}

				if (filters2.size === 0) {
					this.labelMap.delete(labelId);
				}
			}
		},
		async modifyFilter(filter: FilterInfo) {
			const result = await updateFilter({ ...filter });
			if (result) {
				addOrUpdateFilters([filter]);
				const index = this.filterList.findIndex(
					(item) => result.id === item.id
				);
				const queryUpdate =
					this.filterList[index] &&
					this.filterList[index].query !== filter.query;
				this.filterList.splice(index, 1, result);
				this.subscribeFilter(result);
				if (queryUpdate) {
					this.parseFilter(result);
				}
			}
		}
	}
});
