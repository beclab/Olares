import { boot } from 'quasar/wrappers';
import { createI18n } from 'vue-i18n';
import { getLangWithMeta } from '@apps/control-panel-common/src/utils/lang';
import messages from 'src/i18n/controlHub';

const locale = getLangWithMeta();

export type MessageLanguages = keyof typeof messages;
// Type-define 'en-US' as the master schema for the resource
export type MessageSchema = typeof messages;

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

const i18n = createI18n<{ message: MessageLanguages }, MessageLanguages>({
	locale: locale,
	legacy: false,
	fallbackLocale: 'en-US',
	allowComposition: true,
	messages
});

export default boot(({ app }) => {
	// Set i18n instance on app
	app.use(i18n);
});

const { t } = i18n.global;

export { t };
