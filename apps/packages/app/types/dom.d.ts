interface Navigator {
	Backbutton: any;
}

interface Window {
	app: any;
	router: any;
	extension: any;
	getPlatform: any;
}

declare namespace NodeJS {
	export interface Global {
		app: any;
		router: any;
		extension: any;
		getPlatform: any;
	}
}
