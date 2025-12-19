import { useRouter } from 'vue-router';
import { Base64 } from 'js-base64';

import { useDataStore } from '../stores/data';

export function parseToken(token) {
	const store = useDataStore();
	const parts = token.split('.');
	if (parts.length !== 3) {
		throw new Error('token malformed');
	}

	const data = JSON.parse(Base64.decode(parts[1]));

	document.cookie = `auth=${token}; path=/`;
	localStorage.setItem('jwt', token);
	store.setJWT(token);
	store.setUser(data.user);
}

export async function validateLogin() {
	try {
		if (localStorage.getItem('jwt')) {
			await renew(localStorage.getItem('jwt'));
		}
	} catch (_) {
		console.warn('Invalid JWT token in storage'); // eslint-disable-line
	}
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export async function login(username, password) {
	const store = useDataStore();
	store.setUser({
		id: 1,
		locale: 'en',
		viewMode: 'list',
		singleClick: false,
		perm: {
			admin: true,
			execute: true,
			create: true,
			rename: true,
			modify: true,
			delete: true,
			share: true,
			download: true
		},
		commands: [],
		lockPassword: false,
		hideDotfiles: false,
		dateFormat: false
	});
}

export async function renew(jwt) {
	const store = useDataStore();
	const baseURL = store.baseURL();

	const res = await fetch(`${baseURL}/api/renew`, {
		method: 'POST',
		headers: {
			'X-Auth': jwt
		}
	});

	const body = await res.text();

	if (res.status === 200) {
		parseToken(body);
	} else {
		throw new Error(body);
	}
}

export async function signup(username, password) {
	const store = useDataStore();
	const baseURL = store.baseURL();
	const data = { username, password };

	const res = await fetch(`${baseURL}/api/signup`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (res.status !== 200) {
		throw new Error(res.status);
	}
}

export function logout() {
	const store = useDataStore();
	document.cookie = 'auth=; expires=Thu, 01 Jan 1970 00:00:01 GMT; path=/';
	const router = useRouter();
	store.commit('setJWT', '');
	store.commit('setUser', null);
	localStorage.setItem('jwt', null);
	router.push({ path: '/login' });
}

export function getCookie() {
	var cookieValue = null;
	var name = 'sfcsrftoken';
	if (document.cookie && document.cookie != '') {
		var cookies = document.cookie.split(';');
		for (var i = 0; i < cookies.length; i++) {
			var cookie = cookies[i].trim();
			// Does this cookie string begin with the name we want?
			if (cookie.substring(0, name.length + 1) == name + '=') {
				cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
				break;
			}
		}
	}
	return cookieValue;
}
