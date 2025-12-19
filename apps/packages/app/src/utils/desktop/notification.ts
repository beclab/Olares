export interface NotificationItem {
	id?: number;
	createTime: number;
	updateTime: number;
	appName?: string;
	icon?: string;
	childrens: {
		title: string;
		body: string;
		createTime: number;
		event: string;
	}[];
}

export interface MemeryNotificationItem extends NotificationItem {
	open?: boolean;
}
