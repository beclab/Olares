import { get, includes, omit } from 'lodash';
import moment from 'moment-mini';
import { t } from '@apps/control-panel-common/src/boot/i18n';
import {
	safeParseJSON,
	getDescription,
	getAliasName,
	getResourceCreator
} from './index.js';
import { getPodStatusAndRestartCount } from './status.js';
import { safeAtob } from './base64.js';
import { getServiceType } from './service.js';
import { getWorkloadUpdateTime, getJobUpdateTime } from './workload.js';
import { getNodeRoles } from './node.js';

const getContainers = (containers: any, statuses: any, namespace: any) =>
	containers.map((container: any) => {
		const status = omit(
			statuses.find((item: any) => item.name === container.name) || {},
			'image'
		);

		return {
			...container,
			...status,
			namespace
		};
	});

const getOriginData = (item: any) =>
	omit(item, [
		'status',
		'metadata.uid',
		'metadata.selfLink',
		'metadata.generation',
		'metadata.ownerReferences',
		'metadata.resourceVersion',
		'metadata.creationTimestamp',
		'metadata.managedFields'
	]);

const getOriginData2 = (item: any) =>
	omit(item, [
		'metadata.uid',
		'metadata.selfLink',
		'metadata.generation',
		'metadata.ownerReferences',
		'metadata.resourceVersion',
		'metadata.creationTimestamp',
		'metadata.managedFields'
	]);

const getBaseInfo = (item: any) => ({
	uid: get(item, 'metadata.uid'),
	name: get(item, 'metadata.name'),
	creator: getResourceCreator(item),
	description: getDescription(item),
	aliasName: getAliasName(item),
	createTime: get(item, 'metadata.creationTimestamp', ''),
	resourceVersion: get(item, 'metadata.resourceVersion'),
	isFedManaged: get(item, 'metadata.labels["kubefed.io/managed"]') === 'true'
});

export const DefaultMapper = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	spec: get(item, 'spec'),
	_originData: getOriginData(item)
});

export const DefaultMapper2 = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	spec: get(item, 'spec'),
	_originData: getOriginData2(item)
});

export const PodsMapper = (item: any) => ({
	...getBaseInfo(item),
	deletionTime: get(item, 'metadata.deletionTimestamp'),
	labels: get(item, 'metadata.labels'),
	namespace: get(item, 'metadata.namespace'),
	annotations: get(item, 'metadata.annotations'),
	podNums: get(item, 'spec.replicas'),
	status: get(item, 'status', {}),
	podStatus: getPodStatusAndRestartCount(item),
	spec: get(item, 'spec', {}),
	metrics: get(item, 'metrics'),
	node: get(item, 'spec.nodeName', ''),
	nodeIp: get(item, 'status.hostIP', 'none'),
	podIp: get(item, 'status.podIP'),
	networksStatus: safeParseJSON(
		get(item, 'metadata.annotations["k8s.v1.cni.cncf.io/networks-status"]', ''),
		[]
	),
	app: get(item, 'metadata.labels["app.kubernetes.io/name"]'),
	containers: getContainers(
		get(item, 'spec.containers', []),
		get(item, 'status.containerStatuses', []),
		get(item, 'metadata.namespace')
	),
	initContainers: getContainers(
		get(item, 'spec.initContainers', []),
		get(item, 'status.initContainerStatuses', []),
		get(item, 'metadata.namespace')
	),
	startTime: get(item, 'status.startTime'),
	updateTime: get(item, 'status.startTime'),
	volumes: get(item, 'spec.volumes'),
	ownerKind: get(item, 'metadata.ownerReferences[0].kind', ''),
	ownerName: get(item, 'metadata.ownerReferences[0].name', ''),
	_originData: getOriginData(item)
});

export const UserMapper = (item: any) => ({
	...getBaseInfo(item),
	username: get(item, 'metadata.name', ''),
	email: get(item, 'spec.email', ''),
	role: get(item, 'metadata.annotations["iam.kubesphere.io/role"]', ''),
	globalrole: get(
		item,
		'metadata.annotations["iam.kubesphere.io/globalrole"]',
		''
	),
	clusterrole: get(
		item,
		'metadata.annotations["iam.kubesphere.io/clusterrole"]',
		''
	),
	workspacerole: get(
		item,
		'metadata.annotations["iam.kubesphere.io/workspacerole"]',
		''
	),
	roleBind: get(
		item,
		'metadata.annotations["iam.kubesphere.io/role-binding"]',
		''
	),
	groups: get(item, 'spec.groups', []),
	status: get(item, 'status.state', 'Pending'),
	conditions: get(item, 'status.conditions', []),
	lastLoginTime: get(item, 'status.lastLoginTime'),
	_originData: getOriginData(item)
});

export type UserMapperType = ReturnType<typeof UserMapper>;

export const NetworkPoliciesMapper = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	_originData: getOriginData(item),
	key: `${get(item, 'metadata.namespace')}-${get(item, 'metadata.name')}`
});

const getServedVersion = (item: any) => {
	const versions = get(item, 'spec.versions', []);
	if (versions.length === 0) {
		return '';
	}
	let servedVersion = get(versions[versions.length - 1], 'name');
	versions.some((ver: any) => {
		if (get(ver, 'served', false)) {
			servedVersion = get(ver, 'name', servedVersion);
			return true;
		}
		return false;
	});
	return servedVersion;
};

export const CRDMapper = (item: any) => {
	const versions = get(item, 'spec.versions', []);
	return {
		versions,
		...getBaseInfo(item),
		group: get(item, 'spec.group'),
		scope: get(item, 'spec.scope'),
		kind: get(item, 'spec.names.kind'),
		latestVersion: getServedVersion(item),
		module: get(item, 'status.acceptedNames.plural'),
		_originData: getOriginData2(item)
	};
};

export const EndpointMapper = (item: any) => ({
	addresses: item.addresses || [],
	ports: item.ports || []
});

const ServiceMapper = (item: any) => {
	const specType = get(item, 'spec.type');
	const clusterIP = get(item, 'spec.clusterIP');
	const selector = get(item, 'spec.selector', {});

	return {
		type: getServiceType(item),
		clusterIP,
		selector,
		specType,
		...getBaseInfo(item),
		namespace: get(item, 'metadata.namespace'),
		labels: get(item, 'metadata.labels', {}),
		annotations: get(item, 'metadata.annotations', {}),
		status: get(item, 'status'),
		ports: get(item, 'spec.ports', []),
		workloadType: get(
			item,
			'metadata.annotations["kubesphere.io/workloadType"]',
			'Deployment'
		),
		sessionAffinity: get(item, 'spec.sessionAffinity'),
		externalIPs: get(item, 'spec.externalIPs', []),
		externalName: get(item, 'spec.externalName'),
		loadBalancerIngress: get(item, 'status.loadBalancer.ingress', []).map(
			(lb: any) => lb.ip || lb.hostname
		),
		app:
			get(item, 'metadata.labels["app.kubernetes.io/name"]') ||
			get(item, 'metadata.labels.app'),
		_originData: getOriginData(item)
	};
};

const ServiceAccountMapper = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	labels: get(item, 'metadata.labels', {}),
	annotations: get(item, 'metadata.annotations', {}),
	role: get(item, 'metadata.annotations["iam.kubesphere.io/role"]'),
	secrets: get(item, 'secrets', []),
	_originData: getOriginData(item)
});

const ConfigmapMapper = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	labels: get(item, 'metadata.labels', {}),
	annotations: get(item, 'metadata.annotations', {}),
	data: get(item, 'data', {}),
	_originData: getOriginData(item)
});

const EventsMapper = (item: any) => {
	const now = Date.now();

	item.lastTimestamp = item.lastTimestamp || now;

	const age =
		item.count > 1
			? item.count === 2
				? t('EVENT_AGE_DATA_TWICE', {
						lastTime: moment(item.lastTimestamp).fromNow(),
						duration: moment(item.firstTimestamp).to(now, true)
				  })
				: t('EVENT_AGE_DATA', {
						lastTime: moment(item.lastTimestamp).fromNow(),
						count: item.count,
						duration: moment(item.firstTimestamp).to(now, true)
				  })
			: moment(item.firstTimestamp).fromNow();

	return {
		...getBaseInfo(item),
		age,
		type: get(item, 'type'),
		reason: get(item, 'reason'),
		message: get(item, 'message'),
		from: get(item, 'source.component'),
		lastTimestamp: item.lastTimestamp,
		_originData: getOriginData(item)
	};
};

const secretDataParser = (data: any) => {
	if (data.type === 'kubernetes.io/basic-auth') {
		return Object.entries(get(data, 'data', {})).reduce(
			(prev, [key, value]) => ({
				...prev,
				[key]: safeAtob(value) === 'undefined' ? '' : safeAtob(value)
			}),
			{}
		);
	}

	return Object.entries(get(data, 'data', {})).reduce(
		(prev, [key, value]) => ({
			...prev,
			[key]:
				key === '.dockerconfigjson'
					? safeParseJSON(safeAtob(value), {})
					: safeAtob(value)
		}),
		{}
	);
};

const SecretMapper = (item: any) => ({
	...getBaseInfo(item),
	namespace: get(item, 'metadata.namespace'),
	labels: get(item, 'metadata.labels', {}),
	annotations: get(item, 'metadata.annotations', {}),
	type: get(item, 'type', ''),
	data: secretDataParser(item),
	_originData: getOriginData(item)
});

const WorkLoadMapper = (item: any) => ({
	...getBaseInfo(item),
	kind: get(item, 'kind'),
	updateTime: getWorkloadUpdateTime(item),
	labels: get(item, 'metadata.labels', {}),
	namespace: get(item, 'metadata.namespace'),
	annotations: get(item, 'metadata.annotations'),
	status: get(item, 'status', {}),
	availablePodNums: get(item, 'status.availableReplicas', 0),
	readyPodNums: get(item, 'status.readyReplicas', 0),
	spec: get(item, 'spec', {}),
	podNums: get(item, 'spec.replicas', 0),
	selector: get(item, 'spec.selector.matchLabels'),
	containers: get(item, 'spec.template.spec.containers'),
	initContainers: get(item, 'spec.template.spec.initContainers'),
	volumes: get(item, 'spec.template.spec.volumes'),
	strategy: get(item, 'spec.strategy', {}),
	updateStrategy: get(item, 'spec.updateStrategy.type'),
	availableCondition:
		get(item, 'status.conditions', []).find(
			(cd: any) => cd.type === 'Available'
		) || {},
	app: get(item, 'metadata.labels["app.kubernetes.io/name"]'),
	ownerReference: get(item, 'metadata.ownerReferences[0]', {}),
	hasS2i: Object.keys(get(item, 'metadata.labels', {})).some((labelKey) =>
		labelKey.startsWith('s2ibuilder')
	),
	builderNames: Object.entries(get(item, 'metadata.labels', {}))
		.filter((labelArray) => labelArray[0].startsWith('s2ibuilder'))
		.map((array) => array[1]), // polyfill for multi s2i in one workload
	_originData: getOriginData(item)
});

const getVolumePhase = (item: any) => {
	const phase = get(item, 'status.phase');
	const deletionTime = get(item, 'metadata.deletionTimestamp');

	if (deletionTime) {
		return 'Terminating';
	}

	return phase;
};

const VolumeMapper = (item: any) => {
	const deletionTime = get(item, 'metadata.deletionTimestamp');

	return {
		deletionTime,
		phase: getVolumePhase(item),
		...getBaseInfo(item),
		storageProvisioner: get(
			item,
			'metadata.annotations["volume.beta.kubernetes.io/storage-provisioner"]'
		),
		status: get(item, 'status', {}),
		conditions: get(item, 'status.conditions', []),
		namespace: get(item, 'metadata.namespace'),
		labels: get(item, 'metadata.labels'),
		annotations: get(item, 'metadata.annotations'),
		accessMode: get(item, 'spec.accessModes[0]'),
		accessModes: get(item, 'spec.accessModes'),
		storageClassName: get(item, 'spec.storageClassName'),
		resources: get(item, 'spec.resources'),
		capacity: get(
			item,
			'status.capacity.storage',
			get(item, 'spec.resources.requests.storage')
		),
		inUse:
			get(item, 'metadata.annotations["kubesphere.io/in-use"]') === 'true' ||
			get(item, 'status.phase') === 'Bound',
		type: 'pvc',
		_originData: getOriginData(item)
	};
};

const NodeMapper = (item: any) => ({
	...getBaseInfo(item),
	labels: get(item, 'metadata.labels'),
	role: getNodeRoles(get(item, 'metadata.labels')),
	annotations: get(item, 'metadata.annotations'),
	status: get(item, 'status'),
	conditions: get(item, 'status.conditions', []),
	nodeInfo: get(item, 'status.nodeInfo'),
	spec: get(item, 'spec'),
	unschedulable: get(item, 'spec.unschedulable'),
	importStatus: get(
		item,
		'metadata.labels["kubekey.kubesphere.io/import-status"]',
		'success'
	),
	taints: get(item, 'spec.taints', []),
	ip:
		(
			get(item, 'status.addresses', []).find(
				(a: any) => a.type === 'InternalIP'
			) || {}
		).address || '-',
	_originData: getOriginData(item)
});

const JobMapper = (item: any) => ({
	...getBaseInfo(item),
	labels: get(item, 'metadata.labels', {}),
	namespace: get(item, 'metadata.namespace'),
	annotations: get(item, 'metadata.annotations'),
	status: get(item, 'status'),
	updateTime: getJobUpdateTime(item),
	startTime: get(item, 'status.startTime'),
	spec: get(item, 'spec', {}),
	selector: get(item, 'spec.selector.matchLabels'),
	containers: get(item, 'spec.template.spec.containers'),
	volumes: get(item, 'spec.template.spec.volumes'),
	_originData: getOriginData(item)
});

const RevisionMapper = (item: any) => {
	const spec = get(item, 'data.spec', get(item, 'spec', {}));

	return {
		spec,
		...getBaseInfo(item),
		ownerKind: get(item, 'metadata.ownerReferences[0].kind', ''),
		ownerName: get(item, 'metadata.ownerReferences[0].name', ''),
		namespace: get(item, 'metadata.namespace'),
		labels: get(item, 'metadata.labels', {}),
		annotations: get(item, 'metadata.annotations'),
		revision:
			Number(
				get(
					item,
					'metadata.annotations["deployment.kubernetes.io/revision"]',
					item.revision
				)
			) || null,
		status: get(item, 'status'),
		podNums: get(spec, 'replicas'),
		selector: get(spec, 'selector.matchLabels'),
		containers: get(spec, 'template.spec.containers'),
		initContainers: get(spec, 'template.spec.initContainers'),
		volumes: get(spec, 'template.spec.volumes'),
		strategy: get(spec, 'strategy', {}),
		updateStrategy: get(spec, 'updateStrategy.type'),
		_originData: getOriginData(item)
	};
};

const StorageclasscapabilitiesMapper = (item: any) => {
	const { metadata, spec } = item;
	const volumeFeature = get(spec, 'features.volume');
	return {
		metadata,
		spec,
		snapshotFeature: get(spec, 'features.snapshot'),
		volumeFeature,
		supportExpandVolume: includes(
			['OFFLINE', 'ONLINE'],
			volumeFeature.expandMode
		)
	};
};

const CronJobMapper = (item: any) => ({
	...getBaseInfo(item),
	labels: get(item, 'metadata.labels', {}),
	namespace: get(item, 'metadata.namespace'),
	annotations: get(item, 'metadata.annotations'),
	status: get(item, 'status'),
	spec: get(item, 'spec', {}),
	selector: get(item, 'spec.jobTemplate.metadata.labels'),
	suspend: get(item, 'spec.suspend'),
	_originData: getOriginData(item)
});

export const ObjectMapper = {
	pods: PodsMapper,
	endpoints: EndpointMapper,
	users: UserMapper,
	customresourcedefinitions: CRDMapper,
	networkpolicies: NetworkPoliciesMapper,
	namespacenetworkpolicies: NetworkPoliciesMapper,
	default: DefaultMapper,
	services: ServiceMapper,
	serviceaccounts: ServiceAccountMapper,
	configmaps: ConfigmapMapper,
	events: EventsMapper,
	secrets: SecretMapper,
	deployments: WorkLoadMapper,
	daemonsets: WorkLoadMapper,
	statefulsets: WorkLoadMapper,
	workLoadMapper: WorkLoadMapper,
	volumes: VolumeMapper,
	persistentvolumeclaims: VolumeMapper,
	nodes: NodeMapper,
	jobs: JobMapper,
	cronjobs: CronJobMapper,
	revisions: RevisionMapper,
	storageclasscapabilities: StorageclasscapabilitiesMapper
};
