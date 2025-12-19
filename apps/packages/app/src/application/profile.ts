import { NormalApplication } from './base';
import { useUserStore } from '../stores/profileUser';
import axios from 'axios';

export class profileApplication extends NormalApplication {
	applicationName = 'profile';
	async appLoadPrepare(data: any): Promise<void> {
		await super.appLoadPrepare(data);
	}

	async appMounted(): Promise<void> {
		const userStore = useUserStore();
		if (process.env.APPLICATION === 'EDITOR') {
			axios
				.get(window.location.origin + '/api/profile/init')
				.then((response) => {
					userStore.setUser(response.data.data.profile);
					userStore.setInfo(response.data.data.info);
					userStore.getNftAddress();
				});
		}

		const updateLayout = () => {
			userStore.isMobile = window.innerWidth < 768;
		};

		window.addEventListener('resize', updateLayout);
	}

	async appRedirectUrl(): Promise<void> {
		const userStore = useUserStore();
		if (process.env.APPLICATION === 'PREVIEW') {
			return await axios
				.get(window.location.origin + '/api/profile/init')
				.then((response) => {
					userStore.setUser(response.data.data.profile);
					userStore.setInfo(response.data.data.info);
					const profileName = userStore.olaresId;
					document.title = profileName;
				});
		}
	}

	async appUnMounted(): Promise<void> {
		await super.appUnMounted();
	}

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		if (process.env.APPLICATION === 'EDITOR') {
			this.requestIntercepts.push((config) => {
				config.headers['X-Unauth-Error'] = 'Non-Redirect';
				return config;
			});
		}
	}
}
