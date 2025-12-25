import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('src/layouts/profile/ProfilePreviewLayout.vue')
	}
];
export default routes;
