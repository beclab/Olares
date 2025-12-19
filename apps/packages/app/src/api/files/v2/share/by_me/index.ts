// import { DriveType } from 'src/utils/interface/files';
// import ShareAPI from '../base/data';
// import { useFilesStore } from 'src/stores/files';
// import * as filesUtil from '../../common/utils';
// import { appendPath } from '../../path';

// export default class ShareByMeAPI extends ShareAPI {
// 	async formatRepotoPath(item: any): Promise<string> {
// 		console.log('item --->', item);
// 		// const filesStore = useFilesStore();
// 		// if (filesStore.nodes.length == 0) {
// 		// 	await filesUtil.fetchNodeList();
// 		// }

// 		// if (filesStore.onlyMasterNodes[this.origin_id]) {
// 		// 	return appendPath('/Share/by/', filesStore.masterNode, '/');
// 		// }

// 		// if (filesStore.nodes.length > 1) {
// 		// 	return '/Share/by/';
// 		// }
// 		// if (filesStore.nodes.length > 0) {
// 		// 	filesStore.currentNode[this.origin_id] = filesStore.nodes[0];
// 		// }

// 		return appendPath(
// 			'/ShareBy/'
// 			// filesStore.nodes.length == 0 ? '' : filesStore.nodes[0].name,
// 			// '/'
// 		);
// 	}
// }
