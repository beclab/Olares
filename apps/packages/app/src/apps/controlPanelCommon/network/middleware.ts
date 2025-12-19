export type MiddlewareType =
	| 'mongodb'
	| 'postgres'
	| 'redis'
	| 'rabbitmq'
	| 'minio'
	| 'nats'
	| 'mysql'
	| 'mariadb'
	| 'elasticsearch';

export const MIDDLEWARE_ICONS: Record<MiddlewareType, string> = {
	mongodb: 'sym_r_description',
	postgres: 'sym_r_table_chart',
	redis: 'sym_r_view_in_ar',
	rabbitmq: 'sym_r_queue',
	minio: 'sym_r_inventory_2',
	nats: 'sym_r_podcasts',
	mysql: 'sym_r_table_view',
	mariadb: 'sym_r_table_rows',
	elasticsearch: 'sym_r_manage_search'
};

export interface MiddlewareItem {
	name: string;
	namespace: string;
	nodes: number;
	adminUser: string;
	password: string;
	mongos: {
		endpoint: string;
		size: number;
	};
	redisProxy: {
		endpoint: string;
		size: number;
	};
	proxy: {
		endpoint: string;
		size: number;
	};
	type: MiddlewareType;
}

export interface MiddlewareListResponse {
	code: number;
	data: MiddlewareItem[] | [];
}

export interface MiddlewarePasswordParams {
	name: string;
	namespace: string;
	middleware: MiddlewareType;
	user: string;
	password: string;
}

export interface MiddlewarePasswordResponse {
	code: number;
	message: string;
}
