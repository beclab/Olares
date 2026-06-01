import {
	fetchMarketData,
	getMarketDataHash,
	getMarketState
} from 'src/api/market/centerApi';
import {
	isUpgradable,
	nsfwApp,
	suspendApp,
	uninstalledApp
} from 'src/constant/config';
import { useMenuStore } from 'src/stores/market/menu';
import { usePreCheckStore } from 'src/stores/market/preCheck';
import globalConfig from 'src/api/market/config';
import { Version } from 'src/utils/version';
import { defineStore } from 'pinia';
import {
	AppSimpleInfo,
	AppStatusLatest,
	CONTENT_TYPE,
	getAppCombinedId,
	MARKET_SOURCE_OFFICIAL,
	MarketData,
	Page,
	Recommend,
	Topic,
	TOPIC_TYPE
} from 'src/constant/constants';
import {
	clearAppFullInfoMap,
	updateAppFullInfoMapKey
} from 'src/stores/market/marketDB';
import { useAppStore } from 'src/stores/market/appStore';
import { useSettingStore } from 'src/stores/market/setting';
import { intersection } from 'src/utils/utils';
import { ImageUpdater } from 'src/stores/market/ImageUpdater';
import cloneDeep from 'lodash/cloneDeep';

export type CenterState = {
	dataHash: string;
	marketData: MarketData;
	pagesMap: Map<string, any>;
	pollInterval: NodeJS.Timeout | null;
	pollingInterval: number;
	allSourcesApps: Set<string>;
};

export const useCenterStore = defineStore('marketCenter', {
	state: () => {
		return {
			dataHash: '',
			marketData: {},
			pagesMap: new Map<string, any>(),
			pollInterval: null,
			pollingInterval: 180000,
			allSourcesApps: new Set<string>()
		} as CenterState;
	},
	getters: {
		marketSource(state) {
			const settingStore = useSettingStore();
			return state.marketData?.user_data?.sources[settingStore.marketSourceId];
		},
		marketInstalledApps() {
			const apps: string[] = [];
			const settingStore = useSettingStore();
			const appStore = useAppStore();
			appStore.appStatusMap.forEach((statusLatest, combinedId) => {
				if (!uninstalledApp(statusLatest.status)) {
					const list = combinedId.split('_');
					const sourceId = list[0];
					const appName = list[1];
					if (sourceId === settingStore.marketSourceId) {
						apps.push(appName);
					}
				}
			});
			return apps;
		},
		updateList() {
			const result: string[] = [];
			const settingsStore = useSettingStore();
			const appStore = useAppStore();
			appStore.appStatusMap.forEach((statusLatest, combinedId) => {
				if (isUpgradable(statusLatest.status)) {
					const list = combinedId.split('_');
					const sourceId = list[0];
					const simpleLatest = appStore.getAppSimpleInfo(
						statusLatest.status.rawAppName,
						sourceId,
						statusLatest.status
					);
					if (
						simpleLatest &&
						statusLatest.version &&
						simpleLatest.app_simple_info &&
						simpleLatest.app_simple_info.app_version !== statusLatest.version
					) {
						try {
							const latestVersion = new Version(
								simpleLatest.app_simple_info.app_version
							);
							const myVersion = new Version(statusLatest.version);
							if (latestVersion.isGreater(myVersion)) {
								console.log(
									`${combinedId} app ${statusLatest.status.name} ${myVersion} need update to ${latestVersion}`
								);

								//suspend app return
								if (suspendApp(simpleLatest)) {
									return;
								}
								//open nsfw and nsfw app return
								if (settingsStore.nsfw && nsfwApp(simpleLatest)) {
									return;
								}
								result.push(combinedId);
							}
						} catch (e) {
							console.error(`error ${combinedId}`);
						}
					}
				}
			});
			return result;
		}
	},
	actions: {
		init() {
			const appStore = useAppStore();

			appStore.imageUpdater = new ImageUpdater((imageDetails) => {
				console.log('imageDetails update ==> ', imageDetails);

				appStore.appFullInfoMap.forEach((value, key) => {
					const images = value.app_info.image_analysis?.images;
					if (!images) return;

					Object.keys(images).forEach((name) => {
						const targetImage = imageDetails.get(name);
						if (targetImage) {
							images[name] = targetImage;
						}
					});
					updateAppFullInfoMapKey(key, cloneDeep(value));
				});
			});

			if (this.pollInterval) {
				clearInterval(this.pollInterval);
				this.pollInterval = null;
			}
			this.pollInterval = setInterval(() => {
				console.log('pollInterval refresh');
				this.fetchNewData();
			}, this.pollingInterval);
		},
		async loadFirstData() {
			const hasLocalData = this.loadMarketData() && this.loadMarketHash();
			if (hasLocalData) {
				this.calcCategories();
				this.calcMarketData();
			}
		},
		async fetchNewData(forceRefresh = false) {
			try {
				const { hash } = await getMarketDataHash();
				console.log('new hash', hash);

				if ((hash && hash !== this.dataHash) || forceRefresh) {
					console.log('diff hash refresh data');
					const data = await fetchMarketData();
					if (data) {
						this.saveMarketHash(hash);
						this.saveMarketData(data);
						this.calcCategories();
						this.calcMarketData();
					}
				}
				if (!globalConfig.isOfficial) {
					const appStore = useAppStore();
					const data = await getMarketState();
					if (data) {
						appStore.calcLocationSourceAppStateInfo(this.marketData);
						appStore.calcRemoteSourceAppStateInfo(this.marketData);
					}

					console.log('appStatusMap ===>', appStore.appStatusMap);

					const settingStore = useSettingStore();
					const otherSources = appStore.sources.filter(
						(item) => item.id !== settingStore.marketSourceId
					);
					appStore.processSources(
						otherSources,
						'app_info_latest',
						this.marketData,
						(currentSource, item) => {
							if (item.app_simple_info) {
								appStore.addFullInfoQueue(
									item.app_simple_info.app_name,
									currentSource.id,
									false
								);
							}
						}
					);

					console.log('appSimpleInfoMap ===>', appStore.appSimpleInfoMap);
				}
			} catch (error) {
				console.error('Failed to update market data:', error);
			}
		},
		async clear() {
			localStorage.removeItem('marketHash');
			localStorage.removeItem('marketData');
			await clearAppFullInfoMap();
		},
		saveMarketData(data: MarketData) {
			this.marketData = data;
			localStorage.setItem('marketData', JSON.stringify(data));
		},
		loadMarketData(): boolean {
			const data = localStorage.getItem('marketData');
			if (data) {
				this.marketData = JSON.parse(data);
			}
			return !!data;
		},
		saveMarketHash(data: string) {
			this.dataHash = data;
			localStorage.setItem('marketHash', data);
		},
		loadMarketHash(): boolean {
			const data = localStorage.getItem('marketHash');
			if (data) {
				this.dataHash = data;
			}
			return !!data;
		},
		calcCategories() {
			const data: string[] = [];
			if (this.marketSource?.app_info_latest) {
				this.marketSource?.app_info_latest.forEach((item) => {
					if (
						item.app_simple_info.categories &&
						item.app_simple_info.categories.length > 0
					) {
						item.app_simple_info.categories.forEach((category) => {
							if (!data.includes(category)) {
								data.push(category);
							}
						});
					} else {
						console.log(item.app_simple_info);
						console.error(
							`item ${item.app_simple_info.app_name} categories empty`
						);
					}
				});
			}
			const menuStore = useMenuStore();
			menuStore.appCategories = data;
			console.log('app category ===>', menuStore.appCategories);
		},
		calcMarketData() {
			const settingStore = useSettingStore();
			const appStore = useAppStore();
			this.pagesMap.clear();
			appStore.appSimpleInfoMap.clear();
			console.log('marketData ===>', this.marketData);
			if (appStore.appInfoProcessor) {
				appStore.appInfoProcessor.clearNotFoundQueue();
			}

			appStore.calcRemoteSourceSimpleAppInfo(this.marketData);
			appStore.calcLocationSourceSimpleAppInfo(this.marketData);

			console.log('appSimpleInfoMap ===>', appStore.appSimpleInfoMap);

			const serverAppIds = new Set<string>();
			appStore.sources.forEach((source) => {
				const sourceData = this.marketData?.user_data?.sources[source.id];
				if (sourceData?.app_state_latest) {
					sourceData.app_state_latest.forEach((item: AppStatusLatest) => {
						if (item.status?.name) {
							const combinedId = getAppCombinedId(source.id, item.status.name);
							serverAppIds.add(combinedId);
						}
					});
				}
			});

			const keysToRemove: string[] = [];
			appStore.appStatusMap.forEach((statusLatest, combinedId) => {
				if (!serverAppIds.has(combinedId)) {
					const { sourceId } = appStore.splitCombinedId(combinedId);
					if (
						sourceId === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD &&
						uninstalledApp(statusLatest.status) &&
						appStore.appSimpleInfoMap.has(combinedId)
					) {
						return;
					}
					keysToRemove.push(combinedId);
				}
			});
			keysToRemove.forEach((key) => {
				console.log(`[calcMarketData] Removing stale app status: ${key}`);
				appStore.appStatusMap.delete(key);
			});

			appStore.calcRemoteSourceAppStateInfo(this.marketData);
			appStore.calcLocationSourceAppStateInfo(this.marketData);

			console.log('appStatusMap ===>', appStore.appStatusMap);

			if (this.marketSource?.others) {
				if (this.marketSource?.others.tags) {
					const menuStore = useMenuStore();
					menuStore.menuList = this.marketSource?.others.tags ?? [];
					console.log('menuList ===>', menuStore.menuList);
				}

				if (this.marketSource?.others.pages) {
					console.log('market page ===>', this.marketSource?.others.pages);
					console.log(
						'market topic list ===>',
						this.marketSource?.others.topic_lists
					);
					console.log('market topic ===>', this.marketSource?.others.topics);
					this.marketSource?.others.pages.forEach((page: Page) => {
						if (page.content) {
							const data: any[] = [];
							const contentList = JSON.parse(page.content);
							for (const contentData of contentList) {
								if (contentData.type == 'Recommends') {
									const recommend: Recommend =
										this.marketSource?.others.recommends.find(
											(item) => item.name == contentData.id
										);
									if (recommend) {
										const res: any = {
											type: CONTENT_TYPE.RECOMMENDS,
											title: recommend.data ? recommend.data.title : '',
											name: recommend.name,
											description: recommend.description,
											content: []
										};
										res.content = this._splitAppsString(
											recommend.content,
											settingStore.marketSourceId
										);
										data.push(res);
									}
								} else if (contentData.type == 'Topic') {
									const topic: Topic =
										this.marketSource?.others.topic_lists.find(
											(item) => item.name == contentData.id
										);
									if (
										topic &&
										(topic.type === TOPIC_TYPE.TOPIC ||
											topic.type === TOPIC_TYPE.CATEGORY)
									) {
										const res: any = {
											type:
												topic.type === TOPIC_TYPE.TOPIC
													? CONTENT_TYPE.TOPIC
													: CONTENT_TYPE.APP,
											title: topic.title,
											name: topic.name,
											description: topic.description,
											content: [],
											ids: []
										};
										const topicIds = topic.content
											? topic.content.split(',')
											: [];
										for (const topic_id of topicIds) {
											const topicData: any =
												this.marketSource.others.topics.find(
													(item) => item._id == topic_id
												);
											if (topicData && topicData.data) {
												const topicRes = {};
												let hasValidApps = false;
												let validKey = '';
												for (const key in topicData.data) {
													if (
														Object.prototype.hasOwnProperty.call(
															topicData.data,
															key
														)
													) {
														const value = topicData.data[key];
														if (value && value.apps) {
															const processedApps = this._splitAppsString(
																value.apps,
																settingStore.marketSourceId
															);

															if (processedApps && processedApps.length > 0) {
																topicRes[key] = {
																	...value,
																	topicId: topic_id,
																	apps: processedApps
																};
																hasValidApps = true;
																validKey = key;
															}
														}
													}
												}
												if (hasValidApps) {
													res.content.push(topicRes);
													res.ids.push(topicRes[validKey].topicId);
												}
											}
										}
										data.push(res);
									}
								} else if (
									contentData.type === 'Default Topic' &&
									contentData.id === 'Hottest'
								) {
									const res: any = {
										type: CONTENT_TYPE.TOP,
										name: 'Top',
										description: 'Top apps',
										content: []
									};
									res.content = this._getTopList(page.category);
									data.push(res);
								} else if (
									contentData.type === 'Default Topic' &&
									contentData.id === 'Newest'
								) {
									const res: any = {
										type: CONTENT_TYPE.LATEST,
										name: 'Latest',
										description: 'Latest apps',
										content: []
									};
									res.content = this._getLatestList(page.category);
									data.push(res);
								}
							}

							this.pagesMap.set(page.category, data);
						}
					});
				}
			}

			console.log('all ===>', this.pagesMap);

			this._collectAllSourcesLatestApps();
		},

		_collectAllSourcesLatestApps() {
			this.allSourcesApps.clear();

			if (this.marketData?.user_data?.sources) {
				for (const sourceId of Object.keys(this.marketData.user_data.sources)) {
					const sourceData = this.marketData.user_data.sources[sourceId];

					if (
						sourceData?.others?.latest &&
						Array.isArray(sourceData.others.latest)
					) {
						sourceData.others.latest.forEach((appName: string) => {
							this.allSourcesApps.add(appName);
						});
					}
				}
			}

			console.log('allSourcesLatestApps ===>', this.allSourcesApps);
		},
		_getTopList(category: string): string[] {
			if (!this.marketSource?.others?.tops) {
				return [];
			}
			const settingStore = useSettingStore();
			return this.marketSource?.others?.tops
				.map((item) =>
					this._getAppInfo(item.appid, settingStore.marketSourceId)
				)
				.filter((item) => {
					if (!item) return false;
					if (category === 'All') return true;
					return (
						item.categories?.some(
							(cat) => cat.toLowerCase() === category.toLowerCase()
						) || false
					);
				})
				.map((item) => item.app_name);
		},
		_getLatestList(category: string): string[] {
			if (!this.marketSource?.others?.latest) {
				return [];
			}
			const settingStore = useSettingStore();
			return this.marketSource?.others?.latest
				.map((item) => this._getAppInfo(item, settingStore.marketSourceId))
				.filter((item) => {
					if (!item) return false;
					if (category === 'All') return true;
					return (
						item.categories?.some(
							(cat) => cat.toLowerCase() === category.toLowerCase()
						) || false
					);
				})
				.map((item) => item.app_name);
		},
		_splitAppsString(str: string, sourceId: string): string[] {
			if (!str || !sourceId) {
				return [];
			}
			return str
				.split(',')
				.map((app: string) => this._getAppInfo(app.trim(), sourceId))
				.filter(
					(info: AppSimpleInfo | null): info is AppSimpleInfo => info !== null
				)
				.map((info: AppSimpleInfo) => info.app_name);
		},
		_getAppInfo(
			id: string,
			sourceId: string,
			filterNsfw = true,
			key = 'app_name'
		): AppSimpleInfo | null {
			const app = this.marketData.user_data.sources[
				sourceId
			].app_info_latest.find((item) => item.app_simple_info[key] == id);

			if (!app) {
				return null;
			}

			const settingsStore = useSettingStore();
			if (filterNsfw && settingsStore.nsfw && nsfwApp(app)) {
				return null;
			}

			if (!globalConfig.isOfficial) {
				const preCheckStore = usePreCheckStore();
				if (!preCheckStore.initialized || !preCheckStore.systemResource) {
					return null;
				}

				const intersectedArray = intersection(
					preCheckStore.systemResource.nodes,
					app.app_simple_info.support_arch
						? app.app_simple_info.support_arch
						: []
				);

				if (
					preCheckStore.systemResource.nodes.length <= 0 ||
					intersectedArray.length === 0
				) {
					console.error(
						` ${sourceId} ${id} supportArch ->  ${app.app_simple_info.support_arch} intersectedArray length 0 and device nodes ${preCheckStore.systemResource.nodes}`
					);
					return null;
				}
			}

			if (app) {
				const appStore = useAppStore();
				appStore.addFullInfoQueue(id, sourceId, false);
			}
			return app ? app.app_simple_info : null;
		}
	}
});
