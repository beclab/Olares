<template>
	<page-title-component :show-back="false" :title="t('application')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first v-if="applicationStore.installApplication.length > 0">
			<template
				v-for="(application, index) in applicationStore.installApplication"
				:key="index"
			>
				<application-item
					:icon="application.icon"
					:title="application.title || application.name"
					:status="application.state"
					:hide-status="!deviceStore.isMobile"
					:width-separator="
						index !== applicationStore.installApplication.length - 1
					"
					:margin-top="index !== 0"
					@click="onItemClick(application)"
				>
					<template v-slot>
						<application-status
							v-if="!deviceStore.isMobile"
							:status="application.state"
						/>
					</template>
				</application-item>
			</template>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useApplicationStore } from 'src/stores/settings/application';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ApplicationItem from 'src/components/settings/application/ApplicationItem.vue';
import ApplicationStatus from 'src/components/settings/application/ApplicationStatus.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { useRouter } from 'vue-router';
import { TerminusApp } from '@bytetrade/core';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const router = useRouter();
const deviceStore = useDeviceStore();
const applicationStore = useApplicationStore();

const onItemClick = (application: TerminusApp) => {
	router.push(`application/info/${application.name}`);
};
</script>
