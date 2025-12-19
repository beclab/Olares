import { Router } from 'vue-router';
interface AuthParams {
	username: string;
	password: string;
	url: string;
}
export function updateUIToAuthorizationPage(
	router: Router,
	data: AuthParams[]
) {
	router.push({
		path: '/authorization',
		query: {
			users: JSON.stringify(data)
		}
	});
}
