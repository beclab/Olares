// Single source of truth for app-operation enum values.
// Lives in its own tiny module so the policy layer and the service layer
// can both consume it without forming an import cycle.

export enum OPERATION_TYPE {
	INSTALL = 'install',
	CLONE = 'clone',
	UNINSTALL = 'uninstall',
	UPGRADE = 'upgrade',
	CANCEL = 'cancel',
	RESUME = 'resume',
	STOP = 'stop',
	RESTART = 'restart',
	REMOVE = 'remove'
}
