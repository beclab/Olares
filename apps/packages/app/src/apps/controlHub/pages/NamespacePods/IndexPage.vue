<template>
	<Pod :filter="{ namespace }"></Pod>
</template>

<script setup lang="ts">
import Pod from '../Pods/IndexPage.vue';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

const namespace = computed(() => {
	return route.params.namespace as string;
});
</script>

<style></style>
