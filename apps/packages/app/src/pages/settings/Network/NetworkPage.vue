<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Network.toLowerCase()}`)"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature :menu-type="MENU_TYPE.Network" />
		<bt-list :label="t('Advanced network services')">
			<bt-form-item
				@click="gotoPage('/network/reverse_proxy')"
				:title="t('Reverse Proxy')"
				:description="
					t('Manage how external traffic is routed to your internal services.')
				"
				:chevron-right="true"
				:width-separator="adminStore.olaresd"
			/>
			<bt-form-item
				v-if="adminStore.olaresd"
				@click="gotoPage('/network/host')"
				:title="t('hosts')"
				:description="
					t('Customize domain name resolution for your Olares network.')
				"
				:chevron-right="true"
				:width-separator="false"
			/>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { MENU_TYPE } from 'src/constant';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useAdminStore } from 'src/stores/settings/admin';

const router = useRouter();
const adminStore = useAdminStore();
const { t } = useI18n();

function gotoPage(path: string) {
	router.push({ path });
}
</script>

<style scoped lang="scss"></style>
