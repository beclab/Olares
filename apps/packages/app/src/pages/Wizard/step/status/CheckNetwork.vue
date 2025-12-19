<template>
	<div class="row justify-evenly account_box">
		<div
			class="q-pa-md wrap account_center"
			style="word-wrap: break-word"
			v-if="!wizardStatus || wizardStatus != 'network_activate_failed'"
		>
			<div class="boot_justify">
				<animationPage
					:picture="pictureImg"
					:certificate="certificateImg"
					:isAnimated="request_https_certificate"
				/>
			</div>
			<p class="get-text">{{ t('network_title') }}</p>
			<div>
				<p class="account_please_text please_text">
					{{ t('network_content') }}
				</p>
			</div>
			<template v-if="state">
				<p class="State_text">State {{ state }}</p>
			</template>
		</div>
		<div
			class="q-pa-md;wrap account_center"
			style="word-wrap: break-word; min-height: auto"
			v-else
		>
			<div class="boot_justify1">
				<q-img
					src="../../../../assets/wizard-web/wizard_failed.jpg"
					class="account_img"
				/>
			</div>
			<p class="get-text">{{ t('network_failed_title') }}</p>
			<p class="account_please_text please_text">
				{{ t('network_failed_content') }}
			</p>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
// import axios from 'axios';
import { useI18n } from 'vue-i18n';
import animationPage from '../animation.vue';

defineProps({
	wizardStatus: {
		type: String,
		default: '',
		required: false
	}
});

const pictureImg = require('../../../../assets/wizard-web/wizard_waikuang.png');
const certificateImg = require('../../../../assets/wizard-web/wizard_certificate.png');

const { t } = useI18n();

const state = ref<string | null>(null);
const request_https_certificate = ref(true);
</script>
