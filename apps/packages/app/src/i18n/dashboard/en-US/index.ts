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
	CLUSTER_PHYSIC_RESOURCE: "Cluster's physical resources",
	USER_RESOURCES: "{name}'s resources",
	USAGE_RANKING: 'Usage ranking',
	MORE: 'More',
	MORE_DETAILS: 'More details',
	PHYSICAL_RESOURCE_MONTORING: 'Physical resource monitoring',
	TOP_COUNT_CPU_USAGE_APPS: 'Top {count} CPU users',
	TOP_COUNT_MEMORY_USAGE_APPS: 'Top {count} memory users',
	VIEWS_COUNT_VIEW: 'Views in {count} hours',
	VISITORS_COUNT_VISITOR: 'Visitors in {count} hours',
	VISITORS: 'Visitors',
	VIEWS: 'Views',
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
	UPGRADING: 'upgrading',
	RUNNING: 'Running',
	SUSPEND: 'Suspend',
	RESUMING: 'Resuming',
	UNINSTALLING: 'Uninstalling',
	UNINSTALLED: 'Uninstalled',
	OPEN_APP: 'Open App',
	VIEW_DETAIL: 'View details',
	OPERATIONS: 'Operations',
	CPU_DETAILS: 'CPU details',
	MEMORY_DETAILS: 'Memory details',
	DISK_DETAILS: 'Disk details',
	PODS_DETAILS: 'Pods details',
	GPU_DETAILS: 'GPU overview',
	NETWORK_DETAILS: 'Network details',
	FAN_DETAILS: 'Fan details',
	THREAD: 'Thread',
	APP_STATUS: AppStatus.app,
	start_autoplay_metrics_refresh: 'Start auto-refresh',
	stop_autoplay_metrics_refresh: 'Stop auto-refresh'
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
