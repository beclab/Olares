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

export class FilesApplication extends NormalApplication {
	applicationName = 'files';
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
}
