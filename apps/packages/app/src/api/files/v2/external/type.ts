export type { DriveMenuType } from '../drive/type';

export interface SmbMountType {
	password?: string;
	timestamp?: number;
	url: string;
	username?: string;
}

export const normalExternalTypes = ['internal'];
