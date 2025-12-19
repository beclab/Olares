import {
	fetchMarketData,
	getMarketDataHash,
	getMarketState
} from 'src/api/market/centerApi';
import {
	getErrorTextByState,
	isCloneApp,
	isFailed,
	isUpgradable,
	uninstalledApp
} from 'src/constant/config';
import { useMenuStore } from 'src/stores/market/menu';
import { useUserStore } from 'src/stores/market/user';
import globalConfig from 'src/api/market/config';
import { Version } from 'src/utils/version';
import { defineStore } from 'pinia';
import {
	APP_STATUS,
	AppFullInfoLatest,
	AppInfoAggregation,
	AppLocalInfo,
	AppSimpleInfo,
	AppSimpleInfoLatest,
	AppStatusChange,
	AppStatusLatest,
	CONTENT_TYPE,
	getAppCombinedId,
	ImageInfoUpdate,
	MARKET_SOURCE_OFFICIAL,
	MARKET_SOURCE_TYPE,
	MarketData,
	MarketSource,
	MarketSystemChange,
	Page,
	PAYMENT_STATUS,
	Recommend,
	STATUS_OPERATE_TYPE,
	Topic,
	TOPIC_TYPE
} from 'src/constant/constants';
import {
	clearAppFullInfoMap,
	loadAppFullInfoMap,
	updateAppFullInfoMapKey
} from 'src/stores/market/marketDB';
import {
	AppHandler,
	FullInfoProcessor
} from 'src/stores/market/fullInfoProcessor';
import { bus, BUS_EVENT, busEmit } from 'src/utils/bus';
import { i18n } from 'src/boot/i18n';
import { getMarketSource } from 'src/api/market/private/source';
import { useSettingStore } from 'src/stores/market/setting';
import { intersection } from 'src/utils/utils';
import { ImageUpdater } from 'src/stores/market/ImageUpdater';
import cloneDeep from 'lodash/cloneDeep';
import { CacheRequest } from 'src/stores/market/CacheRequest';

export type CenterState = {
	appUrl: string;
	dataHash: string;
	sources: MarketSource[];
	marketData: MarketData;
	pagesMap: Map<string, any>;
	localStatusMap: Map<string, AppLocalInfo>;
	appSimpleInfoMap: Map<string, AppSimpleInfoLatest>;
	appStatusMap: Map<string, AppStatusLatest>;
	appFullInfoMap: Map<string, AppFullInfoLatest>;
	appInfoProcessor: FullInfoProcessor | null;
	imageUpdater: ImageUpdater | null;
	pollInterval: NodeJS.Timeout | null;
	pollingInterval: number;
};

export const useCenterStore = defineStore('marketCenter', {
	state: () => {
		return {
			appUrl: `${globalConfig.url}/app-store/api/v2`,
			dataHash: '',
			sources: [] as MarketSource[],
			marketData: {},
			pagesMap: new Map<string, any>(),
			localStatusMap: new Map<string, AppLocalInfo>(),
			appSimpleInfoMap: new Map<string, AppSimpleInfoLatest>(),
			appStatusMap: new Map<string, AppStatusLatest>(),
			appFullInfoMap: new Map<string, AppFullInfoLatest>(),
			appInfoProcessor: null,
			imageUpdater: null,
			pollInterval: null,
			pollingInterval: 180000
		} as CenterState;
	},
	getters: {
		remoteSource() {
			const centerStore = useCenterStore();
			return centerStore.sources.filter(
				(item) => item.type === MARKET_SOURCE_TYPE.REMOTE
			);
		},
		localSource() {
			const centerStore = useCenterStore();
			return centerStore.sources.filter(
				(item) => item.type === MARKET_SOURCE_TYPE.LOCAL
			);
		},
		marketSource(state) {
			const settingStore = useSettingStore();
			return state.marketData?.user_data?.sources[settingStore.marketSourceId];
		},
		marketInstalledApps(state) {
			const apps: string[] = [];
			const settingStore = useSettingStore();
			state.appStatusMap.forEach((statusLatest, combinedId) => {
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
		installStatusList(state) {
			return Array.from(state.appStatusMap.values()).filter(
				(item) => !uninstalledApp(item.status)
			);
		},
		upgradingList(state) {
			const result: string[] = [];
			state.appStatusMap.forEach((statusLatest, combinedId) => {
				if (
					(statusLatest.status.state === APP_STATUS.UPGRADE.DEFAULT ||
						statusLatest.status.state === APP_STATUS.INITIALIZE.DEFAULT) &&
					statusLatest.status.opType === STATUS_OPERATE_TYPE.UPGRADE
				) {
					result.push(combinedId);
				}
			});
			return result;
		},
		updateList(state) {
			const result: string[] = [];
			state.appStatusMap.forEach((statusLatest, combinedId) => {
				if (isUpgradable(statusLatest.status)) {
					const simpleLatest = state.appSimpleInfoMap.get(combinedId);
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
		splitCombinedId(combinedId: string): {
			sourceId: string;
			appName: string;
		} {
			const list = combinedId.split('_');
			return { sourceId: list[0], appName: list[1] };
		},
		getAppAggregationInfo(
			appId: string,
			sourceId: string
		): AppInfoAggregation | null {
			if (!appId || !sourceId) {
				console.error(
					`appId:${appId} and sourceId:${sourceId} must be a string in getAppAggregationInfo`
				);
				return null;
			}

			const combinedId = getAppCombinedId(sourceId, appId);
			let simpleLatest = this.appSimpleInfoMap.get(combinedId);
			const statusLatest = this.appStatusMap.get(combinedId);
			let fullLatest = this.appFullInfoMap.get(combinedId);

			if (simpleLatest) {
				this.addFullInfoQueue(appId, sourceId, true);
				return {
					app_simple_latest: simpleLatest,
					app_status_latest: statusLatest
						? statusLatest
						: this._getDefaultStatus(simpleLatest.app_simple_info.app_name),
					app_full_info: fullLatest
				} as AppInfoAggregation;
			} else if (statusLatest && isCloneApp(statusLatest.status)) {
				//clone app
				this.addFullInfoQueue(appId, sourceId, true);
				const combinedId2 = getAppCombinedId(
					sourceId,
					statusLatest?.status?.rawAppName
				);
				simpleLatest = this.appSimpleInfoMap.get(combinedId2);
				fullLatest = this.appFullInfoMap.get(combinedId2);
				return {
					app_simple_latest: simpleLatest,
					app_status_latest: statusLatest
						? statusLatest
						: simpleLatest
						? this._getDefaultStatus(simpleLatest.app_simple_info.app_name)
						: this._getDefaultStatus(appId),
					app_full_info: fullLatest
				} as AppInfoAggregation;
			}
			return null;
		},
		updateLocalStatus(
			appId: string,
			sourceId: string,
			statusInfo: AppLocalInfo
		) {
			if (!statusInfo || !statusInfo.status) {
				return;
			}
			if (statusInfo.status === PAYMENT_STATUS.SYNCING) {
				return;
			}
			const combinedId = getAppCombinedId(sourceId, appId);
			this.localStatusMap.set(combinedId, statusInfo);
			busEmit('local_state_update', {
				appId,
				sourceId,
				status: statusInfo.status
			});
		},
		getLocalStatus(appId: string, sourceId: string) {
			const combinedId = getAppCombinedId(sourceId, appId);
			return this.localStatusMap.get(combinedId);
		},
		getAppStatus(appId: string, sourceId: string): AppStatusLatest {
			const combinedId = getAppCombinedId(sourceId, appId);
			if (this.appStatusMap.has(combinedId)) {
				return this.appStatusMap.get(combinedId)!;
			}
			return this._getDefaultStatus(appId);
		},
		getAppSimpleInfo(appId: string, sourceId: string) {
			const combinedId = getAppCombinedId(sourceId, appId);
			return this.appSimpleInfoMap.get(combinedId);
		},
		getAppFullInfo(appId: string, sourceId: string) {
			const combinedId = getAppCombinedId(sourceId, appId);
			return this.appFullInfoMap.get(combinedId);
		},
		getCloneApp(appId: string, sourceId: string) {
			const apps: string[] = [];
			this.appStatusMap.forEach((statusLatest, combinedId) => {
				if (!uninstalledApp(statusLatest.status)) {
					const list = combinedId.split('_');
					const appSourceId = list[0];
					const appName = list[1];
					if (
						sourceId === appSourceId &&
						appId === statusLatest.status.rawAppName &&
						appName !== statusLatest.status.rawAppName
					) {
						apps.push(appName);
					}
				}
			});
			return apps;
		},
		getSourceInstalledApp(sourceId: string) {
			const apps: string[] = [];
			if (sourceId === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD) {
				this.appStatusMap.forEach((statusLatest, combinedId) => {
					if (statusLatest) {
						const list = combinedId.split('_');
						const appSourceId = list[0];
						const appName = list[1];
						if (
							isCloneApp(statusLatest.status) &&
							uninstalledApp(statusLatest.status)
						) {
							return;
						}
						if (sourceId === appSourceId) {
							apps.push(appName);
						}
					}
				});
			} else {
				this.appStatusMap.forEach((statusLatest, combinedId) => {
					if (!uninstalledApp(statusLatest.status)) {
						const list = combinedId.split('_');
						const appSourceId = list[0];
						const appName = list[1];
						if (sourceId === appSourceId) {
							apps.push(appName);
						}
					}
				});
			}
			return apps;
		},
		removeSourceId(sourceId: string): void {
			const filterMapBySourceId = (map: Map<string, any>) => {
				const keysToRemove: string[] = [];
				map.forEach((_, key) => {
					if (key.startsWith(`${sourceId}_`)) {
						keysToRemove.push(key);
					}
				});
				keysToRemove.forEach((key) => map.delete(key));
			};

			filterMapBySourceId(this.appSimpleInfoMap);
			filterMapBySourceId(this.appStatusMap);
			filterMapBySourceId(this.appFullInfoMap);
		},
		addFullInfoQueue(
			appId: string,
			sourceId: string,
			priority: boolean,
			forceRefresh = false,
			forceRequest = false
		) {
			if (!appId || !sourceId) {
				console.error(
					`appId:${appId} and sourceId:${sourceId} must be a string in addFullInfoQueue`
				);
				return;
			}

			if (!this.sources.find((item) => item.id === sourceId)) {
				console.error(`sourceId ${sourceId} must be in`, this.sources);
				return;
			}

			const key = getAppCombinedId(sourceId, appId);
			if (!forceRefresh) {
				const simpleLatest = this.appSimpleInfoMap.get(key);
				if (!simpleLatest) {
					console.error(`app ${key} simple info error`);
					return;
				}
				const fullLatest = this.appFullInfoMap.get(key);
				if (fullLatest) {
					try {
						//version update
						const fullVersion = new Version(
							fullLatest?.app_info?.app_entry?.version
						);
						const simpleVersion = new Version(
							simpleLatest?.app_simple_info?.app_version
						);
						if (
							simpleVersion.isEqual(fullVersion) ||
							simpleVersion.isLess(fullVersion)
						) {
							return;
						} else {
							console.log(
								`app ${simpleLatest?.app_simple_info?.app_name} the new version ${simpleLatest?.app_simple_info?.app_version}!`
							);
						}

						//image size update
						if (simpleLatest?.timestamp === fullLatest.timestamp) {
							return;
						} else {
							console.log(
								`app ${simpleLatest?.app_simple_info?.app_name} the new timestamp ${simpleLatest?.timestamp}!`
							);
						}
					} catch (error) {
						console.error(simpleLatest);
						console.error(fullLatest);
						console.error(error);
						return;
					}
				}
			}

			if (!this.appInfoProcessor) {
				console.error(`appInfoProcessor must be exist`);
				return;
			}

			if (priority) {
				this.appInfoProcessor?.addToPriorityQueue(key, forceRequest);
			} else {
				this.appInfoProcessor?.addToNormalQueue(key, forceRequest);
			}
		},
		updateAppStatus(appId: string, sourceId: string, newState: string) {
			const statusLatest = this.getAppStatus(appId, sourceId);
			if (statusLatest) {
				statusLatest.status.state = newState;
				const combinedId = getAppCombinedId(sourceId, appId);
				this.appStatusMap.set(combinedId, statusLatest);
			}
		},
		updateAppStatusBySocket(data: AppStatusChange) {
			// console.log('market status update => ', data);
			if (!data) {
				console.error('update status null');
				return;
			}
			if (!data.app_name || !data.source) {
				console.error('update status failure: app_name or source empty');
				return;
			}

			const combinedId = getAppCombinedId(data.source, data.app_name);
			if (data.app_state_latest && data.app_state_latest) {
				const oldState = this.appStatusMap.get(combinedId);
				if (
					isFailed(data.app_state_latest.status) &&
					data.app_state_latest.status?.state !== oldState?.status?.state
				) {
					bus.emit(
						BUS_EVENT.APP_BACKEND_ERROR,
						data.app_name +
							' ' +
							i18n.global.t(
								getErrorTextByState(data.app_state_latest.status.state)
							)
					);
				}
				this.appStatusMap.set(combinedId, data.app_state_latest);
			}

			if (
				data.app_state_latest.status.state === APP_STATUS.UNINSTALL.COMPLETED ||
				data.app_state_latest.status.state === APP_STATUS.RUNNING
			) {
				const settingStore = useUserStore();
				settingStore.init().forceRefresh();
			}
		},
		updateMarketSystemBySocket(data: MarketSystemChange) {
			console.log(data);
			if (data.point === 'new_app_ready' && data.extensions) {
				if (data.extensions.source && data.extensions.app_name) {
					this.addFullInfoQueue(
						data.extensions.app_name,
						data.extensions.source,
						true,
						true,
						true
					);
				}
			} else if (data.point === 'local_app_delete' && data.extensions) {
				if (data.extensions.source && data.extensions.app_name) {
					const combinedId = getAppCombinedId(
						data.extensions.source,
						data.extensions.app_name
					);
					this.appSimpleInfoMap.delete(combinedId);
					this.appFullInfoMap.delete(combinedId);
					this.appStatusMap.delete(combinedId);
				}
			}
		},
		updateDownloadedImageSizeBySocket(data: ImageInfoUpdate) {
			this.imageUpdater?.handleSocketUpdate(data);
		},
		async init() {
			const marketSourceRequest = new CacheRequest(
				'cache_market_source',
				getMarketSource,
				{
					onData: (data) => {
						this.sources = data;
						console.log('sources ===>', this.sources);
					}
				}
			);
			const map = await loadAppFullInfoMap();
			if (map) {
				this.appFullInfoMap = map;
			}
			const store = this;
			this.appInfoProcessor = new FullInfoProcessor({
				handleAppInfo(combinedId: string, appInfo: any) {
					store.appFullInfoMap.set(combinedId, appInfo);
					store.appSimpleInfoMap.set(combinedId, {
						app_simple_info: appInfo.app_simple_info,
						type: appInfo.type,
						timestamp: appInfo.timestamp,
						version: appInfo.version
					} as AppSimpleInfoLatest);
					updateAppFullInfoMapKey(combinedId, appInfo);

					//local insert
					const { appName, sourceId } = store.splitCombinedId(combinedId);
					if (sourceId === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD) {
						if (!store.appStatusMap.get(combinedId)) {
							store.appStatusMap.set(
								combinedId,
								store._getDefaultStatus(appName)
							);
						}
					}
				}
			} as AppHandler);
			this.imageUpdater = new ImageUpdater((imageDetails) => {
				console.log('imageDetails update ==> ', imageDetails);

				this.appFullInfoMap.forEach((value, key) => {
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
			this.appInfoProcessor?.processTempAppQueue();
			if (this.pollInterval) {
				clearInterval(this.pollInterval);
				this.pollInterval = null;
			}
			this.pollInterval = setInterval(() => {
				console.log('pollInterval refresh');
				this.fetchNewData();
			}, this.pollingInterval);
			return marketSourceRequest;
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
					const data = await getMarketState();
					if (data) {
						this.calcLocationSourceAppStateInfo();
						this.calcRemoteSourceAppStateInfo();
					}

					console.log('appStatusMap ===>', this.appStatusMap);

					const settingStore = useSettingStore();
					const otherSources = this.sources.filter(
						(item) => item.id !== settingStore.marketSourceId
					);
					this.processSources(
						otherSources,
						'app_info_latest',
						(currentSource, item) => {
							if (item.app_simple_info) {
								this.addFullInfoQueue(
									item.app_simple_info.app_name,
									currentSource.id,
									false
								);
							}
						}
					);

					console.log('appSimpleInfoMap ===>', this.appSimpleInfoMap);
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
			this.pagesMap.clear();
			this.appSimpleInfoMap.clear();
			this.appStatusMap.clear();
			console.log('marketData ===>', this.marketData);
			if (this.appInfoProcessor) {
				this.appInfoProcessor.clearNotFoundQueue();
			}

			this.calcRemoteSourceSimpleAppInfo();

			this.calcLocationSourceSimpleAppInfo();

			console.log('appSimpleInfoMap ===>', this.appSimpleInfoMap);

			this.calcRemoteSourceAppStateInfo();

			this.calcLocationSourceAppStateInfo();

			console.log('appStatusMap ===>', this.appStatusMap);

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
										// console.log(topic);
										const res: any = {
											type:
												topic.type === TOPIC_TYPE.TOPIC
													? CONTENT_TYPE.TOPIC
													: CONTENT_TYPE.APP,
											title: topic.title,
											name: topic.name,
											description: topic.description,
											content: [],
											ids: topic.content ? topic.content.split(',') : []
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
															}
														}
													}
												}
												if (hasValidApps) {
													res.content.push(topicRes);
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
		},
		processSources(sources, latestKey, processor) {
			for (let i = 0; i < sources.length; i++) {
				const currentSource = sources[i];
				const sourceData =
					this.marketData?.user_data?.sources[currentSource.id];

				if (sourceData && sourceData[latestKey]) {
					console.log(`${currentSource.id} latestKey execute`);
					sourceData[latestKey].forEach((item) => {
						processor(currentSource, item);
					});
				}
			}
		},

		calcRemoteSourceSimpleAppInfo() {
			this.processSources(
				this.remoteSource,
				'app_info_latest',
				(currentSource, item) => {
					if (item.app_simple_info) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.app_simple_info.app_name
						);
						this.appSimpleInfoMap.set(combinedId, item);
					}
				}
			);
		},

		calcLocationSourceSimpleAppInfo() {
			this.processSources(
				this.localSource,
				'app_info_latest',
				(currentSource, item) => {
					if (item.app_simple_info) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.app_simple_info.app_name
						);
						this.appSimpleInfoMap.set(combinedId, item);
						this.appStatusMap.set(
							combinedId,
							this._getDefaultStatus(item.app_simple_info.app_name)
						);
					}
				}
			);
		},

		calcLocationSourceAppStateInfo() {
			this.processSources(
				this.localSource,
				'app_state_latest',
				(currentSource, item) => {
					if (item.status) {
						if (
							this._getAppInfo(item.status.name, currentSource.id) ||
							isCloneApp(item?.status)
						) {
							const combinedId = getAppCombinedId(
								currentSource.id,
								item.status.name
							);
							this.appStatusMap.set(combinedId, item);
						}
					}
				}
			);
		},

		calcRemoteSourceAppStateInfo() {
			this.processSources(
				this.remoteSource,
				'app_state_latest',
				(currentSource, item) => {
					if (item.status) {
						if (
							this._getAppInfo(item.status.name, currentSource.id) ||
							isCloneApp(item?.status)
						) {
							const combinedId = getAppCombinedId(
								currentSource.id,
								item.status.name
							);
							this.appStatusMap.set(combinedId, item);
						}
					}
				}
			);
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
			key = 'app_name'
		): AppSimpleInfo | null {
			const app = this.marketData.user_data.sources[
				sourceId
			].app_info_latest.find((item) => item.app_simple_info[key] == id);

			if (!app) {
				return null;
			}

			if (!globalConfig.isOfficial) {
				const userStore = useUserStore();
				if (!userStore.initialized || !userStore.systemResource) {
					return null;
				}

				const intersectedArray = intersection(
					userStore.systemResource.nodes,
					app.app_simple_info.support_arch
						? app.app_simple_info.support_arch
						: []
				);

				if (
					userStore.systemResource.nodes.length <= 0 ||
					intersectedArray.length === 0
				) {
					console.error(
						` ${id} supportArch ->  ${app.app_simple_info.support_arch} intersectedArray length 0 and device nodes ${userStore.systemResource.nodes}`
					);
					return null;
				}
			}

			if (app) {
				this.addFullInfoQueue(id, sourceId, false);
			}
			return app ? app.app_simple_info : null;
		},
		_getDefaultStatus(appName: string): AppStatusLatest {
			return {
				type: 'app-status-default',
				status: {
					entranceStatuses: [],
					lastTransitionTime: '',
					progress: '',
					name: appName,
					rawAppName: appName,
					state: APP_STATUS.UNINSTALL.COMPLETED,
					statusTime: '',
					updateTime: ''
				},
				version: ''
			};
		}
	}
});
