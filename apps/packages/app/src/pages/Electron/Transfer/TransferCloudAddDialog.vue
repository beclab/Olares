<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="dialogTitle"
		:noRouteDismiss="true"
		:ok="false"
		:cancel="false"
		size="medium"
		:okLoading="isLoading"
		:persistent="collectSiteStore.loading"
		@onSubmit="submit"
	>
		<div class="dialog-content dialog-padding q-py-md">
			<div v-if="cloudLineStepRef == CloudLineStep.Input">
				<div class="input-title text-body3 text-ink-3">
					{{ t('Please fill in the link of the file you want to download') }}
				</div>
				<div class="q-mt-xs common-border q-mb-md">
					<q-input
						v-model="linkUrl"
						borderless
						style="max-width: calc(100% - 30px)"
						dense
						type="textarea"
						:placeholder="t('Support HTTP and HTTPS connections')"
						input-class="custom-placeholder"
						input-style="resize: none;padding: 8px !important; max-height: 88px"
						:debounce="1000"
						@update:model-value="queryUrl"
						autogrow
					/>
					<SpinnerLoading
						class="loading"
						v-if="collectSiteStore.loading"
					></SpinnerLoading>
				</div>
				<div
					v-if="!validate.valid && linkUrl"
					class="text-body3 text-negative q-mt-xs"
				>
					{{ validate.reason }}
				</div>
				<CollectionContent :searchUrl="linkUrl" />
			</div>

			<!-- <div v-else-if="cloudLineStepRef == CloudLineStep.Analysising">
				<TransferAnalysis
					:title="t('Parsing download link')"
					:detail="linkUrl"
				/>
				<div class="line-process-bg row items-center">

					<div class="progress"></div>
				</div>

				<div class="slider-container">
					<div class="slider"></div>
				</div>
			</div>

			<div v-else-if="cloudLineStepRef == CloudLineStep.AnalysisError">
				<TransferAnalysis
					:title="
						needCookie
							? t('download.need_cookie_to_download')
							: t('Download link parsing failed')
					"
					:detail="linkUrl"
				/>
			</div>

			<div v-else-if="cloudLineStepRef == CloudLineStep.AnalysisSuccess">
				<TransferAnalysis :detail="linkUrl">
					<template v-slot:title>
						<q-input
							class="prompt-input text-body3"
							v-model="fileName"
							borderless
							no-error-icon
							input-class="text-ink-2 text-body3"
							dense
							placeholder=""
						/>
					</template>
				</TransferAnalysis>

				<div
					class="text-overline text-negative q-mt-md"
					v-if="(!fileInfoRef || !fileInfoRef.file) && !fileName"
				>
					{{ t('dialog.unable_to_retrieve_the_file_name') }}
				</div>

				<div class="text-overline text-negative q-mt-md" v-if="cookieRecommend">
					{{ t('download.recommend_cookie_to_download') }}
				</div>

				<div class="input-title text-body3 text-ink-3 q-mt-lg">
					{{ t('Cloud transfer to') }}
				</div>

				<TransfetSelectTo
					class="q-mt-xs"
					@setSelectPath="setSelectPath"
					:origins="originsRef"
				/>
			</div> -->
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, ref } from 'vue';
import { i18n } from '../../../boot/i18n';
import { useI18n } from 'vue-i18n';
import TransferClient from '../../../services/transfer';
import { CloudFileInfo } from '../../../services/abstractions/transfer/interface';
import { FilePath, useFilesStore } from '../../../stores/files';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { COOKIE_LEVEL, DRIVER_FILE_PREFIX } from '../../../utils/rss-types';
import { DriveType } from '../../../utils/interface/files';
import CollectionContent from './TranserCollectionContent.vue';
import {
	UrlValidationResult,
	validateUrlWithReasonAsync
} from '../../../utils/url2';
import { useCollectSiteStore } from '../../../stores/collect-site';
import { useTerminusStore } from '../../../stores/terminus';
import { useConfigStore } from '../../../stores/rss-config';
import { useBrowserCookieStore } from '../../../stores/settings/browserCookie';
import SpinnerLoading from 'src/components/common/SpinnerLoading.vue';

enum CloudLineStep {
	Input = 1,
	Analysising = 2,
	AnalysisError = 3,
	AnalysisSuccess = 4
}

defineProps({
	confirmBtnTitle: {
		type: String,
		default: i18n.global.t('confirm'),
		required: false
	},
	cancelBtnTitle: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	}
});

const filesStore = useFilesStore();

const CustomRef = ref();

const { t } = useI18n();

const linkUrl = ref('');

const cloudLineStepRef = ref(CloudLineStep.Input);

const fileInfoRef = ref<CloudFileInfo | undefined>(undefined);

const fileSavePathRef = ref<FilePath | undefined>(filesStore.currentPath[1]);

const originsRef = ref([DriveType.Drive]);

const fileName = ref('');

const isLoading = ref(false);

const validateDefault = { valid: false };

const validate = ref<UrlValidationResult>({ ...validateDefault });

const collectSiteStore = useCollectSiteStore();

const browserCookieStore = useBrowserCookieStore();

const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
};

const dialogTitle = computed(() => {
	if (cloudLineStepRef.value == CloudLineStep.Input) {
		return t('Create a new offline connection task');
	}
	return t('Offline link task analysis');
});

const validReset = () => {
	validate.value = { ...validateDefault };
};

const queryUrl = async () => {
	browserCookieStore.current_tab = undefined;
	validate.value = await validateUrlWithReasonAsync(linkUrl.value);
	if (!validate.value.valid || !linkUrl.value) {
		collectSiteStore.reset();
		return;
	}
	validReset();
	collectSiteStore.search(linkUrl.value);
	getCookie(linkUrl.value);
};

const submit = async () => {
	if (cloudLineStepRef.value == CloudLineStep.Input) {
		if (linkUrl.value.length == 0) {
			notifyFailed(t('Please enter the download link'));
			return;
		}
		if (!TransferClient.client.clouder) {
			return;
		}
		cloudLineStepRef.value = CloudLineStep.Analysising;
		const result = await TransferClient.client.clouder.queryUrl(linkUrl.value);
		if (result) {
			fileName.value = result.file;
			fileInfoRef.value = result;
			if (needCookie.value) {
				cloudLineStepRef.value = CloudLineStep.AnalysisError;
				return;
			}
			cloudLineStepRef.value = CloudLineStep.AnalysisSuccess;
		} else {
			fileInfoRef.value = undefined;
			cloudLineStepRef.value = CloudLineStep.AnalysisError;
		}
	}
	if (cloudLineStepRef.value == CloudLineStep.AnalysisError) {
		cloudLineStepRef.value = CloudLineStep.Input;
		return;
	}
	if (cloudLineStepRef.value == CloudLineStep.AnalysisSuccess) {
		if (
			!fileName.value ||
			!TransferClient.client.clouder ||
			!fileSavePathRef.value
		) {
			return;
		}

		const formatPath = fileSavePathRef.value.path.startsWith(DRIVER_FILE_PREFIX)
			? fileSavePathRef.value.path.substring(DRIVER_FILE_PREFIX.length)
			: fileSavePathRef.value.path;

		isLoading.value = true;

		const result = await TransferClient.client.clouder.downloadFile(
			fileName.value,
			fileInfoRef.value.download_url,
			TransferClient.client.clouder.getQueryId(),
			formatPath,
			fileInfoRef.value.file_type
		);
		console.log('result ===>', result);

		if (result) {
			CustomRef.value.onDialogOK();
		} else {
			notifyFailed('add download failed');
		}

		isLoading.value = false;
		return;
	}
};

const getCookie = (urlTarget: string) => {
	const terminusStore = useTerminusStore();
	const olaresId = terminusStore.olaresId.split('@')[0];
	const rssStore = useConfigStore();

	const url =
		process.env.NODE_ENV === 'development'
			? ''
			: rssStore.getModuleSever('settings');

	browserCookieStore.init(
		{
			url: urlTarget
		},
		olaresId,
		url
	);
};

const cookieRecommend = computed(() => {
	return (
		fileInfoRef.value &&
		fileInfoRef.value.cookie_require === COOKIE_LEVEL.RECOMMEND &&
		!fileInfoRef.value.cookie_exist
	);
});

const needCookie = computed(() => {
	return (
		fileInfoRef.value &&
		fileInfoRef.value.cookie_require === COOKIE_LEVEL.REQUIRED &&
		!fileInfoRef.value.cookie_exist
	);
});

onBeforeUnmount(() => {
	collectSiteStore.reset();
});
</script>

<style lang="scss" scoped>
.dialog-padding {
	padding-left: 0px;
	padding-right: 0px;
}

.dialog-content {
	.line-process-bg {
		border: 1px solid $separator;
		height: 16px;
		width: 100%;
		border-radius: 10px;
		overflow: hidden;
		margin-top: 20px;
		margin-bottom: 20px;
		padding-left: 0px;
		position: relative;
	}

	.progress {
		width: 48px;
		height: 8px;
		border-radius: 4px;
		background: $light-blue-default;
		position: absolute;
		top: 3px;
		left: 0px;
		animation: slide 1s linear infinite;
	}

	@keyframes slide {
		0% {
			transform: translateX(0);
		}
		100% {
			transform: translateX(360px);
		}
	}
}

.prompt-input {
	// padding-top: 7px;
	// width: 320px;
	padding-left: 7px;
	padding-right: 7px;
	height: 32px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;

	::v-deep(.q-field__inner) {
		height: 32px;
	}
}

.common-border {
	border: 1px solid $separator;
	border-radius: 8px;
	// height: 92px;
	// overflow: hidden;
	position: relative;

	.loading {
		position: absolute;
		right: 10px;
		top: 10px;
	}
	// padding-top: 8px;
	// padding-bottom: 8px;
}
// ::v-deep(.q-field--dense .q-field__control) {
// 	// height: 32px;
// }

// ::v-deep(.q-field--dense .q-field__marginal) {
// 	// height: 32px;
// }
</style>
