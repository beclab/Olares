<template>
	<div class="item">
		<q-icon name="sym_r_feed" size="16px" />
		<a class="q-ml-sm text-subtitle1" target="_blank" :href="data.link">{{
			t(`home_doc_${data.id}`)
		}}</a>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	}
});
</script>

<style lang="scss" scoped>
.item {
	width: 100%;
	color: $teal-pressed;
	margin-top: 12px;
	text-overflow: ellipsis;
	white-space: nowrap;
	overflow: hidden;
	a {
		color: $teal-pressed;
		text-decoration: none;
	}
}
</style>
