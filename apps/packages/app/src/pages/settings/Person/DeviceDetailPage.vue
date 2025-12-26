<template>
	<page-title-component :show-back="true" :title="device?.description" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="avatar-layout column items-center justify-center">
			<q-img
				class="device-title-image"
				no-spinner
				:src="getDeviceIconName(device)"
			/>
			<div class="text-h4 text-ink-1 q-mt-md">
				{{ device?.description }}
			</div>
			<div class="text-ink-3 text-body1 q-mt-sm">
				{{
					device?.client_type
						? device?.client_type.charAt(0).toUpperCase() +
						  device?.client_type.slice(1)
						: ''
				}}
			</div>
		</div>
		<!-- {{ device?.id }} -->

		<BtList>
			<!--			todo 不确定是不是这个值 -->
			<bt-form-item :title="t('Software type')" :data="device?.platform" />
			<!--			todo 没有 临时用lastSeenTime代替 但是下面有 重复了 -->
			<bt-form-item
				:title="t('Last active time')"
				:data="date.formatDate(device?.lastSeenTime, 'YYYY-MM-DD HH:mm:ss')"
				:width-separator="false"
			/>
			<!--			todo 没有 -->
			<!-- <bt-form-item :title="t('Location')" data="" :width-separator="false" /> -->
		</BtList>

		<BtList class="q-mb-lg">
			<!--			todo 没有 -->
			<!-- <bt-form-item :title="t('Status')" data="" /> -->
			<!--			todo 没有 -->
			<!-- <bt-form-item :title="t('Connection Type')" data="" /> -->
			<bt-form-item
				:title="t('Created At')"
				:data="date.formatDate(device?.createTime, 'YYYY-MM-DD HH:mm:ss')"
			/>
			<bt-form-item
				:title="t('Last Seen')"
				:data="date.formatDate(device?.lastSeenTime, 'YYYY-MM-DD HH:mm:ss')"
			/>
			<!--			todo 接口没有值 但是core中有字段 -->
			<!-- <bt-form-item :title="t('Last IP')" :data="device?.lastIpLocation" /> -->
			<bt-form-item
				:title="t('Tailscale')"
				:data="device?.tailScale_id ? t('Enabled') : t('Deactivated')"
			/>
			<bt-form-item v-if="sso" :title="t('Token Expires')" :data="sso.exp" />
			<bt-form-item
				v-if="sso"
				:title="t('Token Second Factor')"
				:data="sso?.isMfa"
				:width-separator="false"
			/>
		</BtList>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { getDeviceIconName } from 'src/constant/constants';
import { useAdminStore } from 'src/stores/settings/admin';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { computed } from 'vue';
import { date } from 'quasar';
import { Encoder } from '@bytetrade/core';

const route = useRoute();
const { t } = useI18n();
const id = route.params.deviceId;
const adminStore = useAdminStore();

const device = computed(() => {
	return adminStore.devices.find((device) => device.id === id);
});

const sso = computed(() => {
	const device = adminStore.devices.find((device) => device.id === id);

	if (
		!device ||
		!device.sso ||
		device.sso.length == 0 ||
		device.sso.split('.').length != 3
	) {
		return undefined;
	}

	const sso = JSON.parse(
		Encoder.bytesToString(Encoder.base64UrlToBytes(device.sso.split('.')[1]))
	);

	return {
		exp: date.formatDate(sso.exp * 1000, 'YYYY-MM-DD HH:mm:ss'),
		isMfa: sso.mfa == 1 ? 'Yes' : 'No'
	};
});
</script>

<style scoped lang="scss">
.avatar-layout {
	margin-top: 12px;
	margin-bottom: 20px;

	.device-title-image {
		height: 88px;
		width: 88px;
	}
}
</style>
