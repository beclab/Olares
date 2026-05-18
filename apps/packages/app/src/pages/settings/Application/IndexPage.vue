<template>
	<page-title-component :show-back="false" :title="t('Applications')" />

	<bt-list first class="q-mx-lg q-mb-xs">
		<q-input
			dense
			borderless
			class="q-px-lg"
			v-model="searchContent"
			:placeholder="t('Type an app name')"
		>
			<template v-slot:prepend>
				<q-icon name="sym_r_search" class="text-ink-1" size="24px" />
			</template>
		</q-input>
	</bt-list>

	<bt-scroll-area
		class="nav-height-scroll-area-conf"
		style="height: calc(100% - 120px)"
	>
		<bt-list first v-if="showApplications.length > 0">
			<template v-for="(application, index) in showApplications" :key="index">
				<application-item
					:icon="application.icon"
					:title="application.title || application.name"
					:hide-status="false"
					:cs-app="application.isClusterScoped"
					:app-name="application.name"
					:raw-app-name="application.rawAppName"
					:status="application.state"
					:width-separator="index !== showApplications.length - 1"
					:margin-top="index !== 0"
					@click="onItemClick(application)"
				/>
			</template>
		</bt-list>
		<app-menu-empty
			v-else
			:title="t('No applications found')"
			message=""
			image="settings/imgs/root/application.svg"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ApplicationItem from 'src/components/settings/application/ApplicationItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useApplicationStore } from 'src/stores/settings/application';
import { TerminusApp } from '@bytetrade/core';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

const { t } = useI18n();
const router = useRouter();
const applicationStore = useApplicationStore();

const searchContent = ref('');

const showApplications = computed(() => {
	if (!searchContent.value) {
		return applicationStore.installApplication;
	}

	const keyword = searchContent.value.toLowerCase();
	return applicationStore.installApplication.filter((item) => {
		return (
			item.name.toLowerCase().includes(keyword) ||
			(item.title && item.title.toLowerCase().includes(keyword))
		);
	});
});

const onItemClick = (application: TerminusApp) => {
	router.push(`application/info/${application.name}`);
};
</script>
