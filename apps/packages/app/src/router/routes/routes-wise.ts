import { RouteRecordRaw } from 'vue-router';
import { MenuType } from 'src/utils/rss-menu';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		name: 'WiseMain',
		component: () => import('layouts/WiseMainLayout.vue'),
		children: [
			{
				path: MenuType.Trend,
				name: MenuType.Trend,
				component: () => import('pages/Wise/trend/TrendPage.vue')
			},
			// {
			// 	path: MenuType.Feed,
			// 	name: MenuType.Feed,
			// 	component: () => import('pages/Wise/library/FeedPage.vue')
			// },
			{
				path: '/:filterId',
				name: MenuType.Custom,
				component: () => import('pages/Wise/library/CustomPage.vue')
			},
			{
				path: MenuType.History,
				name: MenuType.History,
				component: () => import('src/pages/Wise/manager/RecentlyReadPage.vue')
			},
			{
				path: MenuType.Transmission,
				name: MenuType.Transmission,
				component: () =>
					import('src/pages/Wise/manager/transmission/TransmissionPage.vue')
			},
			{
				path: MenuType.Filtered_Views,
				name: MenuType.Filtered_Views,
				component: () => import('pages/Wise/manager/FilteredViewsPage.vue')
			},
			{
				path: MenuType.RSS_Feeds,
				name: MenuType.RSS_Feeds,
				component: () => import('pages/Wise/manager/RssFeedsPage.vue')
			},
			{
				path: MenuType.Recommend,
				name: MenuType.Recommend,
				component: () => import('pages/Wise/manager/RecommendPage.vue')
			},
			{
				path: MenuType.Tags,
				name: MenuType.Tags,
				component: () => import('pages/Wise/manager/TagsPage.vue')
			},
			{
				path: MenuType.Preferences,
				name: MenuType.Preferences,
				component: () => import('pages/Wise/manager/PreferencesPage.vue')
			},
			{
				path: '/:path/:id/:action?',
				name: MenuType.Entry,
				component: () => import('pages/Wise/reader/EntryReadingPage.vue')
			}
		]
	},

	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/ErrorNotFound.vue')
	},
	{
		path: '/not-found',
		component: () => import('pages/ErrorNotFound.vue')
	}
];

export default routes;
