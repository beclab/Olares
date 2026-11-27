export const PATTERN_UTC_TIME =
	/^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+Z/;

export const defaultScrollAreaThumbStyle = {
	right: '2px',
	borderRadius: '12px',
	backgroundColor: 'rgba(24, 33, 79, 0.2)',
	width: '6px'
};

export const hideScrollAreaBottomThumbStyle = {
	bottom: '2px',
	backgroundColor: 'transparent'
};

export enum BackgroundMode {
	desktop = 'Desktop',
	login = 'Login'
}

export const imgContentModes: 'fill'[] =
	// | 'contain' | 'cover' | 'none' | 'scale-down'
	[
		'fill'
		// 'contain',
		// 'cover',
		// 'none',
		// 'scale-down'
	];

export {
	firstToUpper,
	firstToUpperWith_
} from '@apps/control-panel-common/src/constant';
