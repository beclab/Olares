import { FilesIdType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import OriginV2 from './v2/origin';
import OriginV1 from './v1/origin';
import * as DriveApiV1 from './v1';
import * as DriveApiV2 from './v2';
import { getApplication } from 'src/application/base';
import { useUserStore } from 'src/stores/user';
import SyncDataAPIV2 from './v2/sync/data';
import SyncDataAPIV1 from './v1/sync/data';

import * as filesV1 from './v1/drive/utils';
import * as seahubV1 from './v1/sync/utils';
import * as shareToUserV1 from './v1/shareToUser';
import * as commonV1 from './v1/common/common';

import * as filesV2 from './v2/drive/utils';
import * as seahubV2 from './v2/sync/utils';
import * as shareToUserV2 from './v2/shareToUser';
import * as commonV2 from './v2/common/common';

import * as utilV1 from './v1/utils';
import * as utilV2 from './v2/utils';

import * as syncUtilV1 from './v1/sync/utils';
import * as syncUtilV2 from './v2/sync/utils';

import * as syncFilesFormatV1 from './v1/sync/filesFormat';
import * as syncFilesFormatV2 from './v2/sync/filesFormat';

import * as ai from './ai';

function dataAPIs(
	origin?: DriveType,
	originId: number = FilesIdType.PAGEID
): OriginV1 | OriginV2 {
	if (useFilesVersion() == 'v1') {
		return DriveApiV1.DriveAPI.getAPI(origin, originId);
	}
	return DriveApiV2.DriveAPI.getAPI(origin, originId);
}

const useFilesVersion = () => {
	if (getApplication().platform) {
		const userStore = useUserStore();
		if (!userStore.current_user?.isLargeVersion12) {
			return 'v1';
		}
	}
	return 'v2';
};

const files = () => {
	if (useFilesVersion() == 'v1') {
		return filesV1;
	}
	return filesV2;
};

const seahub = () => {
	if (useFilesVersion() == 'v1') {
		return seahubV1;
	}
	return seahubV2;
};

const shareToUser = () => {
	if (useFilesVersion() == 'v1') {
		return shareToUserV1;
	}
	return shareToUserV2;
};

const common = () => {
	if (useFilesVersion() == 'v1') {
		return commonV1;
	}
	return commonV2;
};

const utils = () => {
	if (useFilesVersion() == 'v1') {
		return utilV1;
	}
	return utilV2;
};

const syncUtil = () => {
	if (useFilesVersion() == 'v1') {
		return syncUtilV1;
	}
	return syncUtilV2;
};

const syncFilesFormat = () => {
	if (useFilesVersion() == 'v1') {
		return syncFilesFormatV1;
	}
	return syncFilesFormatV2;
};

const filesIsV1 = () => {
	return useFilesVersion() == 'v1';
};

const filesIsV2 = () => {
	return useFilesVersion() == 'v2';
};

type SyncDataAPI = SyncDataAPIV1 | SyncDataAPIV2;

export {
	files,
	seahub,
	shareToUser,
	common,
	dataAPIs,
	utils,
	SyncDataAPI,
	syncUtil,
	syncFilesFormat,
	ai,
	filesIsV1,
	filesIsV2,
	commonV2
};
