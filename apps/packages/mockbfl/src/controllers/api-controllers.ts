import { Request, Response, NextFunction } from 'express';
//import {LevelDBStorage,LevelDBStorageConfig } from '../leveldb'
//import { Storable } from '@bytetrade/core';
import level from 'level';

const TERMINUS_NAME = process.env.TERMINUS_NAME || 'did';

const REFRESH_TOKEN = process.env.REFRESH_TOKEN || '';

const ACCESS_TOKEN = process.env.ACCESS_TOKEN || '';

const SESSION_ID = process.env.SESSION_ID || '';

const BASE32_SECRET = process.env.BASE32_SECRET || ''

const OTP_AUTH_URL = process.env.OPT_AUTH_URL || '';

export interface UserInfo {
	olaresId: string;
	wizardStatus: string;
	timestamp: number;
	selfhosted: boolean;
	tailScaleEnable: boolean;
	osVersion: string;
}

export interface LoginRequest {
	username: string;
	password: string;
}

export interface Token {
	access_token: string;
	token_type: string;
	refresh_token: string;
	expires_in: number;
	expires_at: number;
}

export interface Res<T> {
	code: number;
	message: string;
	data?: T;
}

// export class DID {
// 	id = '';
// 	did = '';
// }

//let did: any = null;
const _db = level('./data');

// async function init_did() {
// 	try {
// 		const item: string = await _db.get(TERMINUS_NAME);

// 		console.log(item);
// 		if (item) {
// 			did = JSON.parse(item);
// 		}
// 	} catch (e) {
// 		//console.log(e);
// 		did = null;
// 	}
// }

//init_did();

export class ApiControllers {
	postLogin(request: Request, response: Response, next: NextFunction) {
		console.log('postLogin ');

		const token: Token = {
			access_token: 'abc',
			token_type: 'abc',
			refresh_token: 'abc',
			expires_in: 7200,
			expires_at: 1313131311
		};

		//response.type('text/plain');
		response.set('Content-Type', 'application/json');
		response.send(token);
	}

	async getUserInfo(request: Request, response: Response, next: NextFunction) {
		console.log('getUserInfo ');

		let user: UserInfo = {
			olaresId: TERMINUS_NAME,
			wizardStatus: 'wait_activate_vault',
			timestamp: new Date().getTime(),
			selfhosted: true,
			tailScaleEnable: false,
			osVersion: '0.1.0'
		};

		try {
			const res = await _db.get(TERMINUS_NAME);
			user = JSON.parse(res);

			console.log(user.wizardStatus);
			if (user.wizardStatus == 'system_activating') {
				if (new Date().getTime() - user.timestamp > 5 * 1000) {
					user.wizardStatus = 'wait_activate_network';
					user.timestamp = new Date().getTime();
					_db.put(TERMINUS_NAME, JSON.stringify(user));
				}
			} else if (user.wizardStatus == 'wait_activate_network') {
				if (new Date().getTime() - user.timestamp > 5 * 1000) {
					user.wizardStatus = 'network_activating';
					user.timestamp = new Date().getTime();
					_db.put(TERMINUS_NAME, JSON.stringify(user));
				}
			} else if (user.wizardStatus == 'network_activating') {
				if (new Date().getTime() - user.timestamp > 5 * 1000) {
					user.wizardStatus = 'wait_reset_password';
					user.timestamp = new Date().getTime();
					_db.put(TERMINUS_NAME, JSON.stringify(user));
				}
			}
		} catch (e) {
			_db.put(TERMINUS_NAME, JSON.stringify(user));
		}

		console.log(user);

		const result: Res<UserInfo> = {
			code: 0,
			message: 'success',
			data: user
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	async postSetZone(request: Request, response: Response, next: NextFunction) {
		console.log('postSetDID ' + request.body);

		try {
			const res = await _db.get(TERMINUS_NAME);
			const user: UserInfo = JSON.parse(res);
			user.wizardStatus = 'wait_activate_system';
			user.timestamp = new Date().getTime();

			_db.put(TERMINUS_NAME, JSON.stringify(user));

			response.set('Content-Type', 'application/json');
			response.send({
				code: 0,
				message: 'success',
				data: null
			});
		} catch (e) {
			response.set('Content-Type', 'application/json');
			response.send({
				code: 1,
				message: 'errored',
				data: null
			});
		}
	}

	async postSetSystem(
		request: Request,
		response: Response,
		next: NextFunction
	) {
		console.log('postSetSystem ' + request.body);

		try {
			const res = await _db.get(TERMINUS_NAME);
			const user: UserInfo = JSON.parse(res);
			user.wizardStatus = 'system_activating';
			user.timestamp = new Date().getTime();

			_db.put(TERMINUS_NAME, JSON.stringify(user));

			response.set('Content-Type', 'application/json');
			response.send({
				code: 0,
				message: 'success',
				data: null
			});
		} catch (e) {
			response.set('Content-Type', 'application/json');
			response.send({
				code: 1,
				message: 'errored',
				data: null
			});
		}
	}

	async resetPassword(
		request: Request,
		response: Response,
		next: NextFunction
	) {
		console.log('resetPassword ' + request.params.user);
		console.log('resetPassword ' + request.body);

		try {
			const res = await _db.get(TERMINUS_NAME);
			const user: UserInfo = JSON.parse(res);
			user.wizardStatus = 'completed';
			user.timestamp = new Date().getTime();

			_db.put(TERMINUS_NAME, JSON.stringify(user));

			response.set('Content-Type', 'application/json');
			response.send({
				code: 0,
				message: 'success',
				data: null
			});
		} catch (e) {
			response.set('Content-Type', 'application/json');
			response.send({
				code: 1,
				message: 'errored',
				data: null
			});
		}
	}

	getRoles(request: Request, response: Response, next: NextFunction) {
		console.log('getRoles ');

		const result: Res<null> = {
			code: 0,
			message: 'success',
			data: null
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	async sslTaskState(request: Request, response: Response, next: NextFunction) {
		console.log('getRoles ');

		response.set('Content-Type', 'application/json');
		response.send({
			code: 0,
			message: 'success',
			data: { state: 2 }
		});
	}

	getIP(request: Request, response: Response, next: NextFunction) {
		console.log('getIP ');

		response.set('Content-Type', 'application/json');
		response.send({
			code: 0,
			message: 'success',
			data: {
				masterExternalIP: '127.0.0.1'
			}
		});
	}

	getMFACode(request: Request, response: Response, next: NextFunction) {
		console.log('getMFACode ');

		const result = {
			status: 'OK',
			data: {
				base32_secret: BASE32_SECRET,
				otpauth_url: OTP_AUTH_URL
			}
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	defaultReverseProxy(
		request: Request,
		response: Response,
		next: NextFunction
	) {
		console.log('defaultReverseProxy ');

		const result = {
			status: 'OK',
			data: {
				enable_frp: false
			}
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	getFirstFactor(request: Request, response: Response, next: NextFunction) {
		console.log('getFirstFactor ');
		console.log(JSON.stringify(request.body));

		const username = request.body['username'];
		const password = request.body['password'];

		if (username != TERMINUS_NAME.split('@')[0]) {
			response.set('Content-Type', 'application/json');

			response.status(401).send({
				status: 'Failed'
			});
			return;
		}

		const result = {
			status: 'OK',
			data: {
				access_token: ACCESS_TOKEN,
				refresh_token: REFRESH_TOKEN,
				fa2: true,
				redirect: '',
				session_id: SESSION_ID
			}
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	getSecondFactor(request: Request, response: Response, next: NextFunction) {
		console.log('getSecondFactor ' + request.body.token);

		const result: any = {
			status: 'OK',
			data: {
				access_token: ACCESS_TOKEN,
				refresh_token: REFRESH_TOKEN,
				fa2: true,
				redirect: '/abcd',
				session_id: SESSION_ID
			}
		};

		response.set('Content-Type', 'application/json');
		response.send(result);
	}

	getMoniter(request: Request, response: Response, next: NextFunction) {
		response.set('Content-Type', 'application/json');
		response.send({
			code: 0,
			message: null,
			data: {
				cpu: {
					name: 'cpu',
					color: 'color-FFEB3B',
					uint: '40',
					ratio: 40,
					total: 100,
					usage: 40
				},
				memory: {
					name: 'memory',
					color: 'color-ACF878',
					uint: '40',
					ratio: 40,
					total: 100,
					usage: 40
				},
				disk: {
					name: 'disk',
					color: 'color-5BCCF3',
					uint: '40',
					ratio: 40,
					total: 100,
					usage: 30
				},
				net: {
					received: 100000,
					transmitted: 200000
				}
			}
		});
	}
}
