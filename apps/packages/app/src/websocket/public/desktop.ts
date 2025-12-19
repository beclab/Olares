import { NotificationItem } from 'src/utils/desktop/notification';
import { NotificationDatabase } from 'src/utils/desktop/notificationDB';

export const desktopInsertNotificationItem = (
	item: NotificationItem,
	db: NotificationDatabase,
	callBack?: (item: NotificationItem) => void
) => {
	db.notificationData.add(item).then((id: number) => {
		item.id = id;
		if (callBack) {
			callBack(item);
		}
	});
};
