// types/vue-i18n.d.ts
import { I18n, Composer } from 'vue-i18n';

declare module '@vue/runtime-core' {
	interface ComponentCustomProperties {
		$i18n: I18n;
		$t: Composer['t'];
	}
	interface ComponentCustomProperties {
		$i18n: I18n;
		$te: Composer['te'];
	}
}
