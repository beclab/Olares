import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

export enum MenuType {
	Trend = 'trend',
	History = 'history',
	Custom = 'custom',
	Transmission = 'transmission',
	Filtered_Views = 'filteredViews',
	Tags = 'tags',
	RSS_Feeds = 'edFeeds',
	Recommend = 'recommend',
	Preferences = 'preferences',
	Entry = 'entry'
}

export enum TabType {
	Empty = 'empty',
	UnSeen = 'unseen',
	Seen = 'seen',
	Inbox = 'inbox',
	ReadLater = 'readLater',
	Download = 'download',
	Upload = 'upload'
}

export const MenuInfoMap = computed(() => {
	const { t } = useI18n();
	return {
		[MenuType.Transmission]: {
			title: t('main.transmission'),
			icon: 'sym_r_swap_vert'
		}
	};
});

export const TabInfoMap = computed(() => {
	const { t } = useI18n();
	return {
		[TabType.UnSeen]: { title: t('main.unseen') },
		[TabType.Seen]: { title: t('main.seen') },
		[TabType.ReadLater]: { title: t('main.read_later') },
		[TabType.Inbox]: { title: t('main.inbox') },
		[TabType.Download]: { title: t('main.download') },
		[TabType.Upload]: { title: t('main.upload') }
	};
});

export const SupportDetails: string[] = [
	MenuType.Trend,
	MenuType.History,
	MenuType.Custom,
	MenuType.Entry
];

// export const SupportRouters: string[] = [
// 	...SupportDetails,
// 	MenuType.Recommend,
// 	MenuType.Tags,
// 	MenuType.Preferences,
// 	MenuType.RSS_Feeds,
// 	MenuType.Filtered_Views,
// 	MenuType.Transmission
// ];

export class MenuChoice {
	type: MenuType | string = '';
	tab: TabType | string | number = '';
	params = {};
}
