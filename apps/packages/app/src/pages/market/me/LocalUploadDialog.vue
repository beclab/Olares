<template>
	<bt-custom-dialog
		ref="customDialogRef"
		:title="t('Upload Chart')"
		:ok="okText"
		:cancel="cancelText"
		size="small"
		:persistent="true"
		@onSubmit="onOKClick"
		@onCancel="onCancelClick"
	>
		<div class="row justify-start items-center full-width">
			<q-img
				class="upload-image q-mr-md"
				:src="getRequireImage('setting/chart.svg')"
			/>

			<div
				v-if="chartStatus === UploadStatus.PENDING"
				class="column upload-tips"
			>
				<div class="text-subtitle2 text-ink-1">
					{{ t('Uploading chart', { chart: file?.name || '' }) }}
				</div>
				<q-linear-progress
					rounded
					size="4px"
					class="q-mt-lg"
					:value="progress"
					color="positive"
					track-color="separator"
				/>
			</div>

			<div
				v-else-if="chartStatus === UploadStatus.PROCESSING"
				class="column upload-tips"
			>
				<div class="text-subtitle2 text-ink-1">
					{{ t('Adding Chart to Local Source', { chart: file?.name || '' }) }}
				</div>
				<div class="row justify-start items-center q-mt-md">
					<bt-loading size="24px" :loading="true" />
					<div class="text-subtitle3 text-ink-2 q-ml-sm">
						{{ t('Processing app data') }}
					</div>
				</div>
			</div>

			<div
				v-else-if="chartStatus === UploadStatus.SUCCESS"
				class="column upload-tips"
			>
				<div class="text-subtitle2 text-ink-1">
					{{
						t(
							'App was successfully added to your local source. Do you want to install it now?',
							{ application: appName || '' }
						)
					}}
				</div>
				<div class="row justify-start items-center q-mt-md full-width">
					<q-icon size="20px" name="sym_r_check_circle" color="positive" />
					<div class="text-subtitle3 text-positive upload-message q-ml-sm">
						{{ t('Add app success') }}
					</div>
				</div>
			</div>

			<div
				v-else-if="chartStatus === UploadStatus.UPDATE"
				class="column upload-tips"
			>
				<div class="text-subtitle2 text-ink-1">
					{{
						t(
							'An app named already exists in the local source. Do you want to update it with the one you just uploaded?',
							{ application: appName || '' }
						)
					}}
				</div>
				<div class="row justify-start items-center full-width q-mt-md">
					<q-icon size="20px" name="sym_r_info" color="info" />
					<div class="upload-message text-subtitle3 text-info q-ml-sm column">
						<div>{{ t('Add app success') }}</div>
						<div>{{ t('Installed version:') + localVersion }}</div>
						<div>{{ t('Incoming version:') + appVersion }}</div>
					</div>
				</div>
			</div>

			<div
				v-else-if="chartStatus === UploadStatus.FAILED"
				class="column upload-tips"
			>
				<div class="text-subtitle2 text-negative">
					{{ t('Failed to upload chart', { chart: file?.name || '' }) }}
				</div>
				<div class="row justify-start items-center full-width q-mt-md">
					<q-icon size="20px" name="sym_r_cancel" color="negative" />
					<div class="upload-message text-subtitle3 text-negative q-ml-sm">
						{{ errorMessage || t('unknown') }}
					</div>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import BtLoading from '../../../components/base/BtLoading.vue';
import { isUpgradable, uninstalledApp } from '../../../constant/config';
import { uploadLocalPackage } from '../../../api/market/private/operations';
import { useCenterStore } from '../../../stores/market/center';
import { AppService } from '../../../stores/market/appService';
import { getRequireImage } from '../../../utils/imageUtils';
import { useUserStore } from '../../../stores/market/user';
import { withMountedGuard } from './withMountedGuard';
import { bus, BUS_EVENT, busEmit } from '../../../utils/bus';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
	APP_STATUS,
	MARKET_SOURCE_OFFICIAL
} from '../../../constant/constants';
import {
	ref,
	defineProps,
	defineEmits,
	onMounted,
	onUnmounted,
	computed,
	watch
} from 'vue';

const props = defineProps<{
	file: File;
	appInfo?: any;
}>();

const emit = defineEmits(['close']);
const centerStore = useCenterStore();

const enum UploadStatus {
	PENDING = 'pending',
	SUCCESS = 'success',
	FAILED = 'failed',
	PROCESSING = 'processing',
	UPDATE = 'update',
	REJECTED = 'rejected'
}

const customDialogRef = ref();
const chartStatus = ref(UploadStatus.PENDING);
const progress = ref(0);
const { t } = useI18n();
const $q = useQuasar();
const errorMessage = ref('');
const uploadAbortController = ref<AbortController | null>(null);
const uploadProgressCallback = ref<
	((progressEvent: ProgressEvent) => void) | null
>(null);
const isMounted = ref(false);
const appVersion = ref();
const localVersion = ref();
const appName = ref();

onMounted(() => {
	isMounted.value = true;
	withMountedGuard(isMounted, startUpload)();
});

onUnmounted(() => {
	emit('close');
	isMounted.value = false;
	cancelUpload();
});

watch(
	() => [centerStore.appFullInfoMap, centerStore.appStatusMap],
	() => {
		if (
			appName.value &&
			centerStore
				.getSourceInstalledApp(MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD)
				.includes(appName.value)
		) {
			const appStatus = centerStore.getAppStatus(
				appName.value,
				MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
			);
			const fullLatest = centerStore.getAppFullInfo(
				appName.value,
				MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
			);
			if (appStatus && fullLatest && appStatus.version !== fullLatest.version) {
				busEmit('uploadOK');
				if (!uninstalledApp(appStatus.status)) {
					localVersion.value = appStatus.version;
					chartStatus.value = UploadStatus.UPDATE;
				} else {
					chartStatus.value = UploadStatus.SUCCESS;
				}
			}
		}
	},
	{
		immediate: true,
		deep: true
	}
);

async function startUpload() {
	if (!props.file) {
		chartStatus.value = UploadStatus.FAILED;
		errorMessage.value = 'No file selected.';
		return;
	}

	uploadAbortController.value = new AbortController();
	progress.value = 0;
	uploadProgressCallback.value = (progressEvent) => {
		if (uploadProgressCallback.value) {
			progress.value = Math.round(
				(progressEvent.loaded * 100) / progressEvent.total
			);
		}
	};

	try {
		const { data, message } = await uploadLocalPackage(
			props.file,
			uploadProgressCallback.value,
			{
				signal: uploadAbortController.value.signal
			}
		);

		if (data && data.app_info) {
			appName.value = data.app_info.name;
			appVersion.value = data.app_info.version;
			chartStatus.value = UploadStatus.PROCESSING;
		} else {
			chartStatus.value = UploadStatus.FAILED;
			errorMessage.value = message || 'Upload failed due to an unknown error.';
		}
	} catch (error: any) {
		if (error.name === 'AbortError') {
			errorMessage.value = 'Upload has been canceled.';
		} else {
			errorMessage.value =
				error.response?.data?.message || 'Upload failed. Please try again.';
		}

		chartStatus.value = UploadStatus.FAILED;
	}
}

const okText = computed(() => {
	switch (chartStatus.value) {
		case UploadStatus.PENDING:
		case UploadStatus.PROCESSING:
			return false;
		case UploadStatus.FAILED:
			return t('ok');
		case UploadStatus.SUCCESS:
			return t('Install now');
		case UploadStatus.UPDATE:
			return t('app.update');
		default:
			return false;
	}
});

const cancelText = computed(() => {
	switch (chartStatus.value) {
		case UploadStatus.UPDATE:
			return t('cancel');
		case UploadStatus.SUCCESS:
			return t('Install later');
		default:
			return false;
	}
});

const onOKClick = () => {
	const appStatus = centerStore.getAppStatus(
		appName.value,
		MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
	);
	const userStore = useUserStore();
	switch (chartStatus.value) {
		case UploadStatus.SUCCESS:
			if (
				!appStatus ||
				!appStatus.status ||
				!appName.value ||
				!appVersion.value
			) {
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					'App information is missing. Please install manually.'
				);
				return;
			}
			if (uninstalledApp(appStatus.status)) {
				if (centerStore.marketInstalledApps.includes(appName.value)) {
					bus.emit(BUS_EVENT.APP_BACKEND_ERROR, t('error.source_conflict'));
					customDialogRef.value.onDialogOK();
					return;
				}

				const errorGroup = userStore.frontendPreflight(
					centerStore.getAppFullInfo(
						appName.value,
						MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
					).app_info.app_entry
				);

				if (errorGroup.length === 0) {
					AppService.installApp(
						appStatus.status,
						{
							app_name: appName.value,
							version: appVersion.value,
							source: MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
						},
						$q
					);
				} else {
					bus.emit(
						BUS_EVENT.APP_BACKEND_ERROR,
						'App pre-check failed; unable to install the application.'
					);
					console.error(
						`App pre-check failed; unable to install the application: current status is ${appStatus.status.state}`
					);
				}
			} else {
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					`App installation not supported: current status is ${appStatus.status.state}`
				);
				console.error(
					`App installation not supported: current status is ${appStatus.status.state}`
				);
			}
			customDialogRef.value.onDialogOK();
			break;
		case UploadStatus.UPDATE:
			if (
				!appStatus ||
				!appStatus.status ||
				!appName.value ||
				!appVersion.value
			) {
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					`App status error ${appStatus.status.state} don't support upgrade`
				);
				return;
			}
			if (isUpgradable(appStatus.status)) {
				AppService.upgradeApp(appStatus.status, {
					app_name: appName.value,
					version: appVersion.value,
					source: MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD
				});
			} else {
				bus.emit(BUS_EVENT.APP_BACKEND_ERROR, t('error.source_conflict'));
			}
			customDialogRef.value.onDialogOK();
			break;
		default:
			customDialogRef.value.onDialogCancel();
			break;
	}
};

const onCancelClick = () => {
	customDialogRef.value.onDialogCancel();
};

function cancelUpload() {
	uploadAbortController.value?.abort();
	uploadAbortController.value = null;
	uploadProgressCallback.value = null;
	appName.value = '';
	appVersion.value = '';
	localVersion.value = '';
	progress.value = 0;
	emit('close');
}
</script>

<style scoped lang="scss">
.upload-image {
	width: 72px;
	height: 72px;
}
.upload-tips {
	width: calc(100% - 87px);

	.upload-message {
		width: calc(100% - 36px);
	}
}
</style>
