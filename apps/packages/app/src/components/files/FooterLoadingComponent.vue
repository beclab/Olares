<template>
	<div class="row justify-center items-center loading-root">
		<q-spinner-dots color="blue" v-if="hasData" />
		<div class="no-data text-minor-color" v-else>
			{{ t('files.no_more_data') }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineProps({
	hasData: Boolean
});
</script>

<style lang="scss" scoped>
.loading-root {
	height: 50px;
	width: 100%;

	.no-data {
		font-family: 'Roboto';
		font-style: normal;
		font-weight: 400;
		font-size: 12px;
		line-height: 14px;
		margin-top: 4px;
	}
}
</style>
