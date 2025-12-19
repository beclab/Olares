import { defineStore } from 'pinia';
import axios from 'axios';
import { t } from 'src/boot/studio-i18n';
import { CreateWithOneImageConfig } from '@apps/studio/src/types/core';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDevelopingApps } from './app';
import {
	CreateWithOneDockerConfig,
	APP_STATUS,
	AppState,
	APP_INSTALL_STATE
} from '@apps/studio/src/types/core';
import { bus } from '@apps/studio/src/utils/bus';

const appStore = useDevelopingApps();

export type DockerStore = {
	appStatus: APP_STATUS | undefined;
	appInstallState: APP_INSTALL_STATE | undefined;
	appStatusInfo: any;
};

export const useDockerStore = defineStore('studio-docker', {
	state() {
		return {
			appStatus: undefined,
			appInstallState: undefined,
			appStatusInfo: undefined
		} as DockerStore;
	},
	actions: {
		async create_name(name: string): Promise<void> {
			await axios.post(appStore.url + '/api/command/apps/create', {
				title: name
			});
		},

		async config_app(config: CreateWithOneDockerConfig): Promise<any> {
			return axios.post(
				appStore.url + `/api/command/apps/${config.name}/create`,
				{
					...config,
					title: appStore.current_app?.title
				}
			);
		},

		async create_app(name: string): Promise<void> {
			await axios.post(
				appStore.url + `/api/command/apps/${name}/example/create`,
				{
					title: appStore.current_app?.title
				}
			);
		},

		async get_app_status(name: string): Promise<void> {
			const res: { state: APP_STATUS } = await axios.get(
				appStore.url + `/api/apps/${name}/status`
			);
			try {
				const target = appStore.apps.find((item) => item.appName === name);
				if (target) {
					target.state = res.state;
				}
			} catch (error) {
				//
			}
			this.appStatus = res.state;
		},

		async install_app(name: string): Promise<{ namespace: string }> {
			const res = await axios.post<
				{ namespace: string } | { code: number; message: string }
			>(appStore.url + '/api/command/install-app', {
				name,
				title: appStore.current_app?.title
			});
			try {
				const target = appStore.apps.find((item) => item.appName === name);
				if ('code' in res.data && res.data.code && target) {
					target.reason = res.data.message;
					target.hideErrorFooter = false;
				}
			} catch (error) {
				//
			}

			return res as unknown as { namespace: string };
		},

		async un_install_app(name: string, hideNotify = false): Promise<void> {
			await axios.post(appStore.url + `/api/command/uninstall/${name}`, {});
			if (hideNotify) return;
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.start_uninstalling')
			});
		},

		async delete_app(name: string): Promise<void> {
			await this.un_install_app(name, true);
			await axios.post(appStore.url + '/api/command/delete-app', {
				name
			});
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.delete_file_success')
			});
		},

		async rename_app(name: string, title: string): Promise<void> {
			await axios.put(appStore.url + `/api/command/apps/title/${name}`, {
				title
			});
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.rename_success')
			});
		},

		async get_app_install_state(name: string): Promise<AppState> {
			const res: AppState = await axios.get(
				appStore.url + `/api/app-state?app=${name}`
			);

			this.appInstallState = res.state;
			this.appStatusInfo = res;
			return res;
		},

		async create_app_code_in_olares(
			item: CreateWithOneImageConfig
		): Promise<any> {
			const exposePorts = item?.ports || undefined;
			const gpuVendor = item.requiredGpu ? item.gpuVendor : '';

			return axios.post(
				appStore.url + `/api/command/apps/${item.name}/vscode/create`,
				{
					devEnv: item.devEnv,
					requiredCpu: item.requiredCpu,
					requiredMemory: item.requiredMemory,
					requiredDisk: item.requiredDisk,
					title: appStore.current_app?.title,
					exposePorts,
					requiredGpu: item.requiredGpu,
					gpuVendor
				}
			);
		},

		async init() {
			bus.on('app_installation_event_studio', (message) => {
				console.log('app_installation_event_studio', message);
				if (`${appStore.current_app?.appName}-dev` === message?.name) {
					this.appStatusInfo = message;
				}
			});
		},

		clearMessage() {
			this.appStatusInfo = undefined;
		}
	}
});
