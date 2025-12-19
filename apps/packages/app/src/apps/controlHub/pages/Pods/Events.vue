<template>
	<MyCard2 no-content-gap square flat :title="t('EVENT_PL')">
		<Event v-if="usePod.data" :detail="usePod.data"></Event>
	</MyCard2>
</template>

<script setup lang="ts">
import Event from '@apps/control-panel-common/src/containers/Event.vue';
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import MyCard2 from '@apps/control-panel-common/src/components/MyCard2.vue';
import { t } from '@apps/control-hub/src/boot/i18n';

const usePod = UsePod();
</script>
