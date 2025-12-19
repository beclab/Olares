import enUS from './en-US';
import zhCN from './zh-CN';
import kubeEnUS from 'src/i18n/controlPanelCommon';

export default {
	'en-US': { ...kubeEnUS['en-US'], ...enUS },
	'zh-CN': { ...kubeEnUS['zh-CN'], ...zhCN }
};
