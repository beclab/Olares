export interface BusinessCallback {
	onSuccess(data: any): void;
	onFailure(message: string): void;
}

export interface BusinessAsyncCallback {
	onSuccess(data: any): Promise<void>;
	onFailure(message: string): void;
}
