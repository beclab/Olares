import { DriveDataAPI } from './../index';
import { MenuItem } from 'src/utils/contact';
import { DriveMenuType } from './type';
import { i18n } from 'src/boot/i18n';
import { DriveType } from 'src/utils/interface/files';

export default class ExternalDataAPI extends DriveDataAPI {
	async fetchMenuRepo(): Promise<DriveMenuType[]> {
		return [
			{
				label: i18n.global.t(`files_menu.${MenuItem.EXTERNAL}`),
				key: MenuItem.EXTERNAL,
				icon: 'sym_r_hard_drive',
				driveType: DriveType.External
			}
		];
	}

	async formatRepotoPath(item: any): Promise<string> {
		return '/Files/' + item.key + '/';
	}
}
