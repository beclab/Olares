import { DriveType } from 'src/utils/interface/files';

export interface SyncRepoItemType {
	encrypted: boolean;
	last_modified: string;
	modifier_contact_email: string;
	modifier_email: string;
	modifier_name: string;
	monitored: boolean;
	owner_contact_email: string;
	owner_email: string;
	owner_name: string;
	permission: string;
	repo_id: string;
	repo_name: string;
	salt: string;
	size: number;
	starred: boolean;
	status: string;
	type: string;
}
export interface SyncRepoMineType extends SyncRepoItemType {
	share_type: any;
	repo_id: any;
	label: string;
	key: string;
	icon: string;
	id: string;
	defaultHide: boolean;
	driveType: DriveType;
}

export interface ShareInfoResType {
	shared_group_ids: string[];
	shared_user_emails: string[];
}

export interface SyncRepoSharedItemType {
	contact_email: string;
	encrypted: boolean;
	is_admin: boolean;
	modifier_contact_email: string;
	modifier_email: string;
	modifier_name: string;
	repo_id: string;
	repo_name: string;
	share_permission: string;
	share_type: string;
	user_email: string;
	user_name: string;
}

export interface SyncRepoSharedType extends SyncRepoSharedItemType {
	label: string;
	key: string;
	name: string;
	icon: string;
	id: string;
	defaultHide: boolean;
}

// export interface SyncRepoType {
// 	label: string;
// 	key: string;
// 	name: string;
// 	icon: string;
// 	id: string;
// 	defaultHide: boolean;
// }

export interface SyncRepoItemType {
	encrypted: boolean;
	last_modified: string;
	modifier_contact_email: string;
	modifier_email: string;
	modifier_name: string;
	monitored: boolean;
	owner_contact_email: string;
	owner_email: string;
	owner_name: string;
	permission: string;
	repo_id: string;
	repo_name: string;
	salt: string;
	size: number;
	starred: boolean;
	status: string;
	type: string;
}

export interface SyncResType {
	dir_id: string;
	user_perm: string;
	dirent_list: DirentListType[];
}

export interface DirentListType {
	encoded_thumbnail_src: string;
	id: string;
	modifier_contact_email: string;
	modifier_email: string;
	modifier_name: string;
	mtime: number;
	name: string;
	parent_dir: string;
	permission: string;
	size: number;
	starred: boolean;
	type: string;
}

export interface DeleteItemType {
	dirents: string[];
	parent_dir: string;
	repo_id: string;
}
