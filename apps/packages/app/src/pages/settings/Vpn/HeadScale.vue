<template>
	<page-title-component
		:show-back="true"
		:title="t('headScale_connection_status')"
	/>

	<common-inner-loading :model-value="innerLoading">
		<bt-scroll-area class="nav-height-scroll-area-conf">
			<head-scale-device-card
				:class="
					deviceStore.isMobile ? 'mobile-items-list q-mt-md' : 'q-list-class'
				"
				class="q-mb-lg"
				style="padding: 0"
				v-for="device in headScaleStore.devices"
				:key="device.id"
				:device="device"
			/>
		</bt-scroll-area>
	</common-inner-loading>
</template>

<script lang="ts" setup>
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import CommonInnerLoading from 'src/components/base/CommonInnerLoading.vue';
import HeadScaleDeviceCard from './HeadScaleDeviceCard.vue';
import { useHeadScaleStore } from 'src/stores/settings/headscale';
import { useDeviceStore } from 'src/stores/settings/device';
import { useUsersStore } from 'src/stores/settings/users';
import { useI18n } from 'vue-i18n';
import { ref, onMounted } from 'vue';

const usersStore = useUsersStore();
const headScaleStore = useHeadScaleStore();
const deviceStore = useDeviceStore();
const innerLoading = ref(false);
const { t } = useI18n();

onMounted(async () => {
	try {
		innerLoading.value = true;
		await headScaleStore.getDevices();
		await headScaleStore.getHeadScaleStatus();
	} catch (e) {
		console.log(e);
	} finally {
		innerLoading.value = false;
	}
	usersStore.get_accounts();
});
</script>

<style scoped lang="scss">
.personal-page-root {
	height: 100%;
}
</style>
