// import { DriveType } from 'src/utils/interface/files';
// import ShareAPI from '../base/data';
// import { useFilesStore } from 'src/stores/files';
// // import * as filesUtil from '../../common/utils';
// import { appendPath } from '../../path';

// export default class ShareWithMeAPI extends ShareAPI {
// 	// public driveType: DriveType = DriveType.ShareWithMe;
// 	async formatRepotoPath(item: any): Promise<string> {
// 		console.log('item --->', item);
// 		// return '/Share/with/';
// 		const filesStore = useFilesStore();
// 		// if (filesStore.nodes.length == 0) {
// 		// 	await filesUtil.fetchNodeList();
// 		// }

// 		// if (filesStore.onlyMasterNodes[this.origin_id]) {
// 		// 	return appendPath('/Share/with/', filesStore.masterNode, '/');
// 		// }

// 		// if (filesStore.nodes.length > 1) {
// 		// 	return '/Share/with/';
// 		// }
// 		// if (filesStore.nodes.length > 0) {
// 		// 	filesStore.currentNode[this.origin_id] = filesStore.nodes[0];
// 		// }

// 		return appendPath(
// 			'/ShareWith/'
// 			// filesStore.nodes.length == 0 ? '' : filesStore.nodes[0].name,
// 			// '/'
// 		);
// 	}
// }
