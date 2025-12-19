import { boot } from 'quasar/wrappers';
import { createI18n } from 'vue-i18n';

import messages, { defaultLanguage } from 'src/i18n';
import enUS from '../i18n/en-US';

// Type-define 'en-US' as the master schema for the resource
type MessageSchema = typeof enUS;

// See https://vue-i18n.intlify.dev/guide/advanced/typescript.html#global-resource-schema-type-definition
/* eslint-disable @typescript-eslint/no-empty-interface */
declare module 'vue-i18n' {
	// define the locale messages schema
	export interface DefineLocaleMessage extends MessageSchema {}

	// define the datetime format schema
	export interface DefineDateTimeFormat {}

	// define the number format schema
	export interface DefineNumberFormat {}
}
/* eslint-enable @typescript-eslint/no-empty-interface */

export const i18n = createI18n({
	locale: defaultLanguage,
	legacy: false,
	messages
});

export const t = i18n.global.t;

export default boot(({ app }) => {
	// Set i18n instance on app
	app.use(i18n);
});
