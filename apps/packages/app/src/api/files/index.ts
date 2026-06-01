import { FilesIdType } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { getApplication } from 'src/application/base';
import { useUserStore } from 'src/stores/user';

import Origin from './v2/origin';
import { DriveAPI } from './v2';
import SyncDataAPI from './v2/sync/data';

import * as filesNs from './v2/drive/utils';
import * as seahubNs from './v2/sync/utils';
import * as shareToUserNs from './v2/shareToUser';
import * as commonNs from './v2/common/common';
import * as utilsNs from './v2/utils';
import * as syncUtilNs from './v2/sync/utils';
import * as syncFilesFormatNs from './v2/sync/filesFormat';

import * as ai from './ai';

function dataAPIs(
	origin?: DriveType,
	originId: number = FilesIdType.PAGEID
): Origin {
	return DriveAPI.getAPI(origin, originId);
}

const files = () => filesNs;
const seahub = () => seahubNs;
const shareToUser = () => shareToUserNs;
const common = () => commonNs;
const utils = () => utilsNs;
const syncUtil = () => syncUtilNs;
const syncFilesFormat = () => syncFilesFormatNs;

const isShareEnable = () => {
	if (!getApplication().platform) {
		return true;
	}
	const userStore = useUserStore();
	if (!userStore.current_user) {
		return false;
	}
	return userStore.current_user.isLargeVersion12_3;
};

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
	isShareEnable
};
