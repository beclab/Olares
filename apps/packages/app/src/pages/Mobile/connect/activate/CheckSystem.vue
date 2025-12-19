<template>
	<div class="float-wrap">
		<div class="float" style="background: #fffdc1"></div>
		<div class="float" style="background: #fbffed"></div>
	</div>

	<div class="boot_justify">
		<animationPage
			:picture="system_waikuang_image"
			:certificate="system_image"
			:isAnimated="true"
		/>
	</div>
	<p class="wizard-content__title q-mt-lg">
		{{ t('initialize_olares_system') }}
	</p>
</template>

<script lang="ts" setup>
import animationPage from './animation.vue';
import system_image from '../../../../assets/wizard/system.png';
import system_waikuang_image from '../../../../assets/wizard/system_waikuang.png';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>
