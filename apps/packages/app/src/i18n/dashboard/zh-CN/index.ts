// This is just an example,
// so you can safely delete all default props below
import GPU_OP from './gpu';
import CPU_OP from './cpu';
import MEMORY_OP from './memory';
import DISK_OP from './disk';
import NET_OP from './network';
import settings from '../../settings/zh-CN';
import FAN_OP from './fan';
import AppStatus from '../../zh-CN';

const options = {
	...settings,
	OWNED_NODE: '所属节点',
	OVERVIEW: '概览',
	APPLICATIONS: '应用',
	ANALYTICS: '分析',
	CLUSTER_PHYSIC_RESOURCE: '集群物理资源',
	USER_RESOURCES: '{name} 的资源',
	USAGE_RANKING: '使用排名',
	MORE: '查看全部',
	MORE_DETAILS: '详情',
	PHYSICAL_RESOURCE_MONITORING: '物理资源监控',
	TOP_COUNT_CPU_USAGE_APPS: 'CPU 占用前 {count} 的应用',
	TOP_COUNT_MEMORY_USAGE_APPS: '内存占用前 {count} 的应用',
	VIEWS_COUNT_VIEW: '{count} 小时内浏览量',
	VISITORS_COUNT_VISITOR: '{count} 小时内访客数',
	VISITORS: '访客',
	VIEWS: '浏览量',
	COUNTRIES: '国家',
	EVENTS: '事件',
	AVERAGE_VISIT_TIME: '平均访问时间',
	UNIQUE_VISITORS: '独立访客',
	PAGE_VIEWS: '页面浏览量',
	PAGES: '页面',
	REFERRER: '引荐来源',
	BROWSER: '浏览器',
	OS: '操作系统',
	DEVICE: '设备',
	CITY: '城市',
	CPU: 'CPU',
	ANALYTICS_DATE_OPTION: {
		TODAY: '今天',
		LAST_24_HOURS: '前 24 小时',
		YESTERDAY: '昨天',
		THIS_WEEK: '本周',
		LAST_7_DAYS: '前 7 天',
		THIS_MONTH: '本月',
		LAST_30_DAYS: '前 30 天',
		LAST_90_DAYS: '前 90 天',
		THIS_YEAR: '今年'
	},
	PENDING: '等待中',
	DOWNLOADING: '下载中',
	INSTALLING: '正在安装',
	UPGRADING: '更新中',
	RUNNING: '运行中',
	SUSPEND: '已暂停',
	RESUMING: '恢复中',
	UNINSTALLING: '正在卸载',
	UNINSTALLED: '未安装',
	OPEN_APP: '打开应用',
	VIEW_DETAIL: '查看详情',
	OPERATIONS: '操作',
	CPU_DETAILS: 'CPU 详情',
	MEMORY_DETAILS: '内存详情',
	DISK_DETAILS: '磁盘详情',
	PODS_DETAILS: '容器组详情',
	GPU_DETAILS: 'GPU 详情',
	FAN_DETAILS: '风扇详情',
	NETWORK_DETAILS: '网络详情',
	THREAD: '线程',
	CORES: '{n} 核',
	THREADS: '{n} 线程',
	APP_STATUS: { ...AppStatus.app, downloading: '下载中' },
	start_autoplay_metrics_refresh: '开始自动刷新',
	stop_autoplay_metrics_refresh: '停止自动刷新',
	page_404: '这里还没有内容',
	go_home: '返回首页',
	BOUNCE_RATE: '跳出率',
	REALTIME: '实时',
	START_DATE: '开始日期',
	END_DATE: '结束日期',
	RANGE_SEPARATOR: '至',
	DO_NOT_SHOW_THIS_CONFIRMATION_AGAIN: '不再显示此确认提示'
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
