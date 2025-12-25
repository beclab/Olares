import { defineStore } from 'pinia';
import { MenuData, MenuType, TopicInfo } from 'src/constant/constants';
import { i18n } from 'src/boot/i18n';

export type MenuState = {
	currentItem: string;
	menuList: MenuData[];
	appCategories: string[];
	leftDrawerOpen: boolean;
};

export const useMenuStore = defineStore('menu', {
	state: () => {
		return {
			currentItem: 'All',
			menuList: [],
			appCategories: [],
			leftDrawerOpen: true
		} as MenuState;
	},
	getters: {
		categoryMenu(state): MenuType[] {
			return state.menuList
				.filter((item) => {
					return this.appCategories.includes(item.name) || item.name === 'All';
				})
				.map((item) => {
					return {
						key: item.name,
						label: item.title[i18n.global.locale.value],
						img: item.icon,
						sort: item.sort
					} as MenuType;
				})
				.sort((a, b) => a.sort - b.sort);
		}
	},
	actions: {
		getCategoryName(category: string) {
			const menu = this.menuList.find((item) => item.name === category);
			if (menu) {
				return menu.title[i18n.global.locale.value];
			}
			return '';
		},
		changeItemMenu(item: string) {
			this.currentItem = item;
		},
		setDrawerOpen(open: boolean) {
			this.leftDrawerOpen = open;
		}
	}
});
