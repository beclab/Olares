<template>
	<WebsiteDetailsList :websiteId="websiteId"></WebsiteDetailsList>
</template>

<script setup lang="ts">
import WebsiteDetailsList from './WebsiteDetailsList.vue';
import { useRoute } from 'vue-router';

const route = useRoute();

const { websiteId }: Record<string, any> = route.params;
const { name }: Record<string, any> = route.query;
</script>

<style lang="scss" scoped>
.data-page-wrapper {
	flex: 1;
}
</style>
