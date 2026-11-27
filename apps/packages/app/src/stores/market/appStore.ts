import {
	getErrorTextByState,
	isCloneApp,
	isFailed,
	isTemplateApp,
	uninstalledApp
} from 'src/constant/config';
import globalConfig from 'src/api/market/config';
import { Version } from 'src/utils/version';
import { defineStore } from 'pinia';
import {
	APP_REJECT_REASON,
	APP_STATUS,
	AppFullInfoLatest,
	AppInfoAggregation,
	AppLocalInfo,
	AppSimpleInfoLatest,
	AppStatusChange,
	AppStatusInfo,
	AppStatusLatest,
	getAppCombinedId,
	getEffectiveTime,
	ImageInfoUpdate,
	isSharedV3,
	MARKET_SOURCE_OFFICIAL,
	MARKET_SOURCE_TYPE,
	MarketData,
	MarketSource,
	MarketSystemChange,
	PAYMENT_STATUS,
	STATUS_OPERATE_TYPE
} from 'src/constant/constants';
import {
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
import { CacheRequest } from 'src/stores/market/CacheRequest';
import { ImageUpdater } from 'src/stores/market/ImageUpdater';
import { getInstallAppState, getMarketApps } from 'src/api/market/centerApi';
import { usePreCheckStore } from 'src/stores/market/preCheck';
import { useMeStore } from '../settings/me';

export type AppStoreState = {
	appUrl: string;
	sources: MarketSource[];
	localStatusMap: Map<string, AppLocalInfo>;
	appSimpleInfoMap: Map<string, AppSimpleInfoLatest>;
	appStatusMap: Map<string, AppStatusLatest>;
	appFullInfoMap: Map<string, AppFullInfoLatest>;
	appInfoProcessor: FullInfoProcessor | null;
	imageUpdater: ImageUpdater | null;
};

const sourcesTitleMap: Record<string, string> = {
	cli: 'CLI'
};

export const useAppStore = defineStore('appStore', {
	state: () => {
		return {
			appUrl: `${globalConfig.url}/app-store/api/v2`,
			sources: [] as MarketSource[],
			localStatusMap: new Map<string, AppLocalInfo>(),
			appSimpleInfoMap: new Map<string, AppSimpleInfoLatest>(),
			appStatusMap: new Map<string, AppStatusLatest>(),
			appFullInfoMap: new Map<string, AppFullInfoLatest>(),
			appInfoProcessor: null,
			imageUpdater: null
		} as AppStoreState;
	},
	getters: {
		remoteSource(state) {
			return state.sources.filter(
				(item) => item.type === MARKET_SOURCE_TYPE.REMOTE
			);
		},
		localSource(state) {
			return state.sources.filter(
				(item) => item.type === MARKET_SOURCE_TYPE.LOCAL
			);
		},
		installStatusList(state) {
			return Array.from(state.appStatusMap.values()).filter(
				(item: AppStatusLatest) => !uninstalledApp(item.status)
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
			const simpleLatest = this.appSimpleInfoMap.get(combinedId);
			const statusLatest = this.appStatusMap.get(combinedId);
			const fullLatest = this.appFullInfoMap.get(combinedId);

			if (simpleLatest) {
				this.addFullInfoQueue(appId, sourceId, true);
				return {
					app_simple_latest: simpleLatest,
					app_status_latest: statusLatest
						? statusLatest
						: this._getDefaultStatus(simpleLatest.app_simple_info.app_name),
					app_full_info: fullLatest
				} as AppInfoAggregation;
			}
			return null;
		},
		//Only install app (unique by app name)
		findAppByName(appName: string): {
			status: AppStatusInfo;
			sourceId: string;
			appId: string;
			version: string;
		} | null {
			if (!appName) {
				return null;
			}

			const resolveEntry = (nameFromKey: string, sourceId: string) => {
				const appStatusLatest = this.getAppStatus(nameFromKey, sourceId);
				if (appStatusLatest && !uninstalledApp(appStatusLatest.status)) {
					return {
						status: appStatusLatest.status,
						sourceId,
						appId: nameFromKey,
						version: appStatusLatest.version
					};
				}
				return null;
			};

			for (const [combinedId] of this.appStatusMap.entries()) {
				const { sourceId, appName: nameFromKey } =
					this.splitCombinedId(combinedId);
				if (nameFromKey === appName) {
					const entry = resolveEntry(nameFromKey, sourceId);
					if (entry) return entry;
				}
			}

			for (const [combinedId, statusLatest] of this.appStatusMap.entries()) {
				const { sourceId, appName: nameFromKey } =
					this.splitCombinedId(combinedId);
				if (statusLatest.status.rawAppName === appName) {
					const entry = resolveEntry(nameFromKey, sourceId);
					if (entry) return entry;
				}
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
		setAppStatus(combinedId: string, appStatus: AppStatusLatest) {
			const oldState = this.appStatusMap.get(combinedId);

			if (oldState) {
				const oldTime = getEffectiveTime(oldState);
				const newTime = getEffectiveTime(appStatus);

				if (newTime === 0) {
					console.error(
						`[setAppStatus] Skipped: newTime invalid, combinedId=${combinedId}, oldState=${oldState.status?.state}, newState=${appStatus.status?.state}`
					);
					return;
				}
				if (oldTime > 0 && oldTime > newTime) {
					return;
				}
				// Progress monotonicity only makes sense *within the same state*
				// (e.g. download progress jitter at the same statusTime). On a state
				// transition (downloading -> install failed, etc.) the new state may
				// legitimately carry a reset/lower progress, so it must NOT be blocked
				// here; otherwise a failed install gets dropped and the UI stays stuck
				// at the previous progress (e.g. 100%).
				if (
					oldTime > 0 &&
					oldTime === newTime &&
					oldState.status?.state === appStatus.status?.state
				) {
					const oldProgressStr = oldState.status?.progress;
					const newProgressStr = appStatus.status?.progress;
					if (oldProgressStr && newProgressStr) {
						const oldProgress = parseFloat(oldProgressStr);
						const newProgress = parseFloat(newProgressStr);
						if (
							!isNaN(oldProgress) &&
							!isNaN(newProgress) &&
							oldProgress > newProgress
						) {
							console.error(
								`[setAppStatus] Skipped: progress rollback, combinedId=${combinedId}, oldProgress=${oldProgress}, newProgress=${newProgress}`
							);
							return;
						}
					}
				}
			}
			this.appStatusMap.set(combinedId, appStatus);
		},
		getAppSimpleInfo(appId: string, sourceId: string, status?: AppStatusInfo) {
			const combinedId = getAppCombinedId(sourceId, appId);
			return this.appSimpleInfoMap.get(combinedId);
		},
		getAppFullInfo(appId: string, sourceId: string) {
			const combinedId = getAppCombinedId(sourceId, appId);
			return this.appFullInfoMap.get(combinedId);
		},
		upsertAppFullInfo(combinedId: string, appInfo: AppFullInfoLatest) {
			this.appFullInfoMap.set(combinedId, appInfo);
			this.appSimpleInfoMap.set(combinedId, {
				app_simple_info: appInfo.app_simple_info,
				type: appInfo.type,
				timestamp: appInfo.timestamp,
				version: appInfo.version
			} as AppSimpleInfoLatest);
			updateAppFullInfoMapKey(combinedId, appInfo);

			const { appName, sourceId } = this.splitCombinedId(combinedId);
			if (sourceId === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD) {
				if (!this.appStatusMap.get(combinedId)) {
					this.appStatusMap.set(combinedId, this._getDefaultStatus(appName));
				}
			}
		},
		async fetchAppFullInfoNow(
			appId: string,
			sourceId: string
		): Promise<AppFullInfoLatest | undefined> {
			if (!appId || !sourceId) {
				return undefined;
			}
			if (!this.sources.find((item) => item.id === sourceId)) {
				return undefined;
			}

			const combinedId = getAppCombinedId(sourceId, appId);
			try {
				const data: any = await getMarketApps([
					{
						appid: appId,
						sourceDataName: sourceId
					}
				]);
				const marketApps: AppFullInfoLatest[] = Array.isArray(data?.apps)
					? data.apps
					: Array.isArray(data)
					? data
					: [];
				const latest = marketApps[0];
				if (!latest) {
					return undefined;
				}
				this.upsertAppFullInfo(combinedId, latest);
				return latest;
			} catch (error) {
				console.error(`[fetchAppFullInfoNow] ${combinedId} failed:`, error);
				return undefined;
			}
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
						const meStore = useMeStore();
						const simpleLatest = this.getAppSimpleInfo(appName, appSourceId);
						if (
							isSharedV3(simpleLatest) &&
							!!statusLatest.status.owner &&
							statusLatest.status.owner !== meStore.user.name &&
							uninstalledApp(statusLatest.status) &&
							!isTemplateApp(simpleLatest?.app_simple_info)
						) {
							return;
						}
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
		updateAppStatus(
			appId: string,
			sourceId: string,
			newState: string,
			opType?: string
		) {
			const statusLatest = this.getAppStatus(appId, sourceId);
			if (statusLatest) {
				statusLatest.status.state = newState;
				if (opType) {
					statusLatest.status.opType = opType;
				}
				const combinedId = getAppCombinedId(sourceId, appId);
				this.setAppStatus(combinedId, statusLatest);
			}
		},
		updateAppStatusBySocket(data: AppStatusChange) {
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
					// Print both: reason first (generic state-derived text), then the
					// concrete message detail. The message prefers the status field,
					// falling back to the raw push's top-level message.
					const reason = i18n.global.t(
						getErrorTextByState(data.app_state_latest.status.state)
					);
					const detail =
						data.app_state_latest.status.message?.trim() ||
						data.message?.trim() ||
						'';
					bus.emit(
						BUS_EVENT.APP_BACKEND_ERROR,
						`${data.app_name} ${reason}${detail ? `: ${detail}` : ''}`
					);
				}
				this.setAppStatus(combinedId, data.app_state_latest);
			}

			if (
				data.app_state_latest.status.state === APP_STATUS.STOP.COMPLETED &&
				data.app_state_latest.status.reason ===
					APP_REJECT_REASON.HAMI_UNSCHEDULABLE &&
				process.env.APPLICATION !== 'Desktop'
			) {
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					i18n.global.t(
						'Insufficient VRAM. The application has been automatically stopped as memory usage exceeded the limit.'
					)
				);
				this.setAppStatus(combinedId, data.app_state_latest);
			}

			if (
				data.app_state_latest.status.state === APP_STATUS.UNINSTALL.COMPLETED ||
				data.app_state_latest.status.state === APP_STATUS.RUNNING
			) {
				const preCheckStore = usePreCheckStore();
				preCheckStore.initSystemRequest().forceRefresh();
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
		async init(proxy = false) {
			if (proxy) {
				this.appUrl = '/market-backend/app-store/api/v2';
			}
			const map = await loadAppFullInfoMap();
			if (map) {
				this.appFullInfoMap = map;
			}
			const store = this;
			this.appInfoProcessor = new FullInfoProcessor({
				handleAppInfo(combinedId: string, appInfo: any) {
					store.upsertAppFullInfo(combinedId, appInfo);
				}
			} as AppHandler);
			this.appInfoProcessor?.processTempAppQueue();
		},
		initSourceRequest() {
			return new CacheRequest('cache_market_source', getMarketSource, {
				onData: (data) => {
					data.forEach((e) => {
						if (!e.title) {
							if (sourcesTitleMap[e.name]) {
								e.title = sourcesTitleMap[e.name];
							} else {
								e.title = e.name;
							}
						}
					});
					this.sources = data;
					console.log('sources ===>', this.sources);
				}
			});
		},
		async fetchAppState() {
			try {
				const data = await getInstallAppState();
				if (data) {
					await this.fetchData(data);
				}
			} catch (error) {
				console.error('Failed to fetch app state:', error);
			}
		},
		async fetchData(marketData: MarketData) {
			if (!marketData) {
				return;
			}
			this.processSources(
				this.remoteSource,
				'app_info_latest',
				marketData,
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
			this.processSources(
				this.localSource,
				'app_info_latest',
				marketData,
				(currentSource, item) => {
					if (item.app_simple_info) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.app_simple_info.app_name
						);
						this.appSimpleInfoMap.set(combinedId, item);
						this.setAppStatus(
							combinedId,
							this._getDefaultStatus(item.app_simple_info.app_name)
						);
					}
				}
			);
			this.processSources(
				this.remoteSource,
				'app_state_latest',
				marketData,
				(currentSource, item) => {
					if (item.status) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.status.name
						);
						if (this.appSimpleInfoMap.has(combinedId)) {
							this.setAppStatus(combinedId, item);
						}

						this.addFullInfoQueue(item.status.name, currentSource.id, false);
					}
				}
			);
			this.processSources(
				this.localSource,
				'app_state_latest',
				marketData,
				(currentSource, item) => {
					if (item.status) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.status.name
						);
						if (this.appSimpleInfoMap.has(combinedId)) {
							this.setAppStatus(combinedId, item);
						}

						this.addFullInfoQueue(item.status.name, currentSource.id, false);
					}
				}
			);
		},
		processSources(
			sources: MarketSource[],
			latestKey: string,
			marketData: MarketData,
			processor: (currentSource: MarketSource, item: any) => void
		) {
			for (let i = 0; i < sources.length; i++) {
				const currentSource = sources[i];
				const sourceData = marketData?.user_data?.sources[currentSource.id];

				if (sourceData && sourceData[latestKey]) {
					console.log(`${currentSource.id} latestKey execute`);
					sourceData[latestKey].forEach((item) => {
						processor(currentSource, item);
					});
				}
			}
		},
		calcRemoteSourceSimpleAppInfo(marketData: MarketData) {
			this.processSources(
				this.remoteSource,
				'app_info_latest',
				marketData,
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
		calcLocationSourceSimpleAppInfo(marketData: MarketData) {
			this.processSources(
				this.localSource,
				'app_info_latest',
				marketData,
				(currentSource, item) => {
					if (item.app_simple_info) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.app_simple_info.app_name
						);
						this.appSimpleInfoMap.set(combinedId, item);
						this.setAppStatus(
							combinedId,
							this._getDefaultStatus(item.app_simple_info.app_name)
						);
					}
				}
			);
		},
		calcRemoteSourceAppStateInfo(marketData: MarketData) {
			this.processSources(
				this.remoteSource,
				'app_state_latest',
				marketData,
				(currentSource, item) => {
					if (item.status) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.status.name
						);
						if (this.appSimpleInfoMap.has(combinedId)) {
							this.setAppStatus(combinedId, item);
						}
					}
				}
			);
		},
		calcLocationSourceAppStateInfo(marketData: MarketData) {
			this.processSources(
				this.localSource,
				'app_state_latest',
				marketData,
				(currentSource, item) => {
					if (item.status) {
						const combinedId = getAppCombinedId(
							currentSource.id,
							item.status.name
						);
						if (this.appSimpleInfoMap.has(combinedId)) {
							this.setAppStatus(combinedId, item);
						}
					}
				}
			);
		},
		_getDefaultStatus(appName: string): AppStatusLatest {
			return {
				type: 'app-status-default',
				status: {
					entranceStatuses: [],
					sharedEntrances: [],
					lastTransitionTime: '',
					progress: '',
					name: appName,
					rawAppName: appName,
					state: APP_STATUS.UNINSTALL.COMPLETED,
					statusTime: '1970-01-01T00:00:00.000Z',
					updateTime: '1970-01-01T00:00:00.000Z'
				},
				version: ''
			};
		}
	}
});
