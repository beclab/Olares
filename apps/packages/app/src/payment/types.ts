import { Ref } from 'vue';

export type LogLevel = 'info' | 'warn' | 'error' | 'success';

export interface LogItem {
	time: string;
	message: string;
	level: LogLevel;
}

export interface ProductBase {
	product_id: string;
	metadata?: {
		title: string;
		[key: string]: unknown;
	};
	[key: string]: unknown;
}

export interface ApiResponse<T = unknown> {
	code: number;
	message?: string;
	data: T;
}

export interface ProductServerData {
	app_name: string;
	products: Record<string, Omit<ProductBase, 'product_id'>>;
}

export interface ProductDetailsData {
	app_info: {
		developer: {
			rsa_public: string;
			did: string;
			[key: string]: unknown;
		};
		receive_addresses: {
			optimism: string;
			[key: string]: string;
		};
		[key: string]: unknown;
	};
	[key: string]: unknown;
}

export interface OrderDataBase {
	from: string;
	to: string;
	product: Array<{ product_id: string }>;
	nonce: string;
	signature?: string;
	signer?: string;
}

export interface ChainConfig {
	chainId: string; // '0xa'
	chainName: string;
	nativeCurrency: {
		name: string;
		symbol: string;
		decimals: number;
	};
	rpcUrls: string[];
	blockExplorerUrls: string[];
	maxPriorityFeePerGasGwei?: string;
	maxFeePerGasGwei?: string;
}

export interface ERC20ContractAbiItem {
	inputs?: Array<{
		internalType: string;
		name: string;
		type: string;
	}>;
	name: string;
	outputs?: Array<{
		internalType: string;
		name: string;
		type: string;
	}>;
	stateMutability: string;
	type: string;
}

export interface TxPollTask {
	id: string;
	timer: NodeJS.Timeout;
	status: Ref<PollingStatus>;
	interval: number;
	confirmations: number;
	onSuccess?: (txHash: string, result: TransactionStatusResult) => void;
	onError?: (error: Error) => void;
}

export interface VcPollTask {
	id: string;
	timer: NodeJS.Timeout;
	status: Ref<PollingStatus>;
	apiUrl: string;
	interval: number;
	onSuccess?: (result: VCPollResult) => void;
	onError?: (error: Error) => void;
}

export type PollingStatus = 'idle' | 'polling' | 'completed' | 'failed';
export interface TransactionStatusResult {
	success: boolean;
	confirmed?: boolean;
	receipt?: any;
}
export interface VCPollResult {
	success: boolean;
	vc?: string;
	message?: string;
}
export interface SubmitVCResult {
	success: boolean;
	message?: string;
	transactionHash?: string;
}
