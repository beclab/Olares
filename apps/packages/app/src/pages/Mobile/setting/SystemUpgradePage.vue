<template>
	<div class="upgrade-root">
		<terminus-title-bar :title="title">
			<template v-slot:right>
				<div style="color: transparent" id="debug-trigger">DEBUG</div>
			</template>
		</terminus-title-bar>
		<terminus-scroll-area
			class="restore-scroll-area"
			:style="{
				'--more_btn_height': mdnsStore.upgradingDownloadCompleted
					? '58px'
					: '0px'
			}"
		>
			<template
				v-slot:content
				v-if="mdnsStore.activedMachine && mdnsStore.activedMachine.status"
			>
				<div v-if="mdnsStore.upgradingError" style="margin-top: 120px">
					<scan-local-machine
						:title="t('Update failed to install')"
						:reminder-content="
							t(
								'Before retrying, please check:<br>You have a stable network connection.<br>Your device is powered on.'
							)
						"
						:icon="'upgrade/upgrading_failed.svg'"
					/>
					<div class="text-body2 q-mt-md text-negative">
						{{
							mdnsStore.activedMachine?.status?.upgradingDownloadError ||
							mdnsStore.activedMachine?.status?.upgradingError
						}}
					</div>
				</div>

				<div
					v-else-if="mdnsStore.isUpgrading"
					class="column items-center"
					style="margin-top: 120px"
				>
					<template v-if="mdnsStore.isNewVersionUpgrading">
						<scan-local-machine
							:title="t('Upgrading system')"
							:reminder-content="
								t(
									'Keep your device connected to power during the installation process.'
								)
							"
							:icon="'upgrade/upgrading_system.svg'"
						/>

						<q-linear-progress
							style="margin-top: 64px"
							rounded
							size="4px"
							:value="
								Number(
									mdnsStore.activedMachine?.status?.upgradingProgress
										? mdnsStore.activedMachine.status.upgradingProgress.split(
												'%'
										  )[0]
										: '0'
								) / 100
							"
							color="positive"
							track-color="separator"
						/>

						<div class="text-body2 text-ink-3 q-mt-sm">
							{{ mdnsStore.activedMachine?.status?.upgradingProgress }}
						</div>
					</template>
					<template v-else-if="mdnsStore.upgradingDownloadCompleted">
						<scan-local-machine
							:title="t('Download complete')"
							:reminder-content="
								t('Update package has been downloaded successfully')
							"
							:icon="'upgrade/updating_download.svg'"
						/>
					</template>
					<template v-else-if="mdnsStore.isNewVersionDowaloading">
						<scan-local-machine
							:title="t('Downloading update')"
							:reminder-content="
								t('Please keep your device connected to the internet')
							"
							:icon="'upgrade/updating_downloading.svg'"
						/>
						<div style="margin-top: 64px" class="q-mb-sm text-ink-2 text-body1">
							{{ mdnsStore.activedMachine?.status?.upgradingDownloadStep }}
						</div>

						<q-linear-progress
							rounded
							size="4px"
							:value="
								Number(
									mdnsStore.activedMachine?.status?.upgradingDownloadProgress
										? mdnsStore.activedMachine.status.upgradingDownloadProgress.split(
												'%'
										  )[0]
										: '0'
								) / 100
							"
							color="positive"
							track-color="separator"
						/>

						<div class="text-body2 text-ink-3 q-mt-sm">
							{{ mdnsStore.activedMachine?.status?.upgradingDownloadProgress }}
						</div>
					</template>
				</div>
				<div
					v-else-if="hasUpgrade && !mdnsStore.upgradeEnable"
					style="margin-top: 120px"
				>
					<scan-local-machine
						:title="t('Upgrade successful')"
						:reminder-content="
							t(
								'Congratulations! Your system has been successfully upgraded to version {version}',
								{
									version: mdnsStore.activedMachine?.status?.terminusVersion
								}
							)
						"
						:icon="'upgrade/upgrading_success.svg'"
					/>
				</div>
				<div v-else class="q-mt-md">
					<terminus-item :clickable="false" :borderRadius="12">
						<template v-slot:title>
							<div class="text-subtitle2 text-ink-1">
								{{ t('current_version') }}
							</div>
						</template>
						<template v-slot:side>
							<div class="text-ink-3 text-body3 q-mr-md">
								{{ mdnsStore.activedMachine?.status?.terminusVersion || '--' }}
							</div>
						</template>
					</terminus-item>

					<terminus-item
						:clickable="true"
						:borderRadius="12"
						class="q-mt-md"
						v-if="!isDebug"
					>
						<template v-slot:title>
							<div class="text-subtitle2 text-ink-1" id="debug-trigger">
								{{ t('RC updates') }}
							</div>
						</template>
						<template v-slot:detail>
							<div class="text-overline-m text-ink-3">
								{{ t('Enable to get early access to new features.') }}
							</div>
						</template>
						<template v-slot:side>
							<bt-switch
								size="sm"
								truthy-track-color="light-blue-default"
								v-model="includeRC"
								@update:model-value="updateIncludeRc"
							/>
						</template>
					</terminus-item>

					<terminus-item
						:clickable="false"
						:borderRadius="12"
						class="q-mt-md"
						v-else-if="availableSelect.length > 0"
					>
						<template v-slot:title>
							<div class="text-subtitle2 text-ink-1">
								{{ 'DEBUG' }}
							</div>
						</template>

						<template v-slot:side>
							<bt-select
								v-model="testLevel"
								:options="availableSelect"
								:offset="[40, 10]"
								@update:modelValue="checkLastOlaresVersion"
							/>
						</template>
					</terminus-item>

					<terminus-item :borderRadius="12" class="q-mt-md">
						<template v-slot:title>
							<div class="text-subtitle2 text-ink-1">
								{{ t('new_version') }}
							</div>
						</template>
						<template v-slot:side>
							<div
								class="row item-center justify-end q-mr-md"
								v-if="mdnsStore.upgradeEnable"
							>
								<div class="upgrade-info">
									<q-icon name="sym_r_rocket_launch" color="white"></q-icon>
									<span class="text-subtitle3 text-white">
										{{ mdnsStore.activedMachine?.version?.upgradeableVersion }}
									</span>
								</div>
							</div>
							<div v-else class="text-body3 text-ink-3 q-mr-md">
								{{ t("You've update to date.") }}
							</div>
						</template>
					</terminus-item>
				</div>
			</template>
		</terminus-scroll-area>
		<div class="bottom-view column justify-end">
			<template v-if="mdnsStore.upgradingError">
				<confirm-button
					:btn-title="t('Cancel Upgrade')"
					@on-confirm="cancelUpgrade"
				/>
			</template>
			<template v-else-if="mdnsStore.isUpgrading">
				<template v-if="mdnsStore.isNewVersionUpgrading"></template>
				<template v-else-if="mdnsStore.upgradingDownloadCompleted">
					<confirm-button
						:btn-title="t('Upgrade now')"
						@on-confirm="confirmUpgrade"
					/>
					<div
						class="bluetooth row items-center justify-center t text-body2"
						@click="upgradeLater"
					>
						{{ t('Upgrade later') }}
					</div>
				</template>
				<template v-else-if="mdnsStore.isNewVersionDowaloading">
					<confirm-button
						:btn-title="t('Cancel download')"
						@on-confirm="cancelUpgrade"
					/>
				</template>
			</template>
			<template v-else-if="hasUpgrade && !mdnsStore.upgradeEnable">
				<confirm-button
					:btn-title="t('Finish')"
					@on-confirm="upgradeFinished"
				/>
			</template>
			<template v-else-if="mdnsStore.upgradeEnable">
				<confirm-button :btn-title="t('Upgrade')" @on-confirm="upradeOlares" />
			</template>
		</div>
	</div>
</template>
<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';

import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { useMDNSStore } from '../../../stores/mdns';

import { onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';

import TerminusItem from '../../../components/common/TerminusItem.vue';
import ScanLocalMachine from '../connect/activate/ScanLocalMachine.vue';
import * as semver from 'semver';
import { UpgradeLevel } from '../../../services/abstractions/mdns/service';
import BtSelect from '../../../components/base/BtSelect.vue';
import { useUserStore } from 'src/stores/user';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';
import { useQuasar } from 'quasar';

const { t } = useI18n();

const mdnsStore = useMDNSStore();
mdnsStore.mdnsUsed = false;

const router = useRouter();

const hasUpgrade = ref<boolean>(mdnsStore.isUpgrading);

const userStore = useUserStore();

const $q = useQuasar();

onMounted(() => {
	mdnsStore.startSearchMdnsService();
	checkLastOlaresVersion();

	// 为目标元素添加点击事件监听
	const targetElement = document.querySelector('#debug-trigger');
	if (targetElement) {
		targetElement.addEventListener('click', handleClick);
		console.log('已启用调试模式触发功能，连续点击6下目标元素进入调试模式');
	} else {
		console.warn('未找到目标元素，请检查选择器是否正确');
	}
});

onUnmounted(() => {
	mdnsStore.stopSearchMdnsService();
});

const includeRC = ref(mdnsStore.includeRC);

const updateIncludeRc = () => {
	mdnsStore.updateIncludeRc(includeRC.value);
	checkLastOlaresVersion();
};

const checkLastOlaresVersion = async () => {
	let version: undefined | any = undefined;
	if (!mdnsStore.activedMachine) {
		return;
	}
	if (isDebug.value) {
		version = await mdnsStore.checkLastOsVersion(
			mdnsStore.activedMachine,
			testLevel.value
		);
	} else {
		version = await mdnsStore.checkLastOsVersion(
			mdnsStore.activedMachine,
			includeRC.value ? UpgradeLevel.RC : UpgradeLevel.NONE
		);
	}
	if (version) {
		mdnsStore.activedMachine.version = version;
	}
};

const upradeOlares = () => {
	if (!mdnsStore.activedMachine) {
		return;
	}

	hasUpgrade.value = true;
	mdnsStore.upgradeOlares(mdnsStore.activedMachine!);
};

const cancelUpgrade = () => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	mdnsStore.cancelUpgradeOlares(mdnsStore.activedMachine);
};

const confirmUpgrade = () => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	mdnsStore.confirmUpgradeOlares(mdnsStore.activedMachine);
};

const upgradeLater = () => {
	router.back();
};

const title = computed(() => {
	if (mdnsStore.upgradingError || mdnsStore.isUpgrading) {
		return t('upgrading');
	}
	return t('System update');
});

const isDebug = ref(false);

const isOfficialRelease = (version: string) => {
	const parsed = semver.parse(version);
	if (!parsed) return false;
	return parsed.prerelease.length === 0 && parsed.build.length === 0;
};

const isRCRelease = (version: string) => {
	const parsed = semver.parse(version);
	if (!parsed) return false;

	return parsed.prerelease.some((part) =>
		String(part).toLowerCase().includes('rc')
	);
};

const testLevel = ref(UpgradeLevel.NONE);

const availableSelect = computed(() => {
	if (
		!mdnsStore.activedMachine ||
		!mdnsStore.activedMachine.status ||
		!mdnsStore.activedMachine.status.terminusVersion
	) {
		return [];
	}

	if (isRCRelease(mdnsStore.activedMachine.status.terminusVersion)) {
		return [UpgradeLevel.NONE, UpgradeLevel.RC].map((e) => {
			return {
				label: e == UpgradeLevel.NONE ? 'None' : e,
				value: e
			};
		});
	}

	if (isOfficialRelease(mdnsStore.activedMachine.status.terminusVersion)) {
		return [
			UpgradeLevel.NONE,
			UpgradeLevel.ALPHA,
			UpgradeLevel.BETA,
			UpgradeLevel.RC
		].map((e) => {
			return {
				label: e == UpgradeLevel.NONE ? 'None' : e,
				value: e
			};
		});
	}
	return [];
});

let clickCount = 0;
let lastClickTime = 0;
const REQUIRED_CLICKS = 6;
const TIME_THRESHOLD = 1000;

function handleClick() {
	const now = Date.now();
	if (now - lastClickTime > TIME_THRESHOLD) {
		clickCount = 1;
	} else {
		clickCount++;
	}
	lastClickTime = now;
	if (clickCount === REQUIRED_CLICKS) {
		isDebug.value = !isDebug.value;
		clickCount = 0;
	}
}

watch(
	() => mdnsStore.mdnsMachines,
	() => {
		if (mdnsStore.mdnsMachines.length > 0) {
			const activedMachine = mdnsStore.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			if (
				!mdnsStore.activedMachine ||
				mdnsStore.activedMachine.host != activedMachine?.host
			) {
				mdnsStore.setActivedMachine(activedMachine, undefined, true);
				hasUpgrade.value = mdnsStore.isUpgrading;
				checkLastOlaresVersion();
			}
		}
	},
	{
		deep: true
	}
);

const upgradeFinished = () => {
	hasUpgrade.value = false;
	checkLastOlaresVersion();
};
</script>

<style scoped lang="scss">
.upgrade-root {
	width: 100%;
	height: 100%;

	.restore-scroll-area {
		width: 100%;
		height: calc(100% - 56px - 48px - 48px - var(--more_btn_height));
		padding-left: 20px;
		padding-right: 20px;

		.upgrade-info {
			background: $light-blue-default;
			padding: 2px 12px;
			border-radius: 20px;
		}
	}

	.bottom-view {
		width: 100%;
		padding-bottom: 48px;
		padding-left: 20px;
		padding-right: 20px;

		.bluetooth {
			height: 48px;
			margin-top: 10px;
			width: 100%;
			color: $light-blue-default;
		}

		.bluetooth:hover {
			color: $blue-default;
		}
	}
}
</style>
