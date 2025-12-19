import { RouteRecordRaw } from 'vue-router';
import PluginRootLayout from 'layouts/PluginRootLayout.vue';
import PluginRootLayoutDev from 'layouts/PluginRootLayoutDev.vue';
import PluginMainLayout from 'layouts/PluginMainLayout.vue';
import PluginOptionsLayout from 'layouts/PluginOptionsLayout.vue';
import PluginItemIndex from 'src/pages/Mobile/vault/ItemIndex.vue';
import PluginItemView from 'src/pages/Mobile/vault/ItemView.vue';
import PluginSettingIndex from 'pages/Mobile/setting/SettingIndex.vue';
import PluginTranslateIndex from 'pages/Mobile/translate/TranslateIndex.vue';
// import PluginRssIndexPage from 'pages/Mobile/rss/RssIndexPage.vue';

import UnlockPage from 'src/pages/Mobile/login/unlock/UnlockPage.vue';
import AuthorizationPage from 'src/pages/Mobile/wallet/AuthorizationPage.vue';
import ConnectPage from 'src/pages/Mobile/wallet/ConnectPage.vue';
import SubmitPage from 'src/pages/Mobile/wallet/SubmitVCInfoPage.vue';

import VCCard from 'pages/Mobile/vc/card/VCCard.vue'; //error customElements.define("verifiable-credential",fi)
import ItemIndex from 'pages/Mobile/vault/ItemIndex.vue';
import ItemView from 'pages/Mobile/vault/ItemView.vue';
import Generator from 'pages/Web/Generator.vue';
import IndexPage from 'pages/Web/Security/IndexPage.vue';
import InviteRecipient from 'pages/Items/InviteRecipient.vue';
import OrgIndex from 'src/pages/Mobile/vault/OrgIndex.vue';
import OrgView from 'src/pages/Mobile/vault/OrgView.vue';
import SettingIndex from 'pages/Mobile/setting/SettingIndex.vue';
// import FilesPage from 'pages/Mobile/file/FilesPage.vue'; // error pinia not active
import AccountList from 'src/pages/Mobile/AccountList.vue';
// import AddVCPage from 'pages/Mobile/secret/AddVCPage.vue';
import VCCardList from 'src/pages/Mobile/secret/VCCardList.vue';
import SelectTerminusName from 'pages/Mobile/secret/SelectTerminusName.vue';
import AccountPage from 'src/pages/Mobile/setting/AccountPage.vue';
import DisplayPage from 'pages/Mobile/setting/DisplayPage.vue';
import SecurityPage from 'pages/Mobile/setting/SecurityPage.vue';
import AutoFillPage from 'pages/Mobile/setting/AutoFillPage.vue';
import WebsiteManagerPage from 'pages/Mobile/setting/WebsiteManagerPage.vue';
import ProfilePage from 'pages/Mobile/ProfilePage.vue';
import ChangePwdPage from 'pages/Mobile/setting/ChangePwdPage.vue';
import BackupMnemonicsPage from 'src/pages/Mobile/setting/BackupMnemonicsPage.vue';
// import VCManagePage from 'pages/Mobile/setting/VCManagePage.vue'; //error customElements.define("verifiable-credential",fi)
// import FilePreviewPage from 'pages/Mobile/file/FilePreviewPage.vue'; //error Function("r","regeneratorRuntime = r")
import IndexMobile from 'src/pages/Mobile/cloud/login/IndexMobile.vue';
import CollectPage from 'src/pages/Plugin/collect/IndexPage.vue';
import Home from 'src/pages/Plugin/home/IndexPage.vue';
import Application from 'src/pages/Plugin/application/IndexPage.vue';
import Subscribe from 'src/pages/Plugin/subscribe/IndexPage.vue';
import AppearanceOPtionsPage from 'src/pages/Plugin/options/appearance/IndexPage.vue';
import CollectOPtionsPage from 'src/pages/Plugin/options/collect/IndexPage.vue';
import TranslateOPtionsPage from 'src/pages/Plugin/options/translate/IndexPage.vue';
import SecurityOPtionsPage from 'src/pages/Plugin/options/security/IndexPage.vue';
import AccountOPtionsPage from 'src/pages/Plugin/options/account/IndexPage.vue';

import AccountListBex from 'src/pages/Plugin/options/account/AccountListBex.vue';
import { ROUTE_CONST } from './route-const';
import { isChromeExtension } from 'src/utils/bex/link';

const mobileCommon: RouteRecordRaw[] = [
	{
		path: '/',
		component: PluginMainLayout,
		name: 'MobileMainLayout',
		children: [
			{
				path: '/vc/card',
				component: VCCard
			},
			{
				path: 'secret',
				meta: {
					tabIdentify: 'secret'
				},
				component: ItemIndex
			},

			{
				path: 'items',
				meta: {
					tabIdentify: 'secret'
				},
				component: ItemIndex
			},
			{
				path: 'items/:itemid',
				component: ItemView
			},
			{
				path: 'security/',
				component: IndexPage
			},
			{
				path: 'invite-recipient/:org_id/:invite_id',
				component: InviteRecipient
			},
			{
				path: 'setting',
				meta: {
					tabIdentify: 'setting',
					minimizeApp: 'true'
				},
				component: SettingIndex
			},
			{
				path: '/VC_card_list',
				component: VCCardList
			},
			{
				path: '/select_terminus_name',
				component: SelectTerminusName
			},
			{
				path: '/setting/account',
				component: AccountPage
			},
			{
				path: '/setting/display',
				component: DisplayPage
			},
			{
				path: '/setting/security',
				component: SecurityPage
			},
			{
				path: '/setting/autofill',
				component: AutoFillPage
			},
			{
				path: '/setting/website',
				component: WebsiteManagerPage
			},
			{
				path: '/profile',
				component: ProfilePage
			},
			{
				path: '/change_pwd',
				name: 'changePwd',
				component: ChangePwdPage
			},
			{
				path: '/backup_mnemonics',
				name: 'backupMnemonics',
				component: BackupMnemonicsPage
			},
			{
				path: 'LoginCloud',
				component: IndexMobile
			}
		]
	}
];

const bex: RouteRecordRaw[] = [
	{
		path: '/',
		component: PluginMainLayout,
		children: [
			{
				path: '/collect',
				name: ROUTE_CONST.COLLECT,
				meta: {
					tabIdentify: 'collect'
				},
				component: CollectPage
			},
			{
				path: '/items',
				meta: {
					tabIdentify: 'vault'
				},
				component: PluginItemIndex
			},
			{
				path: '/items/:itemid',
				component: PluginItemView
			},
			{
				path: '/setting',
				meta: {
					tabIdentify: 'setting'
				},
				component: PluginSettingIndex
			},
			{
				path: '/translate',
				name: ROUTE_CONST.TRANSLATE,
				meta: {
					tabIdentify: 'translate'
				},
				component: () => import('src/pages/Mobile/translate/TranslateIndex.vue')
			},
			{
				path: 'org/:org_mode',
				meta: {
					tabIdentify: 'vault'
				},
				component: OrgIndex
			},
			{
				path: 'org/:org_mode/:org_type',
				meta: {
					tabIdentify: 'vault'
				},
				component: OrgView
			},
			{
				path: 'generator/',
				meta: {
					tabIdentify: 'vault'
				},
				component: Generator
			},
			{
				path: '/accounts',
				name: 'accounts',
				meta: {
					noReturn: true
				},
				component: AccountList
			},
			{
				path: '/application',
				name: 'application',
				meta: {
					tabIdentify: 'application'
				},
				component: Application
			},
			{
				path: '/subscribe',
				name: 'subscribe',
				meta: {
					tabIdentify: 'subscribe'
				},
				component: Subscribe
			}
		]
	}
];

const optionLogin: RouteRecordRaw[] = [
	{
		path: '/',
		component: PluginMainLayout,
		children: [
			{
				path: '/welcome',
				name: 'welcome',
				component: () => import('src/pages/Mobile/login/WelcomePage.vue')
			},
			{
				path: '/setUnlockPassword',
				name: 'setUnlockPassword',
				component: () =>
					import('src/pages/Mobile/login/unlock/SetUnlockPwd.vue')
			},
			{
				path: '/import_mnemonic',
				name: 'InputMnemonic',
				component: () => import('pages/Mobile/login/account/InputMnemonic.vue')
			},
			{
				path: 'connectLoading',
				component: () => import('src/pages/Mobile/connect/ConnectLoading.vue')
			},
			{
				path: 'ConnectTerminus',
				component: () => import('src/pages/Mobile/connect/ConnectTerminus.vue')
			}
		]
	}
];

const mobile: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/LarepassLoginLayout.vue'),
		children: [
			{
				path: '/',
				component: PluginMainLayout,
				children: []
			}
		]
	},
	{
		path: '/',
		component: PluginMainLayout,
		children: [
			{
				path: '/authorization',
				component: AuthorizationPage
			},
			{
				path: '/connect',
				component: ConnectPage
			},
			{
				path: '/submit',
				component: SubmitPage
			},
			{
				path: '/unlock',
				name: 'unlock',
				meta: {
					emptyUserDisableBack: true
				},
				component: UnlockPage
			}
		]
	}
];

const mobileExtension: RouteRecordRaw[] = [
	{
		path: '/',
		component: PluginMainLayout,
		children: [
			{
				path: '/home',
				name: 'HOME',
				meta: {
					tabIdentify: 'home'
				},
				component: Home
			}
		]
	}
];

export const optionExtension: RouteRecordRaw[] = [
	{
		path: '/home',
		name: ROUTE_CONST.OPTIONS_ACCOUNT,
		redirect: '/options/account',
		meta: {
			tabIdentify: ROUTE_CONST.OPTIONS_ACCOUNT
		},
		component: AccountOPtionsPage,
		children: [
			{
				path: '/options/account',
				name: ROUTE_CONST.OPTIONS_ACCOUNT_LIST,
				component: AccountListBex
			}
		]
	}
];

const bexOptions: RouteRecordRaw[] = [
	{
		path: '/',
		component: PluginOptionsLayout,
		children: [
			{
				path: '/options/appearance',
				name: ROUTE_CONST.OPTIONS_APPEARANCE,
				meta: {
					tabIdentify: ROUTE_CONST.OPTIONS_APPEARANCE
				},
				component: AppearanceOPtionsPage
			},
			{
				path: '/options/collect',
				name: ROUTE_CONST.OPTIONS_COLLECT,
				meta: {
					tabIdentify: ROUTE_CONST.OPTIONS_COLLECT
				},
				component: CollectOPtionsPage
			},
			{
				path: '/options/translate',
				name: ROUTE_CONST.OPTIONS_TRANSLATE,
				meta: {
					tabIdentify: ROUTE_CONST.OPTIONS_TRANSLATE
				},
				component: TranslateOPtionsPage
			},
			{
				path: '/options/security',
				name: ROUTE_CONST.OPTIONS_SECURITY,
				meta: {
					tabIdentify: ROUTE_CONST.OPTIONS_SECURITY
				},
				component: SecurityOPtionsPage
			}
		]
	}
];

const desktopOptions: RouteRecordRaw[] = [
	{
		path: '/',
		component: () => import('layouts/LarepassLoginLayout.vue'),
		children: [
			{
				path: '/welcome',
				name: 'welcome',
				component: () => import('src/pages/Electron/import/WelcomePage.vue')
			},
			{
				path: '/setUnlockPassword',
				name: 'setUnlockPassword',
				component: () => import('pages/Electron/import/SetUnlockPwdPage.vue')
			},
			{
				path: '/import_mnemonic',
				name: 'InputMnemonic',
				component: () => import('pages/Electron/import/InputMnemonicPage.vue')
			},
			{
				path: 'connectLoading',
				component: () => import('src/pages/Mobile/connect/ConnectLoading.vue')
			},
			{
				path: 'ConnectTerminus',
				component: () => import('src/pages/Mobile/connect/ConnectTerminus.vue')
			}
		]
	}
];

const homeRoutes = isChromeExtension()
	? {
			component: PluginOptionsLayout,
			path: '/optionExtension',
			children: optionExtension
	  }
	: {
			path: '/mobileExtension',
			children: mobileExtension
	  };

const loginRoutes = isChromeExtension()
	? {
			path: '/desktop',
			children: desktopOptions
	  }
	: {
			path: '/optionLogin',
			children: optionLogin
	  };

const routes: RouteRecordRaw[] = [
	{
		path: '/',
		component: process.env.DEV_PLATFORM_BEX
			? PluginRootLayoutDev
			: PluginRootLayout,
		children: [
			homeRoutes,
			{
				path: '/mobile',
				children: mobile
			},
			{
				path: '/mobileCommon',
				children: mobileCommon
			},
			{
				path: '/bex',
				children: bex
			},
			loginRoutes,
			{
				path: '/options',
				children: bexOptions
			}
		]
	}
];

export default routes;
