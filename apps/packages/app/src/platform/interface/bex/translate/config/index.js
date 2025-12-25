import { browser } from 'webextension-polyfill-ts';
import {
	REACT_APP_RULESURL,
	REACT_APP_RULESURL_ON,
	REACT_APP_RULESURL_OFF
} from '../env';
import {
	DEFAULT_SELECTOR,
	DEFAULT_KEEP_SELECTOR,
	GLOBAL_KEY,
	REMAIN_KEY,
	SHADOW_KEY,
	DEFAULT_RULE,
	DEFAULT_OW_RULE,
	BUILTIN_RULES
} from './rules';
import { APP_NAME, APP_LCNAME } from './app';
export {
	GLOBAL_KEY,
	REMAIN_KEY,
	SHADOW_KEY,
	DEFAULT_RULE,
	DEFAULT_OW_RULE,
	BUILTIN_RULES,
	APP_LCNAME
};

export const STOKEY_MSAUTH = `${APP_NAME}_msauth`;
export const STOKEY_BDAUTH = `${APP_NAME}_bdauth`;
export const STOKEY_SETTING = `${APP_NAME}_setting`;
export const STOKEY_RULES = `${APP_NAME}_rules`;
export const STOKEY_WORDS = `${APP_NAME}_words`;
export const STOKEY_SYNC = `${APP_NAME}_sync`;
export const STOKEY_FAB = `${APP_NAME}_fab`;
export const STOKEY_RULESCACHE_PREFIX = `${APP_NAME}_rulescache_`;

export const CMD_TOGGLE_TRANSLATE = 'toggleTranslate';
export const CMD_TOGGLE_STYLE = 'toggleStyle';
export const CMD_OPEN_OPTIONS = 'openOptions';
export const CMD_OPEN_TRANBOX = 'openTranbox';

export const CLIENT_WEB = 'web';
export const CLIENT_CHROME = 'chrome';
export const CLIENT_EDGE = 'edge';
export const CLIENT_FIREFOX = 'firefox';
export const CLIENT_USERSCRIPT = 'userscript';
export const CLIENT_EXTS = [CLIENT_CHROME, CLIENT_EDGE, CLIENT_FIREFOX];

export const KV_RULES_KEY = 'olares-rules.json';
export const KV_WORDS_KEY = 'kiss-words.json';
export const KV_RULES_SHARE_KEY = 'kiss-rules-share.json';
export const KV_SETTING_KEY = 'kiss-setting.json';
export const KV_SALT_SYNC = 'KISS-Translator-SYNC';
export const KV_SALT_SHARE = 'KISS-Translator-SHARE';

export const CACHE_NAME = `${APP_NAME}_cache`;

export const MSG_FETCH = 'fetch';
export const MSG_GET_HTTPCACHE = 'get_httpcache';
export const MSG_OPEN_OPTIONS = 'open_options';
export const MSG_SAVE_RULE = 'save_rule';
export const MSG_TRANS_TOGGLE = 'trans_toggle';
export const MSG_TRANS_TOGGLE_STYLE = 'trans_toggle_style';
export const MSG_OPEN_TRANBOX = 'open_tranbox';
export const MSG_TRANS_GETRULE = 'trans_getrule';
export const MSG_TRANS_PUTRULE = 'trans_putrule';
export const MSG_TRANS_CURRULE = 'trans_currule';
export const MSG_CONTEXT_MENUS = 'context_menus';
export const MSG_COMMAND_SHORTCUTS = 'command_shortcuts';
export const MSG_INJECT_JS = 'inject_js';
export const MSG_INJECT_CSS = 'inject_css';
export const MSG_UPDATE_CSP = 'update_csp';

export const THEME_LIGHT = 'light';
export const THEME_DARK = 'dark';

export const URL_CACHE_TRAN = `https://${APP_LCNAME}/translate`;

// api.cognitive.microsofttranslator.com
export const URL_MICROSOFT_TRAN =
	'https://api-edge.cognitive.microsofttranslator.com/translate';
export let URL_OLARES_TRAN = '';

await browser.storage.local.get(['userId', 'users']).then((result) => {
	const currentUser = result.users?.items?.items?.find(
		(item) => item.id === result.userId
	);
	if (currentUser) {
		const domain = currentUser.name.replace('@', '.');
		URL_OLARES_TRAN = `https://e95da2ac.${domain}/imme`;
	}
});
export const URL_MICROSOFT_AUTH = 'https://edge.microsoft.com/translate/auth';
export const URL_MICROSOFT_LANGDETECT =
	'https://api-edge.cognitive.microsofttranslator.com/detect?api-version=3.0';

export const URL_GOOGLE_TRAN =
	'https://translate.googleapis.com/translate_a/single';
export const URL_BAIDU_LANGDETECT = 'https://fanyi.baidu.com/langdetect';
export const URL_BAIDU_SUGGEST = 'https://fanyi.baidu.com/sug';
export const URL_BAIDU_TTS = 'https://fanyi.baidu.com/gettts';
export const URL_BAIDU_WEB = 'https://fanyi.baidu.com/';
export const URL_BAIDU_TRANSAPI = 'https://fanyi.baidu.com/transapi';
export const URL_BAIDU_TRANSAPI_V2 = 'https://fanyi.baidu.com/v2transapi';
export const URL_DEEPLFREE_TRAN = 'https://www2.deepl.com/jsonrpc';
export const URL_TENCENT_TRANSMART = 'https://transmart.qq.com/api/imt';
export const URL_NIUTRANS_REG =
	'https://niutrans.com/login?active=3&userSource=kiss-translator';

export const DEFAULT_USER_AGENT =
	'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36';

export const OPT_DICT_BAIDU = 'Baidu';

export const OPT_TRANS_GOOGLE = 'Google';
export const OPT_TRANS_MICROSOFT = 'Microsoft';
export const OPT_TRANS_OLARES = 'Olares';
export const OPT_TRANS_DEEPL = 'DeepL';
export const OPT_TRANS_DEEPLX = 'DeepLX';
export const OPT_TRANS_DEEPLFREE = 'DeepLFree';
export const OPT_TRANS_NIUTRANS = 'NiuTrans';
export const OPT_TRANS_BAIDU = 'Baidu';
export const OPT_TRANS_TENCENT = 'Tencent';
export const OPT_TRANS_OPENAI = 'OpenAI';
export const OPT_TRANS_OPENAI_2 = 'OpenAI2';
export const OPT_TRANS_OPENAI_3 = 'OpenAI3';
export const OPT_TRANS_GEMINI = 'Gemini';
export const OPT_TRANS_CLOUDFLAREAI = 'CloudflareAI';
export const OPT_TRANS_OLLAMA = 'Ollama';
export const OPT_TRANS_OLLAMA_2 = 'Ollama2';
export const OPT_TRANS_OLLAMA_3 = 'Ollama3';
export const OPT_TRANS_CUSTOMIZE = 'Custom';
export const OPT_TRANS_CUSTOMIZE_2 = 'Custom2';
export const OPT_TRANS_CUSTOMIZE_3 = 'Custom3';
export const OPT_TRANS_CUSTOMIZE_4 = 'Custom4';
export const OPT_TRANS_CUSTOMIZE_5 = 'Custom5';
export const OPT_TRANS_ALL = [
	OPT_TRANS_GOOGLE,
	OPT_TRANS_MICROSOFT,
	OPT_TRANS_OLARES
	// OPT_TRANS_BAIDU,
	// OPT_TRANS_TENCENT
	// OPT_TRANS_DEEPL,
	// OPT_TRANS_DEEPLFREE,
	// OPT_TRANS_DEEPLX,
	// OPT_TRANS_NIUTRANS,
	// OPT_TRANS_OPENAI,
	// OPT_TRANS_OPENAI_2,
	// OPT_TRANS_OPENAI_3,
	// OPT_TRANS_GEMINI,
	// OPT_TRANS_CLOUDFLAREAI,
	// OPT_TRANS_OLLAMA,
	// OPT_TRANS_OLLAMA_2,
	// OPT_TRANS_OLLAMA_3,
	// OPT_TRANS_CUSTOMIZE,
	// OPT_TRANS_CUSTOMIZE_2,
	// OPT_TRANS_CUSTOMIZE_3,
	// OPT_TRANS_CUSTOMIZE_4,
	// OPT_TRANS_CUSTOMIZE_5
];

export const OPT_LANGDETECTOR_ALL = [
	OPT_TRANS_GOOGLE,
	OPT_TRANS_MICROSOFT,
	OPT_TRANS_BAIDU,
	OPT_TRANS_TENCENT
];

export const OPT_LANGS_TO = [
	['en', 'English - English'],
	['zh-CN', 'Simplified Chinese - 简体中文'],
	['zh-TW', 'Traditional Chinese - 繁體中文'],
	['ar', 'Arabic - العربية'],
	['bg', 'Bulgarian - Български'],
	['ca', 'Catalan - Català'],
	['hr', 'Croatian - Hrvatski'],
	['cs', 'Czech - Čeština'],
	['da', 'Danish - Dansk'],
	['nl', 'Dutch - Nederlands'],
	['fi', 'Finnish - Suomi'],
	['fr', 'French - Français'],
	['de', 'German - Deutsch'],
	['el', 'Greek - Ελληνικά'],
	['hi', 'Hindi - हिन्दी'],
	['hu', 'Hungarian - Magyar'],
	['id', 'Indonesian - Indonesia'],
	['it', 'Italian - Italiano'],
	['ja', 'Japanese - 日本語'],
	['ko', 'Korean - 한국어'],
	['ms', 'Malay - Melayu'],
	['mt', 'Maltese - Malti'],
	['nb', 'Norwegian - Norsk Bokmål'],
	['pl', 'Polish - Polski'],
	['pt', 'Portuguese - Português'],
	['ro', 'Romanian - Română'],
	['ru', 'Russian - Русский'],
	['sk', 'Slovak - Slovenčina'],
	['sl', 'Slovenian - Slovenščina'],
	['es', 'Spanish - Español'],
	['sv', 'Swedish - Svenska'],
	['ta', 'Tamil - தமிழ்'],
	['te', 'Telugu - తెలుగు'],
	['th', 'Thai - ไทย'],
	['tr', 'Turkish - Türkçe'],
	['uk', 'Ukrainian - Українська'],
	['vi', 'Vietnamese - Tiếng Việt']
];
export const OPT_LANGS_FROM = [['auto', 'Auto-detect'], ...OPT_LANGS_TO];
export const OPT_LANGS_SPECIAL = {
	[OPT_TRANS_GOOGLE]: new Map(OPT_LANGS_FROM.map(([key]) => [key, key])),
	[OPT_TRANS_MICROSOFT]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', ''],
		['zh-CN', 'zh-Hans'],
		['zh-TW', 'zh-Hant']
	]),
	[OPT_TRANS_OLARES]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', 'auto'],
		['zh-CN', 'zh'],
		['zh-TW', 'zh']
	]),
	[OPT_TRANS_DEEPL]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key.toUpperCase()]),
		['auto', ''],
		['zh-CN', 'ZH'],
		['zh-TW', 'ZH']
	]),
	[OPT_TRANS_DEEPLFREE]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key.toUpperCase()]),
		['auto', 'auto'],
		['zh-CN', 'ZH'],
		['zh-TW', 'ZH']
	]),
	[OPT_TRANS_DEEPLX]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key.toUpperCase()]),
		['auto', 'auto'],
		['zh-CN', 'ZH'],
		['zh-TW', 'ZH']
	]),
	[OPT_TRANS_NIUTRANS]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', 'auto'],
		['zh-CN', 'zh'],
		['zh-TW', 'cht']
	]),
	[OPT_TRANS_BAIDU]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['zh-CN', 'zh'],
		['zh-TW', 'cht'],
		['ar', 'ara'],
		['bg', 'bul'],
		['ca', 'cat'],
		['hr', 'hrv'],
		['da', 'dan'],
		['fi', 'fin'],
		['fr', 'fra'],
		['hi', 'mai'],
		['ja', 'jp'],
		['ko', 'kor'],
		['ms', 'may'],
		['mt', 'mlt'],
		['nb', 'nor'],
		['ro', 'rom'],
		['ru', 'ru'],
		['sl', 'slo'],
		['es', 'spa'],
		['sv', 'swe'],
		['ta', 'tam'],
		['te', 'tel'],
		['uk', 'ukr'],
		['vi', 'vie']
	]),
	[OPT_TRANS_TENCENT]: new Map([
		['auto', 'auto'],
		['zh-CN', 'zh'],
		['zh-TW', 'zh'],
		['en', 'en'],
		['ar', 'ar'],
		['de', 'de'],
		['ru', 'ru'],
		['fr', 'fr'],
		['fi', 'fil'],
		['ko', 'ko'],
		['ms', 'ms'],
		['pt', 'pt'],
		['ja', 'ja'],
		['th', 'th'],
		['tr', 'tr'],
		['es', 'es'],
		['it', 'it'],
		['hi', 'hi'],
		['id', 'id'],
		['vi', 'vi']
	]),
	[OPT_TRANS_OPENAI]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_OPENAI_2]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_OPENAI_3]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_GEMINI]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_OLLAMA]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_OLLAMA_2]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_OLLAMA_3]: new Map(
		OPT_LANGS_FROM.map(([key, val]) => [key, val.split(' - ')[0]])
	),
	[OPT_TRANS_CLOUDFLAREAI]: new Map([
		['auto', ''],
		['zh-CN', 'chinese'],
		['zh-TW', 'chinese'],
		['en', 'english'],
		['ar', 'arabic'],
		['de', 'german'],
		['ru', 'russian'],
		['fr', 'french'],
		['pt', 'portuguese'],
		['ja', 'japanese'],
		['es', 'spanish'],
		['hi', 'hindi']
	]),
	[OPT_TRANS_CUSTOMIZE]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', '']
	]),
	[OPT_TRANS_CUSTOMIZE_2]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', '']
	]),
	[OPT_TRANS_CUSTOMIZE_3]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', '']
	]),
	[OPT_TRANS_CUSTOMIZE_4]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', '']
	]),
	[OPT_TRANS_CUSTOMIZE_5]: new Map([
		...OPT_LANGS_FROM.map(([key]) => [key, key]),
		['auto', '']
	])
};
export const OPT_LANGS_LIST = OPT_LANGS_TO.map(([lang]) => lang);
export const OPT_LANGS_MICROSOFT = new Map(
	Array.from(OPT_LANGS_SPECIAL[OPT_TRANS_MICROSOFT].entries()).map(([k, v]) => [
		v,
		k
	])
);
export const OPT_LANGS_OLARES = new Map(
	Array.from(OPT_LANGS_SPECIAL[OPT_TRANS_OLARES].entries()).map(([k, v]) => [
		v,
		k
	])
);
export const OPT_LANGS_BAIDU = new Map(
	Array.from(OPT_LANGS_SPECIAL[OPT_TRANS_BAIDU].entries()).map(([k, v]) => [
		v,
		k
	])
);
export const OPT_LANGS_TENCENT = new Map(
	Array.from(OPT_LANGS_SPECIAL[OPT_TRANS_TENCENT].entries()).map(([k, v]) => [
		v,
		k
	])
);
OPT_LANGS_TENCENT.set('zh', 'zh-CN');

export const OPT_STYLE_NONE = 'style_none'; // None
export const OPT_STYLE_LINE = 'under_line'; // Underline
export const OPT_STYLE_DOTLINE = 'dot_line'; // Dotted line
export const OPT_STYLE_DASHLINE = 'dash_line'; // Dashed line
export const OPT_STYLE_WAVYLINE = 'wavy_line'; // Wavy line
export const OPT_STYLE_FUZZY = 'fuzzy'; // Fuzzy
export const OPT_STYLE_HIGHLIGHT = 'highlight'; // Highlight
export const OPT_STYLE_BLOCKQUOTE = 'blockquote'; // Blockquote
export const OPT_STYLE_DIY = 'diy_style'; // Custom style
export const OPT_STYLE_ALL = [
	OPT_STYLE_NONE,
	OPT_STYLE_LINE,
	OPT_STYLE_DOTLINE,
	OPT_STYLE_DASHLINE,
	OPT_STYLE_WAVYLINE,
	OPT_STYLE_FUZZY,
	OPT_STYLE_HIGHLIGHT,
	OPT_STYLE_BLOCKQUOTE,
	OPT_STYLE_DIY
];
export const OPT_STYLE_USE_COLOR = [
	OPT_STYLE_LINE,
	OPT_STYLE_DOTLINE,
	OPT_STYLE_DASHLINE,
	OPT_STYLE_WAVYLINE,
	OPT_STYLE_HIGHLIGHT,
	OPT_STYLE_BLOCKQUOTE
];

export const OPT_TIMING_PAGESCROLL = 'mk_pagescroll'; // Scroll loading translation
export const OPT_TIMING_PAGEOPEN = 'mk_pageopen'; // Direct translation to the end
export const OPT_TIMING_MOUSEOVER = 'mk_mouseover';
export const OPT_TIMING_CONTROL = 'mk_ctrlKey';
export const OPT_TIMING_SHIFT = 'mk_shiftKey';
export const OPT_TIMING_ALT = 'mk_altKey';
export const OPT_TIMING_ALL = [
	OPT_TIMING_PAGESCROLL,
	OPT_TIMING_PAGEOPEN,
	OPT_TIMING_MOUSEOVER,
	OPT_TIMING_CONTROL,
	OPT_TIMING_SHIFT,
	OPT_TIMING_ALT
];

export const DEFAULT_FETCH_LIMIT = 10; // Default maximum number of tasks
export const DEFAULT_FETCH_INTERVAL = 100; // Default task interval time

export const INPUT_PLACE_URL = '{{url}}'; // Placeholder
export const INPUT_PLACE_FROM = '{{from}}'; // Placeholder
export const INPUT_PLACE_TO = '{{to}}'; // Placeholder
export const INPUT_PLACE_TEXT = '{{text}}'; // Placeholder
export const INPUT_PLACE_KEY = '{{key}}'; // Placeholder
export const INPUT_PLACE_MODEL = '{{model}}'; // Placeholder

export const DEFAULT_COLOR = '#209CEE'; // Default highlight background/line color

export const DEFAULT_TRANS_TAG = 'span';
export const DEFAULT_SELECT_STYLE =
	'-webkit-line-clamp: unset; max-height: none; height: auto;';

// Global rules
export const GLOBLA_RULE = {
	pattern: '*', // Match URL
	selector: DEFAULT_SELECTOR, // Selector
	keepSelector: DEFAULT_KEEP_SELECTOR, // Keep element selector
	terms: '', // Professional terms
	translator: OPT_TRANS_MICROSOFT, // Translation service
	fromLang: 'auto', // Source language
	toLang: 'zh-CN', // Target language
	textStyle: OPT_STYLE_DASHLINE, // Translation style
	transOpen: 'false', // Enable translation
	bgColor: '', // Translation color
	textDiyStyle: '', // Custom translation style
	selectStyle: DEFAULT_SELECT_STYLE, // Selector node style
	parentStyle: DEFAULT_SELECT_STYLE, // Selector parent node style
	injectJs: '', // Inject JS
	injectCss: '', // Inject CSS
	transOnly: 'false', // Show translation only
	transTiming: OPT_TIMING_PAGESCROLL, // Translation timing/Mouse hover translation
	transTag: DEFAULT_TRANS_TAG, // Translation element tag
	transTitle: 'false', // Translate page title
	detectRemote: 'false', // Use remote language detection
	skipLangs: [], // Languages not to translate
	fixerSelector: '', // Fixer selector
	fixerFunc: '-', // Fixer function
	transStartHook: '', // Hook function
	transEndHook: '', // Hook function
	transRemoveHook: '' // Hook function
};

// Input box translation
export const OPT_INPUT_TRANS_SIGNS = ['/', '//', '\\', '\\\\', '>', '>>'];
export const DEFAULT_INPUT_SHORTCUT = ['AltLeft', 'KeyI'];
export const DEFAULT_INPUT_RULE = {
	transOpen: false,
	translator: OPT_TRANS_MICROSOFT,
	fromLang: 'auto',
	toLang: 'en',
	triggerShortcut: DEFAULT_INPUT_SHORTCUT,
	triggerCount: 1,
	triggerTime: 200,
	transSign: OPT_INPUT_TRANS_SIGNS[0]
};

// Word selection translation
export const PHONIC_MAP = {
	en_phonic: ['英', 'uk'],
	us_phonic: ['美', 'en']
};
export const OPT_TRANBOX_TRIGGER_CLICK = 'click';
export const OPT_TRANBOX_TRIGGER_HOVER = 'hover';
export const OPT_TRANBOX_TRIGGER_SELECT = 'select';
export const OPT_TRANBOX_TRIGGER_ALL = [
	OPT_TRANBOX_TRIGGER_CLICK,
	OPT_TRANBOX_TRIGGER_HOVER,
	OPT_TRANBOX_TRIGGER_SELECT
];
export const DEFAULT_TRANBOX_SHORTCUT = ['AltLeft', 'KeyS'];
export const DEFAULT_TRANBOX_SETTING = {
	transOpen: false,
	translator: OPT_TRANS_MICROSOFT,
	fromLang: 'auto',
	toLang: 'zh-CN',
	toLang2: 'en',
	tranboxShortcut: DEFAULT_TRANBOX_SHORTCUT,
	btnOffsetX: 10,
	btnOffsetY: 10,
	boxOffsetX: 0,
	boxOffsetY: 10,
	hideTranBtn: false, // Hide translation button
	hideClickAway: false, // Close popup when clicking outside
	simpleStyle: false, // Simple interface
	followSelection: false, // Translation box follows selected text
	triggerMode: OPT_TRANBOX_TRIGGER_CLICK, // Trigger translation mode
	extStyles: '', // Additional styles
	enDict: OPT_DICT_BAIDU // English dictionary
};

// Subscription list
export const DEFAULT_SUBRULES_LIST = [
	{
		url: REACT_APP_RULESURL,
		selected: false
	},
	{
		url: REACT_APP_RULESURL_ON,
		selected: true
	},
	{
		url: REACT_APP_RULESURL_OFF,
		selected: false
	}
];

// Translation API
const defaultCustomApi = {
	url: '',
	key: '',
	customOption: '', // (deprecated)
	reqHook: '', // request hook function
	resHook: '', // response hook function
	fetchLimit: DEFAULT_FETCH_LIMIT,
	fetchInterval: DEFAULT_FETCH_INTERVAL
};
const defaultOpenaiApi = {
	url: 'https://api.openai.com/v1/chat/completions',
	key: '',
	model: 'gpt-4',
	prompt: `You will be provided with a sentence in ${INPUT_PLACE_FROM}, and your task is to translate it into ${INPUT_PLACE_TO}.`,
	temperature: 0,
	maxTokens: 256,
	fetchLimit: 1,
	fetchInterval: 500
};
const defaultOllamaApi = {
	url: 'http://localhost:11434/api/generate',
	key: '',
	model: 'llama3',
	prompt: `Translate the following text from ${INPUT_PLACE_FROM} to ${INPUT_PLACE_TO}:\n\n${INPUT_PLACE_TEXT}`,
	fetchLimit: 1,
	fetchInterval: 500
};
export const DEFAULT_TRANS_APIS = {
	[OPT_TRANS_GOOGLE]: {
		url: URL_GOOGLE_TRAN,
		key: '',
		fetchLimit: DEFAULT_FETCH_LIMIT, // Maximum number of tasks
		fetchInterval: DEFAULT_FETCH_INTERVAL // Task interval time
	},
	[OPT_TRANS_MICROSOFT]: {
		fetchLimit: DEFAULT_FETCH_LIMIT,
		fetchInterval: DEFAULT_FETCH_INTERVAL
	},
	[OPT_TRANS_BAIDU]: {
		fetchLimit: DEFAULT_FETCH_LIMIT,
		fetchInterval: DEFAULT_FETCH_INTERVAL
	},
	[OPT_TRANS_TENCENT]: {
		fetchLimit: DEFAULT_FETCH_LIMIT,
		fetchInterval: DEFAULT_FETCH_INTERVAL
	},
	[OPT_TRANS_DEEPL]: {
		url: 'https://api-free.deepl.com/v2/translate',
		key: '',
		fetchLimit: 1,
		fetchInterval: 500
	},
	[OPT_TRANS_DEEPLFREE]: {
		fetchLimit: 1,
		fetchInterval: 500
	},
	[OPT_TRANS_DEEPLX]: {
		url: 'http://localhost:1188/translate',
		key: '',
		fetchLimit: 1,
		fetchInterval: 500
	},
	[OPT_TRANS_NIUTRANS]: {
		url: 'https://api.niutrans.com/NiuTransServer/translation',
		key: '',
		dictNo: '',
		memoryNo: '',
		fetchLimit: DEFAULT_FETCH_LIMIT,
		fetchInterval: DEFAULT_FETCH_INTERVAL
	},
	[OPT_TRANS_OPENAI]: defaultOpenaiApi,
	[OPT_TRANS_OPENAI_2]: defaultOpenaiApi,
	[OPT_TRANS_OPENAI_3]: defaultOpenaiApi,
	[OPT_TRANS_GEMINI]: {
		url: `https://generativelanguage.googleapis.com/v1/models/${INPUT_PLACE_MODEL}:generateContent?key=${INPUT_PLACE_KEY}`,
		key: '',
		model: 'gemini-pro',
		prompt: `Translate the following text from ${INPUT_PLACE_FROM} to ${INPUT_PLACE_TO}:\n\n${INPUT_PLACE_TEXT}`,
		fetchLimit: 1,
		fetchInterval: 500
	},
	[OPT_TRANS_CLOUDFLAREAI]: {
		url: 'https://api.cloudflare.com/client/v4/accounts/{{ACCOUNT_ID}}/ai/run/@cf/meta/m2m100-1.2b',
		key: '',
		fetchLimit: 1,
		fetchInterval: 500
	},
	[OPT_TRANS_OLLAMA]: defaultOllamaApi,
	[OPT_TRANS_OLLAMA_2]: defaultOllamaApi,
	[OPT_TRANS_OLLAMA_3]: defaultOllamaApi,
	[OPT_TRANS_CUSTOMIZE]: defaultCustomApi,
	[OPT_TRANS_CUSTOMIZE_2]: defaultCustomApi,
	[OPT_TRANS_CUSTOMIZE_3]: defaultCustomApi,
	[OPT_TRANS_CUSTOMIZE_4]: defaultCustomApi,
	[OPT_TRANS_CUSTOMIZE_5]: defaultCustomApi
};

// Default shortcuts
export const OPT_SHORTCUT_TRANSLATE = 'toggleTranslate';
export const OPT_SHORTCUT_STYLE = 'toggleStyle';
export const OPT_SHORTCUT_POPUP = 'togglePopup';
export const OPT_SHORTCUT_SETTING = 'openSetting';
export const DEFAULT_SHORTCUTS = {
	[OPT_SHORTCUT_TRANSLATE]: ['AltLeft', 'KeyQ'],
	[OPT_SHORTCUT_STYLE]: ['AltLeft', 'KeyC'],
	[OPT_SHORTCUT_POPUP]: ['AltLeft', 'KeyK'],
	[OPT_SHORTCUT_SETTING]: ['AltLeft', 'KeyO']
};

export const TRANS_MIN_LENGTH = 5; // Minimum translation length
export const TRANS_MAX_LENGTH = 5000; // Maximum translation length
export const TRANS_NEWLINE_LENGTH = 20; // Newline character count
export const DEFAULT_BLACKLIST = [
	'https://fishjar.github.io/kiss-translator/options.html',
	'https://translate.google.com',
	'https://www.deepl.com/translator',
	'oapi.dingtalk.com',
	'login.dingtalk.com'
]; // Disable translation list
export const DEFAULT_CSPLIST = ['https://github.com']; // Disable CSP list

export const DEFAULT_SETTING = {
	darkMode: false, // Dark mode
	uiLang: 'en', // UI language
	// fetchLimit: DEFAULT_FETCH_LIMIT, // Maximum number of tasks (moved to transApis, deprecated)
	// fetchInterval: DEFAULT_FETCH_INTERVAL, // Task interval time (moved to transApis, deprecated)
	minLength: TRANS_MIN_LENGTH,
	maxLength: TRANS_MAX_LENGTH,
	newlineLength: TRANS_NEWLINE_LENGTH,
	clearCache: false, // Clear cache on next browser startup
	injectRules: true, // Inject subscription rules
	// injectWebfix: true, // Inject fix patches (deprecated)
	// detectRemote: false, // Use remote language detection (moved to rule, deprecated)
	// contextMenus: true, // Add context menu (deprecated)
	contextMenuType: 1, // Context menu type (0: not show, 1: simple menu, 2: multi-level menu)
	// transTag: DEFAULT_TRANS_TAG, // Translation element tag (moved to rule, deprecated)
	// transOnly: false, // Show translation only (moved to rule, deprecated)
	// transTitle: false, // Translate page title (moved to rule, deprecated)
	subrulesList: DEFAULT_SUBRULES_LIST, // Subscription list
	owSubrule: DEFAULT_OW_RULE, // Overwrite subscription rules
	transApis: DEFAULT_TRANS_APIS, // Translation API
	// mouseKey: OPT_TIMING_PAGESCROLL, // Translation timing/Mouse hover translation (moved to rule, deprecated)
	shortcuts: DEFAULT_SHORTCUTS, // Shortcuts
	inputRule: DEFAULT_INPUT_RULE, // Input box settings
	tranboxSetting: DEFAULT_TRANBOX_SETTING, // Word selection translation settings
	touchTranslate: 2, // Touch translation
	blacklist: DEFAULT_BLACKLIST.join(',\n'), // Disable translation list
	csplist: DEFAULT_CSPLIST.join(',\n'), // Disable CSP list
	// disableLangs: [], // Languages not to translate (moved to rule, deprecated)
	transInterval: 500, // Translation interval time
	langDetector: OPT_TRANS_MICROSOFT // Remote language detection service
};

export const DEFAULT_RULES = [GLOBLA_RULE];

export const OPT_SYNCTYPE_WORKER = 'KISS-Worker';
export const OPT_SYNCTYPE_WEBDAV = 'WebDAV';
export const OPT_SYNCTYPE_ALL = [OPT_SYNCTYPE_WORKER, OPT_SYNCTYPE_WEBDAV];

export const DEFAULT_SYNC = {
	syncType: OPT_SYNCTYPE_WORKER, // Sync method
	syncUrl: '', // Data sync API
	syncUser: '', // Data sync username
	syncKey: '', // Data sync key
	syncMeta: {}, // Data update and sync info
	subRulesSyncAt: 0, // Subscription rules sync time
	dataCaches: {} // Cache sync time
};
