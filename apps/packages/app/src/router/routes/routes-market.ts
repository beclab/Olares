import MarkerMainLayout from '../../layouts/MarketMainLayout.vue';
import { TRANSACTION_PAGE } from 'src/constant/constants';

const routes = [
	{
		path: '/',
		component: MarkerMainLayout,
		children: [
			{
				path: '/',
				name: TRANSACTION_PAGE.All,
				component: () => import('pages/market/application/HomePage.vue')
			},
			{
				path: 'category/:categories',
				name: TRANSACTION_PAGE.CATEGORIES,
				component: () => import('pages/market/application/CategoryPage.vue')
			},
			{
				path: 'app/:sourceId/:appName',
				name: TRANSACTION_PAGE.App,
				component: () =>
					import('pages/market/application/detail/AppDetailPage.vue')
			},
			{
				path: 'list/:categories/:type?',
				name: TRANSACTION_PAGE.List,
				component: () => import('pages/market/application/AppListPage.vue')
			},
			{
				path: 'discover/:category/:topicId',
				name: TRANSACTION_PAGE.TOPIC,
				component: () => import('pages/market/application/TopicPage.vue')
			},
			{
				path: '/preview/:sourceId/:appName/:index',
				name: TRANSACTION_PAGE.Preview,
				component: () =>
					import('pages/market/application/AppImagePreviewPage.vue')
			},
			{
				path: '/myapps',
				name: TRANSACTION_PAGE.MyTerminus,
				component: () => import('pages/market/me/MyTerminusPage.vue')
			},
			{
				path: '/settings',
				name: TRANSACTION_PAGE.Preference,
				component: () => import('pages/market/me/PreferencesPage.vue')
			},
			{
				path: '/log',
				name: TRANSACTION_PAGE.Log,
				component: () => import('pages/market/me/LogPage.vue')
			},
			{
				path: '/search',
				name: TRANSACTION_PAGE.Search,
				component: () => import('pages/market/manage/SearchPage.vue')
			},
			{
				path: '/update',
				name: TRANSACTION_PAGE.Update,
				component: () => import('pages/market/manage/UpdatePage.vue')
			},
			{
				path: '/version/:sourceId/:appName',
				name: TRANSACTION_PAGE.Version,
				component: () =>
					import('pages/market/application/VersionHistoryPage.vue')
			}
		]
	},

	// Always leave this as last one,
	// but you can also remove it
	{
		path: '/:catchAll(.*)*',
		component: () => import('pages/ErrorNotFound.vue')
	}
];

export default routes;
