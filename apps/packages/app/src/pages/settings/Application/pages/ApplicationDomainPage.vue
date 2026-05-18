<template>
	<page-title-component :show-back="true" :title="$t('domain_setup')" />
	<bt-scroll-area v-if="isShared" class="nav-height-scroll-area-conf">
		<application-domain
			:shared="isShared"
			:name="applicationName"
			:entrance="entranceName"
		/>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ApplicationDomain from './ApplicationDomain.vue';
import { useRoute } from 'vue-router';
import { ref } from 'vue';

const router = useRoute();
const applicationName = ref(router.params.name as string);
const entranceName = ref(router.params.entrance as string);

const isShared = ref((router.params.shared as string) !== '');
</script>

<style scoped lang="scss"></style>
