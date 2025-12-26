<template>
	<AdaptiveLayout @click="goDeviceDetails">
		<template v-slot:pc>
			<div class="q-pa-lg row justify-between items-center cursor-pointer">
				<div class="row justify-start items-center">
					<q-img
						no-spinner
						class="device-title-image-pc"
						:src="getDeviceIconName(device)"
					/>
					<div class="column justify-start q-ml-sm">
						<div class="text-subtitle1 text-ink-1">
							{{ device?.description }}
						</div>

						<div class="text-body3 text-ink-3">
							{{
								device?.client_type
									? device?.client_type.charAt(0).toUpperCase() +
									  device?.client_type.slice(1)
									: ''
							}}
						</div>
					</div>
				</div>

				<q-icon size="24px" name="sym_r_keyboard_arrow_right" color="ink-3" />
			</div>

			<bt-separator v-if="!isLatest" :offset="20" />
		</template>
		<template v-slot:mobile>
			<div
				class="q-px-lg q-py-md row justify-between items-center cursor-pointer"
			>
				<div class="row justify-start items-center">
					<q-img
						no-spinner
						class="device-title-image-mobile"
						:src="getDeviceIconName(device)"
					/>
					<div class="column justify-start q-ml-sm">
						<div class="text-subtitle3-m text-ink-1">
							{{ device?.description }}
						</div>

						<div class="text-overline-m text-ink-3">
							{{
								device?.client_type
									? device?.client_type.charAt(0).toUpperCase() +
									  device?.client_type.slice(1)
									: ''
							}}
						</div>
					</div>
				</div>

				<q-icon size="24px" name="sym_r_keyboard_arrow_right" color="ink-3" />
			</div>

			<bt-separator v-if="!isLatest" :offset="20" />
		</template>
	</AdaptiveLayout>
</template>

<script lang="ts" setup>
import { PropType, computed } from 'vue';
import { TermiPassDeviceInfo } from '@bytetrade/core';
import { getDeviceIconName } from 'src/constant/constants';
import { useAdminStore } from 'src/stores/settings/admin';
import AdaptiveLayout from '../AdaptiveLayout.vue';
import BtSeparator from '../base/BtSeparator.vue';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();
const adminStore = useAdminStore();

const props = defineProps({
	device: {
		type: Object as PropType<TermiPassDeviceInfo>
	},
	isLatest: {
		type: Boolean,
		default: false,
		required: false
	}
});

const formattedDate = (
	datetime: string | number,
	format = 'YYYY-MM-DD HH:mm:ss'
) => {
	const originalDate = new Date(datetime);
	return date.formatDate(originalDate, format);
};

const formatMobileDetail = () => {
	const formatDat = props.device
		? formattedDate(
				`${props.device.lastSeenTime}`.length > 10
					? props.device.lastSeenTime
					: props.device.lastSeenTime * 1000,
				'MM-DD HH:mm:ss'
		  )
		: '';
	const location = props.device?.lastIpLocation || props.device?.lastIp || '';
	return [formatDat, location].filter((e) => e.length > 0).join('&emsp;');
};

const goDeviceDetails = () => {
	router.push({ path: `/device/${props.device?.id}` });
};

const isMe = computed(() => {
	return (
		adminStore.thisDevice &&
		adminStore.thisDevice.platform == props.device?.platform &&
		adminStore.thisDevice.manufacturer == props.device.manufacturer &&
		adminStore.thisDevice.client_type == props.device.client_type &&
		adminStore.thisDevice.description == props.device.description
	);
});
</script>

<style scoped lang="scss">
.device-title-image-pc {
	height: 40px;
	width: 40px;
}

.device-title-image-mobile {
	height: 45px;
	width: 45px;
}
</style>
