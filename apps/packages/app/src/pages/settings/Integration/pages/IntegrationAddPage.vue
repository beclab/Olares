<template>
	<div class="integration-add-root">
		<page-title-component :title="t('add_account')" />
		<bt-scroll-area class="nav-height-scroll-area-conf">
			<IntegrationAddList
				@itemClick="accountCreate"
				:select-enable="false"
				:backup="isBackup"
			/>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { IntegrationAccountInfo } from 'src/services/abstractions/integration/integrationService';
import IntegrationAddList from '../components/IntegrationAddList.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import integrationService from 'src/services/integration/index';
import { useRoute, useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const $q = useQuasar();
const router = useRouter();
const route = useRoute();

const isBackup = Number(route.query.backup) == 1;

const accountCreate = async (item: IntegrationAccountInfo) => {
	integrationService.getInstanceByType(item.type)?.signIn({
		quasar: $q,
		router
	});
};
</script>

<style scoped lang="scss">
.integration-add-root {
	width: 100%;
	height: 100%;
}
</style>
