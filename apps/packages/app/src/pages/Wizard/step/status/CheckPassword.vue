<template>
	<div class="row justify-evenly account_box" style="min-height: auto">
		<div class="q-pa-md wrap account_center" style="word-wrap: break-word">
			<div class="boot_justify1">
				<img :src="wizard_succeed" class="account_img" />
			</div>
			<p class="get-text">{{ t('reset_password_title') }}</p>
			<p class="account_please_text please_text">
				{{ t('reset_password_content') }}
			</p>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineProps({
	wizard_succeed: {
		type: String,
		default: '',
		required: false
	}
});
</script>
