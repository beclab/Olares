export interface GPUPageRequest {
	pageSize: number;
	pageNo: number;
	sort: 'ASC' | 'DESC';
	sortField: string;
}
export interface GraphicsListParams {
	filters: {
		nodeName?: string;
		type?: string;
		uid?: boolean;
	};
	pageRequest?: Partial<GPUPageRequest>;
}

export interface Graphics {
	uuid: string;
	nodeName: string;
	type: string;
	vgpuUsed: number;
	vgpuTotal: number;
	coreUsed: number;
	coreTotal: number;
	memoryUsed: number;
	memoryTotal: number;
	nodeUid: string;
	health: boolean;
	mode: string;
	temperature?: number;
	powerUsage?: number;
	lastUpdated?: string;
	power: number;
	powerLimit: number;
	status?: 'online' | 'offline' | 'error';
}

export interface GraphicsListResponse {
	list: Graphics[];
}

export type GpuUtilization = {
	vgpu: number;
	core: number;
	memory: number;
};

export type TaskStatus = 'success' | 'running' | 'failed' | 'pending';

export interface TaskListParams {
	filters: {
		name?: string;
		nodeName?: string;
		status?: TaskStatus;
		deviceId?: string;
	};
	pageRequest: Partial<GPUPageRequest>;
}

export interface TaskItem {
	name: string;
	status: TaskStatus;
	appName: string;
	nodeName: string;
	allocatedDevices: number;
	allocatedCores: number;
	allocatedMem: number;
	type: 'NVIDIA' | 'AMD' | 'Intel';
	createTime: string;
	startTime: string | null;
	endTime: string | null;
	podUid: string;
	nodeUid: string;
	resourcePool: string;
	flavor: string;
	priority: string;
	namespace: string;
	deviceIds: string[];
	deviceShareModes: string[];
	devicesMemUtilized: number[];
	devicesCoreUtilizedPercent: number[];
}

export type TaskItemWithShareMode = TaskItem & {
	shareMode: ShareMode;
};

export interface TaskListResponse {
	items: TaskItem[];
}

export interface GraphicsDetailsParams {
	uid: string;
}

export interface GraphicsDetailsResponse
	extends Omit<
		Graphics,
		'status' | 'temperature' | 'powerUsage' | 'lastUpdated'
	> {
	uuid: string;
	nodeName: string;
	type: string;
	vgpuUsed: number;
	vgpuTotal: number;
	coreUsed: number;
	coreTotal: number;
	memoryUsed: number;
	memoryTotal: number;
	nodeUid: string;
	health: boolean;
	mode: 'hami-core' | 'time-slice' | 'mps';
}

export interface InstantVectorParams {
	query: string;
}

export interface InstantVector {
	metric: {
		[key: string]: string;
	};
	value: number;
	timestamp: string;
}

export interface InstantVectorResponse {
	data: InstantVector[];
}

export interface RangeVectorParams {
	range: {
		start: string;
		end: string;
		step: string;
	};
	query: string;
}

export interface RangeVector {
	metric: {
		device_no: 'nvidia0';
		driver_version: '550.144.03';
	};
	values: {
		value: 8.55;
		timestamp: '1744603402000';
	}[];
}

export interface RangeVectorResponse {
	data: RangeVector[];
}

export interface TaskDetailParams {
	name: string;
	podUid: string;
}

export interface TaskDetailResponse {
	name: string;
	status: TaskStatus;
	appName: string;
	nodeName: string;
	allocatedDevices: number;
	allocatedCores: number;
	allocatedMem: number;
	type: 'NVIDIA' | 'AMD' | 'Intel';
	createTime: string;
	startTime: string;
	endTime: string;
	podUid: string;
	nodeUid: string;
	resourcePool: string;
	flavor: string;
	priority: string;
	namespace: string;
	deviceIds: string[];
}

export interface GPUNode {
	ip: string;
	isSchedulable: boolean;
	isReady: boolean;
	type: string[];
	vgpuUsed: number;
	vgpuTotal: number;
	coreUsed: number;
	coreTotal: number;
	memoryUsed: number;
	memoryTotal: number;
	uid: string;
	name: string;
	cardCnt: number;
	osImage: string;
	operatingSystem: string;
	kernelVersion: string;
	containerRuntimeVersion: string;
	kubeletVersion: string;
	kubeProxyVersion: string;
	architecture: string;
	creationTimestamp: string;
	isExternal?: boolean;
}

export interface GPUNodeList {
	list: GPUNode[];
}

export type ShareMode = '0' | '1' | '2';

export interface SettingsGPUItem {
	nodeName: string;
	id: string;
	index: number;
	count: number;
	devmem: number;
	devcore: number;
	type: string;
	mode: string;
	health: boolean;
	sharemode: ShareMode;
	apps: Array<{
		appName: string;
	}>;
}

export interface SettingsGPUResponse {
	code: number;
	message: string | null;
	data: SettingsGPUItem[];
}
