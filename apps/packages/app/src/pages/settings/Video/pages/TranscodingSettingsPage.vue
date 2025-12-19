<template>
	<page-title-component :title="t('Transcoding Settings')" :show-back="true" />
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
										{{
											videoStore.transcodingSettings.encodingThreadCount.name
										}}
									</div>
								</div>
							</template>
							<bt-select
								v-model="
									videoStore.transcodingSettings.encodingThreadCount.value
								"
								:options="encodingThreadCountSelectOption"
								@update:model-value="updateEncodingThreadCount"
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
											'Select the maximum number of threads to use when transcoding. Reducing the thread count will lower CPU usage but may not convert fast enough for a smooth playback experience.'
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
						:chevron-right="true"
						:item-height="56"
						@click="enterEncodingThreadCountSelect()"
					>
						<template v-slot:title>
							<div class="column">
								<div>
									{{ videoStore.transcodingSettings.encodingThreadCount.name }}
								</div>
							</div>
						</template>
					</bt-form-item>
				</template>
			</adaptive-layout>

			<error-message-tip
				:is-error="false"
				:width-separator="false"
				class="q-mt-md"
			>
				<bt-form-item :width-separator="false" :min-item-height="40">
					<template v-slot:title>
						<div class="column">
							<div>
								{{ videoStore.transcodingSettings.cranscodingTempPath.name }}
							</div>
						</div>
					</template>
					<div class="row items-center justify-end">
						<div class="text-body1 text-ink-1 q-mr-sm">
							{{ videoStore.transcodingSettings.cranscodingTempPath.value }}
						</div>
						<q-icon
							@click="editFolder"
							name="sym_r_edit_square"
							size="20px"
							color="ink-1"
						/>
					</div>
				</bt-form-item>
				<template v-slot:reminder>
					<div class="row items-center q-mb-sm" style="width: 100%">
						<div
							class="text-overline-m text-ink-3 item-margin-left"
							style="max-width: 400px"
						>
							{{
								t(
									'Specify a custom path for the transcode files served to clients. Leave blank to use the server default.'
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

import EditTranscodePathDialog from '../dialogs/EditTranscodePathDialog.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import { useVideoStore } from 'src/stores/settings/video';
import VideoService from 'src/services/video';
import {
	encodingThreadCountSelectCommonOption,
	encodingThreadCountSelectMax
} from 'src/services/abstractions/video/service';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import { useRouter } from 'vue-router';
import BtList from 'src/components/settings/base/BtList.vue';
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const videoStore = useVideoStore();
const $q = useQuasar();
const router = useRouter();

const encodingThreadCountSelectOption = ref(
	encodingThreadCountSelectCommonOption
);

const addNumberSelection = () => {
	for (let index = encodingThreadCountSelectMax; index > 0; index--) {
		encodingThreadCountSelectOption.value.splice(1, 0, {
			label: `${index}`,
			value: index,
			enable: true
		});
	}
};
addNumberSelection();

const editFolder = () => {
	$q.dialog({
		component: EditTranscodePathDialog,
		componentProps: {
			folder: videoStore.transcodingSettings.cranscodingTempPath.value
		}
	}).onOk((folder: string) => {
		videoStore.transcodingSettings.cranscodingTempPath.value = folder;
		VideoService.updateInitDataProps({
			TranscodingTempPath: folder
		});
	});
};

const updateEncodingThreadCount = (value: number) => {
	VideoService.updateInitDataProps({
		EncodingThreadCount: value
	});
};
const enterEncodingThreadCountSelect = () => {
	router.push('/video/optionsSelect/transcodingThreadCount');
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
