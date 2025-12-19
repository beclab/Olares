<template v-if="data">
	<my-card flat :title="t('LABEL_PL')">
		<Labels :data="data.labels"></Labels>
	</my-card>
	<my-card no-content-gap flat :title="t('ANNOTATION_PL')">
		<DataDetail :data="data.annotations"></DataDetail>
	</my-card>
</template>

<script setup lang="ts">
import { MetadataInNodes } from '../types/network';
import MyCard from '../components/MyCard2.vue';
import Labels from './Labels.vue';
import DataDetail from './DataDetail.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	data: MetadataInNodes;
}

const props = withDefaults(defineProps<Props>(), {});
</script>

<style></style>
