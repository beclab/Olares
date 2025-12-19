export interface MetadataInNodes {
	name: string;
	uid: string;
	resourceVersion: number;
	creationTimestamp: string;
	labels: {
		[key: string]: string;
	};
	annotations: {
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

export interface MonitoringItem {
	metric_name: string;
	data: MonitoringResponseData[];
}
export interface MonitoringResponse {
	results: MonitoringItem[];
}

export type PodMonitoringParamAll = PodDetailParam & PodsParam;

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

/**
 * node detail
 */

interface NodeSec {
	podCIDR: string;
	podCIDRs: string[];
}

interface NodeStatus {
	addresses: any;
	allocatable: any;
	capacity: any;
	conditions: any;
	daemonEndpoints: any;
	images: any;
	nodeInfo: {
		architecture: 'amd64';
		bootID: '5eae79e5-f868-475e-a730-d4b21408824c';
		containerRuntimeVersion: 'docker://20.10.8';
		kernelVersion: '3.10.0-1160.76.1.el7.x86_64';
		kubeProxyVersion: 'v1.22.10';
		kubeletVersion: 'v1.22.10';
		machineID: 'e887fc2c7f8279b206b7f75ba0a2b0bd';
		operatingSystem: 'linux';
		osImage: 'CentOS Linux 7 (Core)';
		systemUUID: 'EC204FD4-118D-8FEC-F6C4-A62F512EE977';
	};
}
export type NodeDetailResponse = MetadataInNodes & NodeSec & NodeStatus;
