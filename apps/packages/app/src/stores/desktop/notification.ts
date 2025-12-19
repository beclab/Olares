import { defineStore } from 'pinia';
import {
	NotificationItem,
	MemeryNotificationItem
} from 'src/utils/desktop/notification';
import { NotificationDatabase } from 'src/utils/desktop/notificationDB';
import { useAppStore } from './app';

export type DataState = {
	data: MemeryNotificationItem[];
	showNotification: boolean;
	newMessage: boolean;
	currentPage: number;
	pageSize: number;
	hadMore: boolean;
};

const db = new NotificationDatabase();

export const useNotificationStore = defineStore('notification', {
	state: () => {
		return {
			showNotification: false,
			newMessage: false,
			data: [],
			currentPage: 0,
			pageSize: 20,
			hadMore: true
		} as DataState;
	},
	getters: {},
	actions: {
		async initDatas() {
			this.currentPage = 0;

			this.data = await db.notificationData
				.reverse()
				.limit(this.pageSize)
				.toArray();
			this.hadMore = this.data.length == this.pageSize;
		},

		async loadMore() {
			this.currentPage += 1;
			const more = await db.notificationData
				.reverse()
				.offset(this.data.length)
				.limit(this.pageSize)
				.toArray();

			this.hadMore = more.length >= this.pageSize;
			this.data = this.data.concat(more);
		},

		async addItem(item: NotificationItem) {
			await this.updateItemIcon(item);
			const index = this.data.findIndex((e) => e.id == item.id);
			if (index < 0) {
				this.data.splice(0, 0, item);
				if (!this.showNotification) {
					this.newMessage = true;
				}
			} else {
				this.data[index] = {
					...this.data[index],
					...item
				};
				this.update(item.id!, this.data[index]);
			}
		},

		async updateItemIcon(item: NotificationItem) {
			if (item.icon) {
				return;
			}
			if (item.appName) {
				const appStore = useAppStore();
				const app = appStore.myApps.find(
					(e) => e.name == item.appName || e.fatherName == item.appName
				);

				if (app) {
					item.icon = app.icon;
				} else {
					const oldItem = this.data.find(
						(e) => e.appName == item.appName && e.icon != undefined
					);

					if (oldItem) {
						item.icon = oldItem.icon;
					}
				}
			}
			if (!item.icon) {
				item.icon = '/desktop/app-icon/os.svg';
			}

			await this.update(item.id!, {
				icon: item.icon
			});
		},

		async deleteItem(item: NotificationItem, childIndex: number) {
			if (!item.id) {
				return;
			}

			if (childIndex >= 0 && childIndex < item.childrens.length) {
				item.childrens.splice(childIndex, 1);
				if (item.childrens.length > 0)
					await this.update(item.id, {
						childrens: item.childrens
					});
			}
			if (childIndex < 0 || item.childrens.length == 0) {
				await db.notificationData.delete(item.id);
				this.data = this.data.filter((e) => e.id != item.id);
			}
		},

		async update(id: number, item: Partial<NotificationItem>): Promise<number> {
			return await db.notificationData.update(id, item);
		},

		deleteAll() {
			this.data = [];
			db.notificationData.clear();
		},

		toggleNotificaitonDisplay() {
			if (!this.showNotification && this.data.length == 0) {
				return;
			}
			if (!this.showNotification) {
				this.newMessage = false;
			}
			this.showNotification = !this.showNotification;
		}
	}
});
