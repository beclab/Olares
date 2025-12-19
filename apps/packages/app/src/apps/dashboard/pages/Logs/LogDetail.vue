<template>
	<RouteBackBar :title="container" :sub-title="name"></RouteBackBar>
	<Logs theme="light" style="height: calc(100vh - 52px)"></Logs>
</template>

<script setup lang="ts">
import RouteBackBar from '@apps/dashboard/src/components/RouteBackBar.vue';
import Logs from '@apps/control-panel-common/src/containers/Logs.vue';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const { namespace, name, container }: any = route.params;
</script>

<style></style>
