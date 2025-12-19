import { i18n } from 'src/boot/i18n';

export enum UpgradeMode {
	DOWNLOAD_ONLY = 'DownloadOnly',
	DOWNLOAD_AND_UPGRADE = 'DownloadAndUpgrade'
}

export const upgradeModeOptions = () => {
	return [
		{
			label: i18n.global.t('Download only'),
			value: UpgradeMode.DOWNLOAD_ONLY,
			detail: i18n.global.t('Download now and install later'),
			enable: true
		},
		{
			label: i18n.global.t('Download and upgrade'),
			value: UpgradeMode.DOWNLOAD_AND_UPGRADE,
			detail: i18n.global.t('Download and install immediately'),
			enable: true
		}
	];
};
