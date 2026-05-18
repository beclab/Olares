export interface Result<T> {
	code: number;
	message: string | null;
	data?: T;
}

export function returnError(code: number, message: string): Result<null> {
	const result: Result<null> = {
		code: code,
		message: message,
		data: null
	};

	return result;
}

export function returnSucceed(data: any): Result<any> {
	const result: Result<any> = {
		code: 0,
		message: null,
		data: data
	};

	return result;
}

export interface ProviderRequest<T> {
	op: string;
	datatype: string;
	version: string;
	group: string;
	Token: string;
	data: T;
}
