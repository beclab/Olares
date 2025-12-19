import { TerminusCommonPlatform } from './terminalCommonPlatform';
import { isPad } from 'src/utils/platform';
import { useFilesStore, FilesIdType } from '../../stores/files';
import { watch } from 'vue';
import { useTransfer2Store } from 'src/stores/transfer2';
import { useTermipassStore } from '../../stores/termipass';
import { useDeviceStore } from 'src/stores/device';
import { useDeviceStore as useDeviceStore2 } from 'src/stores/settings/device';
import { tabbarItems } from 'src/platform/interface/bex/front/bexTabOptions';
import { useMDNSStore } from 'src/stores/mdns';
import { DeviceType } from '@bytetrade/core';

export class MobileWebPlatform extends TerminusCommonPlatform {
	async appLoadPrepare(data: any): Promise<void> {
		await super.appLoadPrepare(data);
	}

	async appMounted() {
		super.appMounted();
		this.isPad = isPad();

		const transferStore = useTransfer2Store();
		const termipassStore = useTermipassStore();

		const mndsStore = useMDNSStore();
		const deviceStore2 = useDeviceStore2();
		deviceStore2.deviceInfo.device = DeviceType.MOBILE;
		mndsStore.init();

		watch(
			() => [transferStore.downloading, transferStore.uploading],
			(newValue) => {
				const downloadingsNum = newValue[0].length;
				const uploadingNum = newValue[1].length;
				let transferBadge: undefined | string = undefined;
				if (downloadingsNum + uploadingNum > 99) {
					transferBadge = '99+';
				} else {
					transferBadge =
						downloadingsNum + uploadingNum > 0
							? `${downloadingsNum + uploadingNum}`
							: undefined;
				}
				termipassStore.tabItems[1].badge = transferBadge;
			}
		);
	}
	tabbarItems = process.env.DEV_PLATFORM_BEX
		? tabbarItems
		: [
				{
					name: 'files.files',
					identify: 'file',
					normalImage: 'tab_files_normal',
					activeImage: 'tab_files_active',
					darkActiveImage: 'tab_files_dark_active',
					to: '/home',
					tabChanged: () => {
						const filesStore = useFilesStore();
						if (filesStore.backStack[FilesIdType.PAGEID].length > 0) {
							const params = new URLSearchParams(
								filesStore.backStack[FilesIdType.PAGEID][
									filesStore.backStack[FilesIdType.PAGEID].length - 1
								].param
							);
							const query = Object.fromEntries(params);

							filesStore.router.replace({
								path: filesStore.backStack[FilesIdType.PAGEID][
									filesStore.backStack[FilesIdType.PAGEID].length - 1
								].path,
								query
							});
							return true;
						}
						if (filesStore.mobileRepo) {
							filesStore.router.replace(filesStore.mobileRepo);
							return true;
						}
						return false;
					}
				},
				{
					name: 'transmission.title',
					identify: 'transfer',
					normalImage: 'tab_transfer_normal',
					activeImage: 'tab_transfer_active',
					to: '/transfer/history',
					badge: ''
				},
				{
					name: 'Vault',
					identify: 'secret',
					normalImage: 'tab_secret_normal',
					activeImage: 'tab_secret_active',
					to: '/secret'
				},
				{
					name: 'setting',
					identify: 'setting',
					normalImage: 'tab_setting_normal',
					activeImage: 'tab_setting_active',
					to: '/setting'
				}
		  ];

	async homeMounted(): Promise<void> {
		await super.homeMounted();
	}

	isMobile = true;

	isTabbarDisplay() {
		const deviceStore = useDeviceStore();
		if (deviceStore.isScaning) {
			return false;
		}
		const fileStore = useFilesStore();
		return !fileStore.isShard;
	}
}
