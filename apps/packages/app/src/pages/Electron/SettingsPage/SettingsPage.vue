<template>
	<div class="wrap-settings">
		<div class="row items-center q-mt-sm q-pt-sm">
			<q-icon class="q-ml-md q-mr-xs" name="sym_r_settings" size="24px" />
			<q-toolbar-title class="q-pl-none text-body2 text-weight-bold">{{
				t('settings.settings')
			}}</q-toolbar-title>
			<q-avatar
				class="q-mr-lg cursor-pointer"
				icon="sym_r_close"
				@click="goBack"
			>
				<q-tooltip :offset="[0, 0]">Close</q-tooltip>
			</q-avatar>
		</div>
		<q-scroll-area
			ref="scrollAreaRef"
			:thumb-style="scrollBarStyle.thumbStyle"
			style="height: calc(100vh - 100px)"
		>
			<q-list class="settingList">
				<q-item class="setting_item" id="setting_1">
					<q-item-section>
						<div class="column">
							<div class="text-h6 text-ink-1">
								{{ t('settings.autostart_settings') }}
							</div>
							<terminus-check-box
								class="q-mt-md"
								:model-value="settings.automatically"
								:label="t('LarePass automatically starts at boot')"
								@update:modelValue="
									updateAutomatically(!settings.automatically)
								"
							/>
						</div>
					</q-item-section>
				</q-item>
				<TerminusSetTheme class="item-centent setting_item q-mt-xl" />

				<q-item class="q-mt-xl setting_item">
					<q-item-section>
						<TerminusSetLanguage></TerminusSetLanguage>
					</q-item-section>
				</q-item>

				<q-item id="setting_5" class="q-mt-xl setting_item">
					<q-item-section>
						<!-- <TerminusSetSecurity></TerminusSetSecurity> -->
						<div>
							<div>
								<div class="text-h6 text-ink-1">
									{{ t('import') }}
								</div>
								<div class="text-ink-2 text-body2 q-mt-md">
									{{ t('Quickly import data into your Vault') }}
								</div>
								<VaultFileImportComponent> </VaultFileImportComponent>
							</div>
						</div>
					</q-item-section>
				</q-item>

				<q-item
					id="setting_5"
					class="q-mt-xl setting_item"
					v-if="localHostAppsAbilities"
				>
					<q-item-section>
						<!-- <TerminusSetSecurity></TerminusSetSecurity> -->
						<div>
							<div>
								<div class="text-h6 text-ink-1">
									{{ t('Enable local service domain') }}
								</div>
								<div class="text-ink-2 text-body2 q-mt-md">
									{{
										t(
											"Automatically configures the system's Hosts file to direct Olares applications to the local device."
										)
									}}
								</div>
								<div class="q-mt-md text-body2 text-ink-3">
									({{ t('Administrator privileges are required.') }})
								</div>
								<div
									class="adminBtn text-body3 q-mt-md"
									v-if="addedApps.length == 0"
									@click="addAllAppToLocalHost"
								>
									<q-icon name="sym_r_add" class="icon q-mr-xs" size="16px" />
									{{ t('add') }}
								</div>
								<div class="row items-center" v-else>
									<div
										class="adminBtn text-body3 q-mt-md q-mr-md"
										v-if="missingApps.length > 0"
										@click="addAllAppToLocalHost"
									>
										<q-icon
											name="sym_r_sync"
											class="icon q-mr-xs"
											size="16px"
										/>
										{{ t('app.update') }}
									</div>

									<div
										class="adminBtn text-body3 q-mt-md"
										@click="removeAllAppsLocalHost"
									>
										<q-icon
											name="sym_r_delete"
											class="icon q-mr-xs"
											size="16px"
										/>
										{{ t('app.remove') }}
									</div>
								</div>
							</div>
						</div>
					</q-item-section>
				</q-item>

				<q-item id="setting_5" class="q-mt-xl setting_item">
					<q-item-section>
						<TerminusSetSecurity></TerminusSetSecurity>
					</q-item-section>
				</q-item>

				<q-item id="setting_6" class="q-mt-xl setting_item">
					<q-item-section>
						<div>
							<div class="text-h6 text-ink-1">
								{{ t('transmission.title') }}
							</div>
							<terminus-check-box
								class="q-mt-md"
								:model-value="settings.transmissionrKeep"
								:label="t('computer_does_not_sleep_when_there_is_a_task')"
								@update:modelValue="
									transmissionrKeepUpdate(!settings.transmissionrKeep)
								"
							/>
						</div>

						<div class="q-mt-md text-ink-3 text-body3 q-mb-xs">
							{{ t('download_location') }}
						</div>

						<div class="row items-center justify-center">
							<q-input
								outlined
								v-model="settings.downloadLocation"
								dense
								style="flex: 1"
								class="download_location"
							/>
							<div
								class="adminBtn text-body3 q-ml-md"
								@click="updateDownloadLocation"
							>
								<q-icon
									name="sym_r_folder_open"
									class="icon q-mr-xs"
									size="16px"
								/>
								{{ t('select_folder') }}
							</div>
						</div>
					</q-item-section>
				</q-item>

				<q-item id="setting_4" class="q-mt-xl setting_item">
					<!-- <q-item-section> -->
					<div>
						<div class="text-h6 text-ink-1">
							{{ t('account') }}
						</div>
						<div class="text-ink-2 text-body2 q-mt-md">
							{{ t('settings.account_root_message') }}
						</div>
						<div class="adminBtn text-body3 q-mt-md" @click="toAccountCenter">
							<q-icon name="sym_r_account_circle" size="16px" class="q-mr-xs" />
							{{ t('settings.account_administration') }}
						</div>
					</div>
				</q-item>

				<q-item class="q-mt-xl q-mb-xl setting_item">
					<q-item-section>
						<div>
							<div class="text-h6 text-ink-1">
								{{ t('about') }}
							</div>

							<div class="row items-center q-mt-md">
								<div
									class="text-ink-2 text-body2"
									v-if="
										settings.updateStatus.status != 'downloading' &&
										settings.updateStatus.status != 'downloaded'
									"
								>
									<div>
										{{ t('current_version') }}: {{ settings.appVersion }}
									</div>

									<div
										class="adminBtn text-body3 q-mt-md"
										@click="checkNewVersion"
									>
										<q-icon name="sym_r_sync" size="16px" class="q-mr-xs" />
										<q-spinner-dots
											color="ink-2"
											v-if="settings.updateStatus.status == 'checking'"
										/>
										<span v-else>
											{{ t('Check for updates') }}
										</span>
									</div>
								</div>

								<div v-else class="row items-center" style="width: 100%">
									<span class="text-body2 text-ink-2">{{
										t('Updating version {version}', {
											version: settings.updateStatus.version
										})
									}}</span>
									<dir class="line-process-bg row items-center">
										<div
											rounded
											size="8px"
											color="light-blue-default"
											class="progress"
											track-color="background-1"
											:value="(settings.updateStatus.process * 1.0) / 100"
											:style="`width: ${settings.updateStatus.process}%`"
										/>
									</dir>
									<span class="text-body2 text-light-blue-default">{{
										`${settings.updateStatus.process.toFixed(0)}%`
									}}</span>
								</div>
							</div>
						</div>
					</q-item-section>
				</q-item>
			</q-list>
		</q-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { app } from './../../../globals';
import { scrollBarStyle } from '../../../utils/contact';
import { useMenuStore } from '../../../stores/menu';
import { useQuasar } from 'quasar';
import { getPlatform } from '@didvault/sdk/src/core';
import { LayoutMenuIdetify } from '../../../utils/constants';
import TerminusSetTheme from '../../../components/common/TerminusSetTheme.vue';

import { notifySuccess } from '../../../utils/notifyRedefinedUtil';
import { LarePassElectronUpdateStatus } from 'src/platform/interface/electron/interface';
import { busEmit, busOff, busOn, NetworkUpdateMode } from '../../../utils/bus';
import LarePassUpdateReminderDialog from '../../../components/dialog/LarePassUpdateReminderDialog.vue';
import TerminusSetLanguage from 'src/components/common/TerminusSetLanguage.vue';
import TerminusSetSecurity from 'src/components/common/TerminusSetSecurity.vue';
import TerminusCheckBox from 'src/components/common/TerminusCheckBox.vue';
import VaultFileImportComponent from 'src/components/setting/VaultFileImportComponent.vue';
import { useMDNSStore } from 'src/stores/mdns';
import { TerminusStatus } from 'src/services/abstractions/mdns/service';

import { useUserStore } from 'src/stores/user';
import { useSearchStore } from 'src/stores/search';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useTermipassStore } from 'src/stores/termipass';

const menuStore = useMenuStore();
const { t } = useI18n();

const $q = useQuasar();

const userStore = useUserStore();

const searchStore = useSearchStore();

const settings = reactive({
	automatically: true,
	content: true,
	display: true,
	transmissionrKeep: true,
	transmissionDefaultLocation: true,
	weakPassword: app.account?.settings.securityReport.weakPasswords,
	repeatPassword: app.account?.settings.securityReport.reusedPasswords,
	leakPassword: app.account?.settings.securityReport.compromisedPaswords,
	expiredPassword: app.account?.settings.securityReport.expiredItems,
	downloadLocation: '',
	appVersion: '',
	updateStatus: {
		status: 'normal' as LarePassElectronUpdateStatus,
		process: 30.1,
		message: '',
		version: '1.3.10'
	}
});

const router = useRouter();
const route = useRoute();
const scrollAreaRef = ref();
const isScroll = ref(false);
const termipassStore = useTermipassStore();

const olaresSystemStatus = ref<TerminusStatus | undefined>();

watch(
	() => route.params.direction,
	(newVal) => {
		if (!isScroll.value) {
			clicks(newVal);
		}
	}
);

const clicks = (index: string) => {
	let itemScrollTop =
		document.querySelector(`#setting_${index}`) &&
		document.querySelector(`#setting_${index}`)?.offsetTop;

	if (index === '1') {
		itemScrollTop = itemScrollTop - 80;
	} else {
		itemScrollTop = itemScrollTop - 20;
	}
	scrollAreaRef.value.setScrollPosition('vertical', itemScrollTop, 300);
};

const toAccountCenter = () => {
	menuStore.pushTerminusMenuCache(LayoutMenuIdetify.ACCOUNT_CENTER);
	router.push({
		path: '/accountCenter'
	});
};

const goBack = () => {
	menuStore.popTerminusMenuCache();
	router.go(-1);
};

onMounted(async () => {
	if ($q.platform.is.electron) {
		settings.downloadLocation =
			await window.electron.api.download2.getDownloadPath();

		settings.automatically =
			await window.electron.api.settings.getAutomaticallyStartBoot();

		settings.appVersion = (await getPlatform().getDeviceInfo()).appVersion;

		settings.transmissionrKeep =
			await window.electron.api.settings.getTaskPreventSleepBoot();

		settings.updateStatus =
			await window.electron.api.settings.getUpdateStatus();

		busOn('electronVersionUpdate', electronVersionUpdate);
	}
});

onUnmounted(() => {
	if ($q.platform.is.electron) {
		busOff('electronVersionUpdate', electronVersionUpdate);
	}
});

const electronVersionUpdate = (version: any) => {
	settings.updateStatus = version;
};

const updateDownloadLocation = async () => {
	if ($q.platform.is.electron) {
		settings.downloadLocation =
			await window.electron.api.download2.selectDownloadPath();
		await window.electron.api.download2.setDownloadPath(
			settings.downloadLocation
		);
	}
};

const checkNewVersion = async () => {
	if ($q.platform.is.electron) {
		const result = await window.electron.api.settings.checkNewVersion();
		if (!result) {
			notifySuccess(t('Currently the latest version'));
		}
	} else {
		// notifySuccess(t('Currently the latest version'));
		$q.dialog({
			component: LarePassUpdateReminderDialog
		}).onOk((options: { action: 'skip' | 'update'; autoUpdate: boolean }) => {
			console.log(options);
		});
	}
};

const updateAutomatically = async (value: boolean) => {
	settings.automatically = value;
	if ($q.platform.is.electron) {
		await window.electron.api.settings.setAutomaticallyStartBoot(
			settings.automatically
		);
		settings.automatically =
			await window.electron.api.settings.getAutomaticallyStartBoot();
	}
};

const transmissionrKeepUpdate = async (value: boolean) => {
	settings.transmissionrKeep = value;
	if ($q.platform.is.electron) {
		await window.electron.api.settings.setTaskPreventSleepBoot(
			settings.transmissionrKeep
		);
		settings.transmissionrKeep =
			await window.electron.api.settings.getTaskPreventSleepBoot();
	}
};

const localHostAppsAbilities = ref(false);

const addedApps = ref<string[]>([]);
const missingApps = ref<string[]>([]);

const getSyncLocalHostAbilities = async () => {
	if (!$q.platform.is.win && !$q.platform.is.linux) {
		return;
	}
	try {
		const mdnsStore = useMDNSStore();
		const instance = await mdnsStore.getOlaresStatusInfoInstance();
		const res = await instance.get('/api/system/status');

		const status: TerminusStatus = res.data.data;
		olaresSystemStatus.value = status;

		localHostAppsAbilities.value = true;

		getAllAppStatusAddStatus();
	} catch (error) {
		localHostAppsAbilities.value = false;
	}
};

getSyncLocalHostAbilities();

const getAllAppStatusAddStatus = async () => {
	if ($q.platform.is.electron) {
		const addLocalHostsApps = await getAllNeedAddToLocalApps();
		const info = await window.electron.api.settings.getOlaresLocalAppHosts(
			userStore.current_user!.name,
			addLocalHostsApps
		);
		addedApps.value = info.added;
		missingApps.value = info.missing;
	}
};

const addAllAppToLocalHost = async () => {
	if (
		!userStore.current_user ||
		!olaresSystemStatus.value ||
		!olaresSystemStatus.value.hostIp
	) {
		return;
	}

	const addLocalHostsApps = await getAllNeedAddToLocalApps();

	if ($q.platform.is.electron) {
		$q.loading.show();
		try {
			const result =
				await window.electron.api.settings.updateOlaresLocalAppHosts(
					userStore.current_user.name,
					addLocalHostsApps
				);
			$q.loading.hide();
			if (result) {
				notifySuccess(t('success'));
				busEmit('network_update', NetworkUpdateMode.update);
			} else {
				notifyFailed(t('failed'));
			}
			getAllAppStatusAddStatus();
		} catch (error) {
			$q.loading.hide();
			notifyFailed(t('failed'));
		}
	}
};

const removeAllAppsLocalHost = async () => {
	if (
		!userStore.current_user ||
		!olaresSystemStatus.value ||
		!olaresSystemStatus.value.hostIp
	) {
		return;
	}

	if ($q.platform.is.electron) {
		$q.loading.show();
		try {
			const result =
				await window.electron.api.settings.updateOlaresLocalAppHosts(
					userStore.current_user.name,
					[]
				);
			$q.loading.hide();
			if (result) {
				notifySuccess(t('success'));
				if ($q.platform.is.linux) {
					// userStore.current_user.isLocal = false;
					termipassStore.state.publicActions.updateIsLocal(false);
				} else {
					busEmit('network_update', NetworkUpdateMode.update);
				}
			} else {
				notifyFailed(t('failed'));
			}
			getAllAppStatusAddStatus();
		} catch (error) {
			$q.loading.hide();
			notifyFailed(t('failed'));
		}
	}
};

const getAllNeedAddToLocalApps = async () => {
	if (
		!userStore.current_user ||
		!olaresSystemStatus.value ||
		!olaresSystemStatus.value.hostIp
	) {
		return [];
	}
	const appendLine = (host: string) => {
		return olaresSystemStatus.value!.hostIp + ' ' + host;
	};

	const addLocalHostsApps: string[] = [
		appendLine(userStore.current_user.local_default_host),
		appendLine('auth.' + userStore.current_user.local_default_host),
		appendLine('desktop.' + userStore.current_user.local_default_host)
	];

	await searchStore.getAppList();

	searchStore.appList.forEach((e) => {
		addLocalHostsApps.push(
			appendLine(e.id + '.' + userStore.current_user!.local_default_host)
		);
	});

	return addLocalHostsApps;
};
</script>

<style scoped lang="scss">
.wrap-settings {
	width: 100%;
	height: 100%;
	background: $background-1;
	border-radius: 12px;
}

.settingList {
	width: 500px;

	.setting_item {
		padding: 0 32px;

		.adminBtn {
			background-color: $primary;
			display: inline-block;
			color: $grey-10;
			padding: 7px 12px;
			border-radius: 8px;
			cursor: pointer;

			&:hover {
				background-color: $theme-primary-hover;
			}
		}

		.delete {
			border: 1px solid $blue-4;
			color: $blue-4;
			padding: 4px 8px;
			border-radius: 6px;

			&:hover {
				background-color: $theme-primary-hover;
			}
		}

		.download_location {
			// ::v-deep(.q-field--dense .)
			::v-deep(.q-field__inner) {
				height: 32px;
			}
		}
	}

	.line-process-bg {
		border: 1px solid $separator;
		height: 16px;
		border-radius: 10px;
		overflow: hidden;
		margin-top: 0px;
		margin-bottom: 0px;
		padding-left: 5px;
		padding-right: 5px;
		// position: relative;
		flex: 1;
		width: auto;
		margin-left: 10px;
		margin-right: 10px;
		// width: 100%;

		.progress {
			// width: 48px;
			height: 8px;
			border-radius: 4px;
			background: $light-blue-default;
			// position: absolute;
			// top: 3px;
			// left: 0px;
			// animation: slide 1s linear infinite;
		}
	}

	.lock-slider {
		// height: 60px;
		transition: height 0.5s;
		min-height: 0 !important;
		padding-top: 0px !important;
		padding-bottom: 0px !important;
	}

	.hideSlider {
		height: 0 !important;
	}

	.theme-select {
		width: 440px;
		height: 144px;
		// background-color: red;

		.theme-item-common {
			height: 144px;
			width: 210px;
			border: 1px solid $separator;
			border-radius: 12px;
			overflow: hidden;
			.image {
				width: 100%;
				height: 100px;
			}
			.content {
				width: 100%;
				height: 44px;
			}
		}

		.theme-item-select {
			border: 1px solid $yellow-default;
		}
	}

	.item-centent {
		border-radius: 12px;
		width: 100%;
		text-align: left;

		.checkbox-content {
			width: 100%;
			height: 30px;
			.checkbox-common {
				width: 16px;
				height: 16px;
				margin-right: 10px;
				border-radius: 4px;
			}

			.checkbox-unselect {
				border: 1px solid $separator-2;
			}

			.checkbox-selected-yellow {
				background: $yellow-default;
			}
			.checkbox-selected-green {
				background: $positive;
			}
		}

		.lock-slider {
			height: 60px;
			transition: height 0.5s;
			min-height: 0 !important;
			padding-top: 0px !important;
			padding-bottom: 0px !important;
		}

		.hideSlider {
			height: 0 !important;
		}
	}
}
::v-deep(.q-field--dense .q-field__control) {
	height: 32px;
}
</style>
