import Dexie from 'dexie';
import { TransferItem } from './transfer';
export class TransferDatabase extends Dexie {
	transferData: Dexie.Table<TransferItem, number>;

	constructor() {
		super('TransferDatabase');
		this.version(1).stores({
			transferData:
				'++id,task,name,path,type,isFolder,driveType,front,status,url,startTime,endTime,from,to,isPaused,size,message,uniqueIdentifier,repo_id,params,parentPath,userId,relatePath,node,currentPhase,totalPhase,phaseTaskId,pauseDisable,wiseRecordId'
		});
		this.transferData = this.table('transferData');
	}
}
