import { ShareMember, SMBUser } from 'src/stores/files';
import { appendPath } from '../path';
import { commonUrlPrefix } from './utils';
import { CommonFetch } from '../../fetch';
import {
	SharePermission,
	ShareResult,
	ShareType
} from 'src/utils/interface/share';
import { encodeUrl } from 'src/utils/encode';

export const getShareList = async (params: {
	shared_to_me: boolean;
	shared_by_me: boolean;
	expire_in?: number;
	expire_over?: number;
	share_type?: string;
	owner: string;
	permission: string;
}) => {
	const url = appendPath(commonUrlPrefix('share'), 'share_path', '/');
	return await CommonFetch.get(url, {
		params
	});
};

const create = async (
	file: {
		fileType?: string | undefined;
		fileExtend: string;
		oPath?: string | undefined;
	},
	options: {
		name: string;
		share_type: ShareType;
		permission: SharePermission;
		password: string;
		expire_in?: number;
		expire_time?: string;
		users?: { id: string; permission: SharePermission }[];
		public_smb?: boolean;
		upload_size_limit?: number;
	}
): Promise<ShareResult> => {
	const url = appendPath(
		commonUrlPrefix('share'),
		'share_path',
		file.fileType || '',
		file.fileExtend,
		encodeUrl(file.oPath || '/'),
		'/'
	);
	return (await CommonFetch.post(url, options)).data.data;
};

const addMember = async (
	path_id: string,
	members: { share_member: string; permission: SharePermission }[]
) => {
	const url = appendPath(commonUrlPrefix('share'), 'share_member', '/');

	return await CommonFetch.post(url, {
		path_id: path_id,
		share_members: members
	});
};

const getMembers = async (path_id: string): Promise<ShareMember[]> => {
	const url = appendPath(commonUrlPrefix('share'), 'share_member', '/');
	return (
		await CommonFetch.get(url, {
			params: { path_id: path_id }
		})
	).share_members;
};

const remove = async (path_ids: string) => {
	const url = appendPath(commonUrlPrefix('share'), 'share_path', '/');
	return await CommonFetch.delete(url, {
		params: { path_ids: path_ids }
	});
};

const query = async (path_id: string): Promise<ShareResult | undefined> => {
	const url = appendPath(commonUrlPrefix('share'), 'share_path', '/');
	const result = await CommonFetch.get(url, {
		params: { path_id: path_id }
	});
	if (result.share_paths) {
		return result.share_paths[0];
	}
	return undefined;
};

export const getShareToken = async (id: string, password: string) => {
	try {
		const result = await CommonFetch.post(
			appendPath(commonUrlPrefix('share'), 'get_token', '/'),
			{
				id: id,
				pass: password
			}
		);
		if (result.data && result.data.code == 0) {
			return {
				code: 0,
				token: result.data.data,
				message: ''
			};
		}
		return {
			code: result.data.code || -1,
			message: result.data.message || 'Unknown error',
			token: ''
		};
	} catch (error) {
		return undefined;
	}
};

export const getShare = async (path_id: string, token: string) => {
	try {
		const result = await CommonFetch.get(
			appendPath(commonUrlPrefix('share'), 'get_share', '/'),
			{
				params: {
					path_id,
					token
				}
			}
		);
		return result.data;
	} catch (error) {
		return undefined;
	}
};

export const getSMBUsers = async () => {
	try {
		const result = await CommonFetch.get(
			appendPath(commonUrlPrefix('share'), 'smb_share_user', '/'),
			{}
		);
		return result.data as SMBUser[];
	} catch (error) {
		return [];
	}
};

export const createSMBUser = async (user: string, password: string) => {
	try {
		const result = await CommonFetch.post(
			appendPath(commonUrlPrefix('share'), 'smb_share_user', '/'),
			{
				user,
				password
			}
		);
		if (result.status && result.data.code == 0) {
			return true;
		}
		return false;
	} catch (error) {
		// return [];
		return false;
	}
};

export const getShareByFile = async <T>(
	file: {
		fileType?: string | undefined;
		fileExtend: string;
		oPath?: string | undefined;
	},
	share_type: ShareType.SMB | ShareType.INTERNAL
): Promise<T | undefined | null> => {
	//
	const url = appendPath(
		commonUrlPrefix('share'),
		'get_share_internal_smb',
		file.fileType || '',
		file.fileExtend,
		encodeUrl(file.oPath || '/'),
		'/'
	);
	try {
		const result = await CommonFetch.get(url, {
			params: {
				share_type
			}
		});
		return result.data;
	} catch (error) {
		return undefined;
	}
};

export const updateSMBShareMember = async (
	path_id: string,
	users: { id: string; permission: SharePermission }[],
	public_smb: boolean
) => {
	const url = appendPath(commonUrlPrefix('share'), 'smb_share_member', '/');

	return await CommonFetch.post(url, {
		path_id,
		users,
		public_smb
	});
};

export const resetPassword = async (path_id: string, password: string) => {
	const url = appendPath(commonUrlPrefix('share'), 'share_password', '/');
	return await CommonFetch.put(url, {
		path_id,
		password
	});
};

export const updateInternalShareMembers = async (
	path_id: string,
	share_members: { share_member: string; permission: SharePermission }[]
) => {
	const url = appendPath(
		commonUrlPrefix('share'),
		'share_path',
		'share_members',
		'/'
	);

	return await CommonFetch.put(url, {
		path_id,
		share_members
	});
};

export default {
	getShareList,
	create,
	addMember,
	remove,
	getMembers,
	query,
	getShareToken,
	getShare,
	getSMBUsers,
	createSMBUser,
	getShareByFile,
	updateSMBShareMember,
	resetPassword,
	updateInternalShareMembers
};
