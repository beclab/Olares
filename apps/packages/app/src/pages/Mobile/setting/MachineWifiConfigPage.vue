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
						src="../../../assets/wizard/machine-not-found.svg"
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
					<template v-if="wiredNetwork">
						<div class="q-mt-lg text-ink-1 text-h6">
							{{ t('Wired network') }}
						</div>
						<terminus-item
							class="q-mt-sm"
							@click="updateIp(wiredNetwork)"
							:borderClasses="
								wiredNetwork.isHostIp ? 'wifi-item-selectd' : 'wifi-item-normal'
							"
							img-bg-classes="bg-background-3"
							icon-name="sym_r_cable"
							:whole-picture-size="32"
							:icon-size="16"
							:icon-color="
								wiredNetwork.isHostIp ? 'light-blue-default' : 'ink-1'
							"
						>
							<template v-slot:title>
								<div
									class="text-subtitle2"
									:class="
										wiredNetwork.isHostIp ? 'text-light-blue' : 'text-ink-1'
									"
								>
									{{ t('Wired network') }}
								</div>
							</template>
							<template v-slot:detail>
								<div class="column text-body3 text-ink-3">
									{{
										wiredNetwork.isHostIp ? wiredNetwork.ip : t('Not Connected')
									}}
								</div>
							</template>
							<template v-slot:side>
								<div class="q-mr-md">
									<q-img
										src="img/checkbox/check_box_circle.svg"
										width="14px"
										height="14px"
										v-if="wiredNetwork.isHostIp"
									/>
								</div>
							</template>
						</terminus-item>
					</template>

					<div class="q-mt-lg text-ink-1 text-h6" v-if="hasWifiNetwork">
						{{ t('Wi-Fi') }}
					</div>
					<template v-for="(ip, index) in ips" :key="index">
						<terminus-item
							v-if="ipIsWifi(ip)"
							class="q-mt-sm"
							@click="updateIp(ip)"
							:borderClasses="
								ip.isHostIp ? 'wifi-item-selectd' : 'wifi-item-normal'
							"
							img-bg-classes="bg-background-3"
							:imagePath="getWifiIcon(ip)"
							:whole-picture-size="32"
							:icon-size="16"
							:icon-color="ip.isHostIp ? 'light-blue-default' : 'ink-1'"
						>
							<template v-slot:title>
								<div
									class="text-subtitle2"
									:class="ip.isHostIp ? 'text-light-blue' : 'text-ink-1'"
								>
									{{ ip.ssid || t('Wi-Fi') }}
								</div>
							</template>
							<template v-slot:detail>
								<div class="text-body3 text-ink-3">
									{{ ip.isHostIp ? ip.ip : t('Not Connected') }}
								</div>
							</template>
							<template v-slot:side>
								<div class="q-mr-md">
									<q-img
										src="img/checkbox/check_box_circle.svg"
										width="14px"
										height="14px"
										v-if="ip.isHostIp"
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
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { ref, onMounted, watch, onUnmounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
// import { NetworkUpdateMode, busOn, busOff } from '../../../utils/bus';
import { useMDNSStore, SystemIPS } from '../../../stores/mdns';
import { useUserStore } from '../../../stores/user';
import InputPasswordDialog from '../../../components/setting/InputPasswordDialog.vue';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';
import UserStatusCommonDialog from '../../../components/userStatusDialog/UserStatusCommonDialog.vue';
import { useRouter } from 'vue-router';
import {
	TerminusStatusEnum,
	displayStatus
} from '../../../services/abstractions/mdns/service';
import { getNativeAppPlatform } from 'src/application/platform';

const userStore = useUserStore();

const mdnsStore = useMDNSStore();
mdnsStore.mdnsUsed = true;

const { t } = useI18n();

const mobileIsWifi = ref(false);

const ssidRef = ref('');

const $q = useQuasar();

const router = useRouter();

const ips = ref<SystemIPS[]>([]);

const ipIsWifi = (ip: SystemIPS) => {
	if (ip.isWifi != undefined) {
		return ip.isWifi;
	}
	return ip.iface.startsWith('wla') || ip.iface.startsWith('wlo');
};

onMounted(async () => {
	configWifiStatus();
	mdnsStore.startSearchMdnsService();
});

// const appStateChange = async (state: { isActive: boolean }) => {
// 	console.log('app statue update');
// 	if (state.isActive) {
// 		configWifiStatus();
// 	}
// };

// const networkUpdate = async (mode: NetworkUpdateMode) => {
// 	console.log('networkUpdate', mode);

// 	if (mode == NetworkUpdateMode.update) {
// 		configWifiStatus();
// 	}
// };

watch(
	() => mdnsStore.mdnsMachines,
	async () => {
		if (mdnsStore.mdnsMachines.length > 0) {
			const activedMachine = mdnsStore.mdnsMachines.find(
				(e) => e.status && e.status.terminusName == userStore.current_user?.name
			);
			mdnsStore.setActivedMachine(activedMachine, undefined);
			if (ips.value.length == 0 && activedMachine) {
				refreshIps();
			}
		}
	},
	{
		deep: true
	}
);

const refreshIps = async () => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	ips.value = await mdnsStore.getSystemIfs(mdnsStore.activedMachine);
};

const updateIp = async (item: SystemIPS) => {
	if (!mdnsStore.activedMachine) {
		return;
	}
	if (item.isHostIp) {
		return;
	}

	if (
		mdnsStore.activedMachine.status?.terminusState !=
		TerminusStatusEnum.TerminusRunning
	) {
		if (mdnsStore.activedMachine.status) {
			notifyFailed(
				t(
					'Olares current status is {status}. This operation is not allowed. Please try again later.',
					{
						status: displayStatus(mdnsStore.activedMachine.status).status
					}
				)
			);
		}

		return;
	}

	let updateItem = item;
	if (ipIsWifi(item) && (item.ip.length == 0 || item.ssid != ssidRef.value)) {
		if (item.ssid && item.ssid != ssidRef.value) {
			if (!(await reminderNetwork())) {
				return;
			}
		}
		const result = await connectWifi();
		if (!result) {
			return;
		}
		await refreshIps();
		const ip = ips.value.find(
			(e) => e.ssid == ssidRef.value && e.ip.length > 0
		);
		if (!ip) {
			return;
		}
		updateItem = ip;
	}
	const result = await mdnsStore.changeHostMachineTerminus(
		mdnsStore.activedMachine,
		{
			ip: updateItem.ip
		}
	);

	if (result) {
		router.back();
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

onUnmounted(() => {
	mdnsStore.stopSearchMdnsService();
	// busOff('appStateChange', appStateChange);
	// busOff('network_update', networkUpdate);
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
					router.back();
				})
				.onCancel(() => {
					router.back();
				});
		}
	}
};

const connectWifi = async () => {
	const activedMachine = mdnsStore.activedMachine;
	if (!activedMachine) {
		return false;
	}
	return new Promise<boolean>((resolve) => {
		$q.dialog({
			component: InputPasswordDialog,
			componentProps: {
				title: t('Connect to {wifi}', {
					wifi: ssidRef.value
				}),
				passwordTitle: t('Enter password')
			}
		})
			.onOk(async (password: string) => {
				// if (!password) {
				// 	notifyFailed(t('check_password'));
				// 	resolve(false);
				// 	return;
				// }
				const result = await mdnsStore.connectWifiMachineTerminus(
					activedMachine,
					{
						ssid: ssidRef.value,
						password: password
					}
				);
				resolve(result);
			})
			.onCancel(() => {
				// return false;
				resolve(false);
			});
	});
};

const wiredNetwork = computed(() => {
	return ips.value.find((e) => !ipIsWifi(e));
});

const hasWifiNetwork = computed(() => {
	return ips.value.find((e) => ipIsWifi(e)) != undefined;
});

const getWifiIcon = (wifi: SystemIPS) => {
	let prefixStr = 'normal_';
	if (wifi.isHostIp) {
		prefixStr = 'connect_';
	}
	let strength = 'less';
	if (wifi.strength) {
		if (wifi.strength > 80) {
			strength = 'full';
		} else if (wifi.strength > 50) {
			strength = 'middle';
		}
	} else {
		strength = 'full';
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
