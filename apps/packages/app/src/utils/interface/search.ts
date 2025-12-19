import { DriveType } from './files';

export interface TextSearchItem {
	id: number | string;
	author?: string;
	content?: string;
	// meta: any;
	highlight?: string | string[];
	highlight_field: string | string[];
	owner_userid?: string;
	title: string;
	resource_uri?: string;
	fileType: string;
	fileIcon: string;
	repo_id?: string;
	path: string;
	isDir: boolean;
	name?: string;
	driveType?: DriveType;
	meta?: {
		extend: string;
		file_type: string;
		is_dir: boolean;
		path: string;
		image_url: string;
		published_at: number;
		updated: number;
		id: string;
	};
	// created_at: number;
}

export enum ServiceType {
	Files = 'files',
	Knowledge = 'knowledge',
	Sync = 'sync',
	FilesV2 = 'files_v2'
}

export enum SearchType {
	HomePage = 'home',
	FilesPage = 'Files Search',
	AshiaPage = 'Ashia',
	AshiaDocPage = 'Ashia Doc',
	TextSearch = 'Text Search'
}

export enum SearchCategory {
	Suggestion = 'Suggestion',
	Command = 'Command',
	Application = 'Application',
	Result = 'Result',
	Use = 'Use'
}

export interface ServiceParamsType {
	query: string;
	serviceType: ServiceType;
	limit: number;
	offset: number;
	repo_id?: string;
}

export enum SearchV2Type {
	FILE_NAME = 'file_name',
	AGGREGATE = 'aggregate'
}

export interface AppClickInfo {
	appid: string;
	data: any;
	path?: string;
}
