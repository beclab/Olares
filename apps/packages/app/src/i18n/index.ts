import enUS from './en-US';
import zhCN from './zh-CN';
import { SelectorProps } from '../constant';

export default {
	'en-US': enUS,
	'zh-CN': zhCN
};

export const defaultLanguage = 'en-US';

export const supportLanguages: SelectorProps[] = [
	{ value: 'en-US', label: 'English' },
	{ value: 'zh-CN', label: '简体中文' }
];

export const languagesShort = {
	en: 'en-US',
	zh: 'zh-CN'
};

export type SupportLanguageType = 'en-US' | 'zh-CN' | undefined;
