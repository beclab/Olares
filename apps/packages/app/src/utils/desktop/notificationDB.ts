import Dexie from 'dexie';
import { NotificationItem } from './notification';
export class NotificationDatabase extends Dexie {
	notificationData: Dexie.Table<NotificationItem, number>;
	constructor() {
		super('NotificationDatabase');
		this.version(1).stores({
			notificationData: '++id,createTime,updateTime,appName,childrens'
		});
		this.notificationData = this.table('notificationData');
	}
}
