import { defineStore } from 'pinia';
import { MenuLabelType, DocumentType } from '@apps/studio/src/types/core';
import { useDevelopingApps } from './app';
import { t } from 'src/boot/studio-i18n';

export enum MenuLabel {
	STUDIO = 'Studio',
	HOME = 'Home',
	CONTAINERS = 'Containers',
	APPLICATIONS = 'Applications'
}

export type DataState = {
	homeMenu: MenuLabelType[];
	applicationMenu: MenuLabelType[];
	documentList: DocumentType[];
	currentItem: string;
};

export const useMenuStore = defineStore('studio-menu', {
	state() {
		return {
			homeMenu: [
				{
					label: t(`enums.${MenuLabel.STUDIO}`),
					key: MenuLabel.STUDIO,
					icon: '',
					children: [
						{
							label: t(`enums.${MenuLabel.HOME}`),
							key: MenuLabel.HOME,
							icon: 'sym_r_home',
							muted: true
						}
						// {
						// 	label: t(`enums.${MenuLabel.CONTAINERS}`),
						// 	key: MenuLabel.CONTAINERS,
						// 	icon: 'sym_r_deployed_code',
						// 	muted: true
						// }
					]
				}
			],
			applicationMenu: [
				{
					label: t(`enums.${MenuLabel.APPLICATIONS}`),
					key: MenuLabel.APPLICATIONS,
					icon: '',
					children: []
				}
			],

			currentItem: MenuLabel.HOME,
			documentList: [
				{
					id: 1,
					message: 'Studio Tutorial',
					link: 'https://docs.olares.com/developer/develop/tutorial/'
				},
				{
					id: 3,
					message: 'Understand Olares Application Chart',
					link: 'https://docs.olares.com/developer/develop/package/chart.html'
				},
				{
					id: 4,
					message: 'OlaresManifest configuration guide',
					link: 'https://docs.olares.com/developer/develop/package/manifest.html'
				},
				{
					id: 5,
					message: 'How to submit an application',
					link: 'https://docs.olares.com/developer/develop/submit/'
				}
			]
		} as DataState;
	},

	getters: {
		menuList(state) {
			return [...state.homeMenu, ...state.applicationMenu];
		}
	},
	actions: {
		updateApplications() {
			const store = useDevelopingApps();
			this.applicationMenu[0].children = [];
			for (const app of store.apps) {
				this.applicationMenu[0].children.push({
					label: app.title || app.appName,
					key: `/app/${app.appName}`,
					icon: 'sym_r_grid_view'
				});
			}
		},

		updatePathToMenu(path: string) {
			const splitPath = path.split('/');
			if (splitPath.length <= 0) return false;
			switch (splitPath[1]) {
				case 'home':
				case 'create':
					this.currentItem = MenuLabel.HOME;
					break;

				case 'containers':
					this.currentItem = MenuLabel.CONTAINERS;
					break;

				case 'app':
					this.currentItem = `/app/${decodeURIComponent(splitPath[2])}`;
					break;

				default:
					break;
			}
		}
	}
});
