import enUS from './en-US';
import zhCN from './zh-CN';
import ControlHubLang from 'src/i18n/controlHub';
import dashboardLang from 'src/i18n/dashboard';

export default {
	'en-US': { ...ControlHubLang['en-US'], ...dashboardLang['en-US'], ...enUS },
	'zh-CN': { ...ControlHubLang['zh-CN'], ...dashboardLang['zh-CN'], ...zhCN }
};
