/*
 * This file is part of KubeSphere Console.
 * Copyright (C) 2019 The KubeSphere Console Authors.
 *
 * KubeSphere Console is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * KubeSphere Console is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with KubeSphere Console.  If not, see <https://www.gnu.org/licenses/>.
 */

export const WORKLOAD_STATUS = [
	{ text: 'RUNNING', value: 'running' },
	{ text: 'UPDATING', value: 'updating' },
	{ text: 'STOPPED', value: 'stopped' }
];

export const POD_STATUS = [
	{ text: 'RUNNING', value: 'running' },
	{ text: 'WAITING', value: 'waiting' },
	{ text: 'TERMINATED', value: 'terminated' }
];

export const PODS_STATUS = [
	{ text: 'WAITING', value: 'Waiting' },
	{ text: 'RUNNING', value: 'Running' },
	{ text: 'COMPLETED', value: 'Completed' },
	{ text: 'ERROR', value: 'Error' }
];

export const JOB_STATUS = [
	{ text: 'FAILED', value: 'failed' },
	{ text: 'COMPLETED', value: 'completed' },
	{ text: 'RUNNING', value: 'running' }
];

export const S2IJOBS_STATUS = [
	{ text: 'FAILED', value: 'Failed' },
	{ text: 'SUCCESSFUL', value: 'successful' },
	{ text: 'RUNNING', value: 'Running' }
];

export const S2I_STATUS_DESC = {
	Failed: 'IMAGE_BUILDING_FAILED',
	Running: 'BUILDING_IMAGE',
	Successful: 'IMAGE_BUILDING_SUCCESSFUL'
};

export const CRONJOB_STATUS = [
	{ text: 'CRONJOB_PAUSED', value: 'paused' },
	{ text: 'CRONJOB_RUNNING', value: 'running' }
];

export const VOLUME_STATUS = [
	{ text: 'VOLUME_STATUS_BOUND', value: 'bound' },
	{ text: 'VOLUME_STATUS_LOST', value: 'lost' },
	{ text: 'VOLUME_STATUS_PENDING', value: 'pending' }
];

export const PV_STATUS = [
	{ text: 'PV_STATUS_AVAILABLE', value: 'available' },
	{ text: 'PV_STATUS_BOUND', value: 'bound' },
	{ text: 'PV_STATUS_RELEASED', value: 'released' },
	{ text: 'PV_STATUS_FAILED', value: 'failed' }
];

export const NODE_STATUS = [
	{ text: 'NODE_STATUS_UNSCHEDULABLE', value: 'unschedulable' },
	{ text: 'NODE_STATUS_RUNNING', value: 'running' },
	{ text: 'NODE_STATUS_WARNING', value: 'warning' }
];

export const CD_WEATHER_STATUS = [
	{ text: 'HEALTHY', value: 'Healthy' },
	{ text: 'DEGRADED', value: 'Degraded' },
	{ text: 'PROGRESSING', value: 'Progressing' },
	{ text: 'SUSPENDED', value: 'Suspended' },
	{ text: 'UNKNOWN', value: 'Unknown' },
	{ text: 'MISSING', value: 'Missing' }
];

export const CD_SYNC_STATUS = [
	{ text: 'SYNCED', value: 'Synced' },
	{ text: 'OUTOFSYNC', value: 'OutOfSync' },
	{ text: 'UNKNOWN', value: 'Unknown' }
];

export const FLUXCD_APP_TYPES = [
	{
		label: 'HelmRelease',
		value: 'HelmRelease',
		description: 'Helm Application',
		icon: 'helm'
	},
	{
		label: 'Kustomization',
		value: 'Kustomization',
		description: 'Kustomize or pure kubernetes manifests',
		icon: 'kubernetes'
	}
];

export const PRUNE_PROPAGATION_POLICY_OPTIONS = [
	{ label: 'foreground', value: 'foreground' },
	{ label: 'background', value: 'background' },
	{ label: 'orphan', value: 'orphan' }
];

export const ACCESS_MODES = {
	ReadWriteOnce: 'RWO',
	ReadOnlyMany: 'ROX',
	ReadWriteMany: 'RWX'
};

export const SERVICE_TYPES = {
	VirtualIP: 'VIRTUAL_IP',
	Headless: 'HEADLESS',
	ExternalName: 'EXTERNALNAME',
	Unknown: 'UNKNOWN'
};

export const SERVICE_TYPES_VALUE = {
	VIRTUAL_IP: 'virtualIP',
	HEADLESS: 'headlessSelector'
};

export const VOLUME_SNAPSHOT_STATUS = [
	{ text: 'VOLUME_SNAPSHOT_STATUS_CREATING', value: 'creating' },
	{ text: 'VOLUME_SNAPSHOT_STATUS_READY', value: 'ready' },
	{ text: 'VOLUME_SNAPSHOT_STATUS_DELETING', value: 'deleting' }
];

export const VOLUME_SNAPSHOT_CLASS_STATUS = [
	{ text: 'READY', value: 'true' },
	{ text: 'UNREADY', value: 'false' }
];

export const PARMMETER_TYPES = [
	{
		label: 'AUTO_PARAMETER',
		value: 'auto',
		description: 'AUTO_PARAMETER_DESC'
	},
	{
		label: 'HELM_PARAMETER',
		value: 'helm',
		description: 'HELM_PARAMETER_DESC'
	},
	{
		label: 'KUSTOMIZE_PARAMETER',
		value: 'kustomize',
		description: 'KUSTOMIZE_PARAMETER_DESC'
	}
];

export const INGRESS_ANNOTATIONS = [
	'nginx.ingress.kubernetes.io/app-root',
	'nginx.ingress.kubernetes.io/affinity',
	'nginx.ingress.kubernetes.io/auth-realm',
	'nginx.ingress.kubernetes.io/auth-secret',
	'nginx.ingress.kubernetes.io/auth-type',
	'nginx.ingress.kubernetes.io/auth-tls-secret',
	'nginx.ingress.kubernetes.io/auth-tls-verify-depth',
	'nginx.ingress.kubernetes.io/auth-tls-verify-client',
	'nginx.ingress.kubernetes.io/auth-tls-error-page',
	'nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream',
	'nginx.ingress.kubernetes.io/auth-url',
	'nginx.ingress.kubernetes.io/auth-snippet',
	'nginx.ingress.kubernetes.io/enable-global-auth',
	'nginx.ingress.kubernetes.io/backend-protocol',
	'nginx.ingress.kubernetes.io/canary',
	'nginx.ingress.kubernetes.io/canary-by-header',
	'nginx.ingress.kubernetes.io/canary-by-header-value',
	'nginx.ingress.kubernetes.io/canary-by-cookie',
	'nginx.ingress.kubernetes.io/canary-weight',
	'nginx.ingress.kubernetes.io/client-body-buffer-size',
	'nginx.ingress.kubernetes.io/configuration-snippet',
	'nginx.ingress.kubernetes.io/custom-http-errors',
	'nginx.ingress.kubernetes.io/default-backend',
	'nginx.ingress.kubernetes.io/enable-cors',
	'nginx.ingress.kubernetes.io/cors-allow-origin',
	'nginx.ingress.kubernetes.io/cors-allow-methods',
	'nginx.ingress.kubernetes.io/cors-allow-headers',
	'nginx.ingress.kubernetes.io/cors-allow-credentials',
	'nginx.ingress.kubernetes.io/cors-max-age',
	'nginx.ingress.kubernetes.io/force-ssl-redirect',
	'nginx.ingress.kubernetes.io/from-to-www-redirect',
	'nginx.ingress.kubernetes.io/http2-push-preload',
	'nginx.ingress.kubernetes.io/limit-connections',
	'nginx.ingress.kubernetes.io/limit-rps',
	'nginx.ingress.kubernetes.io/permanent-redirect',
	'nginx.ingress.kubernetes.io/permanent-redirect-code',
	'nginx.ingress.kubernetes.io/temporal-redirect',
	'nginx.ingress.kubernetes.io/proxy-body-size',
	'nginx.ingress.kubernetes.io/proxy-cookie-domain',
	'nginx.ingress.kubernetes.io/proxy-cookie-path',
	'nginx.ingress.kubernetes.io/proxy-connect-timeout',
	'nginx.ingress.kubernetes.io/proxy-send-timeout',
	'nginx.ingress.kubernetes.io/proxy-read-timeout',
	'nginx.ingress.kubernetes.io/proxy-next-upstream',
	'nginx.ingress.kubernetes.io/proxy-next-upstream-timeout',
	'nginx.ingress.kubernetes.io/proxy-next-upstream-tries',
	'nginx.ingress.kubernetes.io/proxy-request-buffering',
	'nginx.ingress.kubernetes.io/proxy-redirect-from',
	'nginx.ingress.kubernetes.io/proxy-redirect-to',
	'nginx.ingress.kubernetes.io/enable-rewrite-log',
	'nginx.ingress.kubernetes.io/rewrite-target',
	'nginx.ingress.kubernetes.io/satisfy',
	'nginx.ingress.kubernetes.io/secure-verify-ca-secret',
	'nginx.ingress.kubernetes.io/server-alias',
	'nginx.ingress.kubernetes.io/server-snippet',
	'nginx.ingress.kubernetes.io/service-upstream',
	'nginx.ingress.kubernetes.io/session-cookie-name',
	'nginx.ingress.kubernetes.io/session-cookie-path',
	'nginx.ingress.kubernetes.io/ssl-redirect',
	'nginx.ingress.kubernetes.io/ssl-passthrough',
	'nginx.ingress.kubernetes.io/upstream-hash-by',
	'nginx.ingress.kubernetes.io/x-forwarded-prefix',
	'nginx.ingress.kubernetes.io/load-balance',
	'nginx.ingress.kubernetes.io/upstream-vhost',
	'nginx.ingress.kubernetes.io/whitelist-source-range',
	'nginx.ingress.kubernetes.io/proxy-buffering',
	'nginx.ingress.kubernetes.io/proxy-buffers-number',
	'nginx.ingress.kubernetes.io/proxy-buffer-size',
	'nginx.ingress.kubernetes.io/ssl-ciphers',
	'nginx.ingress.kubernetes.io/connection-proxy-header',
	'nginx.ingress.kubernetes.io/enable-access-log',
	'nginx.ingress.kubernetes.io/lua-resty-waf',
	'nginx.ingress.kubernetes.io/lua-resty-waf-debug',
	'nginx.ingress.kubernetes.io/lua-resty-waf-ignore-rulesets',
	'nginx.ingress.kubernetes.io/lua-resty-waf-extra-rules',
	'nginx.ingress.kubernetes.io/lua-resty-waf-allow-unknown-content-types',
	'nginx.ingress.kubernetes.io/lua-resty-waf-score-threshold',
	'nginx.ingress.kubernetes.io/lua-resty-waf-process-multipart-body',
	'nginx.ingress.kubernetes.io/enable-influxdb',
	'nginx.ingress.kubernetes.io/influxdb-measurement',
	'nginx.ingress.kubernetes.io/influxdb-port',
	'nginx.ingress.kubernetes.io/influxdb-host',
	'nginx.ingress.kubernetes.io/influxdb-server-name',
	'nginx.ingress.kubernetes.io/use-regex',
	'nginx.ingress.kubernetes.io/enable-modsecurity',
	'nginx.ingress.kubernetes.io/enable-owasp-core-rules',
	'nginx.ingress.kubernetes.io/modsecurity-transaction-id',
	'nginx.ingress.kubernetes.io/modsecurity-snippet'
];

export const ICON_TYPES = {
	deployments: 'backup',
	statefulsets: 'stateful-set',
	daemonsets: 'deamon-set',
	jobs: 'backup',
	cronjobs: 'backup',
	services: 'appcenter',
	ingresses: 'loadbalancer',
	routes: 'loadbalancer',
	applications: 'application',
	app_templates: 'documentation',
	repos: 'catalog',
	images: 'cdn',
	registries: 'cdn',
	volumes: 'storage',
	persistentvolumeclaims: 'storage',
	storageclasses: 'database',
	nodes: 'nodes',
	edgenodes: 'nodes',
	devops: 'strategy-group',
	projects: 'project',
	namespaces: 'project',
	users: 'human',
	roles: 'role',
	members: 'group',
	globalroles: 'role',
	clusterroles: 'role',
	workspaceroles: 'role',
	components: 'components',
	accounts: 'human',
	workspaces: 'enterprise',
	clusters: 'cluster',
	pods: 'pod',
	containers: 'container',
	'limits.cpu': 'cpu',
	'limits.memory': 'memory',
	'requests.cpu': 'cpu',
	'requests.memory': 'memory',
	configmaps: 'hammer',
	serviceaccounts: 'client',
	secrets: 'key',
	'alert-messages': 'loudspeaker',
	'alert-policies': 'wrench',
	apiserver: 'api',
	etcd: 'etcd',
	scheduler: 'scheduler',
	s2ibuilders: 'vnas',
	apps: 'appcenter',
	'volume-snapshots': 'snapshot',
	'volume-snapshot-content': 'snapshot',
	customresourcedefinitions: 'select',
	network: 'eip-group',
	networkpolicies: 'firewall',
	namespacenetworkpolicies: 'firewall',
	pipelines: 'blockchain',
	ippools: 'eip-group',
	cluster: 'cluster',
	notification: 'loudspeaker',
	notificationhistory: 'record',
	gpu: 'gpu',
	'gpu.limit': 'gpu',
	cds: 'rocket',
	codeRepos: 'code',
	allowlists: 'allowlist'
};

export const MODULE_KIND_MAP = {
	deployments: 'Deployment',
	statefulsets: 'StatefulSet',
	daemonsets: 'DaemonSet',
	jobs: 'Job',
	cronjobs: 'CronJob',
	pods: 'Pod',
	services: 'Service',
	ingresses: 'Ingress',
	persistentvolumeclaims: 'PersistentVolumeClaim',
	persistentvolumes: 'PersistentVolumes',
	storageclasses: 'StorageClass',
	'alert-policies': 'AlertingPolicy',
	configmaps: 'ConfigMap',
	serviceaccounts: 'ServiceAccount',
	secrets: 'Secret',
	s2ibuilders: 'S2iBuilder',
	nodes: 'Node',
	volumesnapshots: 'VolumeSnapshot',
	volumesnapshotclass: 'VolumeSnapshotClass',
	namespaces: 'Namespace',
	workspaces: 'WorkspaceTemplate',
	clusters: 'Cluster',
	dashboards: 'Dashboard',
	clusterdashboards: 'ClusterDashboard',
	applications: 'Application',
	users: 'User',
	devops: 'DevOpsProject',
	pipelines: 'Pipelines',
	ippools: 'IPPool',
	groups: 'Group',
	volumes: 'Volumes',
	hpas: 'HPA'
};

export const QUOTAS_MAP = {
	'limits.cpu': {
		name: 'limits.cpu',
		placeholder: 'eg: 1 or 1000m'
	},
	'requests.cpu': {
		name: 'requests.cpu',
		placeholder: 'eg: 1 or 1000m'
	},
	'limits.memory': {
		name: 'limits.memory',
		placeholder: 'eg: 100Gi'
	},
	'requests.memory': {
		name: 'requests.memory',
		placeholder: 'eg: 100Gi'
	},
	pods: {
		name: 'count/pods',
		placeholder: 'eg: 100'
	},
	deployments: {
		name: 'count/deployments.apps',
		placeholder: 'eg: 100'
	},
	statefulsets: {
		name: 'count/statefulsets.apps',
		placeholder: 'eg: 100'
	},
	daemonsets: {
		name: 'count/daemonsets.apps',
		placeholder: 'eg: 100'
	},
	jobs: {
		name: 'count/jobs.batch',
		placeholder: 'eg: 100'
	},
	cronjobs: {
		name: 'count/cronjobs.batch',
		placeholder: 'eg: 100'
	},
	volumes: {
		name: 'persistentvolumeclaims',
		placeholder: 'eg: 100'
	},
	services: {
		name: 'count/services',
		placeholder: 'eg: 100'
	},
	routes: {
		name: 'count/ingresses.extensions',
		placeholder: 'eg: 100'
	},
	secrets: {
		name: 'count/secrets',
		placeholder: 'eg: 100'
	},
	configmaps: {
		name: 'count/configmaps',
		placeholder: 'eg: 100'
	}
};

export const WORKSPACE_QUOTAS_MAP = {
	'limits.cpu': {
		name: 'limits.cpu',
		placeholder: 'eg: 1 or 1000m'
	},
	'requests.cpu': {
		name: 'requests.cpu',
		placeholder: 'eg: 1 or 1000m'
	},
	'limits.memory': {
		name: 'limits.memory',
		placeholder: 'eg: 100Gi'
	},
	'requests.memory': {
		name: 'requests.memory',
		placeholder: 'eg: 100Gi'
	}
};

export const REPO_TYPES = [
	{ name: 'GitHub', value: 'github', icon: 'github' },
	{ name: 'GitLab', value: 'gitlab', icon: 'gitlab' },
	{ name: 'Bitbucket', value: 'bitbucket_server', icon: 'bitbucket' },
	{ name: 'Git', value: 'git', icon: 'git' },
	{ name: 'SVN', value: 'svn', icon: 'svn' }
];

export const REPO_KEY_MAP = {
	git: 'git_source',
	svn: 'svn_source',
	single_svn: 'single_svn_source',
	github: 'github_source',
	bitbucket_server: 'bitbucket_server_source',
	gitlab: 'gitlab_source'
};

export const PIPELINE_PARAMS_TYPES = {
	string: 'PARAMS_STRING',
	text: 'PARAMS_TEXT',
	boolean: 'PARAMS_BOOLEAN',
	choice: 'PARAMS_CHOICE',
	password: 'PARAMS_PASSWORD'
};

export const PIPELINE_ACTION_TYPES = {
	discover_branches: 'Discover Branches',
	discover_tags: 'Discover Tag Branches',
	discover_pr_from_origin: 'Discover PR from Origin',
	discover_pr_from_forks: 'Discover PR from Forks'
};

export const TIMETRIGGERINTERVALS = [
	{ label: '1 minute', value: '60000' },
	{ label: '2 minutes', value: '120000' },
	{ label: '5 minutes', value: '300000' },
	{ label: '10 minutes', value: '600000' },
	{ label: '15 minutes', value: '900000' },
	{ label: '20 minutes', value: '1200000' },
	{ label: '25 minutes', value: '1500000' },
	{ label: '30 minutes', value: '1800000' },
	{ label: '1 hour', value: '3600000' },
	{ label: '2 hours', value: '7200000' },
	{ label: '4 hours', value: '14400000' },
	{ label: '8 hours', value: '28800000' },
	{ label: '12 hours', value: '43200000' },
	{ label: '1 day', value: '86400000' },
	{ label: '2 days', value: '172800000' },
	{ label: '1 week', value: '604800000' },
	{ label: '2 weeks', value: '1209600000' },
	{ label: '4 weeks', value: '2419200000' }
];

export const SECRET_TYPES = {
	Opaque: 'DEFAULT',
	'kubernetes.io/tls': 'TLS_INFORMATION',
	'kubernetes.io/dockerconfigjson': 'IMAGE_REGISTRY_INFORMATION',
	'kubernetes.io/basic-auth': 'USERNAME_AND_PASSWORD'
};

export const STRATEGIES = {
	deployments: [
		{
			label: 'SIMULTANEOUS_UPDATE',
			value: 'Recreate',
			description: 'SIMULTANEOUS_UPDATE_DESC'
		},
		{
			label: 'ROLLING_UPDATE_RECOMMENDED',
			value: 'RollingUpdate',
			description: 'ROLLINGUPDATE_DESC'
		}
	],
	daemonsets: [
		{ label: 'ONDELETE', value: 'OnDelete', description: 'ONDELETE_DESC' },
		{
			label: 'ROLLING_UPDATE_RECOMMENDED',
			value: 'RollingUpdate',
			description: 'ROLLINGUPDATE_DESC'
		}
	],
	statefulsets: [
		{ label: 'ONDELETE', value: 'OnDelete', description: 'ONDELETE_DESC' },
		{
			label: 'ROLLING_UPDATE_RECOMMENDED',
			value: 'RollingUpdate',
			description: 'ROLLINGUPDATE_DESC'
		}
	]
};

export const SYNC_STRATEGY = [
	{
		label: 'AUTO_SYNC',
		value: 'automated',
		description: 'AUTO_SYNC_DESC'
	},
	{
		label: 'MANUAL_SYNC',
		value: 'manual',
		description: 'MANUAL_SYNC_DESC'
	}
];

export const PROTOCOLS = [
	{ label: 'GRPC', value: 'GRPC' },
	{ label: 'HTTP', value: 'HTTP' },
	{ label: 'HTTP2', value: 'HTTP2' },
	{ label: 'HTTPS', value: 'HTTPS' },
	{ label: 'MONGO', value: 'MONGO' },
	{ label: 'REDIS', value: 'REDIS' },
	{ label: 'TCP', value: 'TCP' },
	{ label: 'TLS', value: 'TLS' },
	{ label: 'UDP', value: 'UDP' }
];

export const STRATEGIES_PREFIX = {
	deployments: 'spec.strategy',
	statefulsets: 'spec.updateStrategy',
	daemonsets: 'spec.updateStrategy'
};

export const LANG_MAP = {
	zh: 'zh-CN',
	en: 'en'
};

export const GRAY_RELEASE_CATEGORIES = [
	{
		icon: 'blue-green-deployment',
		type: 'Bluegreen',
		title: 'BLUE_GREEN_DEPLOYMENT',
		desc: 'BLUE_GREEN_DEPLOYMENT_DESC'
	},
	{
		icon: 'bird',
		type: 'Canary',
		title: 'CANARY_RELEASE',
		desc: 'CANARY_RELEASE_DESC'
	},
	{
		icon: 'mirroring',
		type: 'Mirror',
		title: 'TRAFFIC_MIRRORING',
		desc: 'TRAFFIC_MIRRORING_DESC'
	}
];

export const PATTERN_NAME = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;
export const PATTERN_USER_NAME =
	/^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
export const PATTERN_IMAGE_NAME =
	/^[a-z0-9]([-a-z0-9]*[a-z0-9])?((\.|\/)[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:([a-z0-9])*)?$/;
export const PATTERN_SERVICE_NAME = /^[a-z]([-a-z0-9]*[a-z0-9])?$/;
export const PATTERN_SERVICE_VERSION = /^[a-z0-9]*$/;
export const PATTERN_LABEL = /(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/;
export const PATTERN_TAG = /^[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?$/;
export const PATTERN_PASSWORD = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,64}$/;
export const PATTERN_IMAGE = /^\S+$/;
export const PATTERN_PORT_NAME = /^[a-z]([-a-z0-9]*[a-z0-9])?$/;
export const PATTERN_COMPONENT_VERSION = /^[a-z0-9]+$/;
export const PATTERN_PIPELINE_NAME =
	/[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/;
export const PATTERN_HOST =
	/^(?=^.{3,255}$)[*a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/;

export const PATTERN_URL =
	/^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)\/?$/;
export const PATTERN_EMAIL = /\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*/;
export const PATTERN_IP =
	/^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$/;
export const PATTERN_IP_MASK = /^[1-9][0-9]*$/;
export const PATTERN_PORT =
	/^([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5])$/;
export const PATTERN_IMAGE_TAG = /^(.*?)([0-9a-zA-Z/]*)(:[-.\w]*[0-9a-zA-Z])*$/;
export const PATTERN_APPTEMPLATE_VERSION =
	/[a-zA-Z0-9](\.?-?[a-zA-Z0-9])+(\s?\[?[a-zA-Z0-9]+\.?-?\]?)*/;
export const PATTERN_UTC_TIME =
	/^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+Z/;
export const PATTERN_ENV_NAME = /^[-._a-zA-Z][-._a-zA-Z0-9]*$/;

export const PIPELINE_CONDITIONS = [
	'branch',
	'environment',
	'expression',
	'not',
	'allOf',
	'anyOf'
];

export const KIND_GROUP_MAP = {
	Deployment: 'apps',
	StatefulSet: 'apps',
	Service: ''
};

export const GRAY_RELEASE_CANARY_CONTENT = [
	{
		label: 'Windows',
		icon: 'windows',
		value: '(Windows NT ([\\d.])+)'
	},
	{ label: 'Linux', icon: 'ubuntu', value: '(Linux )' },
	{ label: 'macOS', icon: 'macos', value: '( Mac OS X ([\\d.])+)' },
	{ label: 'Android', icon: 'android', value: '(Android)' },
	{ label: 'iOS', icon: 'ios', value: '(OS [\\d.]+)' }
];

export const S2IPULLPOLICY_MAP = {
	Always: 'always',
	IfNotPresent: 'if-not-present',
	Never: 'never'
};

export const TRACING_COLORS_HEX = [
	'#17B8BE',
	'#F8DCA1',
	'#B7885E',
	'#FFCB99',
	'#F89570',
	'#829AE3',
	'#E79FD5',
	'#1E96BE',
	'#89DAC1',
	'#B3AD9E',
	'#12939A',
	'#DDB27C',
	'#88572C',
	'#FF9833',
	'#EF5D28',
	'#162A65',
	'#DA70BF',
	'#125C77',
	'#4DC19C',
	'#776E57'
];

export const TIME_MICROSECOND_MAP = {
	'1h': 3600000,
	'2h': 7200000,
	'3h': 10800000,
	'6h': 21600000,
	'12h': 43200000,
	'24h': 86400000,
	'2d': 172800000
};

export const COLORS_MAP = {
	white: '#fff',
	light: '#f9fbfd',
	lightest: '#f9fbfd',
	dark: '#242e42',
	grey: '#e3e9ef',
	green: '#55bc8a',
	blue: '#329dce',
	red: '#ca2621',
	yellow: '#f5a623',
	darkerGreen: '#479e88',
	darkerBlue: '#3385b0',
	darkerRed: '#ab2f29',
	darkerYellow: '#e0992c',
	darkestGreen: '#3b747a',
	darkestBlue: '#326e93',
	darkestRed: '#8c3231',
	darkestYellow: '#8d663e',
	lighterGreen: '#a2d8bb',
	lighterBlue: '#7eb8dc',
	lighterRed: '#ea8573',
	lighterYellow: '#ffc781',
	lightestGreen: '#c4e6d4',
	lightestBlue: '#c7deef',
	lightestRed: '#fae7e5',
	lightestYellow: '#ffe1be'
};

export const MILLISECOND_IN_TIME_UNIT = {
	s: 1000,
	m: 60000,
	h: 3600000,
	d: 86400000,
	w: 604800000
};

export const PROVISIONERS = [
	{
		label: 'QingCloud CSI',
		value: 'disk.csi.qingcloud.com',
		icon: 'qingcloud',
		access_modes: ['ReadWriteOnce'],
		params: [
			{
				key: 'TYPE',
				type: 'input',
				desc: 'QINGCLOUD_CSI_TYPE_DESC'
			},
			{
				key: 'MAXSIZE',
				type: 'input',
				desc: 'CREATE_VOLUME_MAX_SIZE'
			},
			{
				key: 'STEPSIZE',
				type: 'input',
				desc: 'CREATE_VOLUME_STEP_SIZE'
			},
			{
				key: 'MINSIZE',
				type: 'input',
				desc: 'CREATE_VOLUME_MIN_SIZE'
			},
			{
				key: 'FSTYPE',
				type: 'input',
				defaultValue: 'ext4',
				desc: 'VOLUME_FS_TYPE'
			},
			{
				key: 'TAGS',
				type: 'input',
				desc: 'QINGCLOUD_VOLUME_TAGS_DESC'
			}
		],
		description: 'QINGCLOUD_CSI_DESC'
	},
	{
		label: 'GlusterFS',
		value: 'kubernetes.io/glusterfs',
		icon: 'glusterfs',
		access_modes: ['ReadWriteOnce', 'ReadOnlyMany', 'ReadWriteMany'],
		params: [
			{
				key: 'RESTURL',
				_key: 'resturl',
				type: 'input',
				placeholder:
					'https://<Heketi service cluster IP address>:<Heketi service port>',
				desc: 'GLUSTERFS_RESTURL_DESC'
			},
			{
				key: 'CLUSTER_ID',
				_key: 'clusterid',
				type: 'input',
				desc: 'GLUSTERFS_ID_DESC'
			},
			{
				key: 'REST_AUTH_ENABLED',
				_key: 'restauthenabled',
				type: 'select',
				defaultValue: 'true',
				options: [
					{ label: 'TRUE', value: 'true' },
					{ label: 'FALSE', value: 'false' }
				],
				desc: 'GLUSTERFS_RESTAUTHENABLED_DESC'
			},
			{
				key: 'REST_USER',
				_key: 'restuser',
				type: 'input',
				placeholder: 'admin',
				desc: 'GLUSTERFS_RESTUSER_DESC'
			},
			{
				key: 'SECRET_NAMESPACE',
				_key: 'secretNamespace',
				type: 'input',
				desc: 'GLUSTERFS_SECRET_NAMESPACE_DESC'
			},
			{
				key: 'SECRET_NAME',
				_key: 'secretName',
				type: 'input',
				desc: 'GLUSTERFS_SECRET_NAME_DESC'
			},
			{
				key: 'GID_MIN',
				_key: 'gidMin',
				type: 'input',
				placeholder: '2000-2147483647',
				desc: 'GLUSTERFS_GID_MIN_DESC'
			},
			{
				key: 'GID_MAX',
				_key: 'gidMax',
				type: 'input',
				placeholder: '2000-2147483647',
				desc: 'GLUSTERFS_GID_MAX_DESC'
			},
			{
				key: 'VOLUME_TYPE',
				_key: 'volumetype',
				type: 'input',
				desc: 'GLUSTERFS_VOLUME_TYPE_DESC'
			}
		]
	},
	{
		label: 'Ceph RBD',
		value: 'ceph.com/rbd',
		icon: 'ceph',
		access_modes: ['ReadWriteOnce', 'ReadOnlyMany'],
		params: [
			{
				key: 'monitors',
				type: 'input',
				placeholder: 'CEPH_MONITOR_IP',
				desc: 'CEPHRBD_MONITORS_DESC'
			},
			{
				key: 'adminId',
				type: 'input',
				placeholder: 'admin',
				desc: 'CEPHRBD_ADMIN_ID_DESC'
			},
			{
				key: 'adminSecretName',
				type: 'input',
				placeholder: 'Secret Name for adminId',
				desc: 'CEPHRBD_ADMIN_SECRET_NAME_DESC'
			},
			{
				key: 'adminSecretNamespace',
				type: 'input',
				c: 'default',
				desc: 'CEPHRBD_ADMIN_SECRET_NAMESPACE_DESC'
			},
			{
				key: 'pool',
				type: 'input',
				placeholder: 'rbd',
				desc: 'CEPHRBD_POOL_DESC'
			},
			{
				key: 'userId',
				type: 'input',
				desc: 'CEPHRBD_USERID_DESC'
			},
			{
				key: 'userSecretName',
				type: 'input',
				placeholder: 'rbd',
				desc: 'CEPHRBD_USER_SECRET_NAME_DESC'
			},
			{
				key: 'userSecretNamespace',
				type: 'input',
				desc: 'CEPHRBD_USER_SECRET_NAMESPACE_DESC'
			},
			{
				key: 'fsType',
				type: 'input',
				defaultValue: 'ext4',
				desc: 'CEPHRBD_FS_TYPE_DESC'
			},
			{
				key: 'imageFormat',
				type: 'input',
				defaultValue: '2',
				desc: 'CEPHRBD_IMAGE_FORMAT_DESC'
			},
			{
				key: 'imageFeatures',
				type: 'input',
				defaultValue: 'layering',
				desc: 'CEPHRBD_IMAGE_FEATURES_DESC'
			}
		]
	}
];

export const S2I_SUPPORTED_TYPES = ['java', 'nodejs', 'python'];
export const B2I_SUPPORTED_TYPES = ['jar', 'war', 'binary'];

export const MAX_SIZE_UPLOAD = 2 * 1024 * 1024;

export const SCHEME_OPTIONS = [
	{
		label: 'http://',
		value: 'http://'
	},
	{
		label: 'https://',
		value: 'https://'
	}
];

export const SCHEME_REG = /^(http(s)?:\/\/)?(.*)/;

export const LIST_DEFAULT_ORDER = {
	deployments: 'updateTime',
	jobs: 'updateTime',
	pods: 'startTime'
};

export const API_VERSIONS = {
	deployments: 'apis/apps/v1',
	statefulsets: 'apis/apps/v1',
	daemonsets: 'apis/apps/v1',
	jobs: 'apis/batch/v1',
	cronjobs: 'apis/batch/v1beta1',
	pods: 'api/v1',
	namespaces: 'api/v1',
	services: 'api/v1',
	volumes: 'api/v1',
	secrets: 'api/v1',
	configmaps: 'api/v1',
	serviceaccounts: 'api/v1',
	events: 'api/v1',
	resourcequotas: 'api/v1',
	limitranges: 'api/v1',
	persistentvolumeclaims: 'api/v1',
	ingresses: 'apis/networking.k8s.io/v1',
	persistentvolumes: 'api/v1',
	nodes: 'api/v1',
	storageclasses: 'apis/storage.k8s.io/v1',
	accessors: 'apis/storage.kubesphere.io/v1alpha1',
	validatingwebhookconfigurations: '/apis/admissionregistration.k8s.io/v1',
	roles: 'apis/rbac.authorization.k8s.io/v1',
	clusterroles: 'apis/rbac.authorization.k8s.io/v1',
	applications: 'apis/app.k8s.io/v1beta1',
	strategies: 'apis/servicemesh.kubesphere.io/v1alpha2',
	servicepolicies: 'apis/servicemesh.kubesphere.io/v1alpha2',
	horizontalpodautoscalers: 'apis/autoscaling/v2beta2',
	customresourcedefinitions: 'apis/apiextensions.k8s.io/v1',
	clusters: 'apis/cluster.kubesphere.io/v1alpha1',
	workspaces: 'apis/tenant.kubesphere.io/v1alpha2',
	users: 'apis/iam.kubesphere.io/v1alpha2',
	globalroles: 'apis/iam.kubesphere.io/v1alpha2',
	workspaceroles: 'apis/iam.kubesphere.io/v1alpha2',
	dashboards: 'apis/monitoring.kubesphere.io/v1alpha2',
	clusterdashboards: 'apis/monitoring.kubesphere.io/v1alpha2',
	namespacenetworkpolicies: 'apis/network.kubesphere.io/v1alpha1',
	networkpolicies: 'apis/networking.k8s.io/v1',
	ippools: 'apis/network.kubesphere.io/v1alpha1',
	storageclasscapabilities: 'apis/storage.kubesphere.io/v1alpha1',
	meter: 'kapis/metering.kubesphere.io/v1alpha1'
};

export const COMPONENT_ICON_MAP = {
	kubernetes: 'kubernetes',
	kubesphere: 'kubesphere',
	istio: 'istio',
	openpitrix: 'openpitrix',
	devops: 'jenkins',
	logging: 'record',
	monitoring: 'monitor',
	alerting: 'loudspeaker',
	auditing: 'login-servers',
	events: 'thunder',
	notification: 'mail',
	servicemesh: 'istio',
	metrics_server: 'monitor'
};

export const CLUSTER_PROVIDER_ICON = {
	'Aliyun ACK': 'aliyun',
	'Azure Kubernetes Service': 'windows',
	'Huawei Cloud CCE': 'kubernetes',
	'Amazon EKS': 'aws',
	'Google Kubernetes Engine': 'google-plus',
	'QingCloud Kubernetes Engine': 'qingcloud',
	'Tencent Kubernetes Engine': 'kubernetes',
	kubesphere: 'kubernetes'
};

export const CLUSTER_PROVIDERS = [
	{
		label: 'Alibaba Cloud ACK',
		value: 'Aliyun ACK',
		icon: 'aliyun'
	},
	{
		label: 'Azure Kubernetes Service',
		value: 'Aure Kubernetes Service',
		icon: 'windows'
	},
	{
		label: 'Huawei Cloud CCE',
		value: 'Huawei Cloud CCE',
		icon: 'kubernetes'
	},
	{
		label: 'Amazon EKS',
		value: 'Amazon EKS',
		icon: 'aws'
	},
	{
		label: 'Google Kubernetes Engine',
		value: 'Google Kubernetes Engine',
		icon: 'google-plus'
	},
	{
		label: 'QingCloud Kubernetes Engine',
		value: 'QingCloud Kubernetes Engine',
		icon: 'qingcloud'
	},
	{
		label: 'Tencent Kubernetes Engine',
		value: 'Tencent Kubernetes Engine',
		icon: 'kubernetes'
	}
];

export const CLUSTER_PRESET_GROUPS = [
	{
		label: 'production',
		value: 'production'
	},
	{
		label: 'development',
		value: 'development'
	},
	{
		label: 'testing',
		value: 'testing'
	},
	{
		label: 'demo',
		value: 'demo'
	}
];

export const CLUSTER_GROUP_TAG_TYPE = {
	production: 'warning',
	development: 'default',
	testing: 'info',
	demo: 'primary'
};

export const ROLE_QUERY_KEY = {
	globalroles: 'globalrole',
	workspaceroles: 'workspacerole',
	clusterroles: 'clusterrole',
	roles: 'role'
};

export const DEFAULT_CLUSTER = {
	apiVersion: 'cluster.kubesphere.io/v1alpha1',
	kind: 'Cluster',
	metadata: {
		annotations: {
			'cluster.kubesphere.io/is-host-cluster': 'true'
		},
		labels: {
			'cluster.kubesphere.io/visibility': 'public'
		},
		// TODO: name: globals?.config?.defaultClusterName || 'default'
		name: 'default'
	}
};

export const CREDENTIAL_KEY = {
	username_password: 'basic-auth',
	ssh: 'ssh-auth',
	secret_text: 'secret-text',
	kubeconfig: 'kubeconfig'
};

export const CREDENTIAL_TYPE_LIST = [
	'credential.devops.kubesphere.io/basic-auth',
	'credential.devops.kubesphere.io/ssh-auth',
	'credential.devops.kubesphere.io/secret-text',
	'credential.devops.kubesphere.io/kubeconfig'
];

export const CREDENTIAL_DISPLAY_KEY = {
	'basic-auth': 'username_password',
	'ssh-auth': 'ssh',
	'secret-text': 'secret_text',
	kubeconfig: 'kubeconfig'
};

export const FED_ACTIVE_STATUS = {
	FederatedApplication: 'Running',
	FederatedService: 'Running',
	FederatedDeployment: 'Running',
	FederatedStatfulSet: 'Running',
	FederatedVolume: 'Running',
	FederatedIngress: 'Running'
};

export const APP_LABEL_MODULES = [
	'applications',
	'deployments',
	'statefulsets',
	'daemonsets',
	'jobs',
	'cronjobs',
	'services'
];

export const NODE_ROLE_TAG_TYPE = {
	master: 'secondary',
	worker: 'default'
};

export const OMIT_TOTAL_REPLICAS = (kind) => [
	`${kind}.spec.template.totalReplicas`,
	'totalReplicas',
	`${kind}.totalReplicas`,
	`${kind}.Deployment.totalReplicas`,
	`${kind}.workload.totalReplicas`
];

export const PAGESIZE_OPTION = [10, 20, 50, 100];
