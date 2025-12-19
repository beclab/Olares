<template>
	<div class="application-item-root column justify-start">
		<q-item clickable class="application-item item">
			<q-item-section class="item-margin-left">
				<div class="row items-center">
					<div style="position: relative">
						<q-img
							class="application-logo"
							no-spinner
							:src="icon"
							style="border-radius: 10px"
						/>

						<application-status
							v-if="!hideStatus && deviceStore.isMobile"
							:status="status"
						/>
					</div>
					<div
						class="application-name"
						:class="{
							'text-body2': !deviceStore.isMobile,
							'text-subtitle3-m': deviceStore.isMobile
						}"
					>
						{{ title }}
					</div>

					<application-status
						v-if="!hideStatus && !deviceStore.isMobile"
						class="q-ml-md"
						:status="status"
					/>
				</div>
			</q-item-section>
			<q-item-section side class="item-margin-right">
				<div class="row justify-end items-center">
					<slot />
					<q-icon
						v-if="endIcon"
						name="sym_r_chevron_right"
						color="ink-1"
						size="20px"
					/>
				</div>
			</q-item-section>
		</q-item>
		<bt-separator v-if="widthSeparator" :offset="20" />
	</div>
</template>

<script lang="ts" setup>
import BtSeparator from '../base/BtSeparator.vue';
import ApplicationStatus from './ApplicationStatus.vue';
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	icon: {
		type: String,
		require: true
	},
	title: {
		type: String,
		require: true
	},
	hideStatus: {
		type: Boolean,
		default: true
	},
	status: {
		type: String,
		default: ''
	},
	endIcon: {
		type: Boolean,
		default: true
	},
	widthSeparator: {
		type: Boolean,
		default: true
	},
	marginTop: {
		type: Boolean,
		default: true
	}
});

const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.application-item-root {
	width: 100%;
	height: auto;

	.application-item {
		height: 56px;
		min-height: 56px;
		padding: 0;

		.application-logo {
			width: 32px;
			height: 32px;
			border-radius: 8px;
		}
		.application-name {
			color: $ink-1;
			margin-left: 8px;
		}
	}
}
</style>
