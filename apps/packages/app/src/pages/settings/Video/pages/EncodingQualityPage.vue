<template>
	<page-title-component :title="t('Encoding Quality')" :show-back="true" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<adaptive-layout>
				<template v-slot:pc>
					<error-message-tip
						:is-error="false"
						:width-separator="true"
						class="q-mt-md"
					>
						<bt-form-item :width-separator="false" :min-item-height="40">
							<template v-slot:title>
								<div class="column">
									<div>
										{{ videoStore.encodingQuality.encoderPreset.name }}
									</div>
								</div>
							</template>
							<bt-select
								v-model="videoStore.encodingQuality.encoderPreset.value"
								:options="encoderPresetTypeOptions"
								@update:model-value="onUpdateEncoderPreset"
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
											'Pick a faster value to improve performance, or a slower value to improve quality.'
										)
									}}
								</div>
							</div>
						</template>
					</error-message-tip>
				</template>
				<template v-slot:mobile>
					<bt-form-item
						:width-separator="true"
						:chevronRight="true"
						@click="enterEncoderPresetSelect"
					>
						<template v-slot:title>
							<div class="column">
								<div>
									{{ videoStore.encodingQuality.encoderPreset.name }}
								</div>
							</div>
						</template>
					</bt-form-item>
				</template>
			</adaptive-layout>
			<bt-form-item :width-separator="true" :item-height="64">
				<template v-slot:title>
					<div class="column">
						<div>
							{{ videoStore.encodingQuality.h265Crf.name }}
						</div>
					</div>
				</template>
				<bt-time-picker
					v-model="videoStore.encodingQuality.h265Crf.value"
					unit=""
					:input-disabled="true"
					@onUpdate="update265Crf"
				/>
			</bt-form-item>
			<error-message-tip :is-error="false" :width-separator="false">
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.encodingQuality.h264Crf.name }}
							</div>
						</div>
					</template>
					<bt-time-picker
						v-model="videoStore.encodingQuality.h264Crf.value"
						unit=""
						:input-disabled="true"
						@onUpdate="update264Crf"
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
									"The 'Constant Rate Factor' (CRF) is the default quality setting for the x264 and x265 software encoders. You can set the values between 0 and 51, where lower values would result in better quality (at the expense of higher file sizes). Sane values are between 18 and 28. The default for x264 is 23, and for x265 is 28, so you can use this as a starting point. Hardware encoders do not use these settings."
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
import {
	encoderPresetTypeOptions,
	EncoderPresetType
} from 'src/services/abstractions/video/service';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import { useRouter } from 'vue-router';
import { onMounted } from 'vue';
import BtList from 'src/components/settings/base/BtList.vue';

const videoStore = useVideoStore();

const { t } = useI18n();
const router = useRouter();

const onUpdateEncoderPreset = (value: EncoderPresetType) => {
	VideoService.updateInitDataProps({
		EncoderPreset: value
	});
};
const update264Crf = (value: string) => {
	VideoService.updateInitDataProps({
		H264Crf: Number(value)
	});
};

const update265Crf = (value: string) => {
	VideoService.updateInitDataProps({
		H265Crf: Number(value)
	});
};
const enterEncoderPresetSelect = () => {
	router.push('/video/optionsSelect/encoderPreset');
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
