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

						<img
							v-if="sharedApp"
							class="application_cs_icon"
							:style="{ '--csSize': '12px' }"
							:src="getRequireImage('share_app.svg')"
						/>
					</div>
					<div class="column justify-start q-ml-sm">
						<div class="row justify-start items-center">
							<span
								class="text-ink-1"
								:class="{
									'text-body1': !deviceStore.isMobile,
									'text-subtitle3-m': deviceStore.isMobile
								}"
							>
								{{ title }}
							</span>

							<span
								v-if="cloneLabel"
								class="text-info bg-blue-alpha text-caption q-ml-sm clone-label"
							>
								{{ cloneLabel }}
							</span>
						</div>

						<application-status v-if="!hideStatus" :status="status" />
					</div>
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
import { getRequireImage } from 'src/utils/imageUtils';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	icon: {
		type: String,
		require: true
	},
	title: {
		type: String,
		require: true
	},
	sharedApp: {
		type: Boolean,
		require: true
	},
	clonedFrom: {
		type: String,
		require: ''
	},
	appName: {
		type: String,
		require: true
	},
	rawAppName: {
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
const { t } = useI18n();

const cloneLabel = computed(() => {
	if (!props.rawAppName || props.rawAppName === props.appName) {
		return '';
	}
	if (props.clonedFrom === 'template') {
		return t('From Template');
	}
	if (props.clonedFrom === 'app') {
		return t('From App');
	}
	return '';
});
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

		.application_cs_icon {
			background: transparent;
			width: var(--csSize);
			height: var(--csSize);
			position: absolute;
			filter: drop-shadow(0px 2px 4px rgba(0, 0, 0, 0.2));
			right: 0;
			bottom: 0;
			z-index: 3;
		}

		.clone-label {
			padding: 2px 4px;
			border-radius: 4px;
		}
	}
}
</style>
