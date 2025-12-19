<template>
	<page-title-component
		:show-back="true"
		:title="t('headScale_connection_status')"
	/>

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div
			:class="
				deviceStore.isMobile ? 'mobile-items-list q-mt-md' : 'q-list-class'
			"
			style="padding: 0"
			v-for="device in headScaleStore.devices"
			:key="device.id"
		>
			<head-scale-device-card :device="device" />
		</div>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import PageTitleComponent from '../../../components/settings/PageTitleComponent.vue';
import HeadScaleDeviceCard from './HeadScaleDeviceCard.vue';
import { useHeadScaleStore } from '../../../stores/settings/headscale';
import { useDeviceStore } from '../../../stores/settings/device';
import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
import { onMounted } from 'vue';

const accountStore = useUserStore();
const headScaleStore = useHeadScaleStore();
const deviceStore = useDeviceStore();
const { t } = useI18n();

onMounted(async () => {
	try {
		await headScaleStore.getDevices();
		await headScaleStore.getHeadScaleStatus();
	} catch (e) {
		console.log(e);
	}
	accountStore.get_accounts();
});
</script>

<style scoped lang="scss">
.personal-page-root {
	height: 100%;
}
</style>
