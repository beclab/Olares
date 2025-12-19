<template>
	<div class="app-layout">
		<ApplicationHeader @update-notify="updateNotify"></ApplicationHeader>

		<Notify
			v-if="error_message"
			:status="APP_STATUS.ABNORMAL"
			:message="error_message"
			@close="close"
		/>

		<div class="app-container">
			<router-view></router-view>
		</div>
		<div
			v-if="showInstallFooter"
			style="flex: 0 0 48px"
			class="full-width row items-center q-px-lg justify-between install-footer"
		>
			<div class="row items-center q-gutter-sm">
				<q-icon
					size="18px"
					color="ink-2"
					:name="
						isDownloading ? 'sym_r_deployed_code_update' : 'sym_r_deployed_code'
					"
				/>
				<div class="row items-center text-body3 text-ink-2">
					<span>{{ $t('app_state') }}:&nbsp;</span>
					<span v-if="dockerStore.appStatusInfo?.state">
						{{
							isDownloading
								? $t('appInstallStatus.downloading')
								: $t(`appInstallStatus.${dockerStore.appStatusInfo.state}`)
						}}
					</span>
					<span v-if="isDownloading"
						>...{{ dockerStore.appStatusInfo.progress || 0 }}%
					</span>
				</div>
			</div>
			<div
				class="row items-center q-gutter-sm state-right"
				@click="onViewStatus"
			>
				<q-icon
					name="sym_r_mark_chat_unread"
					size="18px"
					style="margin-bottom: -3px"
					color="ink-2"
				/>
				<span class="text-body3 text-ink-2">
					{{
						dockerStore.appStatusInfo?.state === APP_INSTALL_STATE.DOWNLOADING
							? $t('waitingForInstall')
							: dockerStore.appStatusInfo?.state === 'running'
							? $t('running')
							: $t(`appInstallStatus.${dockerStore.appStatusInfo?.state || ''}`)
					}}
				</span>
			</div>
		</div>
		<div
			v-if="showErrorFooter"
			style="flex: 0 0 48px"
			class="full-width bg-red-soft row no-wrap items-center q-px-lg justify-between error-footer"
		>
			<div class="row no-wrap items-center q-gutter-x-xs ellipsis">
				<q-icon name="sym_r_error" color="red-default" size="15px" />
				<span class="text-body3 text-ink-1 ellipsis">{{
					appStore.current_app?.reason
				}}</span>
			</div>
			<div class="row no-wrap items-center q-gutter-md q-ml-lg">
				<q-btn
					dense
					outline
					no-caps
					padding="xs md"
					@click="onView"
					class="header-btn-wrapper"
				>
					<span class="text-caption text-ink-2">{{ $t('view') }}</span>
				</q-btn>
				<q-btn
					dense
					outline
					no-caps
					padding="xs md"
					@click="onDismiss"
					class="header-btn-wrapper"
				>
					<span class="text-caption text-ink-2">{{ $t('dismiss') }}</span>
				</q-btn>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
	APP_STATUS,
	APP_INSTALL_STATE,
	app_install_status_style
} from '../types/core';
import ApplicationHeader from '@apps/studio/src/pages/ApplicationHeader.vue';
import Notify from './../components/common/Notify.vue';
import { useDevelopingApps } from '@apps/studio/src/stores/app';
import { useDockerStore } from '@apps/studio/src/stores/docker';
import AppError from './AppError.vue';
import AppStatusInfo from './AppStatusInfo.vue';

const appStore = useDevelopingApps();
const dockerStore = useDockerStore();
const $q = useQuasar();
const { t } = useI18n();

const error_message = ref();

const showInstallFooter = computed(() => {
	return (
		dockerStore.appStatus === APP_STATUS.DEPLOYED &&
		dockerStore.appStatusInfo?.state &&
		![
			APP_INSTALL_STATE.CANCELED,
			APP_INSTALL_STATE.FAILED,
			APP_INSTALL_STATE.COMPLETED
		].includes(dockerStore.appStatusInfo.state)
	);
});

const isDownloading = computed(
	() => dockerStore.appStatusInfo?.state === APP_INSTALL_STATE.DOWNLOADING
);

const leftStatusText = computed(() => {
	const info = dockerStore.appStatusInfo as any;
	if (info?.state === APP_INSTALL_STATE.DOWNLOADING) {
		return `${t('appInstallStatus.downloading')}... ${info.progress || 0}%`;
	}
	if (info?.state) {
		return t(info.state.toUpperCase());
	}
	return '';
});

const showErrorFooter = computed(() => {
	const errorStatus = [APP_STATUS.ABNORMAL, APP_STATUS.EMPTY];
	return (
		appStore.current_app &&
		appStore.current_app.state &&
		errorStatus.includes(appStore.current_app.state) &&
		appStore.current_app.reason &&
		!appStore.current_app.hideErrorFooter
	);
});

const workloadHeight = computed(() =>
	showInstallFooter.value || showErrorFooter.value
		? `calc(100vh - 104px)`
		: `calc(100vh - 56px)`
);

const updateNotify = (value) => {
	error_message.value = value;
};

const close = () => {
	error_message.value = null;
};

const onView = () => {
	if (appStore.current_app?.reason) {
		$q.dialog({
			title: t('error_details'),
			component: AppError,
			componentProps: {
				errorMessage: appStore.current_app.reason
			}
		});
	}
};

const onViewStatus = () => {
	if (dockerStore.appStatusInfo) {
		$q.dialog({
			title: t('app_status_info'),
			component: AppStatusInfo,
			componentProps: {
				statusInfo: dockerStore.appStatusInfo
			}
		});
	}
};

const onDismiss = () => {
	if (appStore.current_app) {
		appStore.setHideErrorFooter(appStore.current_app.id, true);
	}
};
</script>

<style lang="scss" scoped>
.app-layout {
	height: 100vh;
	display: flex;
	align-items: flex-start;
	justify-content: flex-start;
	flex-direction: column;
}

.app-container {
	flex: 1;
	width: 100%;
	overflow: hidden;
}

.install-footer {
	border-top: 1px solid $separator-2;
	background: $background-1;
}

.my-menu-before-scroll {
	height: calc(100vh - 76px) !important;
}

.my-menu-after-content {
	padding-bottom: 300px !important;
	.q-scrollarea__content {
		padding-bottom: 60px !important;
	}
}
.header-btn-wrapper {
	white-space: nowrap;
}
.header-btn-wrapper::before {
	border-color: $btn-stroke;
}

.state-right {
	cursor: pointer;
	padding: 4px 8px;
	border-radius: 4px;
	transition: background-color 0.2s;

	&:hover {
		background: $background-hover;
	}
}
</style>
<style lang="scss" scoped>
::v-deep(
		.studio-workload-wrapper .my-scroll-area-wrapper.my-scroll-area-no-header
	) {
	height: v-bind(workloadHeight);
}
::v-deep(.studio-workload-wrapper .my-menu-before-scroll) {
	height: v-bind(workloadHeight);
}
::v-deep(.studio-workload-wrapper .my-menu-before-wrapper-no-header) {
	height: 0px;
}
</style>
