import { SearchCategory } from './search';

export interface AppInfo {
	id: string;
	appid: string;
	deployment?: string;
	icon: string;
	title: string;
	target: string;
	name: string;
	namespace?: string;
	owner?: string;
	url?: string;
	//installed: boolean;
	state: string;
	type?: SearchCategory;
	fatherName: string | null;
	openMethod: string;
	isSysApp: boolean;
	fatherState: string;
}
