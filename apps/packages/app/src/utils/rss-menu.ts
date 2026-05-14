import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

export enum MenuType {
	Trend = 'trend',
	History = 'history',
	Custom = 'custom',
	Download = 'download',
	Upload = 'upload',
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
	All = 'all',
	Uploading = 'uploading',
	Downloading = 'downloading',
	Complete = 'complete',
	Failed = 'failed'
}

export const MenuInfoMap = computed(() => {
	const { t } = useI18n();
	return {
		[MenuType.Download]: {
			title: t('main.download_list'),
			icon: 'sym_r_download'
		},
		[MenuType.Upload]: {
			title: t('main.upload_list'),
			icon: 'sym_r_upload'
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
		[TabType.All]: { title: t('main.all') },
		[TabType.Uploading]: { title: t('main.uploading') },
		[TabType.Downloading]: { title: t('main.downloading') },
		[TabType.Complete]: { title: t('main.complete') },
		[TabType.Failed]: { title: t('main.failed') }
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
