/**
 * @param {*} num
 * @param {*} min
 * @param {*} max
 * @returns
 */
export const limitNumber = (num, min = 0, max = 100) => {
	const number = parseInt(num);
	if (Number.isNaN(number) || number < min) {
		return min;
	} else if (number > max) {
		return max;
	}
	return number;
};

export const limitFloat = (num, min = 0, max = 100) => {
	const number = parseFloat(num);
	if (Number.isNaN(number) || number < min) {
		return min;
	} else if (number > max) {
		return max;
	}
	return number;
};

/**
 * @param {*} arr
 * @param {*} val
 * @returns
 */
export const matchValue = (arr, val) => {
	if (arr.length === 0 || arr.includes(val)) {
		return val;
	}
	return arr[0];
};

/**
 * @param {*} delay
 * @returns
 */
export const sleep = (delay) =>
	new Promise((resolve) => {
		const timer = setTimeout(() => {
			clearTimeout(timer);
			resolve();
		}, delay);
	});

/**
 * @param {*} func
 * @param {*} delay
 * @returns
 */
export const debounce = (func, delay = 200) => {
	let timer = null;
	return (...args) => {
		timer && clearTimeout(timer);
		timer = setTimeout(() => {
			func(...args);
			clearTimeout(timer);
			timer = null;
		}, delay);
	};
};

/**
 * @param {*} func
 * @param {*} delay
 * @returns
 */
export const throttle = (func, delay = 200) => {
	let timer = null;
	let cache = null;
	return (...args) => {
		if (!timer) {
			func(...args);
			cache = null;
			timer = setTimeout(() => {
				if (cache) {
					func(...cache);
					cache = null;
				}
				clearTimeout(timer);
				timer = null;
			}, delay);
		} else {
			cache = args;
		}
	};
};

/**
 * @param {*} s
 * @param {*} c
 * @param {*} i
 * @returns
 */
export const isAllchar = (s, c, i = 0) => {
	while (i < s.length) {
		if (s[i] !== c) {
			return false;
		}
		i++;
	}
	return true;
};

/**
 * @param {*} s
 * @param {*} p
 * @returns
 */
export const isMatch = (s, p) => {
	if (s.length === 0 || p.length === 0) {
		return false;
	}

	p = '*' + p + '*';

	let [sIndex, pIndex] = [0, 0];
	let [sRecord, pRecord] = [-1, -1];
	while (sIndex < s.length && pRecord < p.length) {
		if (p[pIndex] === '*') {
			pIndex++;
			[sRecord, pRecord] = [sIndex, pIndex];
		} else if (s[sIndex] === p[pIndex]) {
			sIndex++;
			pIndex++;
		} else if (sRecord + 1 < s.length) {
			sRecord++;
			[sIndex, pIndex] = [sRecord, pRecord];
		} else {
			return false;
		}
	}

	if (p.length === pIndex) {
		return true;
	}

	return isAllchar(p, '*', pIndex);
};

/**
 * @param {*} o
 * @returns
 */
export const type = (o) => {
	const s = Object.prototype.toString.call(o);
	return s.match(/\[object (.*?)\]/)[1].toLowerCase();
};

/**
 * @param {*} text
 * @returns
 */
export const sha256 = async (text, salt) => {
	const data = new TextEncoder().encode(text + salt);
	const digest = await crypto.subtle.digest({ name: 'SHA-256' }, data);
	return [...new Uint8Array(digest)]
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
};

/**
 * @returns
 */
export const genEventName = () => btoa(Math.random()).slice(3, 11);

/**
 * @param {*} a
 * @param {*} b
 * @returns
 */
export const isSameSet = (a, b) => {
	const s = new Set([...a, ...b]);
	return s.size === a.size && s.size === b.size;
};

/**
 * @param {*} s
 * @param {*} c
 * @param {*} count
 * @returns
 */
export const removeEndchar = (s, c, count = 1) => {
	let i = s.length;
	while (i > s.length - count && s[i - 1] === c) {
		i--;
	}
	return s.slice(0, i);
};

/**
 * @param {*} str
 * @param {*} sign
 * @returns
 */
export const matchInputStr = (str, sign) => {
	switch (sign) {
		case '//':
			return str.match(/\/\/([\w-]+)\s+([^]+)/);
		case '\\':
			return str.match(/\\([\w-]+)\s+([^]+)/);
		case '\\\\':
			return str.match(/\\\\([\w-]+)\s+([^]+)/);
		case '>':
			return str.match(/>([\w-]+)\s+([^]+)/);
		case '>>':
			return str.match(/>>([\w-]+)\s+([^]+)/);
		default:
	}
	return str.match(/\/([\w-]+)\s+([^]+)/);
};

/**
 * @param {*} str
 * @returns
 */
export const isValidWord = (str) => {
	const regex = /^[a-zA-Z-]+$/;
	return regex.test(str);
};

/**
 * @param {*} blob
 * @returns
 */
export const blobToBase64 = (blob) => {
	return new Promise((resolve) => {
		const reader = new FileReader();
		reader.onloadend = () => resolve(reader.result);
		reader.readAsDataURL(blob);
	});
};

/**
 * @param {string} text
 * @param {boolean} partial
 * @returns {string}
 */
export const stripMarkdownCodeBlock = (text, partial = false) => {
	if (!text) return '';

	const fullMatch = text.match(/^```(?:json|xml|text)?\s*\n?([\s\S]*?)\n?```$/);
	if (fullMatch) {
		return fullMatch[1];
	}

	if (partial) {
		const startMatch = text.match(/^```(?:json|xml|text)?\s*\n?([\s\S]*)$/);
		if (startMatch) {
			text = startMatch[1];
		}

		const endMatch = text.match(/^([\s\S]*?)\n?```$/);
		if (endMatch) {
			text = endMatch[1];
		}
	}

	return text;
};

/**
 * @param {Array} supportedLangs
 * @param {string} fallback
 * @returns {string}
 */
export const getBrowserPreferredLang = (supportedLangs, fallback = 'en') => {
	const browserLang =
		(typeof navigator !== 'undefined' &&
			(navigator.language || navigator.userLanguage)) ||
		fallback;

	if (supportedLangs.includes(browserLang)) {
		return browserLang;
	}

	const langPrefix = browserLang.split('-')[0];
	if (supportedLangs.includes(langPrefix)) {
		return langPrefix;
	}

	if (langPrefix === 'zh') {
		if (browserLang.includes('TW') || browserLang.includes('HK')) {
			return 'zh-TW';
		}
		return 'zh-CN';
	}

	return fallback;
};

/**
 */
const BLOCK_ELEMENTS = [
	'p',
	'div',
	'h1',
	'h2',
	'h3',
	'h4',
	'h5',
	'h6',
	'li',
	'blockquote',
	'pre',
	'article',
	'section',
	'header',
	'footer',
	'aside',
	'main',
	'td',
	'th',
	'dt',
	'dd'
];

/**
 * @param {HTMLElement} element
 * @param {string} text
 * @param {number} lengthThreshold
 * @returns {boolean}
 */
export const shouldUseNewline = (element, text, lengthThreshold = 20) => {
	if (!element || !text) {
		return false;
	}

	const tagName = element.tagName?.toLowerCase();
	if (tagName && BLOCK_ELEMENTS.includes(tagName)) {
		return true;
	}

	if (text.length >= lengthThreshold) {
		return true;
	}

	return false;
};
