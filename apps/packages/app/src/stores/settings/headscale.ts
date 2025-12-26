import { defineStore } from 'pinia';
import { User, Route } from 'src/headscale/classes';
import { HeadScaleDevice } from '@bytetrade/core';
import { sortDevices } from 'src/headscale/sort';
import axios from 'axios';
import { useTokenStore } from './token';
import { i18n } from 'src/boot/i18n';
import { notifySuccess } from 'src/utils/settings/btNotify';

export type HeadScaleState = {
	status: string;
	headscaleURL: string | null;
	devices: HeadScaleDevice[];
	users: User[];
	headScaleStatus: boolean;
};

export const useHeadScaleStore = defineStore('headScale', {
	state: () => {
		return {
			status: '',
			headscaleURL: null,
			devices: [],
			users: [],
			headScaleStatus: false
		} as HeadScaleState;
	},
	getters: {},
	actions: {
		setUrl(new_url: string | null) {
			this.headscaleURL = new_url;
		},

		async getHeadScaleStatus() {
			const tokenStore = useTokenStore();

			const data: any = await axios.get(
				tokenStore.url + '/api/launcher-public-domain-access-policy'
			);
			this.headScaleStatus = data.deny_all == 1 ? true : false;
		},

		async toggleHeadScaleStatus() {
			const tokenStore = useTokenStore();

			await axios.post(
				tokenStore.url + '/api/launcher-public-domain-access-policy',
				{
					deny_all: this.headScaleStatus ? 0 : 1
				}
			);
			notifySuccess(i18n.global.t('success'));
		},

		async getDevices(): Promise<any> {
			// endpoint url for getting users
			const endpointURL = '/machine';

			//returning variables
			let headscaleDevices = [new HeadScaleDevice()];

			// attempt to get the user data
			try {
				const data: any = await axios.get(this.headscaleURL + endpointURL);

				headscaleDevices = sortDevices(data.machines);

				this.status = 'succeeded';

				for (let i = 0; i < headscaleDevices.length; i++) {
					for (let j = 0; j < headscaleDevices[i].forcedTags.length; ++j) {
						headscaleDevices[i].forcedTags[j] = headscaleDevices[i].forcedTags[
							j
						].replace('tag:', '');
					}
				}

				this.devices = headscaleDevices;
			} catch (e) {
				console.log(e);
				this.status = 'failed';
			}
		},

		async renameDevice(deviceID: string, name: string): Promise<any> {
			// endpoint url for editing users
			const endpointURL = '/machine/' + deviceID + '/rename/' + name;
			try {
				const data: any = await axios.post(this.headscaleURL + endpointURL);
				await this.getDevices();
				return data;
			} catch (e) {
				console.log(e);
				return;
			}
		},

		async removeDevice(deviceID: string): Promise<any> {
			// endpoint url for removing devices
			const endpointURL = '/machine/' + deviceID;
			try {
				const data: any = await axios.delete(this.headscaleURL + endpointURL);

				await this.getDevices();
				return data;
			} catch (e) {
				console.log(e);
				return;
			}
		},

		async getDeviceRoutes(deviceID: string): Promise<Route[]> {
			const endpointURL = '/machine/' + deviceID + '/routes';

			try {
				const data: any = await axios.get(this.headscaleURL + endpointURL);

				return data.routes;
			} catch (e) {
				console.log(e);
				return [];
			}
		},

		async modifyDeviceRoutes(
			routeList: Route[],
			routeID: number
		): Promise<any> {
			let endpointURL = '';

			routeList.forEach((route) => {
				if (route.id == routeID) {
					endpointURL = `/routes/${routeID}/`;
					if (route.enabled) {
						endpointURL += 'disable';
					} else {
						endpointURL += 'enable';
					}
				}
			});
			try {
				await axios.post(this.headscaleURL + endpointURL, {});
			} catch (e) {
				console.log(e);
			}
		},

		async updateTags(deviceID: string, tags: string[]): Promise<any> {
			// endpoint url for editing users
			const endpointURL = '/machine/' + deviceID + '/tags';
			const array: string[] = [];
			for (const t of tags) {
				array.push('tag:' + t);
			}

			try {
				const data = await axios.post(this.headscaleURL + endpointURL, {
					tags: array
				});
				await this.getDevices();
				return data;
			} catch (e) {
				console.log(e);
			}
		}
	}
});
