import { browser } from './libs/browser';
import {
	MSG_FETCH,
	MSG_GET_HTTPCACHE,
	MSG_TRANS_TOGGLE,
	MSG_OPEN_OPTIONS,
	MSG_SAVE_RULE,
	MSG_TRANS_TOGGLE_STYLE,
	MSG_OPEN_TRANBOX,
	MSG_CONTEXT_MENUS,
	MSG_COMMAND_SHORTCUTS,
	MSG_INJECT_JS,
	MSG_INJECT_CSS,
	MSG_UPDATE_CSP,
	DEFAULT_CSPLIST,
	CMD_TOGGLE_TRANSLATE,
	CMD_TOGGLE_STYLE,
	CMD_OPEN_OPTIONS,
	CMD_OPEN_TRANBOX
} from './config';
import { getSettingWithDefault, tryInitDefaultData } from './libs/storage';
import { trySyncSettingAndRules } from './libs/sync';
import { fetchHandle, getHttpCache } from './libs/fetch';
import { sendTabMsg } from './libs/msg';
import { trySyncAllSubRules } from './libs/subRules';
import { tryClearCaches } from './libs';
import { saveRule } from './libs/rules';
import { getCurTabId } from './libs/msg';
import { injectInlineJs, injectInternalCss } from './libs/injector';
import { kissLog } from './libs/log';

globalThis.ContextType = 'BACKGROUND';

const REMOVE_HEADERS = [
	`content-security-policy`,
	`content-security-policy-report-only`,
	`x-webkit-csp`,
	`x-content-security-policy`
];

async function updateCspRules(csplist = DEFAULT_CSPLIST.join(',\n')) {
	try {
		const newRules = csplist
			.split(/\n|,/)
			.map((url) => url.trim())
			.filter(Boolean)
			.map((url, idx) => ({
				id: idx + 1,
				action: {
					type: 'modifyHeaders',
					responseHeaders: REMOVE_HEADERS.map((header) => ({
						operation: 'remove',
						header
					}))
				},
				condition: {
					urlFilter: url,
					resourceTypes: ['main_frame', 'sub_frame']
				}
			}));
		const oldRules = await browser.declarativeNetRequest.getDynamicRules();
		const oldRuleIds = oldRules.map((rule) => rule.id);
		await browser.declarativeNetRequest.updateDynamicRules({
			removeRuleIds: oldRuleIds,
			addRules: newRules
		});
	} catch (err) {
		kissLog(err, 'update csp rules');
	}
}

async function addContextMenus(contextMenuType = 1) {
	try {
		await browser.contextMenus.removeAll();
	} catch (err) {
		//
	}

	switch (contextMenuType) {
		case 1:
			browser.contextMenus.create({
				id: CMD_TOGGLE_TRANSLATE,
				title: browser.i18n.getMessage('app_name'),
				contexts: ['page', 'selection']
			});
			break;
		case 2:
			browser.contextMenus.create({
				id: CMD_TOGGLE_TRANSLATE,
				title: browser.i18n.getMessage('toggle_translate'),
				contexts: ['page', 'selection']
			});
			browser.contextMenus.create({
				id: CMD_TOGGLE_STYLE,
				title: browser.i18n.getMessage('toggle_style'),
				contexts: ['page', 'selection']
			});
			browser.contextMenus.create({
				id: CMD_OPEN_TRANBOX,
				title: browser.i18n.getMessage('open_tranbox'),
				contexts: ['page', 'selection']
			});
			browser.contextMenus.create({
				id: 'options_separator',
				type: 'separator',
				contexts: ['page', 'selection']
			});
			browser.contextMenus.create({
				id: CMD_OPEN_OPTIONS,
				title: browser.i18n.getMessage('open_options'),
				contexts: ['page', 'selection']
			});
			break;
		default:
	}
}

export function translateInit() {
	browser.runtime.onInstalled.addListener(() => {
		tryInitDefaultData();

		addContextMenus();

		updateCspRules();
	});

	browser.runtime.onStartup.addListener(async () => {
		await trySyncSettingAndRules();

		const { clearCache, contextMenuType, subrulesList, csplist } =
			await getSettingWithDefault();

		if (clearCache) {
			tryClearCaches();
		}

		addContextMenus(contextMenuType);

		updateCspRules(csplist);

		trySyncAllSubRules({ subrulesList });
	});

	// browser.commands.onCommand.addListener((command) => {
	// 	// console.log(`Command: ${command}`);
	// 	switch (command) {
	// 		case CMD_TOGGLE_TRANSLATE:
	// 			sendTabMsg(MSG_TRANS_TOGGLE);
	// 			break;
	// 		case CMD_OPEN_TRANBOX:
	// 			sendTabMsg(MSG_OPEN_TRANBOX);
	// 			break;
	// 		case CMD_TOGGLE_STYLE:
	// 			sendTabMsg(MSG_TRANS_TOGGLE_STYLE);
	// 			break;
	// 		case CMD_OPEN_OPTIONS:
	// 			browser.runtime.openOptionsPage();
	// 			break;
	// 		default:
	// 	}
	// });

	browser.contextMenus.onClicked.addListener(({ menuItemId }) => {
		switch (menuItemId) {
			case CMD_TOGGLE_TRANSLATE:
				sendTabMsg(MSG_TRANS_TOGGLE);
				break;
			case CMD_TOGGLE_STYLE:
				sendTabMsg(MSG_TRANS_TOGGLE_STYLE);
				break;
			case CMD_OPEN_TRANBOX:
				sendTabMsg(MSG_OPEN_TRANBOX);
				break;
			case CMD_OPEN_OPTIONS:
				browser.runtime.openOptionsPage();
				break;
			default:
		}
	});
}

export async function translateMessageHandler(msg) {
	const { action, args } = msg;
	switch (action) {
		case MSG_FETCH:
			return await fetchHandle(args);
		case MSG_GET_HTTPCACHE:
			// eslint-disable-next-line no-case-declarations
			const { input, init } = args;
			return await getHttpCache(input, init);
		case MSG_OPEN_OPTIONS:
			return await browser.runtime.openOptionsPage();
		case MSG_SAVE_RULE:
			return await saveRule(args);
		case MSG_INJECT_JS:
			return await browser.scripting.executeScript({
				target: { tabId: await getCurTabId(), allFrames: true },
				func: injectInlineJs,
				args: [args],
				world: 'MAIN'
			});
		case MSG_INJECT_CSS:
			return await browser.scripting.executeScript({
				target: { tabId: await getCurTabId(), allFrames: true },
				func: injectInternalCss,
				args: [args],
				world: 'MAIN'
			});
		case MSG_UPDATE_CSP:
			return await updateCspRules(args);
		case MSG_CONTEXT_MENUS:
			return await addContextMenus(args);
		case MSG_COMMAND_SHORTCUTS:
			return await browser.commands.getAll();
		default:
			throw new Error(`message action is unavailable: ${action}`);
	}
}
