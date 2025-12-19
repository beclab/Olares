<template>
	<page-title-component :title="t('Audio Transcoding')" :show-back="true" />
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
								{{ videoStore.audioTranscoding.enableAudioVbr.name }}
							</div>
						</div>
					</template>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="videoStore.audioTranscoding.enableAudioVbr.value"
						@update:model-value="updateEnableAudioVbr"
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
									'Variable bitrate offers better quality to average bitrate ratio, but in some rare cases may cause buffering and compatibility issues.'
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
								{{ videoStore.audioTranscoding.downMixAudioBoost.name }}
							</div>
						</div>
					</template>
					<bt-time-picker
						v-model="videoStore.audioTranscoding.downMixAudioBoost.value"
						unit=""
						:input-disabled="true"
						@onUpdate="updateDownMixAudioBoost"
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
									'Boost audio when downmixing. A value of one will preserve the original volume.'
								)
							}}
						</div>
					</div>
				</template>
			</error-message-tip>

			<adaptive-layout>
				<template v-slot:pc>
					<error-message-tip
						:is-error="false"
						:width-separator="false"
						class="q-mt-md"
					>
						<bt-form-item :width-separator="false" :min-item-height="40">
							<template v-slot:title>
								<div class="column">
									<div>
										{{
											videoStore.audioTranscoding.downMixStereoAlgorithm.name
										}}
									</div>
								</div>
							</template>
							<bt-select
								v-model="
									videoStore.audioTranscoding.downMixStereoAlgorithm.value
								"
								:options="downMixStereoAlgorithmTypeOptions"
								@update:model-value="onUpdateDownMixStereoAlgorithmType"
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
											'Algorithm used to downmix multi-channel audio to stereo.'
										)
									}}
								</div>
							</div>
						</template>
					</error-message-tip>
				</template>
				<template v-slot:mobile>
					<bt-form-item
						:width-separator="false"
						:chevronRight="true"
						@click="enterDownMixStereoAlgorithmSelect"
					>
						<template v-slot:title>
							<div class="column">
								<div>
									{{ videoStore.audioTranscoding.downMixStereoAlgorithm.name }}
								</div>
							</div>
						</template>
					</bt-form-item>
				</template>
			</adaptive-layout>
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
import {
	downMixStereoAlgorithmTypeOptions,
	DownMixStereoAlgorithmType
} from 'src/services/abstractions/video/service';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import { useRouter } from 'vue-router';
import { onMounted } from 'vue';

const videoStore = useVideoStore();

const { t } = useI18n();
const router = useRouter();

const updateEnableAudioVbr = (value: boolean) => {
	VideoService.updateInitDataProps({
		EnableAudioVbr: value
	});
};
const onUpdateDownMixStereoAlgorithmType = (
	value: DownMixStereoAlgorithmType
) => {
	VideoService.updateInitDataProps({
		DownMixStereoAlgorithm: value
	});
};
const updateDownMixAudioBoost = (value: any) => {
	VideoService.updateInitDataProps({
		DownMixAudioBoost: Number(value)
	});
};

const enterDownMixStereoAlgorithmSelect = () => {
	router.push('/video/optionsSelect/downMixStereoAlgorithm');
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

<style scoped lang="scss"></style>
