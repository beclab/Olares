<template>
	<div class="coming-soon column justify-center items-center">
		<q-img class="coming-soon__img" :src="getRequireImage('coming_soon.svg')" />

		<div class="coming-soon__title">{{ t('base.coming_soon') }}</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from 'src/utils/rss-utils';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
</script>

<style lang="scss" scoped>
.coming-soon {
	width: 100%;
	height: 100%;

	&__img {
		width: 300px;
		height: 180px;
	}

	&__title {
		font-family: Roboto;
		font-size: 16px;
		font-weight: 500;
		line-height: 24px;
		margin-top: 20px;
		letter-spacing: 0em;
		text-align: center;
		color: var(--Grey-05, #adadad);
	}
}
</style>
