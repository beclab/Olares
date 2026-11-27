<template>
	<page-title-component :show-back="false" :title="t('home_menus.advanced')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<bt-form-item
				v-if="terminusStore.olaresd && meStore.isAdmin"
				:title="t('Repository management')"
				:description="t('advanced_repository_tooltip')"
				@click="gotoPage('/developer/mirror')"
				:chevron-right="true"
			/>

			<bt-form-item
				v-if="terminusStore.olaresd && meStore.isAdmin"
				:title="t('Image management')"
				:description="t('advanced_image_tooltip')"
				@click="gotoPage('/developer/images')"
				:chevron-right="true"
			/>

			<bt-form-item
				v-if="terminusStore.olaresd && meStore.isOwner"
				:title="t('advanced_logs_label')"
				:description="t('advanced_logs_tooltip')"
				@click="gotoPage('/developer/log')"
				:chevron-right="true"
			/>

			<bt-form-item
				:title="t('System environment variables')"
				:description="t('advanced_env_tooltip')"
				@click="gotoPage('/developer/environment')"
				:chevron-right="true"
				:width-separator="false"
			/>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useMeStore } from 'src/stores/settings/me';
import { useTerminusStore } from 'src/stores/terminus';

const router = useRouter();
const meStore = useMeStore();
const terminusStore = useTerminusStore();
const { t } = useI18n();

function gotoPage(path: string) {
	router.push({ path });
}
</script>
