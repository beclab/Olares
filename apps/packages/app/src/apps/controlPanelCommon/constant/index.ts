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

export const ACRONYMS = new Set(['CPU', 'GPU', 'API', 'IO', 'RAM', 'DNS']);

export const firstToUpper = (str: string) => {
	const value = str.trim();
	if (value.length === 0) {
		return str;
	}
	return value
		.split(/\s+/)
		.map((item, index) => {
			const upperCaseItem = item.toUpperCase();
			if (ACRONYMS.has(upperCaseItem)) {
				return upperCaseItem;
			}
			const lowerCaseItem = item.toLowerCase();
			return index === 0
				? lowerCaseItem.charAt(0).toUpperCase() + lowerCaseItem.slice(1)
				: lowerCaseItem;
		})
		.join(' ');
};

export const firstToUpperWith_ = (title: string) => {
	return title
		.split('_')
		.map((item, index) => {
			const value = firstToUpper(item);
			return index === 0 || ACRONYMS.has(value.toUpperCase())
				? value
				: value.charAt(0).toLowerCase() + value.slice(1);
		})
		.join(' ');
};
