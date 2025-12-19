/* eslint-disable no-undef */
import {
	MSG_TRANS_TOGGLE,
	MSG_TRANS_TOGGLE_STYLE,
	MSG_TRANS_GETRULE,
	MSG_TRANS_PUTRULE,
	MSG_OPEN_TRANBOX
	// APP_LCNAME,
	// DEFAULT_TRANBOX_SETTING
} from './config';
import { getSettingWithDefault } from './libs/storage';
import { Translator } from './libs/translator';
import { isIframe, sendIframeMsg } from './libs/iframe';
// import Slection from "./views/Selection";
import { touchTapListener } from './libs/touch';
import { debounce, genEventName } from './libs/utils';
import { handlePing, injectScript } from './libs/gm';
import { browser } from './libs/browser';
import { matchRule } from './libs/rules';
import { trySyncAllSubRules } from './libs/subRules';
import { isInBlacklist } from './libs/blacklist';
import inputTranslate from './libs/inputTranslate';
import {
	REACT_APP_NAME,
	REACT_APP_VERSION,
	REACT_APP_OPTIONSPAGE_DEV,
	REACT_APP_OPTIONSPAGE
} from './env';

/**
 * Tampermonkey script settings page
 */
function runSettingPage() {
	if (GM?.info?.script?.grant?.includes('unsafeWindow')) {
		unsafeWindow.GM = GM;
		unsafeWindow.APP_INFO = {
			name: REACT_APP_NAME,
			version: REACT_APP_VERSION
		};
	} else {
		const ping = genEventName();
		window.addEventListener(ping, handlePing);
		const script = document.createElement('script');
		script.textContent = `(${injectScript})("${ping}")`;
		document.head.append(script);
	}
}

/**
 * Extension listens to background events
 * @param {*} translator
 */
function runtimeListener(translator) {
	browser?.runtime.onMessage.addListener((message, sender, sendResponse) => {
		const { action, args } = message;
		switch (action) {
			case MSG_TRANS_TOGGLE:
				translator.toggle();
				sendIframeMsg(MSG_TRANS_TOGGLE);
				break;
			case MSG_TRANS_TOGGLE_STYLE:
				translator.toggleStyle();
				sendIframeMsg(MSG_TRANS_TOGGLE_STYLE);
				break;
			case MSG_TRANS_GETRULE:
				break;
			case MSG_TRANS_PUTRULE:
				translator.updateRule(args);
				sendIframeMsg(MSG_TRANS_PUTRULE, args);
				break;
			case MSG_OPEN_TRANBOX:
				window.dispatchEvent(new CustomEvent(MSG_OPEN_TRANBOX));
				break;
			default:
				return { error: `message action is unavailable: ${action}` };
		}
		sendResponse({ data: translator.rule });
	});
}

/**
 * iframe page execution
 * @param {*} translator
 */
function runIframe(translator) {
	window.addEventListener('message', (e) => {
		const { action, args } = e.data || {};
		switch (action) {
			case MSG_TRANS_TOGGLE:
				translator?.toggle();
				break;
			case MSG_TRANS_TOGGLE_STYLE:
				translator?.toggleStyle();
				break;
			case MSG_TRANS_PUTRULE:
				translator.updateRule(args || {});
				break;
			default:
		}
	});
}

/**
 * Floating button
 * @param {*} translator
 * @returns
 */
// async function showFab(translator) {
//   const fab = await getFabWithDefault();
//   const $action = document.createElement("div");
//   $action.setAttribute("id", APP_LCNAME);
//   $action.style.fontSize = "0";
//   $action.style.width = "0";
//   $action.style.height = "0";
//   document.body.parentElement.appendChild($action);
//   const shadowContainer = $action.attachShadow({ mode: "closed" });
//   const emotionRoot = document.createElement("style");
//   const shadowRootElement = document.createElement("div");
//   shadowContainer.appendChild(emotionRoot);
//   shadowContainer.appendChild(shadowRootElement);
//   const cache = createCache({
//     key: APP_LCNAME,
//     prepend: true,
//     container: emotionRoot,
//   });
//   ReactDOM.createRoot(shadowRootElement).render(
//     <React.StrictMode>
//       <CacheProvider value={cache}>
//         <Action translator={translator} fab={fab} />
//       </CacheProvider>
//     </React.StrictMode>
//   );
// }

/**
 * Word selection translation
 * @param {*} param0
 * @returns
 */
// function showTransbox({
//   contextMenuType,
//   tranboxSetting = DEFAULT_TRANBOX_SETTING,
//   transApis,
//   darkMode,
//   uiLang,
//   langDetector,
// }) {
//   if (!tranboxSetting?.transOpen) {
//     return;
//   }

//   const $tranbox = document.createElement("div");
//   $tranbox.setAttribute("id", "kiss-transbox");
//   $tranbox.style.fontSize = "0";
//   $tranbox.style.width = "0";
//   $tranbox.style.height = "0";
//   document.body.parentElement.appendChild($tranbox);
//   const shadowContainer = $tranbox.attachShadow({ mode: "closed" });
//   const emotionRoot = document.createElement("style");
//   const shadowRootElement = document.createElement("div");
//   shadowRootElement.classList.add(`KT-transbox`);
//   shadowRootElement.classList.add(`KT-transbox_${darkMode ? "dark" : "light"}`);
//   shadowContainer.appendChild(emotionRoot);
//   shadowContainer.appendChild(shadowRootElement);
//   const cache = createCache({
//     key: "kiss-transbox",
//     prepend: true,
//     container: emotionRoot,
//   });
//   ReactDOM.createRoot(shadowRootElement).render(
//     <React.StrictMode>
//       <CacheProvider value={cache}>
//         <Slection
//           contextMenuType={contextMenuType}
//           tranboxSetting={tranboxSetting}
//           transApis={transApis}
//           uiLang={uiLang}
//           langDetector={langDetector}
//         />
//       </CacheProvider>
//     </React.StrictMode>
//   );
// }

/**
 * Display error message at the top of the page
 * @param {*} message
 */
function showErr(message) {
	const $err = document.createElement('div');
	$err.innerText = `KISS-Translator: ${message}`;
	$err.style.cssText = 'background:red; color:#fff;';
	document.body.prepend($err);
}

/**
 * Listen to touch operations
 * @param {*} translator
 * @returns
 */
function touchOperation(translator) {
	const { touchTranslate = 2 } = translator.setting;
	if (touchTranslate === 0) {
		return;
	}

	const handleTap = debounce(() => {
		translator.toggle();
		sendIframeMsg(MSG_TRANS_TOGGLE);
	});
	touchTapListener(handleTap, touchTranslate);
}

/**
 * Entry function
 */
export async function run(isUserscript = false) {
	try {
		const href = document.location.href;

		// Settings page
		if (
			isUserscript &&
			(href.includes(REACT_APP_OPTIONSPAGE_DEV) ||
				href.includes(REACT_APP_OPTIONSPAGE))
		) {
			runSettingPage();
			return;
		}

		// Read settings
		const setting = await getSettingWithDefault();

		// Blacklist
		if (isInBlacklist(href, setting)) {
			return;
		}
		console.log('fffff-1', href, setting);
		// Translate webpage
		const rule = await matchRule(href, setting);
		const translator = new Translator(rule, setting);

		// Adapt iframe
		if (isIframe) {
			runIframe(translator);
			return;
		}

		// Listen to messages
		!isUserscript && runtimeListener(translator);

		// Input box translation
		inputTranslate(setting);

		// Word selection translation
		// showTransbox(setting);

		// Floating ball button
		// await showFab(translator);

		// Touch operation
		touchOperation(translator);

		// Sync subscription rules
		isUserscript && (await trySyncAllSubRules(setting));
	} catch (err) {
		console.error('[KISS-Translator]', err);
		showErr(err.message);
	}
}
