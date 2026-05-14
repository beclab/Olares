export interface TerminusEntrance {
	id: string;
	name: string;
	title: string;
	url: string;
	icon?: string;
	authLevel: string;
	invisible?: boolean;
	openMethod: string;
	state: string;
	reason?: string;
	message?: string;
}

export interface TerminusPorts {
	name: string;
	host: string;
	port: string;
	exposePort: string;
	protocol: string;
}

export interface TerminusApp {
	id: string;
	icon: string;
	isClusterScoped: boolean;
	isSysApp: boolean;
	name: string;
	rawAppName: string;
	namespace?: string;
	target: string;
	owner?: string;
	title: string;
	url?: string;
	state: string;
	deployment?: string;
	entrances: TerminusEntrance[];
	ports: TerminusPorts[];
	requiredGpu: string;
	sharedEntrances?: TerminusEntrance[];
}
