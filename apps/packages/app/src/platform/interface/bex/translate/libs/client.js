import { CLIENT_EXTS, CLIENT_USERSCRIPT, CLIENT_WEB } from '../config';

export const client = 'chrome';
export const isExt = CLIENT_EXTS.includes(client);
export const isGm = client === CLIENT_USERSCRIPT;
export const isWeb = client === CLIENT_WEB;
