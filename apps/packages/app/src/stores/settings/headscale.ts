import { defineStore } from 'pinia';
import { User } from 'src/headscale/classes';
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
			this.headScaleStatus = data.deny_all == 1;
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
			const endpointURL = '/node';

			//returning variables
			let headscaleDevices = [new HeadScaleDevice()];

			// attempt to get the user data
			try {
				const data: any = await axios.get(this.headscaleURL + endpointURL);

				headscaleDevices = sortDevices(data.nodes);

				this.status = 'succeeded';

				// for (let i = 0; i < headscaleDevices.length; i++) {
				// for (let j = 0; j < headscaleDevices[i].forcedTags.length; ++j) {
				// 	headscaleDevices[i].forcedTags[j] = headscaleDevices[i].forcedTags[
				// 		j
				// 	].replace('tag:', '');
				// }
				// }

				this.devices = headscaleDevices;
			} catch (e) {
				console.log(e);
				this.status = 'failed';
			}
		},

		async renameDevice(deviceID: string, name: string): Promise<any> {
			// endpoint url for editing users
			const endpointURL = '/node/' + deviceID + '/rename/' + name;
			try {
				const data: any = await axios.post(this.headscaleURL + endpointURL);
				if (data && data.node) {
					this.devices = this.devices.map((item) =>
						item.id === data.node.id ? data.node : item
					);
				}
				return data;
			} catch (e) {
				console.log(e);
				return;
			}
		},

		async removeDevice(deviceID: string): Promise<any> {
			// endpoint url for removing devices
			const endpointURL = '/node/' + deviceID;
			try {
				const data: any = await axios.delete(this.headscaleURL + endpointURL);

				await this.getDevices();
				return data;
			} catch (e) {
				console.log(e);
				return;
			}
		},

		async approveRoutes(nodeId: string, routeList: string[]): Promise<any> {
			try {
				const data: any = await axios.post(
					this.headscaleURL + '/node/approve_routes',
					{
						id: nodeId,
						routes: routeList
					}
				);
				if (data && data.node) {
					this.devices = this.devices.map((item) =>
						item.id === data.node.id ? data.node : item
					);
				}
			} catch (e) {
				console.log(e);
			}
		},

		async updateTags(deviceID: string, tags: string[]): Promise<any> {
			// endpoint url for editing users
			const endpointURL = '/node/' + deviceID + '/tags';
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
