<template>
	<q-page class="column items-center justify-center">
		<img src="../../assets/desktop/landscape.svg" />
		<p class="title q-mt-lg">Olares</p>
		<p class="content">{{ t('landscape_message') }}</p>
	</q-page>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>
<style lang="scss" scoped>
.title {
	font-size: 28px;
	line-height: 40px;
	color: #111111;
}

.content {
	font-size: 16px;
	line-height: 22px;
	color: #111111;
}
</style>
