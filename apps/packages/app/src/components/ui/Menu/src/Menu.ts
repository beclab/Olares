export interface ItemCell {
	key: string;
	label: string;
	icon?: string;
	img?: string;
	activeImg?: string;
	count?: number;
	muted?: boolean;
	disable?: boolean;
	defaultClose?: boolean;
	disableClickable?: boolean;
}
export interface Item extends ItemCell {
	children?: Array<Item>;
}

export type Size = 'sm' | 'md';

export const defaultSize = 'md';
