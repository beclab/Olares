<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Network.toLowerCase()}`)"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature :menu-type="MENU_TYPE.Network" />
		<bt-list :label="t('Advanced network services')">
			<bt-form-item
				v-if="meStore.isAdmin"
				@click="gotoPage('/network/reverse_proxy')"
				:title="t('Reverse proxy')"
				:description="
					t('Manage how external traffic is routed to your internal services.')
				"
				:chevron-right="true"
				:width-separator="meStore.isAdmin && terminusStore.olaresd"
			/>
			<bt-form-item
				v-if="meStore.isAdmin && terminusStore.olaresd"
				@click="gotoPage('/network/host')"
				:title="t('hosts')"
				:description="
					t('Customize domain name resolution for your Olares network.')
				"
				:chevron-right="true"
				:width-separator="true"
			/>
			<bt-form-item
				v-if="terminusStore.olaresd"
				@click="gotoPage('/network/overlay_gateway')"
				:title="t('Overlay Gateway')"
				:description="
					t(
						'Give supported apps a dedicated local IP for screen mirroring, DLNA, and device discovery'
					)
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
import { useTerminusStore } from 'src/stores/terminus';
import { useMeStore } from 'src/stores/settings/me';
const router = useRouter();
const meStore = useMeStore();
const terminusStore = useTerminusStore();
const { t } = useI18n();

function gotoPage(path: string) {
	router.push({ path });
}
</script>

<style scoped lang="scss"></style>
