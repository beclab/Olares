import axios from 'axios';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { useUserStore } from 'src/stores/user';

const develop = false;
let host = '';
// let base_url = '';

function getParamsOfUser() {
	const userStore = useUserStore();
	const current_user = userStore?.current_user;

	const url = userStore.getModuleSever('agent'); // `https://agent.${current_user.local_name}.${current_user.domain_name}`;

	return {
		current_user,
		url
	};
}

const instance = axios.create({
	// withCredentials: true
});

const auth_token = '';

instance.interceptors.request.use(
	function (config) {
		// const userStore = useUserStore();
		// const current_user = userStore?.current_user;
		if (develop) {
			host = '';
			config.url = host + config.url;
			config.headers['X-Authorization'] = auth_token;
		} else {
			const userOption: any = getParamsOfUser();
			config.headers['X-Authorization'] = userOption.current_user.access_token;
			const base_url = userOption.url;
			config.url = base_url + config.url;
			host = base_url;
		}

		return config;
	},
	function (error) {
		return Promise.reject(error);
	}
);

export const getApps = () => {
	return instance.get('/api/controllers/console/api/apps', {
		params: {
			page: 1
		}
	});
};

export const deleteApp = (app_id) => {
	return instance.delete(`/api/controllers/console/api/apps/${app_id}`);
};

export const createApp = (params) => {
	return instance.post('/api/controllers/console/api/apps', params);
};

export const stopChat = (task_id) => {
	return instance.delete(
		`/api/controllers/console/api/apps/chat-messages/${task_id}/stop`
	);
};

export const getAppById = (app_id) => {
	return instance.get(`/api/controllers/console/api/apps/${app_id}`);
};

export const getMessageById = (app_id, mode, conversation_id) => {
	return instance.get(
		`/api/controllers/console/api/apps/${app_id}/${mode}-messages`,
		{
			params: {
				conversation_id
			}
		}
	);
};
// let userStore = undefined;

export async function getMessage2(params, app_id, mode, callbacks) {
	const userOption: any = getParamsOfUser();
	host = userOption.url;

	fetchEventSource(
		`${host}/api/controllers/console/api/apps/${app_id}/${mode}-messages`,
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-Authorization': userOption.current_user.access_token
			},
			body: JSON.stringify(params),
			...callbacks
		}
	);
}
