import {
	TerminusApp,
	AccountType,
	BackupFrequency,
	BackupLocation,
	IntegrationAccountMiniData
} from '@bytetrade/core';

import { i18n } from '../boot/i18n';
import { getRequireImage } from '../utils/rss-utils';
import { DriveType } from '../utils/interface/files';
import { timeToTimeStamp } from '../pages/settings/Backup2/pages/FormatBackupTime';
import { computed, ref } from 'vue';
import { APP_STATUS, ENTRANCE_STATUS } from 'src/constant/constants';
import axios from 'axios';

export enum BackgroundMode {
	desktop = 'Desktop',
	login = 'Login'
}

export const imgContentModes: 'fill'[] =
	// | 'contain' | 'cover' | 'none' | 'scale-down'
	[
		'fill'
		// 'contain',
		// 'cover',
		// 'none',
		// 'scale-down'
	];

export const firstToUpper = (str: string) => {
	if (str.length === 0) {
		return str;
	}
	return str.trim().toLowerCase().replace(str[0], str[0].toUpperCase());
};

export const getCookie = (cname: string) => {
	const name = cname + '=';
	const ca = document.cookie.split(';');
	for (let i = 0; i < ca.length; i++) {
		const c = ca[i].trim();
		if (c.indexOf(name) == 0) {
			return c.substring(name.length, c.length);
		}
	}
};

export function checkDomainSuffix(domainSuffix: string) {
	return new Promise<boolean>((resolve) => {
		const currentDomain = window.location.hostname;
		const isMatch = currentDomain
			.toLowerCase()
			.endsWith(domainSuffix.toLowerCase());
		resolve(isMatch);
	});
}

export enum MENU_TYPE {
	Root = 'Root',
	Application = 'Application',
	Integration = 'Integration',
	Users = 'Users',
	Backup = 'Backup',
	Appearance = 'Appearance',
	VPN = 'VPN',
	Network = 'Network',
	GPU = 'GPU',
	Restore = 'Restore',
	Developer = 'Developer',
	Video = 'Video',
	Search = 'Search'
}

export interface MenuItem {
	label: string;
	title?: string;
	key: MENU_TYPE;
	img: string;
	description?: string;
}

const BaseMenuItems: Record<MENU_TYPE, MenuItem> = {
	[MENU_TYPE.Root]: {
		label: '',
		key: MENU_TYPE.Root,
		img: ''
	},
	[MENU_TYPE.Users]: {
		label: 'home_menus.users',
		key: MENU_TYPE.Users,
		img: 'settings/imgs/root/account.svg',
		description: 'Manage users within the Olares cluster.'
	},
	[MENU_TYPE.Appearance]: {
		label: 'home_menus.appearance',
		key: MENU_TYPE.Appearance,
		img: 'settings/imgs/root/appearance.svg'
	},
	[MENU_TYPE.Application]: {
		label: 'home_menus.application',
		key: MENU_TYPE.Application,
		img: 'settings/imgs/root/application.svg'
	},
	[MENU_TYPE.Integration]: {
		label: 'home_menus.integration',
		key: MENU_TYPE.Integration,
		img: 'settings/imgs/root/integration.svg',
		title: 'Link Your Accounts & Data',
		description:
			'Add your accounts to access all your personal data in one place'
	},
	[MENU_TYPE.VPN]: {
		label: 'home_menus.vpn',
		key: MENU_TYPE.VPN,
		img: 'settings/imgs/root/vpn.svg',
		description: 'Tailor VPN access to your specific needs.'
	},
	[MENU_TYPE.Network]: {
		label: 'home_menus.network',
		key: MENU_TYPE.Network,
		img: 'settings/imgs/root/network.svg',
		description: 'Manage the network connections for your Olares system.'
	},
	[MENU_TYPE.GPU]: {
		label: 'home_menus.gpu',
		key: MENU_TYPE.GPU,
		img: 'settings/imgs/root/gpu.svg',
		description:
			'Manage and monitor all available GPU resources across your nodes.'
	},
	[MENU_TYPE.Backup]: {
		label: 'home_menus.backup',
		key: MENU_TYPE.Backup,
		img: 'settings/imgs/root/backup.svg'
	},
	[MENU_TYPE.Restore]: {
		label: 'home_menus.restore',
		key: MENU_TYPE.Restore,
		img: 'settings/imgs/root/restore.svg'
	},
	[MENU_TYPE.Developer]: {
		label: 'home_menus.developer',
		key: MENU_TYPE.Developer,
		img: 'settings/imgs/root/developer.svg'
	},
	[MENU_TYPE.Video]: {
		label: 'home_menus.video',
		key: MENU_TYPE.Video,
		img: 'settings/imgs/root/video.svg'
	},
	[MENU_TYPE.Search]: {
		label: 'home_menus.search',
		key: MENU_TYPE.Search,
		img: 'settings/imgs/root/search.svg'
	}
};

export const useMenuItem = (key: MENU_TYPE) => {
	const item = BaseMenuItems[key];
	if (!item) return undefined;
	return {
		...item,
		label: i18n.global.t(item.label),
		description: item.description ? i18n.global.t(item.description) : ''
	};
};

export enum OLARES_ROLE {
	OWNER = 'owner',
	ADMIN = 'admin',
	NORMAL = 'normal'
}

export function getRoleName(role: string) {
	if (role === OLARES_ROLE.NORMAL) {
		return i18n.global.t('members');
	} else if (role === OLARES_ROLE.ADMIN) {
		return i18n.global.t('admin');
	} else if (role === OLARES_ROLE.OWNER) {
		return i18n.global.t('super_admin');
	}
	return role;
}

export interface EntrancePolicy {
	one_time: boolean;
	policy: string;
	uri: string;
	valid_duration: number;
}

export const locationOptions = computed(() => {
	return [
		{
			label: i18n.global.t('server_location.olares_space'),
			value: BackupLocation.TerminusCloud
		},
		{
			label: i18n.global.t('server_location.aws_s3'),
			value: BackupLocation.S3
		}
	];
});

export const frequencyOptions = computed(() => {
	return [
		{
			label: i18n.global.t('frequencys.every_day'),
			value: BackupFrequency.Daily
		},
		{
			label: i18n.global.t('frequencys.every_week'),
			value: BackupFrequency.Weekly
		},
		{
			label: i18n.global.t('frequencys.every_month'),
			value: BackupFrequency.Monthly
		}
	];
});

export const resourcesOptions = computed(() => {
	return [
		{
			label: i18n.global.t('Backup Files'),
			value: BackupResourcesType.files
		},
		{
			label: i18n.global.t('Backup App'),
			value: BackupResourcesType.app
		}
	];
});

export const weekOption = computed(() => {
	return [
		{
			label: i18n.global.t('week.sunday'),
			value: 7
		},
		{
			label: i18n.global.t('week.monday'),
			value: 1
		},
		{
			label: i18n.global.t('week.tuesday'),
			value: 2
		},
		{
			label: i18n.global.t('week.wednesday'),
			value: 3
		},
		{
			label: i18n.global.t('week.thursday'),
			value: 4
		},
		{
			label: i18n.global.t('week.friday'),
			value: 5
		},
		{
			label: i18n.global.t('week.saturday'),
			value: 6
		}
	];
});

export const monthOption = computed(() => {
	const monthOptions: SelectorProps[] = [];
	for (let i = 0; i < 31; i++) {
		const index = i + 1;
		monthOptions.push({
			label: i18n.global.t('monthly_day', { day: index }),
			value: index.toString()
		});
	}
	return monthOptions;
});

export enum FACTOR_MODEL {
	One = 'one_factor',
	Two = 'two_factor',
	Public = 'public',
	System = 'system'
}

export const factorModelOptions = () => {
	return [
		{
			label: i18n.global.t('factor.system'),
			value: FACTOR_MODEL.System,
			enable: true
		},
		{
			label: i18n.global.t('factor.one_factor'),
			value: FACTOR_MODEL.One,
			enable: true
		},
		{
			label: i18n.global.t('factor.two_factor'),
			value: FACTOR_MODEL.Two,
			enable: true
		},
		{
			label: i18n.global.t('factor.none'),
			value: FACTOR_MODEL.Public,
			enable: true
		}
	];
};

export enum AUTH_LEVEL {
	Private = 'private',
	Public = 'public',
	Internal = 'internal'
}

export const authLevelOptions = () => {
	return [
		{
			label: i18n.global.t('private'),
			value: AUTH_LEVEL.Private,
			enable: true
		},
		{
			label: i18n.global.t('public'),
			value: AUTH_LEVEL.Public,
			enable: true
		},
		{
			label: i18n.global.t('Internal'),
			value: AUTH_LEVEL.Internal,
			enable: true
		}
	];
};

export interface Secret {
	Key: string; //  key name
	Workspace: string; // secret workspace name
}

export interface ThirdPartyAccountInterface {
	name: string;
	icon: string;
	email: string;
}

export const ThirdPartyAccountList: ThirdPartyAccountInterface[] = [
	{
		name: 'Google',
		icon: 'Google',
		email: ''
	},
	{
		name: 'iCloud',
		icon: 'iCloud',
		email: ''
	},
	{
		name: 'Yahoo',
		icon: 'Yahoo',
		email: ''
	},
	{
		name: 'Aol',
		icon: 'Aol',
		email: ''
	}
];

export enum MODEL_STATUS {
	//server
	running = 'running',
	installed = 'installed',
	noInstalled = 'no_installed',
	installing = 'installing',

	//local
	pending = 'pending',
	loading = 'loading',
	stopping = 'stopping',
	uninstalling = 'uninstalling',
	installable = 'installable'
}

export interface DifyModelInfo {
	id: string;
	file_name: string;
	progress: number;
	status: MODEL_STATUS;
	folder_path: string;
	model: ModelInfo;
	type: string;
}

export interface ModelMetadata {
	author: string;
	cover: string;
	size: number;
	tags: string[];
}

export interface ModelRuntimeParams {
	frequency_penalty: number;
	max_tokens: number;
	presence_penalty: number;
	stop: any[];
	stream: boolean;
	temperature: number;
	top_p: number;
}

export interface ModelSettingParams {
	ctx_len: number;
	prompt_template: string;
}

export interface ModelInfo {
	object: string;
	format: string;
	source_url: string;
	id: string;
	name: string;
	created: number;
	description: string;
	settings: ModelSettingParams;
	parameters: ModelRuntimeParams;
	metadata: ModelMetadata;
	engine: string;
}

export enum QR_STATUS {
	NORMAL,
	EXPIRED,
	SUCCESSFUL
}

export function getSecondLevelDomain() {
	const domainParts = window.location.hostname.split('.');
	if (domainParts.length >= 3) {
		return domainParts.slice(-3).join('.');
	} else {
		return window.location.hostname;
	}
}

export const getApplicationStatus = (status: string) => {
	let realStatus = '';
	switch (status) {
		case APP_STATUS.UNINSTALL.COMPLETED:
			realStatus = i18n.global.t('app.get');
			break;
		case APP_STATUS.PENDING.DEFAULT:
			realStatus = i18n.global.t('app.pending');
			break;
		case APP_STATUS.INSTALL.DEFAULT:
		case APP_STATUS.DOWNLOAD.DEFAULT:
			realStatus = i18n.global.t('app.installing');
			break;
		case APP_STATUS.MODEL.INSTALLED:
			realStatus = i18n.global.t('app.installed');
			break;
		case APP_STATUS.STOP.COMPLETED:
			realStatus = i18n.global.t('app.stopped');
			break;
		case APP_STATUS.RESUME.DEFAULT:
			realStatus = i18n.global.t('app.resuming');
			break;
		case APP_STATUS.RUNNING:
			realStatus = i18n.global.t('app.running');
			break;
		case APP_STATUS.UNINSTALL.DEFAULT:
			realStatus = i18n.global.t('app.uninstalling');
			break;
		case APP_STATUS.UPGRADE.DEFAULT:
			realStatus = i18n.global.t('app.updating');
			break;
		case APP_STATUS.INITIALIZE.DEFAULT:
			realStatus = i18n.global.t('app.initializing');
			break;
		case APP_STATUS.ENV.APPLYING:
			realStatus = i18n.global.t('app.applying');
			break;
		case APP_STATUS.PENDING.CANCELING:
		case APP_STATUS.DOWNLOAD.CANCELING:
		case APP_STATUS.INITIALIZE.CANCELING:
		case APP_STATUS.INSTALL.CANCELING:
		case APP_STATUS.UPGRADE.CANCELING:
		case APP_STATUS.RESUME.CANCELING:
		case APP_STATUS.ENV.CANCELING:
			realStatus = i18n.global.t('app.canceling');
			break;
		case APP_STATUS.PENDING.CANCEL_FAILED:
		case APP_STATUS.DOWNLOAD.CANCEL_FAILED:
		case APP_STATUS.INSTALL.CANCEL_FAILED:
		case APP_STATUS.ENV.CANCEL_FAILED:
		case APP_STATUS.DOWNLOAD.FAILED:
		case APP_STATUS.INSTALL.FAILED:
		case APP_STATUS.UNINSTALL.FAILED:
		case APP_STATUS.UPGRADE.FAILED:
		case APP_STATUS.RESUME.FAILED:
		case APP_STATUS.STOP.FAILED:
		case APP_STATUS.ENV.APPLY_FAILED:
			realStatus = i18n.global.t('app.failed');
			break;
		default:
			break;
	}

	return realStatus;
};

export const getEntranceStatus = (status: ENTRANCE_STATUS) => {
	let realStatus = '';
	switch (status) {
		case ENTRANCE_STATUS.STOPPED:
			realStatus = i18n.global.t('app.stopped');
			break;
		case ENTRANCE_STATUS.NOT_READY:
			realStatus = i18n.global.t('app.not_ready');
			break;
		case ENTRANCE_STATUS.RUNNING:
			realStatus = i18n.global.t('app.running');
			break;
		default:
			break;
	}

	return realStatus;
};

export enum ReverseProxyMode {
	NoNeed = 1,
	CloudFlare = 2,
	OlaresTunnel = 3,
	SelfBuiltFrp = 4
}

export const reverseProxyOptions = () => {
	return [
		// {
		// 	label: 'No need (IP Direct)',
		// 	value: ReverseProxyMode.NoNeed,
		// 	enable: true
		// },
		{
			label: 'Cloudflare Tunnel',
			value: ReverseProxyMode.CloudFlare,
			enable: true
		},
		{
			label: 'Olares Tunnel',
			value: ReverseProxyMode.OlaresTunnel,
			enable: true
		},
		{
			label: 'Self-built FRP',
			value: ReverseProxyMode.SelfBuiltFrp,
			enable: true
		}
	];
};

export const frpAuthMethod = () => {
	return [
		{
			label: i18n.global.t('None'),
			value: '',
			enable: true
		},
		{
			label: i18n.global.t('Token'),
			value: 'token',
			enable: true
		}
	];
};

export interface OlaresTunnelInterface {
	frp_server: string;
	frp_port: number;
	frp_auth_method: string;
	frp_auth_token: string;
}

export const olaresTunnelDefaultValue = {
	frp_port: 0,
	frp_auth_method: 'jws',
	frp_auth_token: ''
};

export enum VRAMMode {
	Single = '0',
	MemorySlicing = '1',
	TimeSlicing = '2'
}

export enum VRAMModeLabel {
	'App Exclusive',
	'Memory Slicing',
	'Time Slicing'
}

export const VRAMModeOptions = () => {
	return [
		{
			label: i18n.global.t(VRAMModeLabel[0]),
			value: VRAMMode.Single,
			enable: true,
			description: i18n.global.t(
				'Select one application to have dedicated access to this GPU.'
			),
			subTitle: i18n.global.t('Select exclusive app'),
			subDesc: ''
		},
		{
			label: i18n.global.t(VRAMModeLabel[1]),
			value: VRAMMode.MemorySlicing,
			enable: true,
			description: i18n.global.t(
				'Assign a dedicated amount of VRAM to specific applications'
			),
			subTitle: i18n.global.t('Allocate VRAM'),
			subDesc: i18n.global.t(
				'Select your target application and assign VRAM to it'
			)
		},
		{
			label: i18n.global.t(VRAMModeLabel[2]),
			value: VRAMMode.TimeSlicing,
			enable: true,
			description: i18n.global.t(
				"The GPU's  power is shared among multiple applications."
			),
			subTitle: i18n.global.t('Pin application'),
			subDesc: i18n.global.t('Bind an application to this GPU')
		}
	];
};

// export const VRAMTimeSlicingOperations = (
// 	otherVRAMs: { id: string; node: string; name: string }[]
// ) => {
// 	const operations = [
// 		{
// 			label: i18n.global.t('app.stop'),
// 			value: '-1',
// 			enable: true
// 		}
// 	];
// 	if (otherVRAMs.length <= 0) {
// 		return operations;
// 	}

// 	otherVRAMs.forEach((e) => {
// 		operations.push({
// 			label: `[${e.node}]-[${e.name}]`,
// 			value: e.id,
// 			enable: true
// 		});
// 	});

// 	return operations;
// };

export interface HostItem {
	ip: string;
	host: string;
}

export interface VersionInfo {
	current_version: string;
	new_version: string;
	is_new: boolean;

	// current_version: string;
	// new_version: string;
	// is_new: boolean;
	upgradeableVersion: string;
	wizardUrl: string;
	cliUrl: string;
	needRestart?: boolean;
}

export enum UpgradeStatus {
	Running = 'running',
	Completed = 'completed',
	InProgress = 'in-progress',
	Failed = 'failed'
}

export const formatDefaultHost = (host: string) => {
	if (!host.endsWith('/')) {
		return host;
	}
	return host.substring(0, host.length - 1);
};

export const olaresSpaceUrl = formatDefaultHost(
	process.env.OLARES_SPACE_URL || 'https://cloud-api.bttcdn.com'
);

export interface OlaresSpaceRegion {
	cloudName: string;
	regionId: string;
}

export enum BackupResourcesType {
	app = 'app',
	files = 'file'
}

export const SupportBackupAppList = ['wise'];

export enum BackupLocationType {
	fileSystem = 'filesystem',
	awsS3 = 'awss3',
	tencentCloud = 'tencentcloud',
	space = 'space'
}

export function getBackupLocationTypeByIntegrationAccount(
	account: IntegrationAccountMiniData
): BackupLocationType | null {
	switch (account.type) {
		case AccountType.Space:
			return BackupLocationType.space;
		case AccountType.Tencent:
			return BackupLocationType.tencentCloud;
		case AccountType.AWSS3:
			return BackupLocationType.awsS3;
		default:
			return null;
	}
}

export function getBackupIconByLocation(type: BackupLocationType) {
	switch (type) {
		case BackupLocationType.awsS3:
			return getRequireImage('integration/aws.svg');
		case BackupLocationType.fileSystem:
			return '/img/folder-default.svg';
		case BackupLocationType.space:
			return getRequireImage('integration/space.svg');
		case BackupLocationType.tencentCloud:
			return getRequireImage('integration/tencent.svg');
	}
}

export function getBackupStatusImg(status: BackupStatus) {
	switch (status) {
		case BackupStatus.pending:
			return getRequireImage('status/pending.svg');
		case BackupStatus.running:
			return getRequireImage('status/running.svg');
		case BackupStatus.completed:
			return getRequireImage('status/completed.svg');
		case BackupStatus.failed:
			return getRequireImage('status/failed.svg');
		case BackupStatus.canceled:
			return getRequireImage('status/canceled.svg');
		case BackupStatus.rejected:
			return getRequireImage('status/rejected.svg');
		default:
			return '';
	}
}

export function getRestoreColorClass(
	status: BackupStatus,
	classPrefix = 'text'
) {
	switch (status) {
		case BackupStatus.pending:
		case BackupStatus.canceled:
			return `${classPrefix}-ink-3`;
		case BackupStatus.running:
			return `${classPrefix}-info`;
		case BackupStatus.completed:
			return `${classPrefix}-positive`;
		case BackupStatus.failed:
		case BackupStatus.rejected:
			return `${classPrefix}-negative`;
		default:
			return '';
	}
}

export interface BackupPlan {
	backupType: BackupResourcesType;
	backupAppTypeName: string;
	id: string;
	name: string;
	size: string;
	restoreSize: string;
	path: string;
	progress: number;
	nextBackupTimestamp: number;
	location: BackupLocationType;
	locationConfigName: string;
	status: BackupStatus;
	createAt: number;
}

export interface BackupCreat {
	name: string;
	path: string;
	location: BackupLocationType;
	locationConfig:
		| SpaceLocationConfig
		| FileSystemLocationConfig
		| BaseLocationConfig;
	backupPolicy: BackupPolicy;
}

export function createLocationConfig(config: any, region?: OlaresSpaceRegion) {
	let locationConfig;
	if (config.type === BackupLocationType.fileSystem) {
		locationConfig = {
			path: config.data.decodePath
		};
	} else if (config.type === BackupLocationType.space && region) {
		locationConfig = {
			name: config.data.name,
			cloudName: region.cloudName,
			regionId: region.regionId
		};
	} else if (
		config.type === BackupLocationType.tencentCloud ||
		config.type === BackupLocationType.awsS3
	) {
		locationConfig = { name: config.data.name };
	}

	console.log(config);
	console.log(locationConfig);
	return locationConfig;
}

export interface SpaceLocationConfig {
	name: string;
	cloudName: string;
	regionId: string;
}

export interface FileSystemLocationConfig {
	path: string;
}

//aws tencent
export interface BaseLocationConfig {
	name: string;
}

export enum BackupStatus {
	completed = 'Completed',
	failed = 'Failed',
	running = 'Running',
	pending = 'Pending',
	canceled = 'Canceled',
	rejected = 'Rejected'
}

export const backupOriginsRef = ref([
	DriveType.Drive,
	DriveType.Data,
	DriveType.Cache,
	DriveType.External
]);

export interface BackupPlanDetail {
	id: string;
	name: string;
	path: string;
	backupPolicies: BackupPolicy;
	backupAppTypeName: string;
	backupType: string;
	backupSize: string;
	restoreSize: string;
}

export interface BackupPolicy {
	enabled: boolean;
	snapshotFrequency: BackupFrequency;
	timesOfDay: string;
	timespanOfDay?: string;
	dayOfWeek?: number;
	dateOfMonth?: number;
}

export function createPolicy(policy: BackupPolicy): BackupPolicy {
	const { enabled, snapshotFrequency, timesOfDay, dayOfWeek, dateOfMonth } =
		policy;
	console.log(timesOfDay);
	const formatTime = timeToTimeStamp(timesOfDay);
	console.log(formatTime);
	switch (policy.snapshotFrequency) {
		case BackupFrequency.Daily:
			return { enabled, snapshotFrequency, timesOfDay: formatTime };
		case BackupFrequency.Weekly:
			return { enabled, snapshotFrequency, timesOfDay: formatTime, dayOfWeek };
		case BackupFrequency.Monthly:
			return {
				enabled,
				snapshotFrequency,
				timesOfDay: formatTime,
				dateOfMonth: Number(dateOfMonth)
			};
		default:
			return policy;
	}
}

export interface RestorePlan {
	id: string;
	name: string;
	path: string;
	createAt: number;
	endAt: number;
	snapshotTime: number;
	progress: number;
	status: BackupStatus;
	backupAppTypeName: string;
	backupType: string;
}

export interface RestorePlanDetail {
	name: string;
	restorePath: string;
	backupPath: string;
	progress: number;
	snapshotTime: string;
	status: BackupStatus;
	message: string;
	backupAppTypeName: string;
	backupType: string;
}

export const BackupPathOrigins = [
	DriveType.Drive,
	DriveType.Data,
	DriveType.Cache,
	DriveType.External
];

export interface SnapshotInfo {
	id: string;
	createAt: number;
	size: number;
}

export interface RestoreSnapshotInfo extends SnapshotInfo {
	backupPath: string;
}

export interface BackupSnapshot extends SnapshotInfo {
	status: BackupStatus;
	progress: number;
}

export interface BackupSnapshotDetail {
	id: string;
	size: string;
	snapshotType: string;
	progress: number;
	status: BackupStatus;
	message: string;
}

export enum SnapshotType {
	Fully = 'Fully',
	Incremental = 'Incremental',
	Unknown = 'Unknown'
}

export interface SelectorProps {
	label: string;
	value: string | number;
	disable?: boolean;
	hideLabel?: boolean;
	titleClass?: string;
}

export interface ApplicationSelectorState extends SelectorProps {
	app: TerminusApp;
}

export function formatFilePath(filePath: any) {
	return {
		isDir: filePath.isDir,
		path: filePath.path,
		driveType: filePath.driveType,
		param: filePath.param,
		decodePath: filePath.path ? decodeURIComponent(filePath.path) : ''
	};
}

export interface BackupMessage {
	backupId: string;
	id: string;
	message: string;
	progress: number;
	status: BackupStatus;
	size?: string;
	restoreSize?: string;
	totalSize?: string;
	type: string;
}

export interface RestoreMessage {
	id: string;
	progress: number;
	status: BackupStatus;
	message: string;
	endat?: number;
	type: string;
}

export interface AppEnvResponse {
	missingValues: BaseEnv[];
	invalidValues: BaseEnv[];
	missingRefs: BaseEnv[];
}

export interface BaseEnv {
	envName: string;
	value?: string;
	default?: string;
	editable?: boolean;
	type?: 'int' | 'url' | 'ip' | 'domain' | string;
	required?: boolean;
	options?: EnvOption[];
	remoteOptions?: string;
	regex?: string;
	description?: string;
	applyOnChange?: boolean;
	valueFrom?: {
		envName: string;
		status: 'synced' | string;
	};
	//local
	right?: boolean;
	error?: string;
}

export interface EnvOption {
	title: string;
	value: string;
}

export interface UpdateEnvItem {
	envName: string;
	value: string;
}
export type UpdateEnvBody = UpdateEnvItem[];

export interface CloneEntrance {
	name: string;
	title: string;
	message?: string;
}
export interface AppCloneInfoResponse {
	missingValues: CloneEntrance[];
	invalidValues: CloneEntrance[];
	titleValidation: {
		isValid: boolean;
		message: string;
		title: string;
	};
}

export interface EnvDetail extends BaseEnv {
	referencedBy?: Array<{
		appName: string;
		appOwner: string;
		namespace: string;
	}>;
}
