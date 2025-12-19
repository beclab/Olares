import { createApp } from 'vue';
import {
	APP_LCNAME,
	TRANS_MIN_LENGTH,
	TRANS_MAX_LENGTH,
	MSG_TRANS_CURRULE,
	MSG_INJECT_JS,
	MSG_INJECT_CSS,
	OPT_STYLE_DASHLINE,
	OPT_STYLE_FUZZY,
	SHADOW_KEY,
	OPT_TIMING_PAGESCROLL,
	OPT_TIMING_PAGEOPEN,
	OPT_TIMING_MOUSEOVER,
	DEFAULT_TRANS_APIS,
	DEFAULT_FETCH_LIMIT,
	DEFAULT_FETCH_INTERVAL
} from '../config';
import Content from '../views/Content/IndexPage.vue';
import { updateFetchPool, clearFetchPool } from './fetch';
import { debounce, genEventName } from './utils';
import { runFixer } from './webfix';
import { apiTranslate } from '../apis';
import { sendBgMsg } from './msg';
import { isExt } from './client';
import { injectInlineJs, injectInternalCss } from './injector';
import { kissLog } from './log';
import interpreter from './interpreter';

/**
 * Translator class
 */
export class Translator {
	_rule = {};
	_setting = {};
	_rootNodes = new Set();
	_tranNodes = new Map();
	_skipNodeNames = [
		APP_LCNAME,
		'style',
		'svg',
		'img',
		'audio',
		'video',
		'textarea',
		'input',
		'button',
		'select',
		'option',
		'head',
		'script',
		'iframe'
	];
	_eventName = genEventName();
	_mouseoverNode = null;
	_keepSelector = [null, null];
	_terms = [];
	_docTitle = '';

	// Display
	_interseObserver = new IntersectionObserver(
		(entries, observer) => {
			entries.forEach((entry) => {
				if (entry.isIntersecting) {
					observer.unobserve(entry.target);
					this._render(entry.target);
				}
			});
		},
		{
			threshold: 0.1
		}
	);

	// Changes
	_mutaObserver = new MutationObserver((mutations) => {
		mutations.forEach((mutation) => {
			if (
				!this._skipNodeNames.includes(mutation.target.localName) &&
				mutation.addedNodes.length > 0
			) {
				const nodes = Array.from(mutation.addedNodes).filter((node) => {
					if (
						this._skipNodeNames.includes(node.localName) ||
						node.id === APP_LCNAME
					) {
						return false;
					}
					return true;
				});
				if (nodes.length > 0) {
					// const rootNode = mutation.target.getRootNode();
					// todo
					this._reTranslate();
				}
			}
		});
	});

	// Insert shadowroot
	_overrideAttachShadow = () => {
		const _this = this;
		const _attachShadow = HTMLElement.prototype.attachShadow;
		HTMLElement.prototype.attachShadow = function () {
			_this._reTranslate();
			return _attachShadow.apply(this, arguments);
		};
	};

	_updatePool(translator) {
		if (!translator) {
			return;
		}

		const {
			fetchInterval = DEFAULT_FETCH_INTERVAL,
			fetchLimit = DEFAULT_FETCH_LIMIT
		} = this._setting.transApis[translator] || {};
		updateFetchPool(fetchInterval, fetchLimit);
	}

	constructor(rule, setting) {
		this._overrideAttachShadow();

		this._setting = setting;
		this._rule = rule;

		this._keepSelector = (rule.keepSelector || '')
			.split(SHADOW_KEY)
			.map((item) => item.trim());
		this._terms = (rule.terms || '')
			.split(/\n|;/)
			.map((item) => item.split(',').map((item) => item.trim()))
			.filter(([term]) => Boolean(term));

		this._updatePool(rule.translator);

		if (rule.transOpen === 'true') {
			this._register();
		}
	}

	get setting() {
		return this._setting;
	}

	get eventName() {
		return this._eventName;
	}

	get rule() {
		// console.log("get rule", this._rule);
		return this._rule;
	}

	set rule(rule) {
		// console.log("set rule", rule);
		this._rule = rule;

		// Broadcast message
		const eventName = this._eventName;
		window.dispatchEvent(
			new CustomEvent(eventName, {
				detail: {
					action: MSG_TRANS_CURRULE,
					args: rule
				}
			})
		);
	}

	updateRule = (obj) => {
		console.log('updateRule', obj);
		this.rule = { ...this.rule, ...obj };
		this._updatePool(obj.translator);
	};

	toggle = () => {
		if (this.rule.transOpen === 'true') {
			this.rule = { ...this.rule, transOpen: 'false' };
			this._unRegister();
		} else {
			this.rule = { ...this.rule, transOpen: 'true' };
			this._register();
		}
	};

	toggleStyle = () => {
		const textStyle =
			this.rule.textStyle === OPT_STYLE_FUZZY
				? OPT_STYLE_DASHLINE
				: OPT_STYLE_FUZZY;
		this.rule = { ...this.rule, textStyle };
	};

	translateText = async (text) => {
		const { translator, fromLang, toLang } = this._rule;
		const apiSetting =
			this._setting.transApis?.[translator] || DEFAULT_TRANS_APIS[translator];
		const [trText] = await apiTranslate({
			text,
			translator,
			fromLang,
			toLang,
			apiSetting
		});
		return trText;
	};

	_querySelectorAll = (selector, node) => {
		try {
			return Array.from(node.querySelectorAll(selector));
		} catch (err) {
			kissLog(selector, 'querySelectorAll err');
		}
		return [];
	};

	_queryFilter = (selector, rootNode) => {
		return this._querySelectorAll(selector, rootNode).filter(
			(node) => this._queryFilter(selector, node).length === 0
		);
	};

	_queryShadowNodes = (selector, rootNode) => {
		this._rootNodes.add(rootNode);
		this._queryFilter(selector, rootNode).forEach((item) => {
			if (!this._tranNodes.has(item)) {
				this._tranNodes.set(item, '');
			}
		});

		Array.from(rootNode.querySelectorAll('*'))
			.map((item) => item.shadowRoot)
			.filter(Boolean)
			.forEach((item) => {
				this._queryShadowNodes(selector, item);
			});
	};

	_queryNodes = (rootNode = document) => {
		// const childRoots = Array.from(rootNode.querySelectorAll("*"))
		//   .map((item) => item.shadowRoot)
		//   .filter(Boolean);
		// const childNodes = childRoots.map((item) => this._queryNodes(item));
		// const nodes = Array.from(rootNode.querySelectorAll(this.rule.selector));
		// return nodes.concat(childNodes).flat();

		this._rootNodes.add(rootNode);
		this._rule.selector
			.split(';')
			.map((item) => item.trim())
			.filter(Boolean)
			.forEach((selector) => {
				if (selector.includes(SHADOW_KEY)) {
					const [outSelector, inSelector] = selector
						.split(SHADOW_KEY)
						.map((item) => item.trim());
					if (outSelector && inSelector) {
						const outNodes = this._querySelectorAll(outSelector, rootNode);
						outNodes.forEach((outNode) => {
							if (outNode.shadowRoot) {
								// this._rootNodes.add(outNode.shadowRoot);
								// this._queryFilter(inSelector, outNode.shadowRoot).forEach(
								//   (item) => {
								//     if (!this._tranNodes.has(item)) {
								//       this._tranNodes.set(item, "");
								//     }
								//   }
								// );
								this._queryShadowNodes(inSelector, outNode.shadowRoot);
							}
						});
					}
				} else {
					this._queryFilter(selector, rootNode).forEach((item) => {
						if (!this._tranNodes.has(item)) {
							this._tranNodes.set(item, '');
						}
					});
				}
			});
	};

	_register = () => {
		const { fromLang, toLang, injectJs, injectCss, fixerSelector, fixerFunc } =
			this._rule;
		if (fromLang === toLang) {
			return;
		}

		// webfix
		if (fixerSelector && fixerFunc !== '-') {
			runFixer(fixerSelector, fixerFunc);
		}

		// Inject user JS/CSS
		if (isExt) {
			injectJs && sendBgMsg(MSG_INJECT_JS, injectJs);
			injectCss && sendBgMsg(MSG_INJECT_CSS, injectCss);
		} else {
			injectJs && injectInlineJs(injectJs);
			injectCss && injectInternalCss(injectCss);
		}

		// Search nodes
		this._queryNodes();

		this._rootNodes.forEach((node) => {
			// Listen to node changes
			this._mutaObserver.observe(node, {
				childList: true,
				subtree: true
				// characterData: true,
			});
		});

		if (
			!this._rule.transTiming ||
			this._rule.transTiming === OPT_TIMING_PAGESCROLL
		) {
			// Listen to node display
			this._tranNodes.forEach((_, node) => {
				this._interseObserver.observe(node);
			});
		} else if (this._rule.transTiming === OPT_TIMING_PAGEOPEN) {
			// Direct full-text translation
			this._tranNodes.forEach((_, node) => {
				this._render(node);
			});
		} else {
			// Listen to mouse hover
			window.addEventListener('keydown', this._handleKeydown);
			this._tranNodes.forEach((_, node) => {
				node.addEventListener('mouseenter', this._handleMouseover);
				node.addEventListener('mouseleave', this._handleMouseout);
			});
		}

		// Translate page title
		if (this._rule.transTitle === 'true' && !this._docTitle) {
			const title = document.title;
			this._docTitle = title;
			this.translateText(title).then((trText) => {
				document.title = `${trText} | ${title}`;
			});
		}
	};

	_handleMouseover = (e) => {
		// console.log("mouseenter", e);
		if (!this._tranNodes.has(e.target)) {
			return;
		}

		const key = this._rule.transTiming.slice(3);
		if (this._rule.transTiming === OPT_TIMING_MOUSEOVER || e[key]) {
			e.target.removeEventListener('mouseenter', this._handleMouseover);
			e.target.removeEventListener('mouseleave', this._handleMouseout);
			this._render(e.target);
		} else {
			this._mouseoverNode = e.target;
		}
	};

	_handleMouseout = (e) => {
		// console.log("mouseleave", e);
		if (!this._tranNodes.has(e.target)) {
			return;
		}

		this._mouseoverNode = null;
	};

	_handleKeydown = (e) => {
		// console.log("keydown", e);
		const key = this._rule.transTiming.slice(3);
		if (e[key] && this._mouseoverNode) {
			this._mouseoverNode.removeEventListener(
				'mouseenter',
				this._handleMouseover
			);
			this._mouseoverNode.removeEventListener(
				'mouseleave',
				this._handleMouseout
			);

			const node = this._mouseoverNode;
			this._render(node);
			this._mouseoverNode = null;
		}
	};

	_unRegister = () => {
		// Restore page title
		if (this._docTitle) {
			document.title = this._docTitle;
			this._docTitle = '';
		}

		// Remove node change listener
		this._mutaObserver.disconnect();

		// Remove node display listener
		// this._interseObserver.disconnect();

		// Remove keyboard listener
		window.removeEventListener('keydown', this._handleKeydown);

		const { transRemoveHook } = this._rule;
		this._tranNodes.forEach((innerHTML, node) => {
			if (
				!this._rule.transTiming ||
				this._rule.transTiming === OPT_TIMING_PAGESCROLL
			) {
				// Remove node display listener
				this._interseObserver.unobserve(node);
			} else if (this._rule.transTiming !== OPT_TIMING_PAGEOPEN) {
				// Remove mouse hover listener
				// node.style.pointerEvents = "none";
				node.removeEventListener('mouseenter', this._handleMouseover);
				node.removeEventListener('mouseleave', this._handleMouseout);
			}

			// Remove/restore elements
			if (innerHTML) {
				if (this._rule.transOnly === 'true') {
					node.innerHTML = innerHTML;
				} else {
					node.querySelector(APP_LCNAME)?.remove();
				}
				// Hook function
				if (transRemoveHook?.trim()) {
					interpreter.run(`exports.transRemoveHook = ${transRemoveHook}`);
					interpreter.exports.transRemoveHook(node);
				}
			}
		});

		// Remove user JS/CSS
		this._removeInjector();

		// Clear node set
		this._rootNodes.clear();
		this._tranNodes.clear();

		// Clear task pool
		clearFetchPool();
	};

	_removeInjector = () => {
		document
			.querySelectorAll(`[data-source^="KISS-Calendar"]`)
			?.forEach((el) => el.remove());
	};

	_reTranslate = debounce(() => {
		if (this._rule.transOpen === 'true') {
			window.removeEventListener('keydown', this._handleKeydown);
			this._mutaObserver.disconnect();
			this._interseObserver.disconnect();
			this._removeInjector();
			this._register();
		}
	}, this._setting.transInterval);

	_invalidLength = (q) =>
		!q ||
		q.length < (this._setting.minLength ?? TRANS_MIN_LENGTH) ||
		q.length > (this._setting.maxLength ?? TRANS_MAX_LENGTH);

	_render = (el) => {
		let traEl = el.querySelector(APP_LCNAME);

		// Already translated
		if (traEl) {
			if (this._rule.transOnly === 'true') {
				return;
			}

			const preText = this._tranNodes.get(el);
			const curText = el.innerText.trim();
			// const traText = traEl.innerText.trim();

			// todo
			// 1. traText when loading
			// 2. replace startsWith

			if (curText.startsWith(preText)) {
				return;
			}

			traEl.remove();
		}

		let q = el.innerText.trim();
		if (this._rule.transOnly === 'true') {
			this._tranNodes.set(el, el.innerHTML);
		} else {
			this._tranNodes.set(el, q);
		}
		const keeps = [];

		// Translation start hook function
		const { transStartHook } = this._rule;
		if (transStartHook?.trim()) {
			interpreter.run(`exports.transStartHook = ${transStartHook}`);
			interpreter.exports.transStartHook(el, q);
		}

		// Keep elements
		const [matchSelector, subSelector] = this._keepSelector;
		if (matchSelector || subSelector) {
			let text = '';
			el.childNodes.forEach((child) => {
				if (
					child.nodeType === 1 &&
					((matchSelector && child.matches(matchSelector)) ||
						(subSelector && child.querySelector(subSelector)))
				) {
					if (child.nodeName === 'IMG') {
						child.style.cssText += `width: ${child.width}px;`;
						child.style.cssText += `height: ${child.height}px;`;
					}
					text += `[${keeps.length}]`;
					keeps.push(child.outerHTML);
				} else {
					text += child.textContent;
				}
			});

			if (keeps.length > 0) {
				// textContent will keep some useless newline characters, seriously affecting translation quality
				if (q.includes('\n')) {
					q = text;
				} else {
					q = text.replaceAll('\n', ' ');
				}
			}
		}

		// Too long or too short
		if (this._invalidLength(q.replace(/\[(\d+)\]/g, '').trim())) {
			return;
		}

		// Professional terms
		if (this._terms.length > 0) {
			for (const term of this._terms) {
				const re = new RegExp(term[0], 'g');
				q = q.replace(re, (t) => {
					const text = `[${keeps.length}]`;
					keeps.push(`<i class="kiss-trem">${term[1] || t}</i>`);
					return text;
				});
			}
		}

		// Additional styles
		const { selectStyle, parentStyle } = this._rule;
		el.style.cssText += selectStyle;
		if (el.parentElement) {
			el.parentElement.style.cssText += parentStyle;
		}

		// Insert translation node
		traEl = document.createElement(APP_LCNAME);
		traEl.style.visibility = 'visible';
		// if (this._rule.transOnly === "true") {
		//   el.innerHTML = "";
		// }
		el.appendChild(traEl);

		// Render translation node
		const app = createApp(Content, {
			q,
			keeps,
			translator: this,
			element: el
		});
		console.log('app', app);
		app.mount(traEl);
	};
}
