<template>
	<div class="container-header row items-center justify-between">
		<div class="row items-center justify-between">
			<div class="row items-center justify-start">
				<div class="app row items-center justify-center">
					<span class="q-ml-sm text-h6 text-ink-1">{{
						store.current_app?.title || store.current_app?.appName
					}}</span>
				</div>
			</div>
		</div>
		<div class="row items-center justify-end">
			<!-- <div class="status q-mx-md" v-if="dockerStore.appInstallState">
				<q-icon
					class="q-mr-sm"
					:name="app_install_status_style[dockerStore.appInstallState]?.icon"
					size="16px"
					:style="{
						color: app_install_status_style[dockerStore.appInstallState]?.color
					}"
				/>
				<span class="text-subtitle3 text-ink-2">
					{{ t(`appInstallStatus.${dockerStore.appInstallState}`) }}
				</span>
			</div> -->

			<div class="row items-center">
				<div
					class="status"
					v-if="
						dockerStore.appStatus && dockerStore.appStatus != APP_STATUS.EMPTY
					"
				>
					<span
						class="color q-mr-sm"
						:style="{
							background: app_status_style[dockerStore.appStatus].color
						}"
					></span>
					<span
						class="text-subtitle3 text-ink-2 row items-center q-gutter-x-md"
					>
						<span>{{ t(`appStatus.${dockerStore.appStatus}`) }}</span>
					</span>
				</div>
			</div>

			<q-separator
				class="q-mx-md"
				v-if="dockerStore.appStatus"
				vertical
				inset
			/>

			<div
				class="operate-btn q-mr-sm q-px-sm"
				@click="openServerEditor"
				v-if="
					store.current_app.devEnv !== 'default' &&
					dockerStore.appStatus === APP_STATUS.DEPLOYED
				"
			>
				<img src="./../assets/vscode.svg" style="width: 14px; height: 14px" />
				<span class="operate-btn-install">{{ t('vscode_server') }}</span>
			</div>

			<div
				class="operate-btn q-mr-sm q-px-sm"
				@click="onPreview"
				v-if="
					store.current_app.devEnv === 'default' &&
					dockerStore.appStatus === APP_STATUS.DEPLOYED
				"
			>
				<q-icon name="sym_r_preview" size="16px" />
				<span class="operate-btn-install">{{ t('btn_preview') }}</span>
			</div>

			<div
				class="operate-btn operate-disabled q-mr-sm q-px-sm"
				v-if="
					store.current_app.devEnv === 'default' &&
					dockerStore.appStatus === APP_STATUS.UNDEPLOY
				"
			>
				<q-icon name="sym_r_preview" size="16px" />
				<span class="operate-btn-install">{{ t('btn_preview') }}</span>
			</div>

			<div
				v-if="
					!controlFlag &&
					dockerStore.appStatus &&
					[
						APP_STATUS.DEPLOYED,
						APP_STATUS.UNDEPLOY,
						APP_STATUS.ABNORMAL
					].includes(dockerStore.appStatus)
				"
				class="operate-btn q-mr-sm q-px-sm"
				:class="{
					'operate-disabled': uninstalling
				}"
				@click="onInstall"
			>
				<q-icon
					:name="installNeedless ? 'sym_r_low_priority' : 'sym_r_box_edit'"
					size="16px"
				/>

				<span v-if="installNeedless" class="operate-btn-install">{{
					t('cancel')
				}}</span>

				<span v-else class="operate-btn-install">{{
					t('header_btn.apply')
				}}</span>
			</div>

			<div
				v-if="
					!controlFlag &&
					dockerStore.appStatus &&
					[APP_STATUS.EMPTY, APP_STATUS.DEPLOYING].includes(
						dockerStore.appStatus
					)
				"
				class="operate-btn operate-disabled q-mr-sm q-px-sm"
			>
				<q-icon name="sym_r_box_edit" size="16px" />
				<span class="operate-btn-install">{{ t('header_btn.apply') }}</span>
			</div>

			<div
				v-if="controlFlag"
				class="operate-btn q-mr-sm q-px-sm"
				@click="openFiles"
			>
				<q-icon name="sym_r_box_edit" size="16px" />
				<span class="operate-btn-install">{{ t('header_btn.edit') }}</span>
			</div>

			<input
				ref="uploadInput"
				type="file"
				style="display: none"
				accept=".zip,.rar,.7z,.tar,.tgz"
				@change="uploadFile"
			/>

			<div class="operate-more">
				<q-icon name="sym_r_more_vert" color="ink-2" />
				<q-menu class="rounded-borders" flat>
					<q-list dense padding>
						<q-item
							v-if="
								store.current_app.devEnv !== 'default' &&
								dockerStore.appStatus === APP_STATUS.DEPLOYED
							"
							class="row items-center justify-start text-ink-2"
							clickable
							v-ripple
							@click="openEditor"
							v-close-popup
						>
							<div class="q-mr-xs" style="width: 20px; height: 20px">
								<img
									src="./../assets/vscode.svg"
									style="width: 14px; height: 14px"
								/>
							</div>

							{{ t('vscode_desktop') }}
						</q-item>
						<q-item
							class="row items-center justify-start text-ink-2 no-wrap"
							clickable
							v-ripple
							@click="onRename"
							v-close-popup
						>
							<q-icon class="q-mr-xs" name="sym_r_edit_square" size="20px" />
							{{ t('btn_rename') }}
						</q-item>
						<q-item
							class="row items-center justify-start text-ink-2"
							clickable
							v-ripple
							@click="onUploadChart"
							v-close-popup
						>
							<q-icon class="q-mr-xs" name="sym_r_upload" size="20px" />
							{{ t('btn_upload') }}
						</q-item>

						<q-item
							class="row items-center justify-start text-ink-2"
							clickable
							v-ripple
							@click="onDownload"
							v-close-popup
							:disable="
								!dockerStore.appStatus ||
								dockerStore.appStatus === APP_STATUS.EMPTY
							"
						>
							<q-icon class="q-mr-xs" name="sym_r_download" size="20px" />
							{{ t('btn_download') }}
						</q-item>

						<q-item
							class="row items-center justify-start text-ink-2"
							clickable
							v-ripple
							v-close-popup
							:disable="
								dockerStore.appStatus === APP_STATUS.DEPLOYED ? false : true
							"
							@click="onUninstall"
						>
							<q-icon class="q-mr-xs" name="sym_r_reset_tv" size="20px" />
							{{ t('btn_uninstall') }}
						</q-item>

						<q-item
							class="row items-center justify-start text-ink-2"
							clickable
							v-ripple
							@click="onDeleteApplication"
							v-close-popup
						>
							<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
							{{ t('btn_delete') }}
						</q-item>
					</q-list>
				</q-menu>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onUnmounted, ref, watch, computed } from 'vue';
import axios from 'axios';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { ROUTE_NAME } from '../common/router-name';
import {
	APP_STATUS,
	app_status_style,
	app_install_status_style,
	APP_INSTALL_STATE
} from '../types/core';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

import { useDevelopingApps } from '../stores/app';
import { useDockerStore } from '../stores/docker';
import { useFileStore } from '../stores/file';
import { useMenuStore } from '../stores/menu';
import { pushToSystem } from './../utils/utils';

import DialogConfirm from '../components/dialog/DialogConfirm.vue';
import RenameApp from '../components/dialog/RenameApp.vue';

import { componentName } from '@apps/control-hub/src/router/const';
import { on } from 'events';

const emits = defineEmits(['updateNotify']);

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const store = useDevelopingApps();
const dockerStore = useDockerStore();
const fileStore = useFileStore();
const menuStore = useMenuStore();
const $q = useQuasar();

const uploadInput = ref();
const downloading = ref(false);

const timer = ref();

const uninstalling = computed(
	() =>
		dockerStore.appStatusInfo?.state &&
		dockerStore.appStatusInfo.state === APP_INSTALL_STATE.UNINSTALLING
);
const controlFlag = computed(() => {
	const flag = Object.values({ ...componentName, ...ROUTE_NAME }).find(
		(item) => item === route.name
	);
	if (flag) {
		return true;
	} else {
		return false;
	}
});

const installNeedless = computed(() => {
	if (
		fileStore.filesEditStorage.length <= 0 &&
		dockerStore.appStatus === APP_STATUS.DEPLOYED
	) {
		return true;
	} else {
		return false;
	}
});

async function onPreview() {
	if (window.top == window) {
		window.open('//' + store.current_app?.entrance, '_blank');
	} else {
		await store.openApplication({
			appid: String(store.current_app?.id),
			path: ''
		});
	}
}

async function onInstall() {
	if (uninstalling.value) {
		return;
	}
	dockerStore.clearMessage();
	if (installNeedless.value) {
		return openSystem();
	}

	//  hide ErrorFooter
	if (store.current_app?.id) {
		store.setHideErrorFooter(store.current_app.id, true);
	}

	if (fileStore.isEditing) {
		if (controlFlag.value) {
			//
		} else {
			const current_path = route.path.slice(5);
			await fileStore.saveFile(current_path);
		}
	}

	$q.loading.show({
		message: t('installing')
	});

	try {
		let namespace = '';
		if (store.current_app) {
			try {
				const res = await dockerStore.install_app(store.current_app.appName);
				namespace = res.namespace;
				dockerStore.appStatus = null;
				dockerStore.appInstallState = null;
				emits('updateNotify', null);
				// await getAppState();
				await updateAppState();
				await openSystem(namespace);
			} catch (error) {
				dockerStore.appStatus = null;
				dockerStore.appInstallState = null;
				// await getAppState();
				await updateAppState();
			}
		}
		$q.loading.hide();
		fileStore.filesEditStorage = [];
	} catch (error) {
		$q.loading.hide();
	}
}

async function openSystem(value?: string) {
	pushToSystem(route.params.id, router, value);
	// dockerStore.appInstallState = null;
}

async function openFiles() {
	// dockerStore.appInstallState = null;
	router.push({ path: menuStore.currentItem });
}

function onUninstall() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_uninstall'),
			message: t('message.uninstall_app', {
				appName: store.current_app?.title || store.current_app?.appName
			})
		}
	}).onOk(async () => {
		try {
			store.current_app &&
				(await dockerStore.un_install_app(store.current_app.appName));
			dockerStore.appStatus = null;
			dockerStore.appInstallState = null;
			// await getAppState();
			await updateAppState();
			openFiles();
		} catch (error) {
			console.log(error);
		}
	});
}

async function onDeleteApplication() {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_delete'),
			message: t('message.delete_app', {
				appName: store.current_app?.title || store.current_app?.appName
			})
		}
	}).onOk(async () => {
		$q.loading.show();

		try {
			store.current_app &&
				(await dockerStore.delete_app(store.current_app.appName));
			$q.loading.hide();
			await store.getApps();
			await menuStore.updateApplications();
			router.push({ path: '/home' });
		} catch (error) {
			$q.loading.hide();
		}
	});
}

async function onDownload() {
	downloading.value = true;
	window.location.href =
		store.url +
		'/api/command/download-app-chart?app=' +
		store.current_app?.appName;
	downloading.value = false;
}

async function onUploadChart() {
	uploadInput.value.value = null;
	uploadInput.value.click();
}

async function uploadFile(event: any) {
	const file = event.target.files[0];
	if (file) {
		const { status, message } = await upload_dev_file(file);
		if (status) {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: message
			});
		}
	} else {
		console.log('file selected failure');
	}
}

async function upload_dev_file(
	file: any
): Promise<{ status: boolean; message: string }> {
	try {
		const formData = new FormData();
		formData.append('chart', file);
		store.current_app && formData.append('app', store.current_app.appName);
		await axios.post(store.url + '/api/command/upload-app-chart', formData, {
			headers: { 'Content-Type': 'multipart/form-data' }
		});
		return { status: true, message: t('message.successfully') };
	} catch (e: any) {
		console.log(e);
		return { status: false, message: e.message };
	}
}

async function refreshApplication() {
	store.current_app = store.apps.find(
		(item) => item.appName == route.params.id
	);
}

watch(
	() => route.path,
	() => {
		refreshApplication();
	},
	{
		immediate: true
	}
);

watch(
	() => route.params.id,
	async (newVal) => {
		if (newVal) {
			emits('updateNotify', null);

			await updateAppState();

			if (dockerStore.appStatus === APP_STATUS.UNDEPLOY) {
				const path = route.path.split('/').slice(0, 3).join('/');
				router.push({ path });
			}
		}
	},
	{
		immediate: true
	}
);

async function updateAppState() {
	if (timer.value) clearInterval(timer.value);
	await getAppState();
	await getAppInstallState();

	timer.value = setInterval(async () => {
		await getAppState();
	}, 5000);
}

async function getAppState() {
	if (!route.params.id) {
		return false;
	}

	if (
		dockerStore.appStatus &&
		[APP_STATUS.EMPTY, APP_STATUS.ABNORMAL, APP_STATUS.DEPLOYED].includes(
			dockerStore.appStatus
		)
	) {
		return false;
	}

	try {
		await dockerStore.get_app_status(route.params.id as string);
	} catch (error) {
		console.log(error);
	}
}

const getAppInstallState = async () => {
	if (
		dockerStore.appStatus &&
		[APP_STATUS.EMPTY, APP_STATUS.ABNORMAL, APP_STATUS.UNDEPLOY].includes(
			dockerStore.appStatus
		)
	) {
		return false;
	}

	if (
		dockerStore.appInstallState &&
		[
			APP_INSTALL_STATE.CANCELED,
			APP_INSTALL_STATE.FAILED,
			APP_INSTALL_STATE.COMPLETED
		].includes(dockerStore.appInstallState)
	) {
		return false;
	}
	try {
		const res = await dockerStore.get_app_install_state(route.params.id);
		if (res.state === APP_INSTALL_STATE.FAILED) {
			emits('updateNotify', res.message);
		}
	} catch (error) {
		console.log(error);
	}
};

const openServerEditor = () => {
	window.open('//' + store.current_app?.entrance, '_blank');
};

const openEditor = () => {
	const a = document.createElement('a');
	a.href = 'vscode://';
	a.click();
};

async function onRename() {
	$q.dialog({
		component: RenameApp,
		componentProps: {
			currentTitle: store.current_app?.title || store.current_app?.appName
		}
	}).onOk(async (newTitle: string) => {
		if (!newTitle || newTitle === store.current_app?.title) {
			return;
		}

		$q.loading.show({
			message: t('renaming')
		});

		try {
			if (store.current_app) {
				await dockerStore.rename_app(store.current_app.appName, newTitle);
				// 更新本地存储的应用信息
				store.current_app.title = newTitle;
				// 刷新应用列表
				await store.getApps();
				await menuStore.updateApplications();
			}
			$q.loading.hide();
		} catch (error) {
			console.log(error);
			$q.loading.hide();
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('message.rename_failed')
			});
		}
	});
}

onUnmounted(() => {
	clearInterval(timer.value);
});
</script>
<style lang="scss" scoped>
.app-container {
	height: 100vh;
}
.rounded-borders {
	border-radius: 12px !important;
	overflow: hidden;
}

.my-active-class {
	height: 52px !important;
}
</style>
<style scoped lang="scss">
.container-header {
	width: 100%;
	height: 56px;
	padding: 0 20px;
	border-bottom: 1px solid $separator;
	.app {
		margin-right: 20px;

		img {
			width: 32px;
			height: 32px;
			border-radius: 8px;
		}
	}

	.event-btn {
		width: 80px;
		height: 32px;
		line-height: 32px;
		text-align: center;
		display: inline-block;
		border-radius: 8px;
		border: 1px solid $btn-stroke;
		overflow: hidden;
		color: $ink-1;
		cursor: pointer;

		&:hover {
			background: $background-hover;
		}
		&.operate-disabled {
			opacity: 0.5;
		}
	}

	.operate-btn {
		height: 32px;
		line-height: 30px;
		text-align: center;
		border-radius: 8px;
		border: 1px solid $btn-stroke;
		overflow: hidden;
		color: $ink-2;
		display: flex;
		align-items: center;
		justify-content: center;

		&:hover {
			background: $background-hover;
		}
		&.operate-disabled {
			opacity: 0.5;
		}
		.ani-loading {
			animation: rotate 1s linear infinite;
		}
		.operate-btn-install {
			display: inline-block;
			padding-left: 4px;
			color: $ink-2;
			font-size: 12px;
			text-align: center;
			cursor: pointer;
			font-weight: 400;
			line-height: 16px;
			letter-spacing: 0%;
			vertical-align: middle;
		}
	}

	@keyframes rotate {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.operate-more {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		box-sizing: border-box;
		border: 1px solid $btn-stroke;
		cursor: pointer;
		&:hover {
			background: $background-hover;
		}
	}
	.status {
		display: flex;
		align-items: center;
		justify-content: center;
		.color {
			width: 10px;
			height: 10px;
			border-radius: 5px;
			display: inline-block;
		}
	}
}
</style>
../utils/utils
