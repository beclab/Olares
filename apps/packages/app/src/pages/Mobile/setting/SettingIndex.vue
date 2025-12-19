<template>
	<div class="setting">
		<div class="flur1"></div>
		<div class="flur2"></div>
		<TerminusUserHeader :title="t('setting')">
			<template v-slot:add v-if="$q.platform.is.nativeMobile">
				<div
					class="scan-icon row items-center justify-center"
					@click="scanQrCode"
				>
					<q-img
						:src="
							$q.dark.isActive
								? getRequireImage('common/dark_scan_qr.svg')
								: getRequireImage('common/scan_qr.svg')
						"
						width="20px"
					/>
				</div>
			</template>
		</TerminusUserHeader>
		<terminus-scroll-area class="setting-content q-mt-sm">
			<template v-slot:content>
				<div class="terminus-info q-pt-xs">
					<div class="module-content q-mt-md q-pa-lg">
						<div class="row items-center justify-between">
							<div>
								<div class="home-module-title">
									{{ t('main.my_terminus') }}
								</div>
								<div class="text-ink-3 text-body3">
									{{ userName }}
								</div>
							</div>
							<switch-component
								v-if="!userStore.current_user?.offline_mode"
								:title="t('vpn.title')"
								v-model="vpnToggleStatus"
								:disabled="scaleStore.isDisconnecting"
								@update:model-value="updateVpnStatus"
							/>
							<switch-component
								v-else
								:title="t('user_current_status.offline_mode.title')"
								v-model="offLineModeRef"
								@update:model-value="updateOffLineMode"
							/>
						</div>
						<div v-if="mdnsStore.apiMachine && monitorStore.usages">
							<div class="row items-center full-width justify-between q-mt-md">
								<div class="module-content-item items-50 q-pa-md">
									<olares-info :title="t('network')" @click="enterNetworkPage">
										<template v-slot:detail>
											<terminus-user-status hideBg />
										</template>
									</olares-info>
								</div>
								<div class="module-content-item items-50 q-pa-md">
									<olares-info
										v-if="mdnsStore.apiMachine.status"
										:title="t('system')"
										:hiddenArrow="
											mdnsStore.apiMachine.status.terminusName != user.name
										"
										@click="enterScanMachineInfo"
									>
										<template v-slot:detail>
											<olares-status :status="mdnsStore.apiMachine.status" />
										</template>
									</olares-info>

									<olares-info
										v-else
										:title="t('system')"
										icon-name="sym_r_error"
										@click="serviceUnabailableReminder"
									>
										<template v-slot:detail>
											<olares-status :statusStr="t('Service unavailable')" />
										</template>
									</olares-info>
								</div>
							</div>
							<div class="module-content-item usage q-mt-md q-pa-md">
								<div>
									{{ t('Usage') }}
								</div>
								<MonitorInfo2 />
							</div>
						</div>
						<olares-unreachable
							v-else
							class="q-mt-md"
							@search-olares="enterScanMachineInfo"
						/>

						<auth-item-card class="q-mt-lg" />
					</div>

					<div class="module-content q-mt-md q-pa-lg">
						<div class="home-module-title q-mb-md">
							LarePass{{ ' ' + t('settings.settings') }}
						</div>

						<terminus-item
							v-for="(cell, index) in setting"
							:key="index"
							class="q-mt-none"
							img-bg-classes="bg-background-3"
							:icon-name="cell.icon"
							:whole-picture-size="32"
							:side-style="index == 0 ? 'width: 50%' : ''"
							@click="setPath(cell.path)"
							:show-board="false"
							:paddingLeft="0"
						>
							<template v-slot:title>
								<div class="text-subtitle1">
									{{ cell.label }}
								</div>
							</template>
							<template v-slot:side>
								<div class="row items-center justify-end" style="width: 100%">
									<q-icon
										name="sym_r_keyboard_arrow_right"
										size="24px"
										color="ink-2"
									/>
								</div>
							</template>
						</terminus-item>
					</div>

					<div
						style="padding-bottom: 60px; width: 100%; height: 1px"
						v-if="
							termipassStore &&
							termipassStore.totalStatus &&
							termipassStore.totalStatus.isError == 2
						"
					/>
					<div v-else style="padding-bottom: 30px; width: 100%; height: 1px" />
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import { ref, onMounted, onUnmounted, watch, computed } from 'vue';
import { useUserStore } from '../../../stores/user';
import { UserItem } from '@didvault/sdk/src/core';
import { useQuasar } from 'quasar';
import TerminusUserHeader from '../../../components/common/TerminusUserHeader.vue';
import { useMonitorStore } from '../../../stores/monitor';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';

import AuthItemCard from '../items/AuthItemCard2.vue';
import { getRequireImage } from '../../../utils/imageUtils';

import { useI18n } from 'vue-i18n';
import { useTermipassStore } from '../../../stores/termipass';
import SwitchComponent from '../../../components/common/SwitchComponent.vue';
import OlaresInfo from '../../../components/setting/OlaresInfo.vue';
import TerminusUserStatus from '../../../components/common/TerminusUserStatus.vue';
import OlaresStatus from '../../../components/setting/OlaresStatus.vue';
import MonitorInfo2 from '../../../components/setting/MonitorInfo2.vue';
import { useScaleStore } from '../../../stores/scale';
import { useMDNSStore } from '../../../stores/mdns';
import OlaresUnreachable from '../../../components/setting/OlaresUnreachable.vue';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';

const { t } = useI18n();
const isBex = ref(process.env.IS_BEX);

const $router = useRouter();
const settingMenu = [
	{
		label: t('General'),
		icon: 'sym_r_settings_applications',
		path: '/setting/display'
	}
];
const userStore = useUserStore();
const termipassStore = useTermipassStore();
const $q = useQuasar();

const monitorStore = useMonitorStore();

const mdnsStore = useMDNSStore();

const scaleStore = useScaleStore();

const settingsMenus = [...settingMenu];

if ($q.platform.is.nativeMobile) {
	settingsMenus.splice(1, 0, {
		label: t('integration.title'),
		icon: 'sym_r_stacks',
		path: '/integration'
	});
}

settingsMenus.push({
	label: t('settings.safety'),
	icon: 'sym_r_arming_countdown',
	path: '/setting/security'
});

const setting = ref(settingsMenus);

const connected = ref(false);
const userName = ref('');
const userId = ref('');

let user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
if (user.setup_finished) {
	connected.value = true;
	userName.value = user.name;
	userId.value = user.id;
} else {
	connected.value = false;
}

let moniter = ref<any>(null);

onMounted(async () => {
	monitorStore.loadMonitor();
	mdnsStore.getOlaresInfo();
	moniter.value = setInterval(() => {
		monitorStore.loadMonitor();
		mdnsStore.getOlaresInfo();
	}, 1000 * 30);
});

onUnmounted(() => {
	if (moniter.value) {
		clearInterval(moniter.value);
	}
});

const setPath = async (path: string) => {
	$router.push({
		path
	});
};

const scanQrCode = () => {
	$router.push({
		path: '/scanQrCode'
	});
};

const enterScanMachineInfo = async () => {
	if (
		!mdnsStore.apiMachine ||
		!mdnsStore.apiMachine.status ||
		mdnsStore.apiMachine.status.terminusName != user.name
	) {
		return;
	}
	if (!(await userStore.unlockFirst()) || isBex.value) {
		return;
	}
	$router.push({
		path: '/operation/machine'
	});
};

const serviceUnabailableReminder = () => {
	$q.dialog({
		component: TerminusTipDialog,
		componentProps: {
			title: t('Service unavailable'),
			showCancel: false,
			message: t(
				'The Olares system service required for this feature is not yet available on all host platforms.'
			)
		}
	});
};

const enterNetworkPage = () => {
	$router.push({
		path: '/setting/network'
	});
};

const vpnToggleStatus = computed({
	get: () => scaleStore.isOn || scaleStore.isConnecting,
	set: async (value) => {
		if (scaleStore.isDisconnecting) {
			return;
		}
		if (value) {
			await scaleStore.start();
		} else {
			await scaleStore.stop();
		}
	}
});

watch(
	() => userStore.current_user?.offline_mode,
	() => {
		offLineModeRef.value = userStore.current_user?.offline_mode || false;
	}
);

const offLineModeRef = ref(userStore.current_user?.offline_mode || false);

const updateOffLineMode = async () => {
	userStore.updateOfflineMode(offLineModeRef.value);
	monitorStore.loadMonitor();
	mdnsStore.getOlaresInfo();
};
</script>

<style lang="scss" scoped>
.setting {
	height: 100%;
	width: 100%;
	position: relative;
	z-index: 0;

	.flur1 {
		width: 135px;
		height: 135px;
		background: rgba(133, 211, 255, 0.3);
		filter: blur(70px);
		position: absolute;
		right: 20vw;
		top: 20vh;
		z-index: -1;
	}

	.flur2 {
		width: 110px;
		height: 110px;
		background: rgba(217, 255, 109, 0.2);
		filter: blur(70px);
		position: absolute;
		left: 20vw;
		top: 30vh;
		z-index: -1;
	}

	.setting-content {
		height: calc(100% - 56px);
		width: 100%;

		.terminus-info {
			width: 100%;
			padding-left: 20px;
			padding-right: 20px;

			.module-content {
				border: 1px solid $separator;
				border-radius: 20px;
				width: 100%;
				background: $background-1;

				.module-content-item {
					border-radius: 8px;
					background: $background-3;
				}

				.items-50 {
					width: calc(50% - 6px);
					height: 64px;
				}

				.usage {
					// height: 72px;
				}

				.olares-unreachable-bg {
					background: #fff7f2;
					border: 1px solid $separator;
					border-radius: 12px;

					.title {
						color: $ink-1;
					}

					&__introduce {
						width: calc(100% - 135px);
						height: 100%;

						.detail {
							color: $ink-2;
						}

						.backup {
							background: $yellow;
							border-radius: 8px;
							height: 32px;
							text-align: center;
							color: $grey-10;
							width: 93px;
						}
					}
				}
			}

			&__name {
				color: $ink-2;
				max-width: 100%;
			}

			.app-version {
				text-align: right;
			}
		}
	}

	.not-backup {
		text-align: right;
		color: $grey-5;
	}
}
</style>
