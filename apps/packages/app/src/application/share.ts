import { QVueGlobals } from 'quasar';
import { NormalApplication } from './base';
import { useTransfer2Store } from 'src/stores/transfer2';
import { importFilesStyle } from './utils/files';
import { registerFilePreviewEvent } from './utils/preview';
import { useFilesStore } from 'src/stores/files';
import { useFilesCopyStore } from 'src/stores/files-copy';
import { watch } from 'vue';
import { useTokenStore } from 'src/stores/share/token';
import { useShareStore } from 'src/stores/share/share';
import { notifyFailed, notifyWarning } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';
import { getSuitableValue } from 'src/utils/monitoring';
import { Quasar } from 'quasar';

export class ShareApplication extends NormalApplication {
	applicationName = 'share';
	async appLoadPrepare(data: any): Promise<void> {
		//@ts-ignore
		(() => import('../css/styles.css'))();
		importFilesStyle(false);
		await super.appLoadPrepare(data);
		const quasar = data.quasar as QVueGlobals;
		registerFilePreviewEvent(quasar);
	}

	async appMounted(): Promise<void> {
		super.appMounted();
		const transferStore = useTransfer2Store();
		const filesCopyStore = useFilesCopyStore();
		transferStore.init();
		const filesStore = useFilesStore();
		watch(
			() => filesStore.nodes,
			(newName) => {
				if (newName && newName.length > 0) {
					filesStore.nodes.forEach((node) => {
						filesCopyStore.initialize(node.name);
					});
				}
			},
			{ immediate: true }
		);
	}

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
		this.requestIntercepts.push((config) => {
			config.headers['Access-Control-Allow-Origin'] = '*';
			config.headers['Access-Control-Allow-Headers'] =
				'X-Requested-With,Content-Type';
			config.headers['Access-Control-Allow-Methods'] =
				'PUT,POST,GET,DELETE,OPTIONS';

			return config;
		});

		this.responseIntercepts.push((response, router) => {
			if (response && response.status == 401) {
				router.push({ path: '/login' });
				return;
			}

			if (!response || response.status != 200 || !response.data) {
				throw Error('Network error, please try again later');
			}

			const data = response.data;
			if (data.code == 100001) {
				router.push({ path: '/login' });
				throw Error(data.message);
			}

			if (data.status) {
				if (data.status === 'OK') {
					return data.data;
				}
				throw Error(data.status);
			} else {
				if (data.code != 0) {
					throw Error(data.message);
				}

				return data.data;
			}
		});
	}

	async appRedirectUrl(redirect: any): Promise<void> {
		const tokenStore = useTokenStore();
		let host = '';
		if (typeof window !== 'undefined') {
			host = window.location.origin;
		}

		tokenStore.setUrl(host);

		return await tokenStore.loadData().then(async () => {
			if (document.getElementById('Loading'))
				document.getElementById('Loading')?.remove();
		});
	}

	commonTokenInvalidIntercept() {
		this.commonRequestIntercepts = (config) => {
			const shareStore = useShareStore();

			if (shareStore.token) {
				if (!config.params) {
					config.params = {};
				}
				config.params['token'] = shareStore.token;
			}
			return config;
		};
		this.tokenInvalidErrorIntercep = (error) => {
			const shareStore = useShareStore();

			if (error && error.response && error.response.status == 569) {
				shareStore.deleteToken();
				return true;
			}
			if (error && error.response && error.response.status == 559) {
				shareStore.expiredInfo.status = true;
				shareStore.expiredInfo.time = error.response.data.expire;
				return true;
			}

			return false;
		};
	}

	filesUploadConfig = {
		autoBindResumable: false,
		filesUpdate: (origin_id: number, event: any) => {
			const data = this.filterFiles(event.target.files);
			if (data.remove) {
				this.showRemoveFiles(data.remove);
			}
			event.target.files = data.filter;
			const fileStore = useFilesStore();

			fileStore.uploadSelectFile(
				event,
				fileStore.currentPath[origin_id],
				origin_id
			);
		},
		filesFilter: (files: FileList) => {
			const data = this.filterFiles(files);
			if (data.remove) {
				this.showRemoveFiles(data.remove);
			}
			return data.filter;
		},
		toastDeleteFiles: (files: File[]) => {
			this.showRemoveFiles(files);
		}
	};

	private showRemoveFiles(remove: File[]) {
		if (remove.length <= 0) {
			return;
		}

		const shareStore = useShareStore();
		notifyFailed(
			i18n.global.t(
				'share.The following files exceed the file upload limit, the file size cannot exceed {size}<br>{files}',
				{
					files:
						(remove.length > 10 ? remove.slice(0, 10) : remove)
							.map((e) => {
								return (
									(!!e.webkitRelativePath ? e.webkitRelativePath : e.name) +
									`(${getSuitableValue(`${e.size}`, 'disk')})`
								);
							})
							.join('<br>') + (remove.length > 10 ? '<br>...' : ''),
					size: getSuitableValue(
						`${shareStore.share?.upload_size_limit}`,
						'disk'
					)
				}
			)
		);
	}

	private filterFiles = (files: FileList) => {
		const shareStore = useShareStore();
		if (
			!shareStore.share ||
			shareStore.share.upload_size_limit == undefined ||
			shareStore.share.upload_size_limit == 0
		) {
			return {
				filter: files,
				remove: []
			};
		}

		const filterFiles: File[] = [];
		const removeFiles: File[] = [];

		Array.from(files).forEach((file) => {
			if (file.size <= shareStore.share!.upload_size_limit!) {
				filterFiles.push(file);
			} else {
				removeFiles.push(file);
			}
		});

		const dataTransfer = new DataTransfer();

		filterFiles.forEach((file: any) => dataTransfer.items.add(file));

		return {
			filter: dataTransfer.files,
			remove: removeFiles
		};
	};
}
