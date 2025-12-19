<template>
	<div class="column" style="padding: 0 32px 32px 32px">
		<display-item
			:title="i18n.t('recommendation.container')"
			:content="container.name || 'main'"
		/>
		<display-item
			:title="i18n.t('recommendation.image')"
			:content="container.image"
		/>
		<!-- <display-item title="Command" content="" /> -->
		<!-- <display-item title="Args" content="" /> -->
		<display-item :title="i18n.t('recommendation.env')" :env="container.env" />
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { Container } from 'src/stores/argo';
import DisplayItem from './DisplayItem.vue';
import { useI18n } from 'vue-i18n';

const i18n = useI18n();

const props = defineProps({
	container: {
		type: Object as PropType<Container> | undefined,
		required: true
	}
});

console.log(props.container);
</script>
