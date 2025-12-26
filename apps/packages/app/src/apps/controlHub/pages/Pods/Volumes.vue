<template>
	<MyCard no-content-gap square flat :title="t('VOLUMES')">
		<VolumeContainer
			:volumes="volumes"
			:containers="containers"
			:isMultiProject="isMultiProject"
		></VolumeContainer>
	</MyCard>
</template>

<script setup lang="ts">
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import { t } from '@apps/control-hub/src/boot/i18n';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { computed, ref, watchEffect } from 'vue';
import { getWorkloadVolumes } from '@apps/control-panel-common/src/utils/workload';
import VolumeContainer from '@apps/control-panel-common/src/containers/VolumeContainer.vue';
const usePod = UsePod();
const volumes = ref([]);
const containers = computed(() => usePod?.data?.containers ?? []);
const isMultiProject = computed(() => usePod?.data?.isFedManaged);
watchEffect(async () => {
	volumes.value = await getWorkloadVolumes(usePod?.data ?? {});
});
</script>

<style></style>
