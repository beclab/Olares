import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useTransfer2Store } from 'src/stores/transfer2';
import { useTerminusStore } from 'src/stores/terminus';
import { useFilterStore } from 'src/stores/rss-filter';
import { useConfigStore } from 'src/stores/rss-config';
import { busEmit, busOff, busOn, NetworkErrorMode } from 'src/utils/bus';
import { useReaderStore } from '../stores/rss-reader';
import TransferClient from '../services/transfer';
import { importFilesStyle } from './utils/files';
import { useArgoStore } from 'src/stores/argo';
import { useRssStore } from 'src/stores/rss';
import { NormalApplication } from './base';
import {
	DATABASE_VERSION,
	WISE_DATABASE
} from 'src/utils/localStorageConstant';
import { useTransferStore } from 'src/stores/rss-transfer';
import { useCookieStore } from 'src/stores/settings/cookie';
import { useBlacklistStore } from 'src/stores/settings/blacklist';

export class WiseApplication extends NormalApplication {
	applicationName = 'wise';
	private feedUpdate = () => {
		const readerStore = useReaderStore();
		readerStore.feedUpdate();
	};

	private sendMessage = (message: string) => {
		BtNotify.show({ type: NotifyDefinedType.FAILED, message });
	};

	private socketTimer: NodeJS.Timer | null = null;
	private socketTimer2: NodeJS.Timer | null = null;

	async appLoadPrepare(data: any): Promise<void> {
		await super.appLoadPrepare(data);
	}

	async appMounted(): Promise<void> {
		await super.appMounted();

		const configStore = useConfigStore();
		const base_url = window.location.origin;
		configStore.init(base_url, base_url);
		const argoStore = useArgoStore();
		argoStore.setUrl(base_url);

		busOn('feedUpdate', this.feedUpdate);
		busOn('network_error', this.sendMessage);
		importFilesStyle();

		const rssStore = useRssStore();
		const socketStore = useWebsocketManager2Store();
		const terminusStore = useTerminusStore();
		await terminusStore.getTerminusInfo();
		const filterStore = useFilterStore();
		await filterStore.init();
		busEmit('account_update');
		await terminusStore.validateTerminusInfo(
			(currentId, lastId) => {
				const currentVersion = localStorage.getItem(WISE_DATABASE);
				return currentId === lastId && currentVersion === DATABASE_VERSION;
			},
			async () => {
				await Promise.all([configStore.load(), rssStore.load()]);
			},
			async () => {
				await Promise.all([configStore.clear(), rssStore.clear()]);
			}
		);
		if (terminusStore?.olaresId) {
			const cookieStore = useCookieStore();
			await cookieStore.init(
				terminusStore.olaresId.split('@')[0],
				configStore.url + '/knowledge'
			);
		}
		const blacklistStore = useBlacklistStore();
		blacklistStore.init(base_url);
		await rssStore.sync();
		const transferStore2 = useTransfer2Store();
		const rssTransferStore = useTransferStore();
		transferStore2.init();
		rssTransferStore.init();
		socketStore.dispose();
		socketStore.restart();
		socketStore.appMount();
		this.socketTimer = setInterval(() => {
			rssStore.sync();
		}, 30 * 1000);

		this.socketTimer2 = setInterval(() => {
			rssStore.syncWaitFeeds();
		}, 3 * 1000);
	}

	async appUnMounted(): Promise<void> {
		await super.appUnMounted();
		if (this.socketTimer) {
			clearInterval(this.socketTimer);
		}
		if (this.socketTimer2) {
			clearInterval(this.socketTimer2);
		}

		const socketStore = useWebsocketManager2Store();
		socketStore.appUnMounted();

		busOff('feedUpdate', this.feedUpdate);
		busOff('network_error', this.sendMessage);
	}

	getWSConnectUrl() {
		const configStore = useConfigStore();
		return [
			configStore.getModuleSever(
				process.env.NODE_ENV === 'production'
					? ''
					: (process.env.WISE_SUB_DOMAIN as string | ''),
				'wss:',
				'/ws'
			)
		];
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.WISE_NAME,

		externalInfo() {
			return {
				terminusId: TransferClient.client.clouder?.taskBaseIdentify() || ''
			};
		},
		responseShareWorkerMessage(data: any) {
			busEmit('wiseDownloadProcess', data);
			busEmit('CloudTransferUpdate', data);
		}
	};

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			config.headers['X-Unauth-Error'] = 'Non-Redirect';
			return config;
		});
		this.responseIntercepts.push((response) => {
			if (
				!response ||
				(response.status != 200 &&
					response.status != 201 &&
					response.status != 204)
			) {
				busEmit('network_error', {
					type: NetworkErrorMode.axois,
					error: 'Network error, please try again later'
				});
			}

			if (response.data.code === 550) {
				busEmit('appRestore');
				return response.data;
			}

			const opmlIndex = response.config.url?.indexOf('/opml');
			if (opmlIndex && opmlIndex >= 0) {
				return response;
			}

			const terminus = response.config.url?.indexOf('/terminus');
			if (terminus && terminus >= 0) {
				return response.data;
			}

			const logIndex =
				response.config.url?.indexOf('/log') ||
				response.config.url?.indexOf('/download/preview');
			if (logIndex && logIndex >= 0) {
				return response;
			}

			const artifactIndex = response.config.url?.indexOf('artifact-file');
			if (artifactIndex && artifactIndex >= 0) {
				return response.data;
			}

			const downloadIndex = response.config.url?.indexOf('/download');
			if (downloadIndex && downloadIndex >= 0) {
				if (!response.data || response.data.code !== 0) {
					busEmit(
						'network_error',
						response.data.message || 'Network error, please try again later'
					);
					throw new Error(response.data.message);
				}
				return response.data.data;
			}

			const index = response.config.url?.indexOf('api/v1');
			if (index && index >= 0) {
				return response.data;
			} else {
				const isNoToast = response.config.noToast === true;
				if ((!response.data || response.data.code !== 0) && !isNoToast) {
					busEmit(
						'network_error',
						response.data.message || 'Network error, please try again later'
					);
					throw new Error(response.data.message);
				}
			}

			return response.data.data;
		});

		this.responseErrorInterceps = (error: any) => {
			busEmit('network_error', error.message);
			throw error;
		};
	}
}
