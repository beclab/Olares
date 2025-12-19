import {
	PodsParam,
	MonitoringResponse,
	PodDetailParam,
	PodItemResponse,
	PodMonitoringParamAll,
	ContainersMonitoringParamAll,
	ResourcesResponse,
	ComponenthealthResponse,
	NamespacesResponse,
	NamespacesParam,
	NodeDetailResponse,
	UsersParam,
	GlobalrolesParams,
	GlobalrolesResponse,
	checkoutUserNameReponse,
	AppDetailResponse,
	Pagination,
	NetworkpoliciesResponse,
	NetworkpoliciesParam,
	CustomresourcesResponse,
	CustomresourceDetailParam,
	SecretsDataResponse,
	ServiceaccountsItemResponse,
	DeploymentsDetailResponse,
	jobType
} from './network';
import { AxiosResponse } from 'axios';
import { api } from '../boot/axios';
import { joinSelector } from '../utils';
import {
	MiddlewareListResponse,
	MiddlewarePasswordParams,
	MiddlewarePasswordResponse,
	MiddlewareType
} from './middleware';
import { isEmpty } from 'lodash';
export * from './user-service';

export const getNodeMonitoring = (
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get('/kapis/monitoring.kubesphere.io/v1alpha3/nodes', { params });
};

export const getNodesList = (
	params?: PodsParam
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/nodes', { params });
};

export const getPodsList = (
	params: PodsParam
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/pods', { params });
};

export const getPodDetail = (
	params: PodDetailParam
): Promise<AxiosResponse<PodItemResponse>> => {
	return api.get(
		`/api/v1/namespaces/${params.namespace}/pods/${params.podName}`
	);
};

export const getMetrics = (
	namespace: string,
	params: PodsParam,
	config?: Record<string, any>
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/pods`,
		{
			params,
			...config
		}
	);
};

export const getPodMonitoring = (
	params: PodMonitoringParamAll
): Promise<AxiosResponse<MonitoringResponse>> => {
	const { namespace, podName, ...rest } = params;
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/pods/${podName}`,
		{
			params: rest
		}
	);
};

export const getContainersMonitoring = (
	params: ContainersMonitoringParamAll
): Promise<AxiosResponse<MonitoringResponse>> => {
	const { namespace, podName, container } = params;
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/pods/${podName}/containers/${container}`,
		{
			params
		}
	);
};

export const getContainersLogs = (params: any): Promise<AxiosResponse<any>> => {
	const { namespace, podName, ...rest } = params;
	console.log(rest);
	return api.get(`/api/v1/namespaces/${namespace}/pods/${podName}/log`, {
		params
	});
};

export const getClusterMonitoring = (
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get('/kapis/monitoring.kubesphere.io/v1alpha3/cluster', {
		params
	});
};

export const getComponenthealth = (): Promise<
	AxiosResponse<ComponenthealthResponse>
> => {
	return api.get('kapis/resources.kubesphere.io/v1alpha2/componenthealth');
};

export const getComponentsApiserver = (
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		'/kapis/monitoring.kubesphere.io/v1alpha3/components/apiserver',
		{
			params
		}
	);
};

export const getComponentsScheduler = (
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		'/kapis/monitoring.kubesphere.io/v1alpha3/components/scheduler',
		{
			params
		}
	);
};

export const getNamespaces = (
	params: NamespacesParam,
	config?: Record<string, any>
): Promise<AxiosResponse<NamespacesResponse>> => {
	return api.get('/kapis/monitoring.kubesphere.io/v1alpha3/namespaces', {
		params,
		...config
	});
};

export const getUsers = <T = any>(
	params: UsersParam
): Promise<AxiosResponse<T>> => {
	return api.get('/kapis/iam.kubesphere.io/v1alpha2/users', { params });
};

export const getNodeDetail = (
	node: string,
	params?: PodsParam
): Promise<AxiosResponse<NodeDetailResponse>> => {
	return api.get(`/kapis/resources.kubesphere.io/v1alpha3/nodes/${node}`, {
		params
	});
};

export const getYaml = (
	namespace: string,
	podName: string
): Promise<AxiosResponse<NodeDetailResponse>> => {
	return api.get(`/api/v1/namespaces/${namespace}/pods/${podName}`, {});
};

export const getLoginrecords = (
	params: UsersParam
): Promise<AxiosResponse<any>> => {
	const { user, ...rest } = params;
	return api.get(
		`/kapis/iam.kubesphere.io/v1alpha2/users/${user}/loginrecords`,
		{ params: rest }
	);
};

export const getGlobalroles = (
	params: GlobalrolesParams
): Promise<AxiosResponse<GlobalrolesResponse>> => {
	return api.get('/kapis/iam.kubesphere.io/v1alpha2/globalroles', {
		params
	});
};

export const checkoutUserName = (
	name: string
): Promise<AxiosResponse<checkoutUserNameReponse>> => {
	return api.get(`/kapis/iam.kubesphere.io/v1alpha2/users/${name}`, {
		headers: { 'x-check-exist': true }
	});
};

export const checkEmail = <T = any>(
	params: UsersParam
): Promise<AxiosResponse<T>> => {
	return api.get('/kapis/iam.kubesphere.io/v1alpha2/users', {
		params,
		headers: { 'x-check-exist': true }
	});
};

export const PostUsers = <T = any>(params: any): Promise<AxiosResponse<T>> => {
	return api.post('/kapis/iam.kubesphere.io/v1alpha2/users', params);
};

export const getAppDetail = (): Promise<AxiosResponse<AppDetailResponse>> => {
	return api.get('/capi/app/detail', {});
};

export const getNetworkpolicies = (
	params: Pagination
): Promise<AxiosResponse<NetworkpoliciesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/networkpolicies', {
		params
	});
};

export const getNetworkpoliciesDetail = (
	params: NetworkpoliciesParam
): Promise<AxiosResponse<NetworkpoliciesResponse>> => {
	const { namespace, name } = params;
	return api.get(
		`/apis/networking.k8s.io/v1/namespaces/${namespace}/networkpolicies/${name}`
	);
};

export const getCustomresources = (
	params: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(
		'/kapis/resources.kubesphere.io/v1alpha3/customresourcedefinitions',
		{ params }
	);
};

export const getCustomresourceDetail = (
	params: CustomresourceDetailParam
): Promise<AxiosResponse<any>> => {
	const { version, kind } = params;
	return api.get(`/apis/sys.bytetrade.io/${version}/${kind}`);
};

export const getNamespacesList = (
	params: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/namespaces', {
		params
	});
};

export const getNamespacesGroup = (
	params: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	console.log(params);
	return api.get('/capi/namespaces/group');
};

export const getNamespaceDeployments = (
	namespace: string,
	params?: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/deployments`,
		{
			params
		}
	);
};

export const getNamespacesDetail = (
	namespaces: string
): Promise<AxiosResponse<PodItemResponse>> => {
	return api.get(`/api/v1/namespaces/${namespaces}`);
};

export const getDeployments = (
	params: Pagination & { name?: string }
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/deployments', {
		params
	});
};

export const getWorkloadList = (
	params: Pagination & { name?: string },
	kind: string
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(`/kapis/resources.kubesphere.io/v1alpha3/${kind}`, {
		params
	});
};

export const getNamespaceStatefulsets = (
	namespace: string,
	params?: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/statefulsets`,
		{
			params
		}
	);
};

export const getNamespaceDaemonsets = (
	namespace: string,
	params?: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/daemonsets`,
		{
			params
		}
	);
};

export const getNameSpacePodsList = (
	params: PodsParam & { namespace: string },
	config?: any
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	console.log(rest);
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/pods`,
		{
			params,
			...config
		}
	);
};

export const getNameSpaceServicesList = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	console.log(rest);
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/services`,
		{ params }
	);
};

export const getEndpoints = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<any>> => {
	const { namespace, name, ...rest } = params;
	console.log(rest);
	return api.get(`/api/v1/namespaces/${namespace}/endpoints/${name}`);
};

export const getWorkloads = (
	params: PodsParam & { namespace: string; type: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, type, ...rest } = params;
	console.log(rest);
	return api.get(`/apis/apps/v1/namespaces/${namespace}/${type}`, { params });
};

export const patchWorkloadsControler = (
	namespace: string,
	kind: string,
	name: string,
	params?: any
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.patch(
		`/apis/apps/v1/namespaces/${namespace}/${kind}/${name}`,
		params,
		{ headers: { 'Content-Type': 'application/merge-patch+json' } }
	);
};
export const getWorkloadsControler = (
	namespace: string,
	kind: string,
	name: string
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get(`/apis/apps/v1/namespaces/${namespace}/${kind}/${name}`);
};

export const getDetail = (
	apiVersion: string,
	params: { namespace: string; kind: string; name: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get(
		`${apiVersion}/namespaces/${params.namespace}/${params.kind}/${params.name}`
	);
};

export const updateWorkloadsControler = (
	namespace: string,
	kind: string,
	name: string,
	params: any
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.put(
		`/apis/apps/v1/namespaces/${namespace}/${kind}/${name}`,
		params
	);
};

export const updateDetail = (
	apiVersion: string,
	obj: { namespace: string; kind: string; name: string },
	params: any
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.put(
		`${apiVersion}/namespaces/${obj.namespace}/${obj.kind}/${obj.name}`,
		params
	);
};

export const updateCustomResources = (
	namespace: string,
	module: string,
	name: string,
	apiVersion: string,
	params: any,
	cluster?: string
): Promise<AxiosResponse<ResourcesResponse>> => {
	let path = '';
	if (cluster) {
		// path += `/klusters/${cluster}`
	}
	if (namespace) {
		path += `/namespaces/${namespace}`;
	}

	return api.put(`/apis/${apiVersion}${path}/${module}/${name}`, params);
};

export const getEvent = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	return api.get(`/api/v1/namespaces/${namespace}/events`, { params: rest });
};

export const getAllSecrets = (
	params: Pagination
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/secrets', { params });
};

export const getConfigmapsAll = (
	params: Pagination
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/configmaps', {
		params
	});
};

export const getServiceaccountsAll = (
	params: Pagination
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/kapis/resources.kubesphere.io/v1alpha3/serviceaccounts', {
		params
	});
};

export const getSecretsData = (params: {
	namespace: string;
	name: string;
}): Promise<AxiosResponse<SecretsDataResponse>> => {
	return api.get(
		`/api/v1/namespaces/${params.namespace}/secrets/${params.name}`
	);
};

export const getServicesData = (
	namespace: string,
	name: string,
	params?: any
): Promise<AxiosResponse<SecretsDataResponse>> => {
	return api.get(`/api/v1/namespaces/${namespace}/services/${name}`, {
		params
	});
};

export const putServicesData = (
	namespace: string,
	name: string,
	params?: any
): Promise<AxiosResponse<SecretsDataResponse>> => {
	return api.put(`/api/v1/namespaces/${namespace}/services/${name}`, params);
};

export const getConfigmapsData = (params: {
	namespace: string;
	name: string;
}): Promise<AxiosResponse<SecretsDataResponse>> => {
	return api.get(
		`/api/v1/namespaces/${params.namespace}/configmaps/${params.name}`
	);
};

export const getSecrets = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	console.log(rest);
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/secrets`,
		{ params }
	);
};

export const getConfigmaps = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	console.log(rest);
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/configmaps`,
		{ params }
	);
};

export const getServiceaccounts = (
	params: PodsParam & { namespace: string }
): Promise<AxiosResponse<ResourcesResponse>> => {
	const { namespace, ...rest } = params;
	console.log(rest);
	return api.get(
		`/kapis/resources.kubesphere.io/v1alpha3/namespaces/${namespace}/serviceaccounts`,
		{ params }
	);
};

export const getServiceaccountsItem = (
	namespace: string,
	name: string
): Promise<AxiosResponse<ServiceaccountsItemResponse>> => {
	return api.get(`api/v1/namespaces/${namespace}/serviceaccounts/${name}`);
};

export const getDeploymentsDetail = (
	namespace: string,
	name: string,
	params?: PodsParam
): Promise<AxiosResponse<DeploymentsDetailResponse>> => {
	return api.get(`/apis/apps/v1/namespaces/${namespace}/deployments/${name}`, {
		params
	});
};

export const getCRDItemList = (
	params: Pagination & {
		namespace?: string;
		group: string;
		version: string;
		kind: string;
	}
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { namespace, group, version, kind, ...rest } = params;
	const path = namespace
		? `/apis/${group}/${version}/namespaces/${namespace}/${kind}`
		: `/apis/${group}/${version}/${kind}`;
	return api.get(path, {
		params: rest
	});
};

export const getNodeEvent = (
	params: PodsParam
): Promise<AxiosResponse<ResourcesResponse>> => {
	return api.get('/api/v1/events', { params });
};

const getPath = ({
	cluster,
	namespace
}: {
	cluster?: string;
	namespace: string;
}) => {
	let path = '';
	if (cluster) {
		path += `/klusters/${cluster}`;
	}
	if (namespace) {
		path += `/namespaces/${namespace}`;
	}
	return path;
};

export const getRevisions = (
	module: string,
	{
		cluster,
		namespace,
		kind,
		selector
	}: {
		cluster: string;
		namespace: string;
		kind: string;
		selector: any;
	}
) => {
	const labelSelector = joinSelector(selector);
	console.log(kind);
	const prefix =
		module === 'deployments'
			? `apis/apps/v1${getPath({ cluster, namespace })}/replicasets`
			: `apis/apps/v1${getPath({
					cluster,
					namespace
			  })}/controllerrevisions`;
	return api.get(`${prefix}?labelSelector=${labelSelector}`);
};

export const getMiddlewareListAll = (): Promise<
	AxiosResponse<MiddlewareListResponse>
> => {
	return api.get(`/middleware/v1/list`);
};

export const getMiddlewareList = (
	type: MiddlewareType
): Promise<AxiosResponse<MiddlewareListResponse>> => {
	return api.get(`/middleware/v1/${type}/list`);
};

export const getMiddlewareAll = (): Promise<AxiosResponse<any>> => {
	return api.get('/middleware/v1/requests');
};

export const updateMiddlewarePassword = (
	middleware: MiddlewareType,
	params: MiddlewarePasswordParams
): Promise<AxiosResponse<MiddlewarePasswordResponse>> => {
	return api.post(`/middleware/v1/${middleware}/password`, params);
};

export const getContainersMonitoringOfPod = (
	namespace: string,
	pod: string,
	params: any
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/pods/${pod}/containers`,
		{
			params
		}
	);
};

export const getUserMetric = (
	username: string,
	params: any
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(`/kapis/monitoring.kubesphere.io/v1alpha3/users/${username}`, {
		params
	});
};

export const getWorkloadsMetrics = (
	namespace: string,
	pods: string,
	params: PodsParam,
	config?: Record<string, any>
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/workloads/deployment/${pods}/pods`,
		{
			params,
			...config
		}
	);
};

export const getNamespaceMetrics = (
	namespace: string,
	params: PodsParam,
	config?: Record<string, any>
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}`,
		{
			params,
			...config
		}
	);
};

export const getTypesMetrics = (
	type: string,
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(`/capi/monitoring.kubesphere.io/v1alpha3/${type}`, {
		params
	});
};

export const getCustomresourcedefinitions = (
	type: string
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/apis/apiextensions.k8s.io/v1/customresourcedefinitions/${type}`
	);
};

export const deleteCustomResources = (params: {
	apiVersion: string;
	namespaces: string;
	cluster?: string;
	module: string;
	name: string;
}): Promise<AxiosResponse<any>> => {
	const { apiVersion, cluster, namespaces } = params;
	let path = '';
	if (cluster) {
		// path += `/klusters/${cluster}`
	}
	if (namespaces) {
		path += `/namespaces/${namespaces}`;
	}

	return api.delete(
		`/apis/${apiVersion}${path}/${params.module}/${params.name}`
	);
};

export const deleteCustomApplications = (params: {
	apiVersion: string;
	name: string;
}): Promise<AxiosResponse<any>> => {
	return api.delete(`/${params.apiVersion}/klusters/${params.name}`);
};

export const getCustomresourceItem = (
	apiVersion: string,
	namespace: string,
	module: string,
	name: string,
	cluster?: string
): Promise<AxiosResponse<any>> => {
	return api.get(
		`/apis/${apiVersion}${getPath({ cluster, namespace })}/${module}/${name}`
	);
};

export const deletePod = (
	namespace: string,
	pod: string
): Promise<AxiosResponse<any>> => {
	return api.delete(`/api/v1/namespaces/${namespace}/pods/${pod}`);
};

export const getPersistentvolumeclaims = (
	params: { namespace: string } & PodsParam
) => {
	const { namespace, ...rest } = params;
	const path = namespace
		? `/namespaces/${namespace}/persistentvolumeclaims`
		: '/persistentvolumeclaims';
	return api.get(`/kapis/resources.kubesphere.io/v1alpha3${path}`, {
		params: rest
	});
};

export const deletePersistentvolumeclaims = (params: {
	apiVersion: string;
	namespace: string;
	cluster?: string;
	module: string;
	name: string;
}): Promise<AxiosResponse<any>> => {
	const { apiVersion, cluster, namespace, name, module } = params;
	const path = getPath({ cluster, namespace });
	return api.delete(`/${apiVersion}${path}/${module}/${name}`);
};

export const patchPersistentvolumeclaims = (
	params: {
		apiVersion: string;
		namespace: string;
		cluster?: string;
		module: string;
		name: string;
	},
	data: any
): Promise<AxiosResponse<any>> => {
	const { apiVersion, cluster, namespace, name, module } = params;
	const path = getPath({ cluster, namespace });
	return api.patch(`/${apiVersion}${path}/${module}/${name}`, data, {
		headers: {
			'Content-Type': 'application/merge-patch+json'
		}
	});
};

export const getJobs = (
	module: jobType,
	params?: Pagination
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(`/kapis/resources.kubesphere.io/v1alpha3/${module}`, {
		params
	});
};

export const getCornJobsDetail = (
	apiVersion: string,
	namespace: string,
	module: string,
	name: string,
	params: any
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.get(`${apiVersion}/namespaces/${namespace}/${module}/${name}`, {
		params
	});
};

export const getJobEvent = (
	params: PodMonitoringParamAll
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { name, cluster, namespace, ...rest } = params;

	const clusterPath = cluster ? `/klusters/${cluster}` : '';
	const namespacePath = namespace ? `/namespaces/${namespace}` : '';

	return api.get(`api/v1${clusterPath}${namespacePath}/events`, {
		params: rest
	});
};

export const fetchJobRecords = (
	params: any
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { cluster, module, namespace, selector, name, ...rest } = params;
	if (!isEmpty(selector)) {
		params.labelSelector = joinSelector(selector);
	}
	return api.get(
		`apis/batch/v1${getPath({ cluster, namespace })}/${module}/${name}`,
		{
			params: rest
		}
	);
};

export const fetchListByK8s = (
	params: any
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { cluster, module, namespace, selector, ...rest } = params;
	if (!isEmpty(selector)) {
		params.labelSelector = joinSelector(selector);
	}
	return api.get(`apis/batch/v1${getPath({ cluster, namespace })}/${module}`, {
		params: rest
	});
};

export const jobRerun = (
	resourceVersion: string,
	params: PodMonitoringParamAll
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { name, cluster, namespace } = params;

	return api.post(
		`kapis/operations.kubesphere.io/v1alpha2${getPath({
			cluster,
			namespace
		})}/jobs/${name}?action=rerun&resourceVersion=${resourceVersion}`
	);
};

export const deleteJob = (
	apiVersion: string,
	namespace: string,
	module: string,
	name: string,
	params: any
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	return api.delete(`${apiVersion}/namespaces/${namespace}/${module}/${name}`, {
		params
	});
};

export const getWorkloadsMetrics2 = (
	namespace: string,
	workload: string,
	pod: string,
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/workloads/${workload}/${pod}/pods`,
		{
			params
		}
	);
};

export const getWorkloadsPods = (
	namespace: string,
	workload: string,
	pod: string,
	params: PodsParam
): Promise<AxiosResponse<MonitoringResponse>> => {
	return api.get(
		`kapis/monitoring.kubesphere.io/v1alpha3/namespaces/${namespace}/workloads/${workload}/${pod}/pods`,
		{
			params
		}
	);
};

export const toggleJob = (
	urlParams: PodMonitoringParamAll & { apiVersion: string },
	params: any
): Promise<AxiosResponse<CustomresourcesResponse>> => {
	const { apiVersion, name, cluster, namespace } = urlParams;

	return api.patch(
		`${apiVersion}${getPath({
			cluster,
			namespace
		})}/cronjobs/${name}`,
		params,
		{ headers: { 'Content-Type': 'application/merge-patch+json' } }
	);
};

export const restartPods = (
	namespace: string,
	name: string
): Promise<AxiosResponse<any>> => {
	return api.delete(`/api/v1/namespaces/${namespace}/pods/${name}`);
};
