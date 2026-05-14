export enum BackupLocation {
	S3 = 's3',
	TerminusCloud = 'terminus-cloud'
}

export enum BackupFrequency {
	Daily = '@daily',
	Weekly = '@weekly',
	Monthly = '@monthly',
	Yearly = '@yearly'
}

export interface BackupConfig {
	region?: string;
	bucket?: string;
	s3Url: string;
	prefix: string;
	accessKey?: string;
	secretKey?: string;
}

export interface BackupPolicy {
	name?: string;
	snapshotFrequency?: BackupFrequency;
	timesOfDay?: string;
	enabled: boolean;
	dayOfWeek: number;
}

export interface BackupCreateReq {
	name: string;
	location: BackupLocation;
	config?: BackupConfig;
	backupPolicies: BackupPolicy;
	password: string;
	confirmPassword: string;
}

export interface BackupUpdateReq {
	config?: BackupConfig;
	backupPolicies: BackupPolicy;
	password: string;
	confirmPassword: string;
}

export interface BackupPlanItem {
	backupPolicies: BackupPolicy;
	name: string;
	size: number;
}

export interface BackupPlanItems {
	total: number;
	items: BackupPlanItem[];
}

export enum BackupSnapshotPhase {
	Pending = 'Pending',
	Running = 'Running',
	Failed = 'Failed',
	Canceled = 'Canceled',
	Started = 'Started',
	Succeed = 'Succeed'
}

export interface BackupSnapshot {
	name: string;
	creationTimestamp: number;
	size: number;
	phase: BackupSnapshotPhase;
}

export interface BackupSnapshots {
	total: number;
	items: BackupSnapshot[];
}

export enum BackupSnapshotType {
	Fully = 'fully',
	Incremental = 'incremental'
}

export interface BackupSnapshotDetail {
	name: string;
	creationTimestamp: number;
	size: number;
	phase: BackupSnapshotPhase;
	failedMessage: string;
	owner: string;
	backupType: BackupSnapshotType;
	repositoryPasswordHash: string;
}
