<template>
	<div class="row items-center justify-center text-ink-3 empty-application">
		<q-icon name="sym_r_block" size="24px" class="q-mr-xs" />
		<div class="text-body1">
			{{ t('No app binded') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>

<style scoped lang="scss">
.empty-application {
	height: 44px;
}
</style>
