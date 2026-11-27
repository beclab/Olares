<template>
	<div v-show="modelValue" class="full-width full-height relative-position">
		<base-inner-loading
			:showing="modelValue"
			width="202px"
			height="136px"
			:src="srcPath"
			:absolute="true"
			:label="t('Some important process is in progress. Please wait...')"
		/>
	</div>
	<slot v-if="!modelValue" />
</template>

<script setup lang="ts">
import BaseInnerLoading from './BaseInnerLoading.vue';
import { useI18n } from 'vue-i18n';

defineProps({
	modelValue: {
		type: Boolean,
		default: false
	}
});

const { t } = useI18n();
</script>

<style scoped lang="scss"></style>
