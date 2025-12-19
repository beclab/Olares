interface MetadataInNodes {
	name: string;
	uid: string;
	resourceVersion: number;
	creationTimestamp: string;
	labels: {
		[key: string]: string;
	};
}

interface Spec {
	containers: Container[];
	dnsPolicy: string;
	enableServiceLinks: true;
	hostname: string;
	nodeName: string;
	preemptionPolicy: string;
	priority: 0;
	restartPolicy: string;
	schedulerName: string;
	securityContext: any;
	serviceAccount: string;
	serviceAccountName: string;
	subdomain: string;
	terminationGracePeriodSeconds: number;
	tolerations: null[];
	volumes: null[];
}

interface Status {
	conditions: null[];
	containerStatuses: null[];
	hostIP: string;
	phase: string;
	podIP: string;
	podIPs: null[];
	qosClass: string;
	startTime: string;
}

export type MetadataInPods = MetadataInNodes & { namespace: string };

export interface Container {
	env: null[];
	image: string;
	imagePullPolicy: string;
	name: string;
	ports: null[];
	resources: any;
	terminationMessagePath: string;
	terminationMessagePolicy: string;
	volumeMounts: null[];
}
export interface PodItem {
	metadata: MetadataInPods;
	spec: Spec;
	status: Status;
}

export type PodItemResponse = PodItem & {
	apiVersion: string;
	kind: string;
};

export interface MonitoringResponseData {
	resultType: string;
	result: {
		avg_value: string;
		currency_unit: string;
		fee: string;
		max_value: string;
		min_value: string;
		resource_unit: string;
		sum_value: string;
		values: [number, string][];
	}[];
}
export interface MonitoringResponse {
	results: {
		metric_name: string;
		data: MonitoringResponseData[];
	}[];
}

export type PodMonitoringParamAll = PodsParam;

export interface PodDetailParam {
	namespace: string;
	podName: string;
}

export interface PodsParam {
	metrics_filter?: string;
	resources_filter?: string;
	start?: string | number;
	end?: string | number;
	step?: string;
	time?: string;
	sort_metric?: string;
	// 'desc' | 'asc'
	sort_type?: string;
	page?: number;
	limit?: number;
	sortBy?: string;
	// can not find
	times?: number;
	cluster?: string;
	name?: string;
}

export interface ResourcesResponse {
	items: PodItem[];
	totalItems: number;
}

export type ContainersMonitoringParamAll = PodDetailParam &
	PodsParam & { container: string };

export interface kubesphereStatusItem {
	healthyBackends: number;
	label: { [key: string]: string };
	name: string;
	namespace: string;
	selfLink: string;
	startedAt: string;
	totalBackends: number;
}
export interface ComponenthealthResponse {
	kubesphereStatus: kubesphereStatusItem[];
	nodeStatus: {
		healthyNodes: number;
		totalNodes: number;
	};
}

export type NamespacesResponse = MonitoringResponse & {
	total_item: number;
	total_page: number;
	page: number;
};

export type NamespacesParam = PodsParam & { type: string };

export const IP_METHOD_OPTION = {
	auto: 'IP_METHOD_AUTO',
	manual: 'IP_METHOD_MANUAL',
	none: 'IP_METHOD_NONE'
};

type IP_METHOD = keyof typeof IP_METHOD_OPTION;

export interface SystemIFSItem {
	iface: string;
	ip: string;
	isHostIp: boolean;
	isWifi: boolean;
	mtu: number;
	internetConnected: boolean;
	ipv4Gateway: string;
	ipv6Gateway: string;
	ipv4DNS: string;
	ipv6DNS: string;
	ipv6Address: string;
	ipv4Mask: string;
	method: IP_METHOD;
	ipv6Connectivity: boolean;
	rxRate: number;
	txRate: number;
}

export type SystemIFSResponse = SystemIFSItem[];

export interface SystemIFSParams {
	testConnectivity?: boolean;
}

export interface SystemFanData {
	gpu_fan_speed: number;
	gpu_temperature: number;
	cpu_fan_speed: number;
	cpu_temperature: number;
}

export interface SystemFanResponse {
	code: number;
	data: SystemFanData;
	message: string;
}
