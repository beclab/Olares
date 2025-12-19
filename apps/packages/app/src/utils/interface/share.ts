export enum ShareType {
	INTERNAL = 'internal',
	PUBLIC = 'external',
	SMB = 'smb'
}

export enum SharePermission {
	EMPTY = 0,
	View = 1,
	UploadOnly = 2,
	Edit = 3,
	ADMIN = 4
}

export interface ShareResult {
	id: string;
	owner: string;
	file_type: string;
	extend: string;
	path: string;
	share_type: ShareType;
	name: string;
	expire_in: number;
	expire_time: string;
	permission: SharePermission;
	create_time: string;
	update_time: string;
	shared_by_me: boolean;
	smb_link?: string;
	smb_user?: string;
	smb_password?: string;
	upload_size_limit?: number;
	sync_repo_name: string;
	node?: string;
}
