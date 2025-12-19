import { Err, ErrorCode } from './core';
import { marshal, unmarshal } from './core';
import { Request, Response, Sender } from './core/transport';
import { RequestProgress } from './core';
import { translate as $l } from './util';

export type Method = 'GET' | 'POST' | 'PUT' | 'DELETE';

export async function request(
	method: Method,
	url: string,
	body?: string,
	headers?: Map<string, string>,
	progress?: RequestProgress
): Promise<XMLHttpRequest> {
	const req = new XMLHttpRequest();

	return new Promise<XMLHttpRequest>((resolve, reject) => {
		req.onreadystatechange = () => {
			if (req.readyState === 4) {
				if (!req.status) {
					reject(
						new Err(
							ErrorCode.FAILED_CONNECTION,
							$l(
								'The app could not establish a connection with our servers, please check your internet connection or try again later!'
							)
						)
					);
				} else {
					resolve(req);
				}
			}
		};

		try {
			req.open(method, url, true);
			if (headers) {
				headers.forEach((value, key) => req.setRequestHeader(key, value));
			}
			if (progress) {
				req.onprogress = (pg: { loaded: number; total: number }) =>
					(progress.downloadProgress = pg);
				req.upload.onprogress = (pg: { loaded: number; total: number }) =>
					(progress.uploadProgress = pg);
			}
			req.send(body);
		} catch (e) {
			reject(
				new Err(
					ErrorCode.FAILED_CONNECTION,
					$l(
						'The app could not establish a connection with our servers, please check your internet connection or try again later!'
					)
				)
			);
		}
	});
}

export class AjaxSender implements Sender {
	constructor(public url: string) {}

	async send(req: Request, progress?: RequestProgress): Promise<Response> {
		const body = marshal(req.toRaw());
		const headers = new Map<string, string>([
			['Content-Type', 'application/json'],
			['Accept', 'application/json']
		]);
		try {
			const response = await request('POST', this.url, body, headers, progress);
			const res = unmarshal(response.responseText);
			return new Response().fromRaw(res);
		} catch (e) {
			throw new Err(ErrorCode.SERVER_ERROR);
		}
	}
}
