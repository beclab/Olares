<template>
	<div class="landscape-screen fixed-full column items-center justify-center">
		<img src="../../assets/desktop/landscape.svg" />
		<p class="title q-mt-lg">Olares</p>
		<p class="content">{{ t('landscape_message') }}</p>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>
<style lang="scss" scoped>
.landscape-screen {
	z-index: 9999;
	background: var(--Background-1, #ffffff);
}

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
