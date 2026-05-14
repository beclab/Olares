import { QVueGlobals } from 'quasar';
import { NormalApplication } from './base';
import { useTransfer2Store } from 'src/stores/transfer2';
import { importFilesStyle } from './utils/files';
import { registerFilePreviewEvent } from './utils/preview';
import { useFilesStore } from 'src/stores/files';
import { useFilesCopyStore } from 'src/stores/files-copy';
import { watch } from 'vue';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { useDeviceStore } from 'src/stores/settings/device';
import { DeviceType } from '@bytetrade/core';
import { WebsocketSharedWorkerEnum } from 'src/websocket/interface';
import { useWebsocketManager2Store } from 'src/stores/websocketManager2';
import { busOff, busOn } from 'src/utils/bus';
import { useOperateinStore, CopyStoragesType } from 'src/stores/operation';
import { FilesWSType } from 'src/websocket/public/files';

export class FilesApplication extends NormalApplication {
	applicationName = 'files';
	async appLoadPrepare(data: any): Promise<void> {
		//@ts-ignore
		// (() => import('../css/styles.css'))();
		// importFilesStyle(false);
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

		setTimeout(() => {
			const websocketStore = useWebsocketManager2Store();
			websocketStore.start();
		}, 1000);

		busOn('updateCopyItems', this.updateCopyitems);
		busOn('resetCopyItems', this.resetCopyitems);
	}
	async appUnMounted() {
		super.appUnMounted();

		busOff('updateCopyItems', this.updateCopyitems);
		busOff('resetCopyItems', this.resetCopyitems);
	}

	updateCopyitems() {
		const operationStore = useOperateinStore();
		const websocketStore = useWebsocketManager2Store();
		websocketStore.apply(
			FilesWSType.UpdateCopyItems,
			JSON.stringify({
				files: operationStore.copyFiles,
				isCut: operationStore.isCut
			})
		);
	}

	resetCopyitems() {
		const websocketStore = useWebsocketManager2Store();
		websocketStore.apply(FilesWSType.ResetCopyItems, '');
	}

	async appRedirectUrl(
		_redirect: any,
		_currentRoute: RouteLocationNormalizedLoaded
	) {
		const deviceStore = useDeviceStore();
		deviceStore.init(
			(state: { device: DeviceType; isVerticalScreen: boolean }) => {
				console.log(state);
				// if (!deviceStore.isMobile && state.device === DeviceType.MOBILE) {
				// 	router.replace('/');
				// }
			}
		);
	}

	initAxiosIntercepts(): void {
		super.initAxiosIntercepts();
	}

	applicationLanguageUpdate(terminusLanguage: string): void {
		const filesStore = useFilesStore();
		filesStore.getMenu();
	}

	websocketConfig = {
		useShareWorker: true,
		shareWorkerName: WebsocketSharedWorkerEnum.FILES_NAME,

		externalInfo() {
			return {};
		},
		responseShareWorkerMessage(data: { type: 'ws' | FilesWSType; data: any }) {
			if (data.type == FilesWSType.UpdateCopyItems) {
				const { files, isCut } = JSON.parse(data.data);
				const operationStore = useOperateinStore();
				operationStore.updateCopyFiles(files, isCut, false);
			} else if (data.type == FilesWSType.ResetCopyItems) {
				const operationStore = useOperateinStore();
				operationStore.resetCopyFiles(true, false);
			}
		}
	};
}
