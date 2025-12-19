import { date } from 'quasar';
import { i18n } from 'src/boot/i18n';
import { useI18n } from 'vue-i18n';

export enum EllipsisPositon {
	left = 1,
	middile = 2,
	right = 3
}

export const generateStringEllipsis = (
	text: string,
	maxLength = 10,
	ellipsisString = '...',
	position = EllipsisPositon.middile
) => {
	if (!text) {
		return '';
	}
	if (text.length + ellipsisString.length <= maxLength) {
		return text;
	}
	if (ellipsisString.length >= maxLength) {
		return ellipsisString;
	}
	const subEllStringMaxLength = maxLength - ellipsisString.length;

	if (position === EllipsisPositon.left) {
		return text.slice(0, subEllStringMaxLength) + ellipsisString;
	} else if (position === EllipsisPositon.right) {
		return ellipsisString + text.slice(-subEllStringMaxLength);
	}
	return (
		text.slice(0, Math.floor(subEllStringMaxLength / 2)) +
		ellipsisString +
		text.slice(Math.floor(subEllStringMaxLength / 2) * -1)
	);
};

export const hiddenChar = (str: string, frontLen: number, endLen: number) => {
	if (str.length > frontLen + endLen + 3) {
		const ellipsis = '...';
		return (
			str.substring(0, frontLen) + ellipsis + str.substring(str.length - endLen)
		);
	} else {
		return str;
	}
};

export const getPastTime = (stamp1: Date, stamp2: Date) => {
	const { t } = useI18n();

	const time = stamp1.getTime() - stamp2.getTime();
	const second = time / 1000;
	const minute = second / 60;
	if (minute < 1) {
		return t('just_now');
	}
	if (minute < 60) {
		return `${minute.toFixed(0)}` + ' ' + t('minutes_ago');
	}

	const hour = minute / 60;
	if (hour < 24) {
		return `${hour.toFixed(0)}` + ' ' + t('hours_ago');
	}

	const day = hour / 24;
	if (day < 30) {
		return `${day.toFixed(0)} ` + ' ' + t('days_ago');
	}

	const month = day / 30;

	if (month < 12) {
		return `${month.toFixed(0)} months ago` + ' ' + t('months_ago');
	}

	const year = month / 12;
	return `${year.toFixed(0)} years ago` + ' ' + t('years_ago');
};

export const formatStampTime = (
	stamp: number,
	compare = new Date().getTime()
) => {
	const { t } = useI18n();

	const time = compare - stamp;
	const second = time / 1000;
	const minute = second / 60;

	if (minute <= 1) {
		return t('just_now');
	}

	const hour = minute / 60;
	if (hour < 24) {
		return date.formatDate(stamp, 'MM-DD HH:mm');
	}

	const day = hour / 24;

	if (day < 30) {
		return date.formatDate(stamp, 'MM-DD HH:mm');
	}

	return date.formatDate(stamp, 'MMM DD, YYYY');
};

export const formatDateToDMHM = (stamp1: Date, stamp2: Date) => {
	const time = stamp1.getTime() - stamp2.getTime();
	const seconds = Math.floor(time / 1000);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);

	const leftDays = Math.floor(hours / 24);
	const leftHours = Math.floor(hours % 24);
	const leftMinute = Math.floor(minutes % 60);
	const leftSeconds = Math.floor(seconds % 60);

	let secondsStr = '';
	let minutesStr = '';
	let hoursStr = '';
	let daysStr = '';

	if (leftDays > 0) {
		daysStr = leftDays + i18n.global.t('time.days_short').toUpperCase();
	}

	if (leftDays > 0 || leftHours > 0) {
		hoursStr = leftHours + i18n.global.t('time.hour_short').toUpperCase();
	}

	if (leftDays > 0 || leftHours > 0 || leftMinute > 0) {
		minutesStr = leftMinute + i18n.global.t('time.minutes_short').toUpperCase();
	}

	if (leftDays > 0 || leftHours > 0 || leftMinute > 0 || leftSeconds > 0) {
		secondsStr =
			leftSeconds + i18n.global.t('time.seconds_short').toUpperCase();
	}

	const formatTimes = [daysStr, hoursStr, minutesStr, secondsStr]
		.filter((e) => e.length > 0)
		.join(' ');

	return formatTimes;
};

export const UI_TYPE = {
	Tab: 'tab',
	Pop: 'index',
	Notification: 'notification'
};

type UiTypeCheck = {
	isTab: boolean;
	isNotification: boolean;
	isPop: boolean;
};

export const getUiType = (route?: any): UiTypeCheck => {
	// const location = window.location;
	// const pathname = location.pathname;
	// return Object.entries(UI_TYPE).reduce((m: any, [key, value]) => {
	// 	m[`is${key}`] = pathname.endsWith(`/${value}.html`);
	// 	return m;
	// }, {} as UiTypeCheck);
	const typeUnDefined =
		!route ||
		!route.query ||
		!route.query.type ||
		!(route.query.type as string);
	if (typeUnDefined) {
		return {
			isTab: true,
			isNotification: false,
			isPop: false
		};
	}

	return Object.entries(UI_TYPE).reduce((m: any, [key, value]) => {
		m[`is${key}`] = route.query.type == value;
		return m;
	}, {} as UiTypeCheck);
};

export const getUITypeName = (): string => {
	const UIType = getUiType();
	if (UIType.isPop) return 'popup';
	if (UIType.isNotification) return 'notification';
	if (UIType.isTab) return 'tab';
	return 'popup';
};

export function isIPV4OrIPv6Address(url: string): boolean {
	if (url.toLowerCase() === 'localhost') {
		return true;
	}
	const ipv4Pattern =
		/^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

	const ipv6Pattern = /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;

	return ipv4Pattern.test(url) || ipv6Pattern.test(url);
}
export const stopScrollMove = (m?: (e: any) => void) => {
	document.body.style.overflow = 'hidden';
	if (m) {
		document.addEventListener('touchmove', m, { passive: false });
	}
};

export const startScrollMove = (m?: (e: any) => void) => {
	document.body.style.overflow = '';
	if (m) {
		document.removeEventListener('touchmove', m);
	}
};

export const utcToDate = (utc_datetime: string) => {
	const T_pos = utc_datetime.indexOf('T');
	const Z_pos = utc_datetime.indexOf('Z');
	const year_month_day = utc_datetime.substr(0, T_pos);
	const hour_minute_second = utc_datetime.substr(T_pos + 1, Z_pos - T_pos - 1);
	const new_datetime = year_month_day + ' ' + hour_minute_second;
	return new Date(Date.parse(new_datetime));
};

export const showItemIcon = (name: string) => {
	switch (name) {
		case 'vault':
			return 'sym_r_language';

		case 'web':
			return 'sym_r_language';

		case 'computer':
			return 'sym_r_computer';

		case 'creditCard':
			return 'sym_r_credit_card';

		case 'bank':
			return 'sym_r_account_balance';

		case 'wifi':
			return 'sym_r_wifi_password';

		case 'passport':
			return 'sym_r_assignment_ind';

		case 'authenticator':
			return 'sym_r_password';

		case 'document':
			return 'sym_r_list_alt';

		case 'custom':
			return 'sym_r_chrome_reader_mode';

		default:
			return 'sym_r_language';
	}
};

export const formatMinutesTime = (minutes: number) => {
	const { t } = useI18n();

	if (minutes < 60) {
		return `${minutes}` + t('time.minutes_short');
	}

	if (minutes < 60 * 24) {
		const hours = Math.floor(minutes / 60);
		const min = minutes % 60;
		return (
			`${hours}` +
			t('time.hour_short') +
			' ' +
			`${min}` +
			t('time.minutes_short')
		);
	}

	const days = Math.floor(minutes / (60 * 24));
	const hours = Math.floor((minutes - days * (60 * 24)) / 60);
	const min = minutes - days * (60 * 24) - hours * 60;

	return (
		`${days}` +
		t('time.days_short') +
		' ' +
		`${hours}` +
		t('time.hour_short') +
		' ' +
		`${min}` +
		t('time.minutes_short')
	);
};

export const getParams = (url: string, params: string) => {
	const res = new RegExp('(?:&|/?)' + params + '=([^&$]+)').exec(url);
	return res ? res[1] : '';
};

export const detectType = (mimetype: string) => {
	if (mimetype.startsWith('video')) return 'video';
	if (mimetype.startsWith('audio')) return 'audio';
	if (mimetype.startsWith('image')) return 'image';
	if (mimetype.startsWith('pdf')) return 'pdf';
	if (mimetype.startsWith('text')) return 'text';
	return 'blob';
};

// Determine if two object arrays contain the same value
export function containsSameValue<T>(
	arr1: T[],
	arr2: string[],
	key: string
): boolean {
	return arr1
		? arr1.some(
				(item1) => arr2.find((item2) => item1[key] === item2) !== undefined
		  )
		: false;
}

export const getextension = (name: string) => {
	return name.indexOf('.') > -1 ? name.substring(name.lastIndexOf('.')) : '';
};

//https://github.com/bhowell2/binary-insert-js
export type Comparator<T> = (a: T, b: T) => number;

/**
 * Takes in a __SORTED__ array and inserts the provided value into
 * the correct, sorted, position.
 * @param array the sorted array where the provided value needs to be inserted (in order)
 * @param insertValue value to be added to the array
 * @param comparator function that helps determine where to insert the value (
 */
export function binaryInsert<T>(
	array: T[],
	insertValue: T,
	comparator: Comparator<T>
) {
	/*
	 * These two conditional statements are not required, but will avoid the
	 * while loop below, potentially speeding up the insert by a decent amount.
	 * */
	if (array.length === 0 || comparator(array[0], insertValue) >= 0) {
		array.splice(0, 0, insertValue);
		return array;
	} else if (
		array.length > 0 &&
		comparator(array[array.length - 1], insertValue) <= 0
	) {
		array.splice(array.length, 0, insertValue);
		return array;
	}
	let left = 0,
		right = array.length;
	let leftLast = 0,
		rightLast = right;
	while (left < right) {
		const inPos = Math.floor((right + left) / 2);
		const compared = comparator(array[inPos], insertValue);
		if (compared < 0) {
			left = inPos;
		} else if (compared > 0) {
			right = inPos;
		} else {
			right = inPos;
			left = inPos;
		}
		// nothing has changed, must have found limits. insert between.
		if (leftLast === left && rightLast === right) {
			break;
		}
		leftLast = left;
		rightLast = right;
	}
	// use right, because Math.floor is used
	array.splice(right, 0, insertValue);
	return array;
}

export enum DefaultType {
	Limit = 20
}

const hp = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
export function generatePasword() {
	let password = '';
	for (let i = 0; i < 16; ++i) {
		const index = Math.floor(Math.random() * hp.length);
		password = password + hp[index];
	}
	return password;
}

export function isValidDate(value: string | number): boolean {
	const date = new Date(value);
	const timestamp = date.getTime();

	return !isNaN(timestamp) && value === date.toISOString().split('T')[0];
}

export const utcToStamp = (utc_datetime: string) => {
	const T_pos = utc_datetime.indexOf('T');
	const Z_pos = utc_datetime.indexOf('Z');
	const year_month_day = utc_datetime.substr(0, T_pos);
	const hour_minute_second = utc_datetime.substr(T_pos + 1, Z_pos - T_pos - 1);
	const new_datetime = year_month_day + ' ' + hour_minute_second;
	return new Date(Date.parse(new_datetime));
};

export const getSliceArray = (list: any, size: number) => {
	if (list.length > size) {
		return list.slice(0, size);
	} else {
		return list;
	}
};

export function humanStorageSize(bytes: any) {
	const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
	let u = 0;

	while (parseInt(bytes, 10) >= 1024 && u < units.length - 1) {
		bytes /= 1024;
		++u;
	}

	return { value: bytes.toFixed(1), unit: units[u] };
}

export function formatTimeDifference(utcTimeString: string): string {
	const utcDate = new Date(utcTimeString);
	const currentDate = new Date();

	const timeDifference = currentDate.getTime() - utcDate.getTime();
	const secondsDifference = Math.floor(timeDifference / 1000);
	const minutesDifference = Math.floor(secondsDifference / 60);
	const hoursDifference = Math.floor(minutesDifference / 60);
	const daysDifference = Math.floor(hoursDifference / 24);
	const weeksDifference = Math.floor(daysDifference / 7);
	const monthsDifference = Math.floor(daysDifference / 30);
	const yearsDifference = Math.floor(daysDifference / 365);

	if (daysDifference === 0) {
		return 'Today';
	} else if (daysDifference === 1) {
		return 'Yesterday';
	} else if (daysDifference < 7) {
		return `${daysDifference} days ago`;
	} else if (weeksDifference === 1) {
		return '1 week ago';
	} else if (weeksDifference < 4) {
		return `${weeksDifference} weeks ago`;
	} else if (monthsDifference === 1) {
		return '1 month ago';
	} else if (monthsDifference < 12) {
		return `${monthsDifference} months ago`;
	} else if (yearsDifference === 1) {
		return '1 year ago';
	} else if (yearsDifference <= 2) {
		return '2 years ago';
	} else {
		return utcDate.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
}

export function capitalizeFirstLetter(str: string): string {
	if (!str) {
		return '';
	}
	if (str.length > 0) {
		return str.charAt(0).toUpperCase() + str.slice(1);
	}
	return str;
}

export function decodeUnicode(str) {
	return str.replace(/\\u[\dA-F]{4}/gi, function (match) {
		return String.fromCharCode(parseInt(match.replace(/\\u/g, ''), 16));
	});
}

export function intersection<T>(array1: T[], array2: T[]): T[] {
	return array1.filter((value) => array2.includes(value));
}

const languageMap: LanguageMap = {
	af: 'Afrikaans',
	'af-ZA': 'Afrikaans (South Africa)',
	ar: 'Arabic',
	'ar-AE': 'Arabic (U.A.E.)',
	'ar-BH': 'Arabic (Bahrain)',
	'ar-DZ': 'Arabic (Algeria)',
	'ar-EG': 'Arabic (Egypt)',
	'ar-IQ': 'Arabic (Iraq)',
	'ar-JO': 'Arabic (Jordan)',
	'ar-KW': 'Arabic (Kuwait)',
	'ar-LB': 'Arabic (Lebanon)',
	'ar-LY': 'Arabic (Libya)',
	'ar-MA': 'Arabic (Morocco)',
	'ar-OM': 'Arabic (Oman)',
	'ar-QA': 'Arabic (Qatar)',
	'ar-SA': 'Arabic (Saudi Arabia)',
	'ar-SY': 'Arabic (Syria)',
	'ar-TN': 'Arabic (Tunisia)',
	'ar-YE': 'Arabic (Yemen)',
	az: 'Azeri (Latin)',
	'az-AZ': 'Azeri (Latin) (Azerbaijan)',
	// 'az-AZ': 'Azeri (Cyrillic) (Azerbaijan)',
	be: 'Belarusian',
	'be-BY': 'Belarusian (Belarus)',
	bg: 'Bulgarian',
	'bg-BG': 'Bulgarian (Bulgaria)',
	'bs-BA': 'Bosnian (Bosnia and Herzegovina)',
	ca: 'Catalan',
	'ca-ES': 'Catalan (Spain)',
	cs: 'Czech',
	'cs-CZ': 'Czech (Czech Republic)',
	cy: 'Welsh',
	'cy-GB': 'Welsh (United Kingdom)',
	da: 'Danish',
	'da-DK': 'Danish (Denmark)',
	de: 'German',
	'de-AT': 'German (Austria)',
	'de-CH': 'German (Switzerland)',
	'de-DE': 'German (Germany)',
	'de-LI': 'German (Liechtenstein)',
	'de-LU': 'German (Luxembourg)',
	dv: 'Divehi',
	'dv-MV': 'Divehi (Maldives)',
	el: 'Greek',
	'el-GR': 'Greek (Greece)',
	en: 'English',
	'en-AU': 'English (Australia)',
	'en-BZ': 'English (Belize)',
	'en-CA': 'English (Canada)',
	'en-CB': 'English (Caribbean)',
	'en-GB': 'English (United Kingdom)',
	'en-IE': 'English (Ireland)',
	'en-JM': 'English (Jamaica)',
	'en-NZ': 'English (New Zealand)',
	'en-PH': 'English (Republic of the Philippines)',
	'en-TT': 'English (Trinidad and Tobago)',
	'en-US': 'English (United States)',
	'en-ZA': 'English (South Africa)',
	'en-ZW': 'English (Zimbabwe)',
	eo: 'Esperanto',
	es: 'Spanish',
	'es-AR': 'Spanish (Argentina)',
	'es-BO': 'Spanish (Bolivia)',
	'es-CL': 'Spanish (Chile)',
	'es-CO': 'Spanish (Colombia)',
	'es-CR': 'Spanish (Costa Rica)',
	'es-DO': 'Spanish (Dominican Republic)',
	'es-EC': 'Spanish (Ecuador)',
	'es-ES': 'Spanish (Castilian)',
	// 'es-ES': 'Spanish (Spain)',
	'es-GT': 'Spanish (Guatemala)',
	'es-HN': 'Spanish (Honduras)',
	'es-MX': 'Spanish (Mexico)',
	'es-NI': 'Spanish (Nicaragua)',
	'es-PA': 'Spanish (Panama)',
	'es-PE': 'Spanish (Peru)',
	'es-PR': 'Spanish (Puerto Rico)',
	'es-PY': 'Spanish (Paraguay)',
	'es-SV': 'Spanish (El Salvador)',
	'es-UY': 'Spanish (Uruguay)',
	'es-VE': 'Spanish (Venezuela)',
	et: 'Estonian',
	'et-EE': 'Estonian (Estonia)',
	eu: 'Basque',
	'eu-ES': 'Basque (Spain)',
	fa: 'Farsi',
	'fa-IR': 'Farsi (Iran)',
	fi: 'Finnish',
	'fi-FI': 'Finnish (Finland)',
	fo: 'Faroese',
	'fo-FO': 'Faroese (Faroe Islands)',
	fr: 'French',
	'fr-BE': 'French (Belgium)',
	'fr-CA': 'French (Canada)',
	'fr-CH': 'French (Switzerland)',
	'fr-FR': 'French (France)',
	'fr-LU': 'French (Luxembourg)',
	'fr-MC': 'French (Principality of Monaco)',
	gl: 'Galician',
	'gl-ES': 'Galician (Spain)',
	gu: 'Gujarati',
	'gu-IN': 'Gujarati (India)',
	he: 'Hebrew',
	'he-IL': 'Hebrew (Israel)',
	hi: 'Hindi',
	'hi-IN': 'Hindi (India)',
	hr: 'Croatian',
	'hr-BA': 'Croatian (Bosnia and Herzegovina)',
	'hr-HR': 'Croatian (Croatia)',
	hu: 'Hungarian',
	'hu-HU': 'Hungarian (Hungary)',
	hy: 'Armenian',
	'hy-AM': 'Armenian (Armenia)',
	id: 'Indonesian',
	'id-ID': 'Indonesian (Indonesia)',
	is: 'Icelandic',
	'is-IS': 'Icelandic (Iceland)',
	it: 'Italian',
	'it-CH': 'Italian (Switzerland)',
	'it-IT': 'Italian (Italy)',
	ja: 'Japanese',
	'ja-JP': 'Japanese (Japan)',
	ka: 'Georgian',
	'ka-GE': 'Georgian (Georgia)',
	kk: 'Kazakh',
	'kk-KZ': 'Kazakh (Kazakhstan)',
	kn: 'Kannada',
	'kn-IN': 'Kannada (India)',
	ko: 'Korean',
	'ko-KR': 'Korean (Korea)',
	kok: 'Konkani',
	'kok-IN': 'Konkani (India)',
	ky: 'Kyrgyz',
	'ky-KG': 'Kyrgyz (Kyrgyzstan)',
	lt: 'Lithuanian',
	'lt-LT': 'Lithuanian (Lithuania)',
	lv: 'Latvian',
	'lv-LV': 'Latvian (Latvia)',
	mi: 'Maori',
	'mi-NZ': 'Maori (New Zealand)',
	mk: 'FYRO Macedonian',
	'mk-MK': 'FYRO Macedonian (Former Yugoslav Republic of Macedonia)',
	mn: 'Mongolian',
	'mn-MN': 'Mongolian (Mongolia)',
	mr: 'Marathi',
	'mr-IN': 'Marathi (India)',
	ms: 'Malay',
	'ms-BN': 'Malay (Brunei Darussalam)',
	'ms-MY': 'Malay (Malaysia)',
	mt: 'Maltese',
	'mt-MT': 'Maltese (Malta)',
	nb: 'Norwegian (Bokmål)',
	'nb-NO': 'Norwegian (Bokmål) (Norway)',
	nl: 'Dutch',
	'nl-BE': 'Dutch (Belgium)',
	'nl-NL': 'Dutch (Netherlands)',
	'nn-NO': 'Norwegian (Nynorsk) (Norway)',
	ns: 'Northern Sotho',
	'ns-ZA': 'Northern Sotho (South Africa)',
	pa: 'Punjabi',
	'pa-IN': 'Punjabi (India)',
	pl: 'Polish',
	'pl-PL': 'Polish (Poland)',
	ps: 'Pashto',
	'ps-AR': 'Pashto (Afghanistan)',
	pt: 'Portuguese',
	'pt-BR': 'Portuguese (Brazil)',
	'pt-PT': 'Portuguese (Portugal)',
	qu: 'Quechua',
	'qu-BO': 'Quechua (Bolivia)',
	'qu-EC': 'Quechua (Ecuador)',
	'qu-PE': 'Quechua (Peru)',
	ro: 'Romanian',
	'ro-RO': 'Romanian (Romania)',
	ru: 'Russian',
	'ru-RU': 'Russian (Russia)',
	sa: 'Sanskrit',
	'sa-IN': 'Sanskrit (India)',
	se: 'Sami (Northern)',
	'se-FI': 'Sami (Northern) (Finland)',
	// 'se-FI': 'Sami (Skolt) (Finland)',
	// 'se-FI': 'Sami (Inari) (Finland)',
	'se-NO': 'Sami (Northern) (Norway)',
	// 'se-NO': 'Sami (Lule) (Norway)',
	// 'se-NO': 'Sami (Southern) (Norway)',
	'se-SE': 'Sami (Northern) (Sweden)',
	// 'se-SE': 'Sami (Lule) (Sweden)',
	// 'se-SE': 'Sami (Southern) (Sweden)',
	sk: 'Slovak',
	'sk-SK': 'Slovak (Slovakia)',
	sl: 'Slovenian',
	'sl-SI': 'Slovenian (Slovenia)',
	sq: 'Albanian',
	'sq-AL': 'Albanian (Albania)',
	'sr-BA': 'Serbian (Latin) (Bosnia and Herzegovina)',
	// 'sr-BA': 'Serbian (Cyrillic) (Bosnia and Herzegovina)',
	'sr-SP': 'Serbian (Latin) (Serbia and Montenegro)',
	// 'sr-SP': 'Serbian (Cyrillic) (Serbia and Montenegro)',
	sv: 'Swedish',
	'sv-FI': 'Swedish (Finland)',
	'sv-SE': 'Swedish (Sweden)',
	sw: 'Swahili',
	'sw-KE': 'Swahili (Kenya)',
	syr: 'Syriac',
	'syr-SY': 'Syriac (Syria)',
	ta: 'Tamil',
	'ta-IN': 'Tamil (India)',
	te: 'Telugu',
	'te-IN': 'Telugu (India)',
	th: 'Thai',
	'th-TH': 'Thai (Thailand)',
	tl: 'Tagalog',
	'tl-PH': 'Tagalog (Philippines)',
	tn: 'Tswana',
	'tn-ZA': 'Tswana (South Africa)',
	tr: 'Turkish',
	'tr-TR': 'Turkish (Turkey)',
	tt: 'Tatar',
	'tt-RU': 'Tatar (Russia)',
	ts: 'Tsonga',
	uk: 'Ukrainian',
	'uk-UA': 'Ukrainian (Ukraine)',
	ur: 'Urdu',
	'ur-PK': 'Urdu (Islamic Republic of Pakistan)',
	uz: 'Uzbek (Latin)',
	'uz-UZ': 'Uzbek (Latin) (Uzbekistan)',
	// 'uz-UZ': 'Uzbek (Cyrillic) (Uzbekistan)',
	vi: 'Vietnamese',
	'vi-VN': 'Vietnamese (Viet Nam)',
	xh: 'Xhosa',
	'xh-ZA': 'Xhosa (South Africa)',
	zh: 'Chinese',
	'zh-CN': 'Chinese (S)',
	'zh-HK': 'Chinese (Hong Kong)',
	'zh-MO': 'Chinese (Macau)',
	'zh-SG': 'Chinese (Singapore)',
	'zh-TW': 'Chinese (T)',
	zu: 'Zulu',
	'zu-ZA': 'Zulu (South Africa)'
};

interface LanguageMap {
	[key: string]: string;
}

export function convertLanguageCodeToName(code: string) {
	return languageMap[code] || '';
}

export function convertLanguageCodesToNames(codes: string[]): string[] {
	if (!codes || codes.length === 0) {
		return [];
	}

	return codes.map((code) => languageMap[code] || 'code');
}

export const delay = async (ms: number) => {
	return new Promise((resolve) => {
		const timer = setTimeout(() => {
			resolve(undefined);
			clearTimeout(timer);
		}, ms);
	});
};
