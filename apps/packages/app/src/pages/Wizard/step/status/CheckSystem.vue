<template>
	<div class="row justify-evenly account_box">
		<div class="q-pa-md;wrap account_center" style="word-wrap: break-word">
			<div class="boot_justify">
				<animationPage
					:picture="wizard_system"
					:certificate="wizard_systema"
					:isAnimated="true"
				/>
			</div>
			<p class="get-text">{{ t('system_title') }}</p>
			<p class="account_please_text please_text">
				{{ t('system_content_1') }}
			</p>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import animationPage from '../animation.vue';

const wizard_systema = require('../../../../assets/wizard-web/wizard_systema.png');
const wizard_system = require('../../../../assets/wizard-web/wizard_system.png');

export default defineComponent({
	name: 'CheckVault',
	components: { animationPage },
	setup() {
		const { t } = useI18n();

		return {
			wizard_systema,
			wizard_system,
			t
		};
	}
});
</script>
