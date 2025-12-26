import { defineStore } from 'pinia';
import axios from 'axios';
// import { TerminusInfo, DefaultTerminusInfo } from '@bytetrade/core';
import { Env } from 'src/utils/rss-types';

const fields =
	'metadata,items.metadata.uid,items.metadata.name,items.metadata.namespace,items.metadata.creationTimestamp,items.metadata.labels,items.metadata.annotations,items.status.phase,items.status.message,items.status.finishedAt,items.status.startedAt,items.status.estimatedDuration,items.status.progress,items.spec.suspend';

interface CronWorkflowMetadata {
	name: string;
	namespace: string;
	uid: string;
	resourceVersion: string;
	generation: number;
	creationTimestamp: string;
	labels: {
		[key: string]: string;
	};
	annotations: {
		[key: string]: string;
	};
	managedFields: {
		manager: string;
		operation: string;
		apiVersion: string;
		time: string;
		fieldsType: string;
		fieldsV1: {
			[key: string]: any;
		};
	}[];
}

interface CronWorkflowSpec {
	workflowSpec: {
		templates: {
			name: string;
			inputs: any;
			outputs: any;
			metadata: any;
			steps: [
				{
					name: string;
					template: string;
					arguments: any;
				}[]
			];
		}[];
		entrypoint: string;
		arguments: any;
		volumes: {
			name: string;
			hostPath: {
				path: string;
				type: string;
			};
		}[];
	};
	schedule: string;
	concurrencyPolicy: string;
	startingDeadlineSeconds: number;
	successfulJobsHistoryLimit: number;
	failedJobsHistoryLimit: number;
}

interface CronWorkflowStatus {
	active: [
		{
			kind: string;
			namespace: string;
			name: string;
			uid: string;
			apiVersion: string;
			resourceVersion: string;
		}
	];
	lastScheduledTime: string;
	conditions: null;
}

interface CronWorkflowData {
	metadata: CronWorkflowMetadata;
	spec: CronWorkflowSpec;
	status: CronWorkflowStatus;
}

interface MetaData {
	resourceVersion: string;
}

interface GetCronWorkflowResponse {
	items: CronWorkflowData[];
	metadata: MetaData;
}

interface WorkflowMetadata {
	name: string;
	namespace: string;
	uid: string;
	creationTimestamp: string;
	labels: {
		'workflows.argoproj.io/creator': string;
		'workflows.argoproj.io/phase': string;
		'workflows.argoproj.io/cron-workflow': string;
		[key: string]: string;
	};
	annotations: {
		'workflows.argoproj.io/pod-name-format': string;
		'workflows.argoproj.io/scheduled-time': string;
		[key: string]: string;
	};
}

interface WorkflowSpec {
	arguments: {
		[key: string]: any;
	};
}

interface WorkflowStatus {
	phase: string;
	startedAt: string;
	finishedAt: string | null;
	message: string;
	progress: string;
}

interface WorkflowDetailMetadata {
	name: string;
	namespace: string;
	uid: string;
	resourceVersion: string;
	generation: number;
	creationTimestamp: string;
	labels: {
		[key: string]: string;
	};
	annotations: {
		[key: string]: string;
	};
	ownerReferences: OwnerReference[];
	managedFields: ManagedField[];
}

interface OwnerReference {
	apiVersion: string;
	kind: string;
	name: string;
	uid: string;
	controller: boolean;
	blockOwnerDeletion: boolean;
}

interface ManagedField {
	manager: string;
	operation: string;
	apiVersion: string;
	time: string;
	fieldsType: string;
	fieldsV1: {
		'f:metadata': {
			'f:annotations': any;
			'f:labels:': any;
			'f:ownerReferences': any;
			[key: string]: any;
		};
		'f:spec': any;
		'f:status': any;
	};
}

interface WorkflowDetailSpec {
	templates: Template[];
	entrypoint: string;
	arguments: any;
	volumes: Volume[];
}

export interface Container {
	name: string;
	image: string;
	env: Env[];
	resources: Record<string, unknown>;
	volumeMounts: VolumeMount[];
	imagePullPolicy: string;
}

interface VolumeMount {
	name: string;
	mountPath: string;
}

interface Template {
	name: string;
	inputs: any;
	outputs: any;
	metadata: any;
	steps: Step[][];
	container?: Container;
}

type Step = {
	name: string;
	template: string;
	arguments: any;
};

interface Volume {
	name: string;
	hostPath: {
		path: string;
		type: string;
	};
}

interface WorkflowDetailStatus {
	phase: string;
	startedAt: string;
	finishedAt: string;
	estimatedDuration: number;
	progress: string;
	nodes: {
		[key: string]: NodeStatus;
	};
	conditions: Condition[];
	resourcesDuration: {
		cpu: number;
		memory: number;
	};
	artifactRepositoryRef: {
		default: boolean;
		artifactRepository: any;
	};
	artifactGCStatus: {
		notSpecified?: boolean;
	};
}

export interface NodeStatus {
	id: string;
	name: string;
	displayName?: string;
	type?: string;
	templateName?: string;
	templateScope?: string;
	phase?: string;
	startedAt?: string;
	finishedAt?: string;
	estimatedDuration?: number;
	progress?: string;
	resourcesDuration?: {
		cpu?: number;
		memory?: number;
	};
	children: string[];
	outboundNodes: string[];
}

interface Condition {
	type: string;
	status: string;
}

export interface WorkflowDetail {
	metadata: WorkflowDetailMetadata;
	spec: WorkflowDetailSpec;
	status: WorkflowDetailStatus;
}

export interface Workflow {
	metadata: WorkflowMetadata;
	spec: WorkflowSpec;
	status: WorkflowStatus;
}

interface GetWorkflowResponse {
	items: Workflow[];
	metadata: MetaData;
}

export type ArgoState = {
	//
	cronLabel: string;
	namespace: string;
	workflow_id: string;
	// metadata: MetaData | null;
	// terminus_info: TerminusInfo;
	cron_workflows: CronWorkflowData[];
	cronWorkflowsLoading: boolean;
	workflowsLoading: boolean;
	workflows: Workflow[];
	url: string;
};

export const useArgoStore = defineStore('argo', {
	state: () => {
		return {
			//
			cronLabel: '',
			namespace: '',
			workflow_id: '',
			// terminus_info: DefaultTerminusInfo,
			cron_workflows: [],
			cronWorkflowsLoading: false,
			workflowsLoading: false,
			workflows: [],
			url: ''
			// metadata: null,
		} as ArgoState;
	},
	getters: {
		// namespace(): string {
		// return 'user-space-' + this.terminus_info.terminusName.split('@')[0];
		// },
	},
	actions: {
		setUrl(new_url: string) {
			this.url = new_url;
		},

		async get_cron_workflows() {
			this.cronWorkflowsLoading = true;
			const data: GetCronWorkflowResponse = await axios.get(
				this.url + '/api/v1/cron-workflows/'
			);
			// this.url + '/api/v1/cron-workflows/' + this.namespace

			console.log(data.items);
			this.cron_workflows = data.items;
			this.cronWorkflowsLoading = false;
			// this.metadata = data.metadata;
		},
		async get_workflows(
			namespace: string,
			label?: string,
			resourceVersion?: string
		) {
			this.workflowsLoading = true;
			const data: any = {
				'listOptions.limit': '50',
				fields
			};
			if (label) {
				data['listOptions.labelSelector'] =
					'workflows.argoproj.io/cron-workflow=' + label;
			}
			if (resourceVersion) {
				data['getOptions.resourceVersion'] = resourceVersion;
			}
			const params = new URLSearchParams(data);

			const res: GetWorkflowResponse = await axios.get(
				this.url + '/api/v1/workflows/' + namespace + '?' + params
			);
			console.log(res);
			this.workflows = res.items;
			this.workflowsLoading = false;
		},
		async get_workflow_detail(
			namespace: string,
			name: string,
			resourceVersion?: string
		): Promise<WorkflowDetail> {
			const data: any = {};
			if (resourceVersion) {
				data['getOptions.resourceVersion'] = resourceVersion;
			}
			const params = new URLSearchParams(data);
			const res: WorkflowDetail = await axios.get(
				this.url + '/api/v1/workflows/' + namespace + '/' + name + '?' + params
			);
			return res;
		},

		async getArchivedLog(namespace: string, uid: string, id: string) {
			return await axios.get(this.getArchivedLogUrl(namespace, uid, id));
		},

		getArchivedLogUrl(namespace: string, uid: string, id: string) {
			return (
				this.url +
				'/artifact-files/' +
				namespace +
				'/archived-workflows/' +
				uid +
				'/' +
				id +
				'/outputs/main-logs'
			);
		}
	}
});
