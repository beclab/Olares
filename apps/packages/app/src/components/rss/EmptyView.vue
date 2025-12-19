<template>
	<div
		class="empty-root column justify-center items-center"
		:style="{ minHeight: isTable ? '680px' : '' }"
	>
		<img
			class="empty-img"
			:src="
				getRequireImage(
					$q.dark.isActive ? 'content_empty_dark.svg' : 'content_empty.svg'
				)
			"
		/>
		<span class="empty-text text-subtitle1 text-ink-1 q-mt-lg text-center">{{
			content ? content : t('main.no_content_desc')
		}}</span>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { getRequireImage } from '../../utils/rss-utils';

const { t } = useI18n();
const $q = useQuasar();

defineProps({
	content: {
		type: String,
		require: false
	},
	isTable: {
		type: Boolean,
		default: false
	}
});
</script>

<style lang="scss" scoped>
.empty-root {
	width: 100%;
	height: 100%;

	.empty-img {
		width: 240px;
		height: 240px;
	}

	.empty-text {
		max-width: 600px;
	}
}
</style>
