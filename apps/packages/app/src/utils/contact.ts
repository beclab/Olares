export enum MenuItem {
	DRIVE = 'Drive',
	SYNC = 'Sync',
	APPLICATION = 'Application',
	HOME = 'Home',
	DOCUMENTS = 'Documents',
	PICTURES = 'Pictures',
	MOVIES = 'Movies',
	DOWNLOADS = 'Downloads',
	MYLIBRARIES = 'My Libraries',
	SHAREDWITH = 'Shared Libraries',
	DATA = 'Data',
	CACHE = 'Cache',
	CODE = 'Code',
	MUSIC = 'Music',
	EXTERNAL = 'External',
	CLOUDDRIVE = 'CloudDrive',
	SHARE = 'Share'
	// SHAREWITHME = 'ShareWithMe',
	// SHAREBYME = 'ShareByMe'
}

export enum VaultMenuItem {
	VAULTCLASSIFICATION = 'Vault Classification',
	ALLVAULTS = 'All Vaults',
	AUTHENTICATOR = 'Authenticator',
	RECENTLYUSED = 'Recently Used',
	FAVORITES = 'Favorites',
	ATTACHMENTS = 'Attachments',
	MyVault = 'My Vault',
	TAGS = 'Tags',
	MYTEAMS = 'Team Vault',
	TEAMS = 'Teams',
	INVITES = 'Invites',
	TOOLS = 'Tools',
	SECURITYREPORT = 'Security Report',
	PASSWORDGENERATOR = 'Password Generator',
	UTILITY = 'Utility',
	LOCKSCREEN = 'Lock Screen',
	SETTINGS = 'Settings'
}

export enum OPERATE_ACTION {
	SHARE = 1,
	DELETE,
	RENAME,
	ATTRIBUTES,
	MOVE,
	COPY,
	CUT,
	PASTE,
	DOWNLOAD,
	UPLOAD_FILES,
	UPLOAD_FOLDER,
	LINK_SHARING,
	VERSION_HISTORY,
	CREATE_REPO,
	CREATE_FOLDER,
	REFRESH,
	SYNCHRONIZE_TO_LOCAL,
	OPEN_LOCAL_SYNC_FOLDER,
	UNSYNCHRONIZE,
	SYNC_IMMEDIATELY,
	SHARE_WITH,
	EXIT_SHARING,
	REPO_DELETE,
	REPO_RENAME,
	CANCEL,
	UNMOUNT,
	MD5,
	BACKUP,
	SHARE_IN_INTERNAL,
	SHARE_IN_SMB,
	SHARE_IN_PUBLIC,
	EDIT_PERMISSIONS,
	RESET_PASSWORD,
	REVOKE_SHARING
}

type PopupItem = {
	name: string;
	icon?: string;
	requiredSync: boolean | undefined;
	action: OPERATE_ACTION;
};

export const popupMenu: PopupItem[] = [
	{
		name: 'files_popup_menu.open_local_sync_folder',
		icon: 'sym_r_folder_open',
		requiredSync: true,
		action: OPERATE_ACTION.OPEN_LOCAL_SYNC_FOLDER
	},
	{
		name: 'files_popup_menu.synchronize_to_local',
		icon: 'sym_r_sync',
		requiredSync: false,
		action: OPERATE_ACTION.SYNCHRONIZE_TO_LOCAL
	},
	{
		name: 'files_popup_menu.unsynchronize',
		icon: 'sym_r_sync_disabled',
		requiredSync: true,
		action: OPERATE_ACTION.UNSYNCHRONIZE
	},
	{
		name: 'files_popup_menu.sync_immediately',
		icon: 'sym_r_autoplay',
		requiredSync: true,
		action: OPERATE_ACTION.SYNC_IMMEDIATELY
	},
	{
		name: 'files_popup_menu.share_with',
		requiredSync: undefined,
		icon: 'sym_r_share',
		action: OPERATE_ACTION.SHARE_WITH
	},
	{
		name: 'files_popup_menu.exit_sharing',
		requiredSync: undefined,
		icon: 'sym_r_share_off',
		action: OPERATE_ACTION.EXIT_SHARING
	},
	{
		name: 'files_popup_menu.rename',
		icon: 'sym_r_edit_square',
		requiredSync: undefined,
		action: OPERATE_ACTION.RENAME
	},
	{
		name: 'files_popup_menu.delete',
		icon: 'sym_r_delete',
		requiredSync: undefined,
		action: OPERATE_ACTION.DELETE
	},
	{
		name: 'files_popup_menu.attributes',
		icon: 'sym_r_ballot',
		requiredSync: undefined,
		action: OPERATE_ACTION.ATTRIBUTES
	}
];

export interface SyncItem {
	id: string;
	key: string | number;
	label: string;
	icon: string;
	children?: SyncItem[];
}

export enum SYNC_STATE {
	DISABLE,
	WAITING,
	INIT,
	ING,
	DONE,
	ERROR,
	UNKNOWN
}

export enum FilesSortType {
	DEFAULT = 0,
	NAME = 1, // a-z
	SIZE = 2,
	TYPE = 3, //A-Z
	Modified = 4 //By modification time
}

export const filesSortTypeInfo: Record<
	FilesSortType,
	{
		name: string;
		introduce: { asc: string; desc: string };
		icon: string;
		by: string;
	}
> = {
	[FilesSortType.NAME]: {
		name: 'sort.name.name',
		introduce: {
			asc: 'sort.name.asc',
			desc: 'sort.name.desc'
		},
		icon: 'sym_r_sort_by_alpha',
		by: 'name'
	},
	[FilesSortType.SIZE]: {
		name: 'sort.size.name',
		introduce: {
			asc: 'sort.size.asc',
			desc: 'sort.size.desc'
		},
		icon: 'sym_r_sort',
		by: 'size'
	},
	[FilesSortType.TYPE]: {
		name: 'sort.type.name',
		introduce: {
			asc: 'sort.type.asc',
			desc: 'sort.type.desc'
		},
		icon: 'sym_r_auto_awesome_motion',
		by: 'type'
	},
	[FilesSortType.Modified]: {
		name: 'sort.modified.name',
		introduce: {
			asc: 'sort.modified.asc',
			desc: 'sort.modified.desc'
		},
		icon: 'sym_r_acute',
		by: 'modified'
	},
	[FilesSortType.DEFAULT]: {
		name: 'By default',
		introduce: {
			asc: '',
			desc: ''
		},
		icon: '',
		by: ''
	}
};

export const scrollBarStyle = {
	contentStyle: {},
	contentActiveStyle: {},
	horizontalThumbStyle: {
		right: '2px',
		borderRadius: '3px',
		backgroundColor: '#BCBDBE',
		height: '6px',
		opacity: '1'
	},
	thumbStyle: {
		right: '2px',
		borderRadius: '3px',
		backgroundColor: '#BCBDBE',
		width: '6px',
		height: '6px',
		opacity: '1'
	}
};

export type DefaultDomainValueType = 'global' | 'cn';

export const defaultDomains: {
	name: string;
	value: DefaultDomainValueType;
}[] = [
	{
		name: 'olares.com',
		value: 'global'
	},
	{
		name: 'olares.cn',
		value: 'cn'
	}
];

export const getDomainNameByType = (value: DefaultDomainValueType) => {
	return defaultDomains.find((e) => e.value == value)?.name || 'olares.com';
};
