<template>
	<div class="wifi-config-root">
		<terminus-title-bar :title="t('Select connection method')" />
		<terminus-scroll-area class="wifi-config-scroll-area">
			<template v-slot:content>
				<div
					v-if="!mobileIsWifi"
					class="column items-center"
					style="width: 100%"
				>
					<q-img
						class="q-mt-xl"
						src="../../../../assets/wizard/machine-not-found.svg"
						width="124px"
						height="124px"
					/>
					<div class="text-ink-1 text-h5 q-mt-sm">
						{{ t('No available Wi-Fi network found.') }}
					</div>
					<div class="text-body3 text-ink-2 q-mt-sm" style="text-align: center">
						{{
							t(
								'Your phone is not connected to Wi-Fi. To ensure the intranet experience, the Wi-Fi connected to Olares must be the same as the Wi-Fi network connected to your phone. Please connect your phone to the Wi-Fi network first.'
							)
						}}
					</div>
				</div>
				<div v-else>
					<div class="q-mt-lg text-ink-1 text-h6" v-if="ips.length > 0">
						{{ t('Wi-Fi') }}
					</div>

					<template v-for="(ip, index) in ips" :key="index">
						<terminus-item
							class="q-mt-sm"
							@click="updateIp(ip)"
							:borderClasses="ip.co ? 'wifi-item-selectd' : 'wifi-item-normal'"
							img-bg-classes="bg-background-3"
							:imagePath="getWifiIcon(ip)"
							:whole-picture-size="32"
							:icon-size="16"
						>
							<template v-slot:title>
								<div
									class="text-subtitle2"
									:class="ip.co ? 'text-light-blue' : 'text-ink-1'"
								>
									{{ ip.ss }}
								</div>
							</template>
							<template v-slot:detail>
								<div class="text-body3 text-ink-3">
									{{ ip.co ? t('Wi-Fi Connected') : t('Not Connected') }}
								</div>
							</template>
							<template v-slot:side>
								<div class="q-mr-md">
									<q-img
										src="img/checkbox/check_box_circle.svg"
										width="14px"
										height="14px"
										v-if="ip.co"
									/>
								</div>
							</template>
						</terminus-item>
					</template>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>
<script setup lang="ts">
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import TerminusItem from '../../../../components/common/TerminusItem.vue';
import { ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useMDNSStore } from '../../../../stores/mdns';
import InputPasswordDialog from '../../../../components/setting/InputPasswordDialog.vue';
import TerminusTipDialog from '../../../../components/dialog/TerminusTipDialog.vue';
import { useRouter } from 'vue-router';
import UserStatusCommonDialog from '../../../../components/userStatusDialog/UserStatusCommonDialog.vue';
import {
	notifyFailed,
	notifySuccess
} from '../../../../utils/notifyRedefinedUtil';
import { BluetoothWifiInfo } from 'src/utils/interface/bluetooth';
import { getNativeAppPlatform } from 'src/application/platform';

const mdnsStore = useMDNSStore();

const { t } = useI18n();

const mobileIsWifi = ref(true);

const ssidRef = ref('');

const $q = useQuasar();

const router = useRouter();

const ips = ref<BluetoothWifiInfo[]>([]);

onMounted(async () => {
	configWifiStatus();

	if (mdnsStore.blueConfigNetworkMachine) {
		refreshWifiList();
		startAutoRefreshWifi();
	}

	// busOn('network_update', networkUpdate);

	// busOn('appStateChange', appStateChange);
});

// const appStateChange = async (state: { isActive: boolean }) => {
// 	if (state.isActive) {
// 		configWifiStatus();
// 	}
// };

// const networkUpdate = async (mode: NetworkUpdateMode) => {
// 	if (mode == NetworkUpdateMode.update) {
// 		configWifiStatus();
// 	}
// };

let refreshIpsTimer: any;

let refreshWifiListIng = false;

const refreshWifiList = async () => {
	if (refreshWifiListIng) {
		return;
	}
	refreshWifiListIng = true;
	ips.value = await mdnsStore.getBluetoothMachineWifi(
		mdnsStore.blueConfigNetworkMachine
	);
	refreshWifiListIng = false;
};

const startAutoRefreshWifi = () => {
	clearRefreshIpsTimer();
	refreshIpsTimer = setInterval(async () => {
		if (isWifiUpdating) {
			return;
		}
		refreshWifiList();
	}, 10 * 1000);
};

const clearRefreshIpsTimer = () => {
	if (refreshIpsTimer) {
		clearInterval(refreshIpsTimer);
		refreshIpsTimer = undefined;
	}
};

const updateIp = async (item: BluetoothWifiInfo) => {
	if (item.co) {
		return;
	}
	if (item.ss != ssidRef.value) {
		const result = await reminderNetwork();
		if (!result) {
			return;
		}
	}
	connectWifi(item);
};

onUnmounted(() => {
	// busOff('appStateChange', appStateChange);
	// busOff('network_update', networkUpdate);
	clearRefreshIpsTimer();
	mdnsStore.blueConfigNetworkMachine = undefined;
});

const configWifiStatus = async () => {
	if (!$q.platform.is.nativeMobile) {
		return;
	}
	mobileIsWifi.value =
		(await getNativeAppPlatform().netConnectionStatus()).connectionType ==
		'wifi';

	if (mobileIsWifi.value) {
		try {
			ssidRef.value = await getNativeAppPlatform().getWifiSSID();
		} catch (error) {
			$q.dialog({
				component: TerminusTipDialog,
				componentProps: {
					title: t('Local precise location permission'),
					message: t(
						'LarePass needs to access local precise location permission. Go to permission settings to turn it on?'
					)
				}
			})
				.onOk(async () => {
					getNativeAppPlatform().openLocationSettings();
				})
				.onCancel(() => {
					router.back();
				});
		}
	}
};
const reminderNetwork = () => {
	return new Promise<boolean>((resolve) => {
		$q.dialog({
			component: UserStatusCommonDialog,
			componentProps: {
				title: t('Network Tips'),
				message: t(
					'If your phone and Olares are not on the same Wi-Fi network, LarePass will not be able to detect Olares.'
				),
				messageClasses: 'text-ink-2',
				messageCenter: true,
				btnTitle: t('confirm'),
				resetDomainTitle: t('Reselect Network'),
				resetDomainClasses: 'item-common-border text-subtitle1 text-ink-1',
				setDomain: true
			}
		})
			.onOk(async (value: string) => {
				if (value == 'reset') {
					resolve(false);
				} else if (value == 'confirm') {
					resolve(true);
				}
			})
			.onCancel(() => {
				resolve(false);
			});
	});
};

let isWifiUpdating = false;

const connectWifi = async (item: BluetoothWifiInfo) => {
	return new Promise<boolean>((resolve) => {
		$q.dialog({
			component: InputPasswordDialog,
			componentProps: {
				title: t('Connect to {wifi}', {
					wifi: item.ss
				}),
				passwordTitle: t('Enter password')
			}
		})
			.onOk(async (password: string) => {
				// if (!password) {
				// 	resolve(false);
				// 	return;
				// }
				$q.loading.show();
				isWifiUpdating = true;
				const result = await mdnsStore.bluetoothConnectWifi(
					mdnsStore.blueConfigNetworkMachine!,
					item.ss,
					password
				);
				isWifiUpdating = false;
				$q.loading.hide();
				if (result.status) {
					notifySuccess(
						t('Connected to Wi-Fi {wifi}', {
							wifi: item.ss
						})
					);
					setTimeout(() => {
						router.go(-2);
					}, 1000);
				} else {
					if (result.message && result.message.length > 0) {
						notifyFailed(result.message);
					}
				}
			})
			.onCancel(() => {
				resolve(false);
			});
	});
};

const getWifiIcon = (wifi: BluetoothWifiInfo) => {
	let prefixStr = 'normal_';
	if (wifi.co) {
		prefixStr = 'connect_';
	}

	let strength = 'less';
	if (wifi.st > 80) {
		strength = 'full';
	} else if (wifi.st > 50) {
		strength = 'middle';
	}
	return `wifi/${prefixStr}${strength}.svg`;
};
</script>

<style scoped lang="scss">
.wifi-config-root {
	width: 100%;
	height: 100%;

	.wifi-config-scroll-area {
		width: 100%;
		height: calc(100% - 56px);
		padding-left: 20px;
		padding-right: 20px;

		.wifi-item-selectd {
			border: 1px solid $light-blue-default;
		}
		.wifi-item-normal {
			border: 1px solid $separator;
		}
	}
}
</style>
