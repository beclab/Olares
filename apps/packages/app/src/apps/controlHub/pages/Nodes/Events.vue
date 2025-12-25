<template>
	<MyCard square flat :title="t('EVENT_PL')">
		<Event
			v-if="detail"
			:detail="detail"
			style="max-height: 800px; position: relative"
		>
		</Event>
	</MyCard>
</template>

<script setup lang="ts">
import { t } from '@apps/control-hub/src/boot/i18n';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import Event from '@apps/control-panel-common/src/containers/NodeEvent.vue';

interface Props {
	detail: Record<string, any>;
}
const props = withDefaults(defineProps<Props>(), {});
</script>

<style></style>
