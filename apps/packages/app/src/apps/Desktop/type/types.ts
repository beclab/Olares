import { AppInfo } from 'src/utils/interface/applications';

export interface WindowInfo {
	id: string;
	url: string;
	ssl_type: string;
	max_width: number;
	max_height: number;
	min_width: number;
	min_height: number;
	width: number;
	height: number;
	top: number;
	left: number;
	title: string;
	is_show: boolean;
	icon: string;
	z: number;
	active: boolean;
	isResizable: boolean;
}

export interface DesktopAppInfo {
	id: string;
	appid: string;
	index: number;
	page: number;
	page_num: number;
	left: number;
	top: number;
	size: number;
	icon: string;
	name: string;
	title: string;
	namespace?: string;
	owner?: string;
	url?: string;
	//installed: boolean;
	state: string;
	fatherName: string | null;
	isSysApp: boolean;
	fatherState: string;
}

export interface DockerAppInfo {
	id: string;
	left: number;
	top: number;
	size: number;
	icon: string;
	name: string;
	title: string;
	namespace?: string;
	owner?: string;
	url?: string;
	is_temp: boolean;
	show_dot: boolean;
	disabled?: boolean;
}

export interface DesktopPosition {
	index: number;
	page: number;
	page_num: number;
	left: number;
	top: number;
}

export interface AppStoreInfo {
	appid: string;
	desc: string;
	icon: string;
	name: string;
	status: string;
	title: string;
	version: string;
	categories: string;
	fullDescription: string;
	promoteImage: string[];
	subCategory: string;
}

export interface Config {
	inner_window: WindowInfo[];
	app_links: AppInfo[];
	bg_img: string;
}

export interface AccountInfo {
	uid: string;
	state: string;
	roles: string[];
	name: string;
	email: string;
	display_name: string;
	did: string;
	description: string;
	creation_timestamp: number;
}

export interface UserInfo {
	name: string;
	did: string;
	is_ephemeral: boolean;
	owner_role: string;
	zone: string;
	wizard_complete: boolean;
}

export interface WizardInfo {
	step: number;

	username: string | null;
	lang: string | null;
	timezone: string | null;

	did: string | null;

	use_frps: boolean;
	frps_region: string;
	external_ip: string | null;

	request_https_certificate: boolean;
	https_certificate_ready: boolean;

	dns_checked: boolean;
	// requested
	// //let progress = ref(0);
	// let checked = ref(false)

	finished: boolean;
	authLimit: {
		authority: number | null;
		appointIP: string;
		privateIP: string;
	};
}

export interface AppClickInfo {
	appid: string;
	data: any;
	path?: string;
}

export interface va1apha1Request {
	access_level: number | null;
	allow_cidrs: string[] | null;
}

export interface AuthorityType {
	name: string;
	level: number; //
	limit: number; //	1 - All IP; 2 - Private IP
	tokenAuthentication: boolean;
	awsDefault: boolean;
	intranetDefault: boolean;
}

export enum UpgradeState {
	StatusFailed = 'failed',
	StatusRunning = 'running',
	StatusComplete = 'completed',
	StatusStart = 'started',
	NotRunning = 'not_running'
}

export interface MessageData {
	message: string;
	type: string;
}

export enum SearchType {
	HomePage = 'home',
	FilesPage = 'Files Search',
	AshiaPage = 'Ashia',
	AshiaDocPage = 'Ashia Doc',
	TextSearch = 'Text Search'
}
