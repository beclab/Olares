export enum NotifyDefinedType {
	DEFAULT = 'bt-default',
	LOGIN_NOTIFY = 'bt-login',
	POSITIVE = 'bt-positive',
	SUCCESS = 'bt-success',
	FAILED = 'bt-failed',
	WARNING = 'bt-warning',
	LOADING = 'bt-loading',
	PROGRESS = 'bt-progress',
	MESSAGE = 'bt-message'
}

export interface BtNotifyProps {
	alwaysShow: any;
	show: any;
	hide: any;
	init(notify: any): void;
	notify: any;
}
