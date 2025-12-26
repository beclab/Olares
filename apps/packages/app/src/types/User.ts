export interface Social {
	username: string;
	platform: string;
	url: string;
}

export interface Block {
	id: string;
	type: string;
	nickName: string;
	enabled: boolean;
}

export interface LinkBlock extends Block {
	url: string;
	title: string;
	subTitle: string;
	size: SIZE_TYPE;
	textAlignment: string;
	outline: boolean;
	shadow: boolean;
	icon: string;
}

export interface TextBlock extends Block {
	title: string;
	description: string;
	textAlignment: ALIGNMENT_TYPE;
	transparent: boolean;
}

export interface ImageBlock extends Block {
	title: string;
	description: string;
	link: string;
	img: string;
	ratio: string;
}

export function getGradientColor(top: string, bottom: string): string {
	return `linear-gradient(180deg, ${top} 0%, ${bottom} 100%)`;
}

export interface User {
	header: UserHeader;
	appearance: {
		theme: AppearanceTheme;
		block: {
			cornerRadius: number;
			style: BLOCK_STYLE_TYPE;
			shadow: boolean;
			outline: boolean;
			transparency: number;
		};
		font: string;
	};
	social: UserSocial;
	block: {
		data: Block[];
	};
	layout: {
		style: number;
	};
}

export interface UserHeader {
	style: HEADER_STYLE_TYPE;
	banner: string;
	avatarUrl: string;
	profileShape: PROFILE_SHAPE_TYPE;
	profileSize: number;
	profileOutline: boolean;
	nickName: string;
	description: string;
	textSize: SIZE_TYPE;
	format: HEADER_FORMAT_TYPE;
}

export interface UserSocial {
	data: Social[];
	size: SIZE_TYPE;
}

export interface AppearanceTheme {
	style: THEME_TYPE;
	preset: string;
	uploadImg: string;
	background: string;
	filter: IMAGE_FILTER;
	gradientTopColor: string;
	gradientBottomColor: string;
	useUpload: boolean;
	localImg: string;
	header: {
		textColor: string;
	};
	link: {
		background: string;
		textColor: string;
	};
	block: {
		background: string;
		textColor: string;
	};
}

export function generateUniqueId(existingIds) {
	let newId;
	do {
		newId = Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
	} while (existingIds.includes(newId));

	return newId;
}

export enum HEADER_FORMAT_TYPE {
	COLUMN = 'column',
	ROW = 'row'
}

export enum THEME_TYPE {
	SOLID = 'solid',
	GRADIENT = 'gradient',
	IMAGE = 'image'
}

export enum BLOCK_STYLE_TYPE {
	SQUARE = 'square',
	ROUND = 'round',
	CUSTOM = 'custom'
}

export enum SIZE_TYPE {
	SMALL = 'small',
	MEDIUM = 'medium',
	LARGER = 'larger'
}

export enum IMAGE_FILTER {
	NONE = 'none',
	DARK = 'dark',
	Light = 'light'
}

export enum HEADER_STYLE_TYPE {
	CLASSIC = 'classic',
	PORTRAIT = 'portrait',
	BANNER = 'banner'
}

export enum PROFILE_SHAPE_TYPE {
	CIRCULAR = 'circular',
	SQUARE = 'square'
}

export enum BLOCK_TYPE {
	LINK = 'link',
	TEXT = 'text',
	IMAGE = 'image'
}

export enum RATIO_TYPE {
	AUTO = '',
	RATIO_1_1 = '1',
	RATIO_3_2 = '1.5',
	RATIO_16_9 = '1.77778',
	RATIO_3_1 = '3'
}

export enum ALIGNMENT_TYPE {
	LEFT = 'left',
	CENTER = 'center',
	RIGHT = 'right'
}

export enum FONT_TYPE {
	ROBOTO = 'Roboto',
	POPINS = 'Popins',
	ARVO = 'Arvo',
	BITTER = 'Bitter',
	RIGHTEOUS = 'Righteous',
	LOBSTER = 'Lobster',
	ORBITRON = 'Orbitron',
	AUDIOWIDE = 'Audiowide',
	COURGETTE = 'Courgette'
}
