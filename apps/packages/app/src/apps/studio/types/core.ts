export interface ApplicationInfo {
	id: number;
	appName: string;
	devEnv: string;
	createTime: string; //"2023-12-11 09:00:00"
	updateTime: string;

	chart: string;
	entrance: string;
	ide: string;
	podContainers: string;
	state: APP_STATUS;
	title: string;
	description: string;
	reason?: string;
	hideErrorFooter?: boolean;
}

export interface OpenApplicationConfig {
	appid: string;
	path: string;
}

export interface MenuLabelType {
	label: string;
	key: string;
	icon: string;
	muted?: boolean;
	children?: MenuLabelType[];
}

export interface DocumentType {
	id: string | number;
	message: string;
	link: string;
}

export type FilesSelectType = {
	label: string;
	icon: string;
	path: string;
	expandable: boolean;
	selectable: boolean;
	children: any;
	isDir: boolean;
	lazy?: boolean;
};

export type FilesCodeType = {
	code: string;
	lang: string;
	name: string;
};

export enum APP_STATUS {
	EMPTY = 'empty',
	DEPLOYED = 'deployed',
	DEPLOYING = 'deploying',
	UNDEPLOY = 'undeploy',
	ABNORMAL = 'abnormal'
}

export const app_status_style = {
	empty: {
		color: '#ADADAD',
		notify_bg: '#ADADAD'
	},
	deployed: {
		color: '#29CC5F',
		notify_bg: '#29CC5F'
	},
	deploying: {
		color: '#FEBE01',
		notify_bg: '#FEBE01'
	},
	undeploy: {
		color: '#ADADAD',
		notify_bg: '#ADADAD'
	},
	abnormal: {
		color: '#FF4D4D',
		notify_bg: '#FFF2F2'
	}
};

export enum VENDOR {
	NVIDIA = 'nvidia',
	AMD = 'amd',
	INTEL = 'intel'
}

export interface ContainerType {
	image: string;
	port: number;
	startCmd?: string;
	startCmdArgs?: string;
}

export interface EnvType {
	[key: string]: string;
}

export interface CreateWithOneDockerConfig {
	id?: number;
	name: string;
	container: ContainerType;
	requiredCpu: string;
	limitedCpu?: string;
	requiredMemory: string;
	limitedMemory?: string;
	requiredGpu: boolean;
	needPg: boolean;
	needRedis: boolean;
	gpuVendor: VENDOR;
	env: EnvType;
	mounts: EnvType;
}

export interface CreateWithOneImageConfig {
	name: string;
	devEnv: string;
	requiredCpu: string;
	requiredMemory: string;
	requiredDisk: string;
	ports?: number[];
	requiredGpu: boolean;
	gpuVendor?: VENDOR.NVIDIA;
}

export interface VariableType {
	id?: number;
	key: string;
	value: string;
}

export enum OPERATE_ACTION {
	ADD_FOLDER = 'ADD_FOLDER',
	ADD_FILE = 'ADD_FILE',
	RENAME = 'RENAME',
	DELETE = 'DELETE'
}

export interface SortingType {
	asc: boolean;
	by: string;
}
export interface FileItem {
	content: string;
	extension: string;
	isDir: boolean;
	isSymlink: boolean;
	mode: number;
	modified: string;
	name: string;
	path: string;
	size: number;
	type: string;
	items?: FileItem[];
	sorting?: SortingType;
}

export enum APP_INSTALL_STATE {
	PENDING = 'pending',
	DOWNLOADING = 'downloading',
	PROCESSING = 'processing',
	CANCELED = 'canceled',
	FAILED = 'failed',
	COMPLETED = 'completed',

	INSTALLING = 'installing',
	UPGRADING = 'upgrading',
	UNINSTALLING = 'uninstalling',
	RESUMING = 'resuming',
	RESUMED = 'resumed',
	CANCELING = 'canceling',
	STOPPING = 'stopping',
	RUNNING = 'running'
}

export const app_install_status_style = {
	pending: {
		color: '#5C5C5C',
		icon: 'sym_r_download'
	},
	downloading: {
		color: '#FEBE01',
		icon: 'sym_r_downloading'
	},
	processing: {
		color: '#3377FF',
		icon: 'sym_r_download_for_offline'
	},
	canceled: {
		color: '#ADADAD',
		icon: 'sym_r_block'
	},
	failed: {
		color: '#FF4D4D',
		icon: 'sym_r_file_download_off'
	},
	completed: {
		color: '#29CC5F',
		icon: 'sym_r_move_to_inbox'
	}
};

export interface AppState {
	appName: string;
	appNamespace: string;
	appOwner: string;
	creationTimestamp: string;
	message: string;
	opType: string;
	resourceType: string;
	source: string;
	state: APP_INSTALL_STATE;
}
