import bcrypt from 'bcrypt';
import axios from 'axios';
import {
	NotificationChannel,
	NotificationData,
	ResolutionResult
} from '@bytetrade/core';
//NotificationToken
const providerUrl = 'http://' + process.env.OS_SYSTEM_SERVER;
const OS_APP_KEY = process.env.OS_APP_KEY;
const OS_APP_SECRET = process.env.OS_APP_SECRET;

console.log('providerUrl ' + providerUrl);
console.log('OS_APP_KEY' + OS_APP_KEY);
console.log('OS_APP_SECRET' + OS_APP_SECRET);

export async function getAccessToken(
	group: string,
	dataType: string,
	op: string
) {
	const instance = axios.create({
		baseURL: providerUrl,
		timeout: 1000,
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS'
		}
	});

	const timestamp = (new Date().getTime() / 1000).toFixed(0);
	const text = OS_APP_KEY! + timestamp + OS_APP_SECRET;
	const token = await bcrypt.hash(text, 10);

	console.log('token ' + token);

	const response = await instance.post('/permission/v1alpha1/access', {
		app_key: OS_APP_KEY,
		timestamp: Number(timestamp),
		token: token,
		perm: {
			group: group,
			dataType: dataType,
			version: 'v1',
			ops: [op]
		}
	});
	console.log(response);
	if (!response || response.status != 200 || !response.data) {
		throw Error('Network error, error 10001');
	}

	const data = response.data;
	if (data.code === 0) {
		return data.data.access_token;
	}

	throw Error('Network error, error 10002');
}

export async function createToken(
	access_token: string,
	name: string,
	firebase_token: string
) {
	const instance = axios.create({
		baseURL: providerUrl,
		timeout: 1000,
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
			'X-Access-Token': access_token
		}
	});

	const url = '/system-server/v1alpha1/token/service.notification/v1';

	console.log(url);
	const response = await instance.post(url, {
		token: firebase_token,
		name: name
	});
	console.log(response);
	if (!response || response.status != 200 || !response.data) {
		// throw Error('Network error, please try again later')
		throw Error('Network error, error 10003');
	}

	const data = response.data;
	if (data.code === 0) {
		return data.data;
	}

	throw Error('Network error, error 10004');
}

// export async function createMessage(
// 	access_token: string,
// 	message: NotificationData
// ) {
// 	const instance = axios.create({
// 		baseURL: providerUrl,
// 		timeout: 1000,
// 		headers: {
// 			'Access-Control-Allow-Origin': '*',
// 			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
// 			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
// 			'X-Access-Token': access_token
// 		}
// 	});

// 	const url = '/system-server/v1alpha1/message/service.notification/v1';

// 	console.log(url);
// 	const response = await instance.post(url, message);
// 	console.log(response);
// 	if (!response || response.status != 200 || !response.data) {
// 		// throw Error('Network error, please try again later')
// 		throw Error('Network error, error 10003');
// 	}

// 	const data = response.data;
// 	if (data.code === 0) {
// 		return data.data;
// 	}

// 	throw Error('Network error, error 10004');
// }

// export async function resolveByName(
// 	access_token: string,
// 	name: string
// ): Promise<ResolutionResult> {
// 	const instance = axios.create({
// 		baseURL: providerUrl,
// 		timeout: 1000,
// 		headers: {
// 			'Access-Control-Allow-Origin': '*',
// 			'Access-Control-Allow-Headers': 'X-Requested-With,Content-Type',
// 			'Access-Control-Allow-Methods': 'PUT,POST,GET,DELETE,OPTIONS',
// 			'X-Access-Token': access_token
// 		}
// 	});

// 	const url = '/system-server/v1alpha1/message/service.did/v1';

// 	console.log(url);
// 	const response = await instance.post(url, { name });
// 	console.log(response);
// 	if (!response || response.status != 200 || !response.data) {
// 		// throw Error('Network error, please try again later')
// 		throw Error('Resolve did by name falied');
// 	}

// 	const data = response.data;
// 	console.log('resolve_by_name');
// 	console.log(data.data);
// 	if (data.code === 0) {
// 		return data.data;
// 	}

// 	throw Error('Resolve did by name falied');
// }
