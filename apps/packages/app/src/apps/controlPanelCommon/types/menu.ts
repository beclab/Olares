export interface MenuItem {
	title: string;

	code: string;

	icon?: string;

	children?: MenuItem[];
}

export interface MenuOptions {
	title?: string;
	code?: string;
	icon?: string;
	children?: string;
}

export const defaultMenuOptions: MenuOptions = {
	title: 'title',
	code: 'code',
	icon: 'icon',
	children: 'children'
};
