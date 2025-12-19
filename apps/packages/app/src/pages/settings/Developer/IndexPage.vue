<template>
	<page-title-component :show-back="false" :title="t('system')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<bt-form-item
				v-if="adminStore.olaresd && adminStore.isAdmin"
				:title="t('Repository management')"
				@click="gotoPage('/developer/mirror')"
				:chevron-right="true"
			/>

			<bt-form-item
				v-if="adminStore.olaresd && adminStore.isAdmin"
				:title="t('Image management')"
				@click="gotoPage('/developer/images')"
				:chevron-right="true"
			/>

			<bt-form-item
				v-if="adminStore.olaresd && adminStore.isAdmin"
				:title="t('logs')"
				@click="gotoPage('/developer/log')"
				:chevron-right="true"
			/>

			<bt-form-item
				:title="t('System Environment Variables')"
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
import { useAdminStore } from 'src/stores/settings/admin';

const router = useRouter();
const adminStore = useAdminStore();
const { t } = useI18n();

function gotoPage(path: string) {
	router.push({ path });
}
</script>
