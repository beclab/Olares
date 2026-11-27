// This is just an example,
// so you can safely delete all default props below
import GPU_OP from './gpu';
import CPU_OP from './cpu';
import MEMORY_OP from './memory';
import DISK_OP from './disk';
import NET_OP from './network';
import settings from '../../settings/en-US';
import FAN_OP from './fan';
import AppStatus from '../../en-US';

const options = {
	...settings,
	OWNED_NODE: 'Host node',
	OVERVIEW: 'Overview',
	APPLICATIONS: 'Applications',
	ANALYTICS: 'Analytics',
	CLUSTER_PHYSIC_RESOURCE: 'Cluster’s physical resources',
	USER_RESOURCES: '{name}’s resources',
	USAGE_RANKING: 'Usage ranking',
	MORE: 'View all',
	MORE_DETAILS: 'More details',
	PHYSICAL_RESOURCE_MONITORING: 'Physical resource monitoring',
	TOP_COUNT_CPU_USAGE_APPS: 'Top {count} apps by CPU usage',
	TOP_COUNT_MEMORY_USAGE_APPS: 'Top {count} apps by memory usage',
	VIEWS_COUNT_VIEW: 'Views in {count} hours',
	VISITORS_COUNT_VISITOR: 'Visitors in {count} hours',
	VISITORS: 'Visitors',
	VIEWS: 'Views',
	COUNTRIES: 'Countries',
	EVENTS: 'Events',
	AVERAGE_VISIT_TIME: 'Average visit time',
	UNIQUE_VISITORS: 'Unique visitors',
	PAGE_VIEWS: 'Page views',
	PAGES: 'Pages',
	REFERRER: 'Referrer',
	BROWSER: 'Browser',
	OS: 'OS',
	DEVICE: 'Device',
	CITY: 'City',
	CPU: 'CPU',
	ANALYTICS_DATE_OPTION: {
		TODAY: 'Today',
		LAST_24_HOURS: 'Last 24 hours',
		YESTERDAY: 'Yesterday',
		THIS_WEEK: 'This week',
		LAST_7_DAYS: 'Last 7 days',
		THIS_MONTH: 'This month',
		LAST_30_DAYS: 'Last 30 days',
		LAST_90_DAYS: 'Last 90 days',
		THIS_YEAR: 'This year'
	},
	PENDING: 'Pending',
	DOWNLOADING: 'Downloading',
	INSTALLING: 'Installing',
	UPGRADING: 'Updating',
	RUNNING: 'Running',
	SUSPEND: 'Suspended',
	RESUMING: 'Resuming',
	UNINSTALLING: 'Uninstalling',
	UNINSTALLED: 'Uninstalled',
	OPEN_APP: 'Open app',
	VIEW_DETAIL: 'View details',
	OPERATIONS: 'Actions',
	CPU_DETAILS: 'CPU details',
	MEMORY_DETAILS: 'Memory details',
	DISK_DETAILS: 'Disk details',
	PODS_DETAILS: 'Pods details',
	GPU_DETAILS: 'GPU details',
	NETWORK_DETAILS: 'Network details',
	FAN_DETAILS: 'Fan details',
	THREAD: 'thread',
	CORES: '{n} core | {n} cores',
	THREADS: '{n} thread | {n} threads',
	APP_STATUS: { ...AppStatus.app, downloading: 'Downloading' },
	start_autoplay_metrics_refresh: 'Start auto-refresh',
	stop_autoplay_metrics_refresh: 'Stop auto-refresh',
	page_404: 'Nothing here yet',
	go_home: 'Go home',
	BOUNCE_RATE: 'Bounce rate',
	REALTIME: 'Real-time',
	START_DATE: 'Start date',
	END_DATE: 'End date',
	RANGE_SEPARATOR: 'to',
	DO_NOT_SHOW_THIS_CONFIRMATION_AGAIN: 'Do not show this confirmation again'
};

export default {
	...options,
	GPU_OP,
	CPU_OP,
	MEMORY_OP,
	DISK_OP,
	NET_OP,
	FAN_OP
};
