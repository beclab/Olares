export interface SSIResult {
	error: string;
	response: any;
}

export interface Operation {
	id: string;
	done: boolean;
	result: SSIResult;
}
