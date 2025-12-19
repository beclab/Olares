<template>
	<!-- <div class="column"> -->
	<page-title-component
		:show-back="true"
		:title="t('device_of_vault_trustee')"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<template v-for="(device, index) in devices" :key="'device' + index">
			<trust-device-item
				:device="device"
				class="mobile-items-list"
				style="padding-top: 16px; padding-bottom: 16px"
			/>
		</template>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useAdminStore } from 'src/stores/settings/admin';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import TrustDeviceItem from 'src/components/settings/person/TrustDeviceItem.vue';
import { useI18n } from 'vue-i18n';

const adminStore = useAdminStore();
const { t } = useI18n();

const devices = ref([]);
onMounted(async () => {
	devices.value = await adminStore.get_vault_devices();
});
</script>
