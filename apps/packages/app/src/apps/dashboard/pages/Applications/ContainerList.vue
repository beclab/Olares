<template>
	<FullPageWithBack :title="title">
		<Containers hide-terminal hide-log bordered></Containers>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import Containers from '@apps/control-panel-common/src/containers/ContainersList/Containers.vue';
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack.vue';

const route = useRoute();
const { namespace, name }: Record<string, any> = route.params;
const title = computed(() => `${namespace}/${name}`);
</script>

<style></style>
