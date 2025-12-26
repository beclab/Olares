import { createServer, IncomingMessage } from 'http';
import {
	Receiver,
	Request,
	Sender,
	Response
} from '@didvault/sdk/src/core/transport';
import { marshal, unmarshal } from '@didvault/sdk/src/core/';
import { Err, ErrorCode } from '@didvault/sdk/src/core/error';
//import { getLocation } from '../geoip';
import { request as requestHttps } from 'https';
import { request as requestHttp } from 'http';
import { Config, ConfigParam } from '@didvault/sdk/src/core/config';

export function readBody(
	request: IncomingMessage,
	maxSize = 1e7
): Promise<string> {
	return new Promise((resolve, reject) => {
		const body: Buffer[] = [];
		let size = 0;

		request
			.on('data', (chunk) => {
				size += chunk.length;
				if (size > maxSize) {
					console.error('Max request size exceeded!', size, maxSize);
					request.destroy(new Err(ErrorCode.MAX_REQUEST_SIZE_EXCEEDED));
				}
				body.push(chunk);
			})
			.on('error', (e) => {
				reject(e);
			})
			.on('end', () => {
				resolve(Buffer.concat(body).toString());
			});
	});
}

export class HTTPReceiverConfig extends Config {
	@ConfigParam('number')
	port = 3000;

	@ConfigParam('number')
	maxRequestSize = 1e9;

	@ConfigParam()
	allowOrigin = '*';

	/** Path on the HTTP server for responding with 200, to be used in health checks (e.g. load balancers) */
	@ConfigParam()
	healthCheckPath = '/healthcheck';
}

export class HTTPReceiver implements Receiver {
	constructor(public readonly config: HTTPReceiverConfig) {}

	async listen(handler: (req: Request) => Promise<Response>) {
		const server = createServer(async (httpReq, httpRes) => {
			httpRes.on('error', (e) => {
				// todo
				console.error(e);
			});

			httpRes.setHeader('Access-Control-Allow-Origin', this.config.allowOrigin);
			httpRes.setHeader('Access-Control-Allow-Methods', 'OPTIONS, POST');
			httpRes.setHeader('Access-Control-Allow-Headers', 'Content-Type');

			switch (httpReq.method) {
				case 'OPTIONS':
					httpRes.end();
					break;
				case 'GET':
					try {
						// httpReq.url will include searchParams, so we parse the pathname properly
						const url = new URL(`http://localhost${httpReq.url || ''}`);
						if (url.pathname === this.config.healthCheckPath) {
							httpRes.statusCode = 200;
						} else {
							// Legacy server response for GET requests
							httpRes.statusCode = 405;
						}
					} catch (error) {
						console.error(error);
						httpRes.statusCode = 400;
					}
					httpRes.end();
					break;
				case 'POST':
					try {
						const body = await readBody(httpReq, this.config.maxRequestSize);
						const req = new Request().fromRaw(unmarshal(body));
						const ipAddress =
							httpReq.headers['x-forwarded-for'] ||
							httpReq.socket?.remoteAddress;
						req.ipAddress = Array.isArray(ipAddress) ? ipAddress[0] : ipAddress;
						// const location =
						// 	req.ipAddress && (await getLocation(req.ipAddress));
						// req.location = location
						// 	? {
						// 			country: location.country?.names['en'],
						// 			city: location.city?.names['en']
						// 	  }
						// 	: undefined;
						req.location = undefined;

						const clientVersion =
							(req.device && req.device.appVersion) || undefined;
						const res = await handler(req);
						const resBody = marshal(res.toRaw(clientVersion));
						httpRes.setHeader(
							'Content-Type',
							'application/json; charset=utf-8'
						);
						httpRes.setHeader('Content-Length', Buffer.byteLength(resBody));
						httpRes.write(resBody);
					} catch (error) {
						console.error(error);
						httpRes.statusCode = 400;
					}
					httpRes.end();
					break;
				default:
					httpRes.statusCode = 405;
					httpRes.end();
			}
		});

		server.listen(this.config.port);
	}
}

export function request(
	urlString: string,
	method: 'GET' | 'POST' = 'GET',
	body?: string,
	headers: { [header: string]: string } = {}
): Promise<string> {
	return new Promise((resolve, reject) => {
		const url = new URL(urlString);
		const fn = url.protocol === 'https:' ? requestHttps : requestHttp;
		const req = fn(
			url,
			{
				method,
				headers
			},
			(res) => {
				res.setEncoding('utf8');
				let resBody = '';

				res.on('data', (data) => {
					resBody += data;
				});

				res.on('end', () => {
					if (res.statusCode === 200) {
						resolve(resBody);
					} else {
						reject(
							`${res.statusCode} ${res.statusMessage} - Message:\n${resBody}`
						);
					}
				});

				res.on('error', (e) => reject(e));
			}
		);

		req.write(body);
		req.end();
	});
}

export class HTTPSender implements Sender {
	constructor(public url: string) {}

	async send(req: Request): Promise<Response> {
		const body = marshal(req.toRaw());

		const resBody = await request(this.url, 'POST', body, {
			'Content-Type': 'application/json',
			Accept: 'application/json'
		});

		return new Response().fromRaw(unmarshal(resBody));
	}
}
