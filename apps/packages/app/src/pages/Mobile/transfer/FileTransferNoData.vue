<template>
	<div class="column justify-center items-center">
		<img src="../../../assets/layout/nodata.svg" style="margin-top: 180px" />
		<span class="q-mb-md text-ink-1" style="margin-top: 32px">
			{{ t('there_is_nothing_for_now') }}
		</span>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>

<style scoped lang="scss"></style>
