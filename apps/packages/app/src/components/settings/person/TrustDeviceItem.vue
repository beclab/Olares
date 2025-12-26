<template>
	<bt-grid>
		<template v-slot:title>
			<div class="device-item-title row justify-between items-center">
				<div class="row justify-start items-center">
					<q-img
						class="device-title-image"
						:src="`settings/devices/${getIconName()}.png`"
					/>
					<div class="text-subtitle1 title-left-label">
						{{ device?.description }}
					</div>
				</div>
			</div>
		</template>
		<template v-slot:grid>
			<bt-grid-item
				:label="t('system')"
				:revert="true"
				:value="device?.osVersion"
			/>
			<bt-grid-item
				:label="t('manufacturer')"
				:revert="true"
				:value="device?.manufacturer"
			/>
			<bt-grid-item
				:label="t('locale')"
				:revert="true"
				:value="device?.locale"
			/>
		</template>
	</bt-grid>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { VaultActiveSessionInfo } from '@bytetrade/core';
import BtGrid from '../base/BtGrid.vue';
import BtGridItem from '../base/BtGridItem.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps({
	device: {
		type: Object as PropType<VaultActiveSessionInfo>
	}
});

const getIconName = () => {
	if (props.device) {
		switch (props.device.platform) {
			case 'android':
			case 'Android':
				switch (props.device.manufacturer) {
					case 'Xiaomi':
						return 'xiaomi';
					case 'HUAWEI':
						return 'huawei';
					case 'Honor':
						return 'honor';
					case 'Oppo':
						return 'oppo';
					case 'Vivo':
						return 'vivo';
					case 'Samsung':
						return 'samsung';
					case 'Google':
						return 'google';
					default:
						return 'android';
				}
			case 'ios':
			case 'iOS':
				return 'iPhone';
			case 'chrome extends':
				return 'chrome';
			case 'MacOS':
				return 'mac';
			case 'Windows':
				return 'windows';
		}
	}
};
</script>

<style scoped lang="scss">
.device-item-title {
	margin-bottom: 12px;

	.device-title-image {
		height: 32px;
		width: 32px;
	}

	.title-left-label {
		margin-left: 8px;
		color: $ink-1;
	}
}
</style>
