export interface Usage {
	total: number;
	usage: number;
	ratio: number;
	uint: string;
	name?: string;
	color?: string;
}

export interface Net {
	transmitted: number;
	received: number;
}

export interface Moniter {
	cpu: Usage;
	memory: Usage;
	disk: Usage;
	net: Net;
}
