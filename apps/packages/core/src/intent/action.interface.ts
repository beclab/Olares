export enum Action {
	// Standard Activity Actions
	// These are the current standard actions that Intent defines for launching activities
	// (usually through Context#startActivity. The most important, and by far most frequently used,
	// are ACTION_MAIN and ACTION_EDIT.
	ACTION_MAIN = 'main',
	ACTION_VIEW = 'view',
	ACTION_ATTACH_DATA = 'attach_data',
	ACTION_EDIT = 'edit',
	ACTION_PICK = 'pick',
	ACTION_CHOOSER = 'chooser',
	ACTION_GET_CONTENT = 'get_content',
	ACTION_DIAL = 'dial',
	ACTION_CALL = 'call',
	ACTION_SEND = 'send',
	ACTION_SENDTO = 'sendto',
	ACTION_ANSWER = 'answer',
	ACTION_INSERT = 'insert',
	ACTION_DELETE = 'delete',
	ACTION_RUN = 'run',
	ACTION_SYNC = 'sync',
	ACTION_PICK_ACTIVITY = 'pick_activity',
	ACTION_SEARCH = 'search',
	ACTION_WEB_SEARCH = 'web_search',
	ACTION_FACTORY_TEST = 'factory_test',

	// Standard Broadcast Actions
	ACTION_TIME_TICK = 'time_tick',
	ACTION_TIME_CHANGED = 'time_changed',
	ACTION_TIMEZONE_CHANGED = 'timezone_changed',
	ACTION_BOOT_COMPLETED = 'boot_completed',
	ACTION_PACKAGE_ADDED = 'package_added',
	ACTION_PACKAGE_CHANGED = 'package_changed',
	ACTION_PACKAGE_REMOVED = 'package_removed',
	ACTION_PACKAGE_RESTARTED = 'package_restarted',
	ACTION_PACKAGE_DATA_CLEARED = 'package_data_cleared',
	ACTION_PACKAGES_SUSPENDED = 'packages_suspended',
	ACTION_PACKAGES_UNSUSPENDED = 'packages_unsuspended',
	ACTION_UID_REMOVED = 'uid_removed',
	ACTION_BATTERY_CHANGED = 'battery_changed',
	ACTION_POWER_CONNECTED = 'power_connected',
	ACTION_POWER_DISCONNECTED = 'power_disconnected',
	ACTION_SHUTDOWN = 'shutdown'
}
