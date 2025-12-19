<template>
	<page-title-component :title="t('Others')" :show-back="true" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<error-message-tip
				:is-error="false"
				:width-separator="true"
				class="q-mt-sm"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.others.enableThrottling.name }}
							</div>
						</div>
					</template>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="videoStore.others.enableThrottling.value"
						@update:model-value="updateEnableThrottling"
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
									'When a transcode or remux gets far enough ahead from the current playback position, pause the process so it will consume fewer resources. This is most useful when watching without seeking often. Turn this off if you experience playback issues.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>
			<error-message-tip
				:is-error="false"
				:width-separator="true"
				class="q-mt-md"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.others.enableSegmentDeletion.name }}
							</div>
						</div>
					</template>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="videoStore.others.enableSegmentDeletion.value"
						@update:model-value="updateEnableSegmentDeletion"
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
									'Delete old segments after they have been downloaded by the client. This prevents having to store the entire transcoded file on disk. Turn this off if you experience playback issues.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>
			<error-message-tip
				:is-error="false"
				:width-separator="true"
				class="q-mt-md"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.others.throttleDelaySeconds.name }}
							</div>
						</div>
					</template>
					<bt-time-picker
						v-model="videoStore.others.throttleDelaySeconds.value"
						unit=""
						:max="1000"
						:input-disabled="true"
						@onUpdate="updateThrottleDelaySeconds"
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
									'Time in seconds after which the transcoder will be throttled. Must be large enough for the client to maintain a healthy buffer. Only works if throttling is enabled.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>

			<error-message-tip
				:is-error="false"
				:width-separator="true"
				class="q-mt-md"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.others.segmentKeepSeconds.name }}
							</div>
						</div>
					</template>
					<bt-time-picker
						v-model="videoStore.others.segmentKeepSeconds.value"
						unit=""
						:input-disabled="true"
						:max="1000"
						@onUpdate="updateSegmentKeepSeconds"
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

			<error-message-tip
				:is-error="false"
				:width-separator="false"
				class="q-mt-md"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.others.maxMuxingQueueSize.name }}
							</div>
						</div>
					</template>
					<bt-time-picker
						v-model="videoStore.others.maxMuxingQueueSize.value"
						unit=""
						:max="4096"
						:input-disabled="true"
						@onUpdate="updateMaxMuxingQueueSize"
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
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';

import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import { useVideoStore } from 'src/stores/settings/video';
import VideoService from 'src/services/video';
import BtTimePicker from 'src/components/settings/base/BtTimePicker.vue';
import { onMounted } from 'vue';

const videoStore = useVideoStore();

const { t } = useI18n();

const updateEnableThrottling = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableThrottling: value
	});
};

const updateEnableSegmentDeletion = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableSegmentDeletion: value
	});
};

const updateThrottleDelaySeconds = (value: string) => {
	VideoService.updateInitDataProps({
		ThrottleDelaySeconds: Number(value)
	});
};

const updateSegmentKeepSeconds = (value: string) => {
	VideoService.updateInitDataProps({
		SegmentKeepSeconds: Number(value)
	});
};

const updateMaxMuxingQueueSize = (value: string) => {
	VideoService.updateInitDataProps({
		MaxMuxingQueueSize: Number(value)
	});
};

onMounted(async () => {
	try {
		if (!VideoService.initData) {
			const config = await videoStore.getVideoConfig();
			VideoService.configInitData(config);
		}
	} catch (error) {
		console.log(error);
	}
});
</script>

<style scoped lang="scss"></style>
