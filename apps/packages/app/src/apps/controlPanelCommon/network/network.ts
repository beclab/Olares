import { TerminusStatus } from 'src/services/abstractions/mdns/service';

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

interface Role {
	apiVersion: string;
	kind: string;
}

export type PodItemResponse = PodItem & Role;

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

export type PodMonitoringParamAll = PodDetailParam & PodsParam;

export interface PodDetailParam {
	namespace: string;
	podName?: string;
}

export type PodsParam = {
	metrics_filter?: string;
	resources_filter?: string;
	start?: string | number;
	end?: string | number;
	step?: string;
	time?: string;
	sort_metric?: string;
	// 'desc' | 'asc'
	sort_type?: string;
	// can not find
	times?: number;
	cluster?: string;
	name?: string;
	fieldSelector?: string;
	labelSelector?: string;
} & Pagination;

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

export type NamespacesParam = PodsParam & { type?: string };

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

export interface Pagination {
	limit?: number;
	page?: number;
	sortBy?: string; //createTime
}

export type UsersParam = {
	globalrole?: string;
	email?: string;
	user?: string;
} & Pagination;

export interface PasswordChangeParams {
	password: string;
	rePassword: string;
}

export interface PasswordChangeResponse {
	message: string;
}

export interface GlobalrolesParams {
	sortBy?: string;
	limit?: number;
	annotation?: string;
}

interface Rule {
	apiGroups: string[];
	resources: string[];
	verbs: string[];
}
export type GlobalrolesResponse = {
	items: Array<Role & { metadata: MetadataInNodes } & Rule[]>;
	totalItems: number;
};

export interface checkoutUserNameReponse {
	exist: boolean;
}

export interface AppDetailResponse {
	clusterRole: string;
	config: any;
	ksConfig: { [key: string]: any };
	user: {
		email: string;
		globalrole: string;
		grantedClusters: string[];
		lang: string;
		lastLoginTime: string;
		username: string;
		globalRules: { [key: string]: string[] };
	};
	workspaces: string[];
	systemNamespaces: string[];
}

export interface NetworkpoliciesResponse {
	items: Array<{
		metadata: { [key: string]: any };
		spec: { [key: string]: any };
	}>;
}

export interface NetworkpoliciesParam {
	namespace: string;
	name: string;
}

export interface CustomresourcesResponse {
	items: any[];
	totalItems: number;
}

export interface CustomresourceDetailParam {
	version: string;
	kind: string;
}

export type SecretsDataResponse = {
	data: {
		release: string;
	};
	metadata: MetadataInNodes;
	type: string;
} & Role;

export type ServiceaccountsItemResponse = Role & {
	metadata: MetadataInPods;
	secrets: { name: string }[];
};

export type DeploymentsDetailResponse = Role & {
	metadata: MetadataInNodes;
	spec: any;
	status: any;
};

export interface AppListItem {
	id: string;
	name: string;
	namespace: string;
	deployment: string;
	owner: string;
	url: string;
	icon: string;
	title: string;
	target: string;
	entrances: {
		id: string;
		name: string;
		title: string;
		url: string;
		icon: string;
		invisible?: boolean;
		authLevel?: string;
		openMethod?: string;
		state?: string;
		reason?: string;
	}[];
	state: string;
	isSysApp: boolean;
	isClusterScoped: boolean;
}

interface Entrance {
	icon: string;
	id: string;
	name: string;
	title: string;
}

export interface AppWS {
	app: string;
	appid: string;
	entrances: Entrance[];
	icon: string;
	id: string;
	op: string;
	status: string;
	title: string;
	uid: string;
}

export interface AppListResponse {
	code: number;
	message: string | null;
	data: AppListItem[];
}

export interface AppListAllResponse {
	code: number;
	message: string | null;
	data: Record<string, PodItem>;
}

export enum jobType {
	cronjobs,
	jobs
}

export interface SystemStatusResponse {
	code: number;
	message: string | null;
	data: TerminusStatus;
}
