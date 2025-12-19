import { defineStore } from 'pinia';
import axios from 'axios';
import { t } from 'src/boot/studio-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDevelopingApps } from './app';
import { FileItem, FilesCodeType } from '@apps/studio/src/types/core';

const appStore = useDevelopingApps();

export type DockerStore = {
	cached: Record<string, FileItem[]>;
	currentFileData: FileItem;
	filesEditStorage: [];
	isEditing: boolean;
	fileInfo: FilesCodeType;
};

export const useFileStore = defineStore('studio-file', {
	state() {
		return {
			cached: {},
			currentFileData: {},
			filesEditStorage: [],
			isEditing: false,
			fileInfo: {
				code: '',
				lang: 'json',
				name: ''
			}
		} as DockerStore;
	},
	actions: {
		async getFile(path: string): Promise<void> {
			if (path in this.cached) {
				this.currentFileData = this.cached[path];
			} else {
				this.currentFileData = [];
			}

			const res: FileItem = await axios.get(
				appStore.url + '/api/files/' + path
			);

			this.currentFileData = res;
			this.cached[path] = res;
		},

		async putFile(path: string): Promise<void> {
			await axios.post(appStore.url + '/api/files/' + path);

			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.create_file_success')
			});
		},

		async deleteFile(path: string): Promise<void> {
			await axios.delete(appStore.url + '/api/files/' + path);

			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.delete_file_success')
			});
		},

		async createFolder(path: string): Promise<void> {
			await axios.post(appStore.url + '/api/files/' + path + '?file_type=dir');

			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.create_folder_success')
			});
		},

		async pathFile(path: string, params: any, headers: any): Promise<void> {
			await axios.patch(appStore.url + '/api/files/' + path, params, headers);

			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('message.rename_folder_success')
			});
		},

		async saveFile(current_path: string): Promise<boolean> {
			try {
				const res = await axios.put(
					appStore.url + '/api/files/' + current_path,
					this.fileInfo.code,
					{
						headers: { 'content-type': 'text/plain' }
					}
				);

				this.isEditing = false;
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('message.save_file_success')
				});
				return Promise.resolve(true);
			} catch (error) {
				throw Error('save error');
			}
		},

		addEditStorage(path) {
			this.filesEditStorage = [...new Set([...this.filesEditStorage, path])];
		},

		deleteEditStorage(path) {
			this.filesEditStorage = this.filesEditStorage.filter(
				(item) => item !== path
			);
		}
	}
});
