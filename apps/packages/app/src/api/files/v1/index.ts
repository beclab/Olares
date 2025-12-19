import * as files from './drive/utils';
import * as seahub from './sync/utils';
import * as shareToUser from './shareToUser';
import * as common from './common/common';

import Origin from './origin';
import DriveDataAPI from './drive/data';
import SyncDataAPI from './sync/data';
import DataDataAPI from './data/data';
import CacheDataAPI from './cache/data';
import GoogleDataAPI from './google/data';
import DropboxDataAPI from './dropbox/data';
import Awss3DataAPI from './awss3/data';
import TencentDataAPI from './tencent/data';
import ExternalDataAPI from './external/data';
import url from '../../../utils/url';

import { FilesIdType } from './../../../stores/files';
import { DriveType } from 'src/utils/interface/files';

class DriveAPI {
	private static driveInstances: Map<string, Origin> = new Map();

	private static readonly API_CONSTRUCTORS: Record<
		DriveType,
		new (id: number) => Origin
	> = {
		[DriveType.Sync]: SyncDataAPI,
		[DriveType.Drive]: DriveDataAPI,
		[DriveType.Data]: DataDataAPI,
		[DriveType.Cache]: CacheDataAPI,
		[DriveType.GoogleDrive]: GoogleDataAPI,
		[DriveType.Dropbox]: DropboxDataAPI,
		[DriveType.Awss3]: Awss3DataAPI,
		[DriveType.Tencent]: TencentDataAPI,
		[DriveType.External]: ExternalDataAPI,
		[DriveType.Share]: DriveDataAPI,
		[DriveType.PublicShare]: DriveDataAPI
	};

	public static getAPI(
		origin?: DriveType,
		originId: number = FilesIdType.PAGEID
	): Origin {
		const finalDriveType: DriveType =
			origin && this.API_CONSTRUCTORS[origin]
				? origin
				: this.getDriveTypeFromUrl();

		if (!this.driveInstances.has(finalDriveType + '_' + originId)) {
			const Constructor = this.API_CONSTRUCTORS[finalDriveType];
			this.driveInstances.set(
				finalDriveType + '_' + originId,
				new Constructor(originId)
			);
		}

		return this.driveInstances.get(finalDriveType + '_' + originId)!;
	}

	private static getDriveTypeFromUrl(): DriveType {
		const driveTypeStr = common.formatUrltoDriveType(url.getWindowPathname());
		if (driveTypeStr && Object.values(DriveType).includes(driveTypeStr)) {
			return driveTypeStr;
		}

		return DriveType.Drive;
	}
}

function dataAPIs(
	origin?: DriveType,
	originId: number = FilesIdType.PAGEID
): Origin {
	return DriveAPI.getAPI(origin, originId);
}

export {
	files,
	seahub,
	shareToUser,
	common,
	dataAPIs,
	DriveDataAPI,
	SyncDataAPI,
	DataDataAPI,
	CacheDataAPI,
	GoogleDataAPI,
	DropboxDataAPI,
	Awss3DataAPI,
	TencentDataAPI,
	ExternalDataAPI,
	DriveAPI
};
