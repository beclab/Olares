import { CFG_TYPE } from 'src/constant/config';
import { ErrorGroup } from 'src/constant/errorGroupHandler';
import { TerminusEntrance, TermiPassDeviceInfo } from '@bytetrade/core';
import { OrderDataBase } from 'src/payment/types';
import { BaseEnv } from 'src/constant/index';
export interface MarketData {
	user_data: {
		sources: {
			'Official-Market-Sources': SourceData;
			local: SourceData;
		};
		hash: string;
	};
	user_id: string;
	timestamp: number;
}

export interface MarketSource {
	id: string;
	name: string;
	base_url: string;
	type: string;
	description: string;
	is_active: boolean;
	priority: number;
	updated_at: string;
}

export enum MARKET_SOURCE_TYPE {
	LOCAL = 'local',
	REMOTE = 'remote'
}

export const MARKET_SOURCE_PREFIX = 'market.';
export const MARKET_SOURCE_OFFICIAL = {
	LOCAL: {
		UPLOAD: 'upload',
		STUDIO: 'studio',
		CLI: 'cli'
	},
	REMOTE: {
		OLARES: 'market.olares'
	}
};

function collectAllSources(sourceObj) {
	const sources = new Set();

	function traverse(obj) {
		Object.values(obj).forEach((value) => {
			if (typeof value === 'object' && value !== null) {
				traverse(value);
			} else if (typeof value === 'string') {
				sources.add(value);
			}
		});
	}

	traverse(sourceObj);
	return sources;
}
export const ALL_MARKET_OFFICIAL_SOURCES = collectAllSources(
	MARKET_SOURCE_OFFICIAL
);

interface SourceData {
	type: string;
	app_state_latest: AppStatusLatest[];
	app_info_latest: AppSimpleInfoLatest[];
	others: SourceOthers;
}

export interface SourceOthers {
	hash: string;
	version: string;
	topics: {
		createdAt: string;
		source: string;
		name: string;
		updated_at: string;
		_id: string;
		data: TopicInfo[];
	};
	topic_lists: Topic[];
	recommends: Recommend[];
	latest: string[];
	tops: RankApp[];
	pages: Page[];
	tags: MenuData[];
}

export class MenuData {
	createdAt: string;
	icon: string;
	name: string;
	sort: number;
	source: string;
	title: AppI18n;
	updated_at: string;
	_id: string;
}

export interface RankApp {
	appid: string;
	rank: number;
}

export interface AppSimpleInfoLatest {
	type: string;
	timestamp: number;
	version: string;
	app_simple_info: AppSimpleInfo;
}

export interface AppSimpleInfo {
	app_id: string;
	app_name: string;
	app_icon: string;
	app_description: AppI18n;
	app_version: string;
	app_title: AppI18n;
	categories: string[];
	support_arch: string;
}

export interface AppInfoAggregation {
	app_simple_latest: AppSimpleInfoLatest;
	app_status_latest: AppStatusLatest;
	app_full_info: AppFullInfoLatest;
	app_error_group?: ErrorGroup[];
}

export interface AppFullInfo {
	app_entry: AppEntry;
	image_analysis: ImageAnalysis;
	price: PriceConfig;
}

export interface PaymentOrderData {
	from: string;
	to: string;
	ras_public_key: string;
	product: Array<{ product_id: string }>;
	price_config: PriceConfig;
	token_info: SupportToken[];
}

export interface SupportToken {
	chain: string;
	receive_wallet: string;
	token_amount: number;
	token_contract: string;
	token_decimals: number;
	token_icon: string;
	token_symbol: string;
}

export interface PriceConfig {
	developer: string;
	paid?: PaidInfo | null;
	products: Product[];
}

export interface PaidInfo {
	product_id: string;
	price: PriceEntry[];
	description?: PriceDescription[];
}

export interface Product {
	product_id: string;
	type: string;
	price: PriceEntry[];
	description: PriceDescription[];
}

export interface PriceEntry {
	chain: string;
	token_symbol: string;
	receive_wallet: string;
	product_price: number;
}

export interface PriceDescription {
	lang: string;
	title?: string;
	description?: string;
	icon?: string;
}

export interface ImageAnalysis {
	source_id: string;
	total_images: number;
	user_id: string;
	images: Record<string, ImageDetails>;
}

export interface ImageDetails {
	analyzed_at: string;
	architecture: string;
	create_at: string;
	download_progress: number;
	downloaded_layers: number;
	downloaded_size: number;
	error_message: string;
	layer_count: number;
	name: string;
	node: any[];
	status: string;
	tag: string;
	total_size: number;
}

export interface AppStatusLatest {
	type: string;
	status: AppStatusInfo;
	version: string;
}

export interface AppStatusInfo {
	entranceStatuses: TerminusEntrance[];
	lastTransitionTime: string;
	name: string;
	state: string;
	opType?: string;
	progress: string;
	statusTime: string;
	rawAppName: string;
	updateTime: string;
}

export enum STATUS_OPERATE_TYPE {
	INSTALL = 'install',
	UPGRADE = 'upgrade'
}

export interface MarketAppRequest {
	appid: string;
	sourceDataName: string;
}

export interface AppFullInfoLatest {
	type: string;
	timestamp: number;
	version: string;
	raw_package: any;
	values: any;
	app_info: AppFullInfo;
	rendered_package: any;
	app_simple_info: AppSimpleInfo;
}

interface AppI18n {
	'en-US': string;
	'zh-CN': string;
}

export function getI18nValue<T = string>(
	i18nObject: Record<string, T> | undefined,
	locale: string,
	defaultValue?: T
): T | undefined {
	if (!i18nObject) return defaultValue;

	if (i18nObject[locale] !== undefined) {
		return i18nObject[locale];
	}

	if (i18nObject['en-US'] !== undefined) {
		return i18nObject['en-US'];
	}

	return Object.values(i18nObject)[0] ?? defaultValue;
}

export interface TopicInfo {
	apps: string;
	des: string;
	detailimg: string;
	group: string;
	iconimg: string;
	isdelete: boolean;
	richtext: string;
	title: string;
	topicId: string;
	mobileDetailImg: string;
	mobileRichtext: string;
	backgroundColor: string;
}

export interface Topic {
	name: string;
	type: string;
	description: string;
	content: string;
	createdAt: string;
	updated_at: string;
	title: AppI18n;
}

export interface Recommend {
	name: string;
	description: string;
	content: string;
	createdAt: string;
	updated_at: string;
	data: {
		title: AppI18n;
	};
}

export interface Page {
	category: string;
	content: string;
	createdAt: string;
	updated_at: string;
}

export interface AppEntry {
	appID: string;
	categories: string[];
	count: string;
	name: string;
	cfgType: CFG_TYPE;
	chartName: string;
	icon: string;
	description: AppI18n;
	appid: string;
	title: AppI18n;
	version: string;
	curVersion: string;
	needUpdate: boolean;
	versionName: string;
	fullDescription: AppI18n;
	upgradeDescription: AppI18n;
	promoteImage: string[];
	promoteVideo: string;
	supportArch: string[];
	subCategory: string;
	developer: string;
	requiredMemory: string;
	requiredDisk: string;
	apiVersion: string;
	envs: BaseEnv[];
	subCharts: { name: string; shared?: boolean }[];
	supportClient: {
		edge: string;
		android: string;
		ios: string;
		windows: string;
		mac: string;
		linux: string;
		chrome: string;
	};
	requiredGPU: string;
	requiredCPU: string;
	rating: number;
	target: string;
	namespace: string;
	onlyAdmin: boolean;
	permission: {
		appData: boolean;
		appCache: boolean;
		userData: string[];
		sysData: SysDataCfg[];
		provider: AppProvider[];
	};
	versionHistory: VersionRecord[];
	entrances: Entrance[];
	middleware: Middleware;
	options: {
		mobileSupported: boolean;
		allowMultipleInstall: boolean;
		analytics: {
			enable: boolean;
		};
		dependencies: Dependency[];
		conflicts: Conflict[];
		policies: Policy[];
		appScope: {
			clusterScoped: boolean;
			appRef: string[];
		};
		websocket: {
			port: number;
			url: string;
		};
	};
	i18n: any;
	ports: Port[];
	tailscale: {
		acls: [
			{
				dst: string[];
				proto: string;
			}
		];
	};
	lastCommitHash: string;
	createTime: number;
	updateTime: number;
	installTime: string;
	uid: string;
	status: string;
	locale: string[];
	submitter: string;
	doc: string;
	website: string;
	featuredImage: string;
	sourceCode: string;
	license: License[];
	legal: null;
	appLabels: string[];
	source: string;
	modelSize: string;
}

export interface Middleware {
	[middlewareName: string]: MiddleWareCfg;
}

export interface AppProvider {
	appName: string;
	providerName: string;
}

export interface Port {
	exposePort: number;
	host: string;
	name: string;
	port: number;
	protocol: string;
	addToTailscaleAcl: boolean;
}

export function isCSV2(fullInfo: AppFullInfoLatest) {
	return (
		fullInfo &&
		fullInfo.app_info &&
		fullInfo.app_info.app_entry &&
		fullInfo.app_info.app_entry.subCharts &&
		fullInfo.app_info.app_entry.subCharts.length > 0 &&
		fullInfo.app_info.app_entry.apiVersion === 'v2'
	);
}

export interface Conflict {
	name: string;
	type: string;
}

export interface Entrance {
	authLevel: string;
	host: string;
	icon: string;
	name: string;
	port: 0;
	title: string;
	invisible: boolean;
}

export interface SysDataCfg {
	group: string;
	dataType: string;
	version: string;
	ops: string[];
}

export interface MiddleWareCfg {
	database: any;
	username: string;
	password: string;
}

export interface Dependency {
	name: string;
	type: DEPENDENCIES_TYPE;
	version: string;
	mandatory: boolean;
}

export interface ClusterApp {
	name: string;
	type: CLUSTER_TYPE;
	version: string;
	state: string;
}

export interface Policy {
	entranceName: string;
	description: string;
	level: string;
	oneTime: boolean;
	uriRegex: string;
	validDuration: string;
}

export interface License {
	text: string;
	url: string;
}

export interface VersionRecord {
	appName: string;
	mergedAt: string;
	version: string;
	versionName: string;
	upgradeDescription: string;
}

export interface PermissionNode {
	label: string;
	icon?: string;
	children: PermissionNode[];
}

export interface Token {
	access_token: string;
	token_type: string;
	refresh_token: string;
	expires_in: number;
	expires_at: number;
}

export interface Resource {
	total: number;
	usage: number;
	ratio: number;
	unit: string;
}

export interface UserResource {
	cpu: Resource;
	memory: Resource;
	disk: Resource;
	gpu: Resource;
}

export interface TerminusResource {
	apps: ClusterApp[];
	metrics: {
		cpu: Resource;
		memory: Resource;
		disk: Resource;
		gpu: Resource;
	};
	nodes: string[];
}

export interface User {
	role: string;
	username: string;
}

export interface MenuType {
	label: string;
	key: string;
	img: string;
	sort: number;
}

export const CONTENT_TYPE = {
	TOP: 'Top',
	LATEST: 'Latest',
	RECOMMENDS: 'Recommends',
	TOPIC: 'Topic',
	APP: 'App'
};
export const TOPIC_TYPE = {
	TOPIC: 'Topic',
	CATEGORY: 'Category'
};

export enum VERSION_DISPLAY_MODE {
	NONE = 'none',
	ONLY_MY = 'my',
	ONLY_TARGET = 'target',
	PRIORITY_MY = 'priority-my',
	PRIORITY_TARGET = 'priority-target'
}

export enum TRANSACTION_PAGE {
	All = 'All',
	CATEGORIES = 'Category',
	App = 'App',
	List = 'List',
	TOPIC = 'Topic',
	Preview = 'Preview',
	Version = 'Version',
	Log = 'Log',
	Update = 'Update',
	Search = 'Search',
	MyTerminus = 'MyTerminus',
	Preference = 'Preference'
}

export enum DEPENDENCIES_TYPE {
	application = 'application',
	system = 'system',
	middleware = 'middleware'
}

export enum CLUSTER_TYPE {
	app = 'app',
	middleware = 'middleware'
}

export const APP_STATUS = {
	PENDING: {
		DEFAULT: 'pending',
		CANCELING: 'pendingCanceling',
		CANCELED: 'pendingCanceled',
		CANCEL_FAILED: 'pendingCancelFailed'
	},
	DOWNLOAD: {
		DEFAULT: 'downloading',
		CANCELING: 'downloadingCanceling',
		CANCELED: 'downloadingCanceled',
		CANCEL_FAILED: 'downloadingCancelFailed',
		FAILED: 'downloadFailed'
	},
	INSTALL: {
		DEFAULT: 'installing',
		CANCELING: 'installingCanceling',
		CANCELED: 'installingCanceled',
		CANCEL_FAILED: 'installingCancelFailed',
		FAILED: 'installFailed'
	},
	INITIALIZE: {
		DEFAULT: 'initializing',
		CANCELING: 'initializingCanceling'
	},
	STOP: {
		DEFAULT: 'stopping',
		COMPLETED: 'stopped',
		FAILED: 'stopFailed'
	},
	RESUME: {
		DEFAULT: 'resuming',
		CANCELING: 'resumingCanceling',
		FAILED: 'resumeFailed'
	},
	UNINSTALL: {
		DEFAULT: 'uninstalling',
		FAILED: 'uninstallFailed',
		COMPLETED: 'uninstalled'
	},
	ENV: {
		APPLYING: 'applyingEnv',
		APPLY_FAILED: 'applyEnvFailed',
		CANCELING: 'applyingEnvCanceling',
		CANCEL_FAILED: 'applyingEnvCancelFailed',
		CANCELED: 'applyingEnvCanceled'
	},
	RUNNING: 'running',
	UPGRADE: {
		DEFAULT: 'upgrading',
		CANCELING: 'upgradingCanceling',
		FAILED: 'upgradeFailed'
	},
	MODEL: {
		INSTALLED: 'installed'
	}
} as const;

export enum ENTRANCE_STATUS {
	NOT_READY = 'notReady',
	STOPPED = 'stopped',
	RUNNING = 'running'
}

export enum LOCAL_STATUS {
	PRE_CHECK_FINISHED = 'preCheckFinished'
}

export enum PAYMENT_STATUS {
	NOT_EVALUATED = 'not_evaluated',
	PURCHASED = 'purchased',
	WAITING_DEVELOPER_CONFIRMATION = 'waiting_developer_confirmation',
	PAYMENT_RETRY_REQUIRED = 'payment_retry_required',
	PAYMENT_REQUIRED = 'payment_required',
	SIGNATURE_REQUIRED = 'signature_required',
	SIGNATURE_NEED_RESIGN = 'signature_need_resign',
	SYNCING = 'syncing',
	NOT_BUY = 'not_buy'
}

export enum APP_PAYMENT_TYPE {
	APP_TYPE_FREE = 'FREE',
	APP_TYPE_PAID = 'PAID',
	APP_TYPE_UNKNOWN = 'UNKNOWN'
}

export interface AppLocalInfo {
	status: LOCAL_STATUS | PAYMENT_STATUS;
	data: any;
}

export interface AppStatusChange {
	app_info_latest: AppSimpleInfoLatest;
	app_state_latest: AppStatusLatest;
	timestamp: string;
	user: string;
	source: string;
	app_name: string;
	notify_type: string;
}

export interface MarketSystemChange {
	extensions: MarketSystemExtensions;
	notify_type: string;
	point: string;
	timestamp: string;
	user: string;
}

export interface ProductOrder {
	extensions: MarketSystemExtensions;
	extensions_obj: {
		payment_data: OrderDataBase;
	};
	notify_type: string;
	point: string;
	timestamp: string;
	user: string;
}

export interface ImageInfoUpdate {
	image_info: ImageDetails;
	timestamp: string;
	user: string;
	notify_type: string;
}

export interface MarketSystemExtensions {
	app_name: string;
	app_version: string;
	source: string;
	task_status: string;
	task_type: string;
}

export enum CLIENT_TYPE {
	ios = 'ios',
	android = 'android',
	edge = 'edge',
	windows = 'windows',
	mac = 'mac',
	linux = 'linux',
	chrome = 'chrome'
}

export function getDeviceIconName(device: TermiPassDeviceInfo): string {
	let name = '';
	if (device) {
		switch (device.platform) {
			case 'android':
			case 'Android':
				switch (device.manufacturer) {
					case 'Xiaomi':
						name = 'xiaomi';
						break;
					case 'HUAWEI':
						name = 'huawei';
						break;
					case 'Honor':
						name = 'honor';
						break;
					case 'Oppo':
						name = 'oppo';
						break;
					case 'Vivo':
						name = 'vivo';
						break;
					case 'Samsung':
						name = 'samsung';
						break;
					case 'Google':
						name = 'google';
						break;
					default:
						name = 'android';
				}
				break;
			case 'ios':
			case 'iOS':
				name = 'iPhone';
				break;
			case 'chrome extends':
				name = 'chrome';
				break;
			case 'MacOS':
				name = 'mac';
				break;
			case 'Windows':
				name = 'windows';
				break;
		}
	}
	if (name) {
		return `settings/devices/${name}.png`;
	} else {
		return '';
	}
}

export interface DomainCookieRecord {
	domain: string;
	name: string;
	value: string;
	expires: number | string | Date;
	path: string;
	secure: boolean;
	httpOnly?: boolean;
	sameSite?: string;
	other?: string;
}

export interface AggregatedDomain {
	mainDomain: string;
	subDomains: DomainCookie[];
}

export class DomainCookie {
	domain = '';
	account = '';
	updateTime = 0;
	records: DomainCookieRecord[] = [];

	get_store_key(): string {
		if (this.account) {
			return 'cookie:' + this.domain + ':' + this.account;
		} else {
			return 'cookie:' + this.domain;
		}
	}

	hasExpiredRecords(isExpiresInSeconds = true): boolean {
		if (this.records.length === 0) return false;

		const currentTime = isExpiresInSeconds
			? Math.floor(Date.now() / 1000)
			: Date.now();

		return this.records.some((record) => {
			if (
				record.expires === 0 ||
				record.expires == undefined ||
				record.expires == ''
			)
				return false;
			return currentTime > record.expires;
		});
	}

	constructor(props?: Partial<DomainCookie>) {
		props && Object.assign(this, props);
	}
}

export interface NetscapeCookie {
	domain: string;
	includeSubdomain: boolean;
	path: string;
	secure: boolean;
	expiration: number;
	name: string;
	value: string;
}

export function generatePageDataItemKey(item: any, index: number) {
	const parts = [
		item.type || '',
		item.name || '',
		item.description || '',
		Array.isArray(item.content) ? item.content.join(',') : '',
		Array.isArray(item.ids) ? item.ids.join(',') : '',
		index
	];
	return parts.join('_');
}

export function getAppCombinedId(sourceId: string, appName: string): string {
	return `${sourceId}_${appName}`;
}
