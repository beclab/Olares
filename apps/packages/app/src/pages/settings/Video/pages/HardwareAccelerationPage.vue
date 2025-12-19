<template>
	<page-title-component :title="t('Hardware Acceleration')" :show-back="true" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<adaptive-layout>
			<template v-slot:pc>
				<q-list class="q-list-class q-mt-md">
					<bt-form-item :width-separator="false">
						<template v-slot:title>
							<div class="column justify-start">
								<div class="row">
									<div class="text-subtitle2 text-ink-1 q-mt-md">
										{{ t('Hardware Acceleration') }}
									</div>
								</div>
								<div class="text-body3 text-ink-3 q-mt-xs q-mb-md">
									{{
										t(
											'Hardware acceleration requires additional configuration.'
										)
									}}
								</div>
							</div>
						</template>
						<bt-select
							v-model="videoStore.hardwareAcceleration.type"
							:options="hardwareAccelerationOptions"
							@update:model-value="onInput"
						/>
					</bt-form-item>
				</q-list>
			</template>
			<template v-slot:mobile>
				<div class="text-body3-m text-ink-3 q-mt-md q-mb-md">
					{{ t('Hardware acceleration requires additional configuration.') }}
				</div>
				<q-list class="mobile-items-list">
					<bt-form-item
						v-for="(option, index) in hardwareAccelerationOptions"
						:key="index"
						:width-separator="index + 1 < hardwareAccelerationOptions.length"
						:clickable="true"
						@click="onInput(option.value)"
					>
						<template v-slot:title>
							<div
								:class="
									videoStore.hardwareAcceleration.type == option.value
										? 'text-blue-default'
										: 'text-ink-1'
								"
							>
								{{ option.label }}
							</div>
						</template>
						<q-icon
							name="sym_r_check_circle"
							size="18px"
							color="blue-default"
							v-if="videoStore.hardwareAcceleration.type == option.value"
						/>
					</bt-form-item>
				</q-list>
			</template>
		</adaptive-layout>

		<!-- preferSystemNativeHwDecoder -->
		<div
			class="text-subtitle1 details-title"
			v-if="
				videoStore.hardwareAcceleration.decodingCodeList &&
				videoStore.hardwareAcceleration.decodingCodeList.length > 0
			"
		>
			{{ t('Enable hardware decoding') }}
		</div>
		<q-list
			v-if="
				videoStore.hardwareAcceleration.decodingCodeList &&
				videoStore.hardwareAcceleration.decodingCodeList.length > 0
			"
			:class="deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class'"
		>
			<bt-form-item
				v-for="(decoding, index) in videoStore.hardwareAcceleration
					.decodingCodeList"
				:key="index"
				:title="decoding.name"
				:width-separator="
					index + 1 < videoStore.hardwareAcceleration.decodingCodeList.length ||
					videoStore.hardwareAcceleration
						.support_enableDecodingColorDepth10Hevc ||
					videoStore.hardwareAcceleration.support_enableDecodingColorDepth10Vp9
				"
			>
				<bt-switch
					size="sm"
					truthy-track-color="light-blue-default"
					v-model="decoding.value"
					@update:model-value="updateItem(decoding)"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="
					videoStore.hardwareAcceleration.support_enableDecodingColorDepth10Hevc
				"
				:title="
					videoStore.hardwareAcceleration.enableDecodingColorDepth10Hevc.name
				"
				:width-separator="
					videoStore.hardwareAcceleration.support_enableDecodingColorDepth10Vp9
				"
			>
				<bt-switch
					size="sm"
					truthy-track-color="light-blue-default"
					v-model="
						videoStore.hardwareAcceleration.enableDecodingColorDepth10Hevc.value
					"
					@update:model-value="updateHevc"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="
					videoStore.hardwareAcceleration.support_enableDecodingColorDepth10Vp9
				"
				:title="
					videoStore.hardwareAcceleration.enableDecodingColorDepth10Vp9.name
				"
				:width-separator="false"
			>
				<bt-switch
					size="sm"
					truthy-track-color="light-blue-default"
					v-model="
						videoStore.hardwareAcceleration.enableDecodingColorDepth10Vp9.value
					"
					@update:model-value="updateVP9"
				/>
			</bt-form-item>
		</q-list>

		<q-list
			v-if="videoStore.hardwareAcceleration.support_enableEnhancedNvdecDecoder"
			:class="deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class'"
		>
			<error-message-tip :is-error="false" :width-separator="false">
				<bt-form-item
					:width-separator="false"
					:item-height="40"
					:title="
						videoStore.hardwareAcceleration.enableEnhancedNvdecDecoder.name
					"
				>
					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="
							videoStore.hardwareAcceleration.enableEnhancedNvdecDecoder.value
						"
						@update:model-value="updateEnableEnhancedNvdecDecoder"
					/>
				</bt-form-item>
				<template v-slot:reminder>
					<div class="row items-center q-mb-sm" style="width: 100%">
						<div
							class="text-overline-m text-ink-3 item-margin-left"
							style="max-width: 400px"
						>
							{{
								t(
									'Time in seconds for which segments should be kept after they are downloaded by the client. Only works if segment deletion is enabled.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>
		</q-list>

		<q-list
			v-if="videoStore.hardwareAcceleration.support_preferSystemNativeHwDecoder"
			:class="deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class'"
		>
			<error-message-tip :is-error="false" :width-separator="false">
				<bt-form-item
					:width-separator="false"
					:title="
						videoStore.hardwareAcceleration.preferSystemNativeHwDecoder.name
					"
				>
					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="
							videoStore.hardwareAcceleration.preferSystemNativeHwDecoder.value
						"
						@update:model-value="updatePreferSystemNativeHwDecoder"
					/>
				</bt-form-item>
			</error-message-tip>
		</q-list>
		<div
			class="text-subtitle1 details-title"
			v-if="
				videoStore.hardwareAcceleration.support_enableHardwareEncoding ||
				videoStore.hardwareAcceleration
					.support_enableIntelLowPowerH264HwEncoder ||
				videoStore.hardwareAcceleration.support_enableIntelLowPowerHevcHwEncoder
			"
		>
			{{ t('Hardware encoding options') }}
		</div>
		<q-list
			v-if="
				videoStore.hardwareAcceleration.support_enableHardwareEncoding ||
				videoStore.hardwareAcceleration
					.support_enableIntelLowPowerH264HwEncoder ||
				videoStore.hardwareAcceleration.support_enableIntelLowPowerHevcHwEncoder
			"
			:class="deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class'"
		>
			<bt-form-item
				v-if="videoStore.hardwareAcceleration.support_enableHardwareEncoding"
				:title="videoStore.hardwareAcceleration.enableHardwareEncoding.name"
				:width-separator="
					videoStore.hardwareAcceleration
						.support_enableIntelLowPowerH264HwEncoder ||
					videoStore.hardwareAcceleration
						.support_enableIntelLowPowerHevcHwEncoder
				"
			>
				<bt-switch
					size="sm"
					truthy-track-color="light-blue-default"
					v-model="videoStore.hardwareAcceleration.enableHardwareEncoding.value"
					@update:model-value="updateEnableHardwareEncoding"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="
					videoStore.hardwareAcceleration
						.support_enableIntelLowPowerH264HwEncoder
				"
				:title="
					videoStore.hardwareAcceleration.enableIntelLowPowerH264HwEncoder.name
				"
				:width-separator="
					videoStore.hardwareAcceleration
						.support_enableIntelLowPowerHevcHwEncoder
				"
			>
				<bt-switch
					size="sm"
					truthy-track-color="light-blue-default"
					v-model="
						videoStore.hardwareAcceleration.enableIntelLowPowerH264HwEncoder
							.value
					"
					@update:model-value="updateEnableIntelLowPowerH264HwEncoder"
				/>
			</bt-form-item>

			<error-message-tip
				v-if="
					videoStore.hardwareAcceleration
						.support_enableIntelLowPowerHevcHwEncoder
				"
				:is-error="false"
				:width-separator="false"
				class="q-mt-sm"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div>
							{{
								videoStore.hardwareAcceleration.enableIntelLowPowerHevcHwEncoder
									.name
							}}
						</div>
					</template>
					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="
							videoStore.hardwareAcceleration.enableIntelLowPowerHevcHwEncoder
								.value
						"
						@update:model-value="updateEnableIntelLowPowerHevcHwEncoder"
					/>
				</bt-form-item>
				<template v-slot:reminder>
					<div class="row items-center q-mb-sm" style="width: 100%">
						<div
							class="text-overline-m text-ink-3 item-margin-left"
							style="max-width: 400px"
						>
							{{
								t(
									'Low-power encoding saves unnecessary CPU-GPU synchronization. The i915 HuC firmware must be configured correctly before use in Linux.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>
		</q-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';

import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';

import { useDeviceStore } from 'src/stores/settings/device';
import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import { useVideoStore } from 'src/stores/settings/video';
import VideoService from 'src/services/video';
import {
	HardwareAccelerationType,
	HardwareDecodingCodeTypeInfo
} from 'src/services/abstractions/video/service';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';

const { t } = useI18n();
const deviceStore = useDeviceStore();

const videoStore = useVideoStore();

const hardwareAccelerationOptions = ref(
	VideoService.supportHardwareAccelerationTypeList
);

const onInput = (value: HardwareAccelerationType) => {
	VideoService.updateHardwareAccelerationType(value);
};
const updateItem = (value: HardwareDecodingCodeTypeInfo) => {
	if (value.value) {
		if (
			VideoService.initData.HardwareDecodingCodecs.find((e) => e == value.type)
		) {
			return;
		}
		VideoService.initData.HardwareDecodingCodecs.push(value.type);
		VideoService.updateInitDataProps({
			HardwareDecodingCodecs: VideoService.initData.HardwareDecodingCodecs
		});
	} else {
		VideoService.updateInitDataProps({
			HardwareDecodingCodecs:
				VideoService.initData.HardwareDecodingCodecs.filter(
					(a) => a != value.type
				)
		});
	}
};
const updateHevc = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableDecodingColorDepth10Hevc: value
	});
};
const updateVP9 = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableDecodingColorDepth10Vp9: value
	});
};
const updateEnableHardwareEncoding = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableHardwareEncoding: value
	});
};
const updateEnableIntelLowPowerH264HwEncoder = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableIntelLowPowerH264HwEncoder: value
	});
};
const updateEnableIntelLowPowerHevcHwEncoder = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableIntelLowPowerHevcHwEncoder: value
	});
};

const updateEnableEnhancedNvdecDecoder = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableEnhancedNvdecDecoder: value
	});
};
const updatePreferSystemNativeHwDecoder = (value: boolean) => {
	VideoService.updateInitDataProps({
		PreferSystemNativeHwDecoder: value
	});
};

onMounted(async () => {
	try {
		if (!VideoService.initData) {
			const config = await videoStore.getVideoConfig();
			VideoService.configInitData(config);
			console.log(config);
		}
	} catch (error) {
		console.log(error);
	}
});
</script>

<style scoped lang="scss">
.details-title {
	margin-top: 20px;
	margin-bottom: 11px;
	color: $ink-1;
}
</style>
