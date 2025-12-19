// import { CLIENT_EXTS, CLIENT_USERSCRIPT, CLIENT_WEB } from "../config";

/**
 * 浏览器兼容插件，另可用于判断是插件模式还是网页模式，方便开发
 * @returns
 */

import { browser as browserTemp } from 'webextension-polyfill-ts';

export const browser = browserTemp;

export const isBg = () => globalThis?.ContextType === 'BACKGROUND';
