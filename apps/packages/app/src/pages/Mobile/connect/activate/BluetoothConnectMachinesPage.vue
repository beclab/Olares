<template>
	<div class="bluetooth-local-machine-root">
		<terminus-title-bar
			:title="isScaning || isNotFound ? '' : t('Bluetooth network setup')"
		/>

		<terminus-scroll-area class="local-machine-scroll">
			<template v-slot:content>
				<div v-if="isScaning || isNotFound" class="wizard-content">
					<scan-local-machine
						:icon="isNotFound ? 'wizard/machine-not-found.svg' : undefined"
						:title="
							isNotFound
								? t('Olares not found')
								: t('Scanning Olares in Bluetooth')
						"
						:reminder-content="
							isNotFound
								? t('No Olares device pending activation found')
								: t('Bluetooth on the phone must be turned on')
						"
					/>
				</div>
				<dir
					v-else-if="mdnsStore.bluetoothPageDisplayMachines.length == 0"
					class="wizard-content"
				>
					<scan-local-machine
						:title="t('Olares not found')"
						:icon="'wizard/machine-not-found.svg'"
						:reminder-content="t('No Olares device pending activation found')"
					/>
				</dir>
				<q-list class="machines-list" v-else>
					<local-machine-item
						v-for="machine in mdnsStore.bluetoothPageDisplayMachines"
						:key="machine.host"
						:machine="machine"
						:isBluetooth="true"
						@bluetooth-config-network="configNetwork"
					/>
				</q-list>
			</template>
		</terminus-scroll-area>

		<div class="bottom-view column justify-end" v-if="!isScaning">
			<confirm-button
				:btn-title="t('Rescan Olares via Bluetooth')"
				:btnStatus="
					mdnsStore.bluetoothPageDisplayMachines.length > 0
						? ConfirmButtonStatus.disable
						: ConfirmButtonStatus.normal
				"
				@onConfirm="rescanTerminus"
			/>
		</div>
	</div>
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useMDNSStore } from '../../../../stores/mdns';

import { useQuasar } from 'quasar';

import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import ConfirmButton from '../../../../components/common/ConfirmButton.vue';
import ScanLocalMachine from './ScanLocalMachine.vue';
import LocalMachineItem from '../../../../components/wizard/LocalMachineItem.vue';
import { useI18n } from 'vue-i18n';

import { ConfirmButtonStatus } from '../../../../utils/constants';

import './wizard.scss';
import { useRouter } from 'vue-router';
import { TerminusServiceInfo } from '../../../../services/abstractions/mdns/service';

const mdnsStore = useMDNSStore();

const { t } = useI18n();

const isScaning = ref(true);

const isNotFound = ref(false);

const router = useRouter();

onMounted(() => {
	if ($q.platform.is.nativeMobile) {
		mdnsStore.startSearchBluetoothDevice();
	}
	startFoundMachines();
});

onUnmounted(() => {
	if ($q.platform.is.nativeMobile) {
		mdnsStore.stopSearchBluetoothDevice();
	}
	clearfoundMachinesTimer();
});

let foundMachinesTimer: any = undefined;

const clearfoundMachinesTimer = () => {
	if (foundMachinesTimer) {
		clearInterval(foundMachinesTimer);
		foundMachinesTimer = undefined;
	}
};

let leftTime = 10;
const startFoundMachines = () => {
	if (foundMachinesTimer) {
		clearfoundMachinesTimer();
	}
	leftTime = 10;
	isNotFound.value = false;
	isScaning.value = true;

	foundMachinesTimer = setInterval(async () => {
		leftTime = leftTime - 1;
		if (leftTime == 0) {
			clearfoundMachinesTimer();
			if (mdnsStore.bluetoothMachines.length == 0) {
				isNotFound.value = true;
			}
			isScaning.value = false;
		}
	}, 1 * 1000);
};

watch(
	() => mdnsStore.bluetoothMachines,
	() => {
		if (mdnsStore.bluetoothMachines.length > 0) {
			isScaning.value = false;
			isNotFound.value = false;
		}
	},
	{
		deep: true
	}
);

const rescanTerminus = async () => {
	if ($q.platform.is.nativeMobile) {
		await mdnsStore.stopSearchBluetoothDevice();
		await mdnsStore.startSearchBluetoothDevice();
	}
	startFoundMachines();
};

const configNetwork = (device: TerminusServiceInfo) => {
	mdnsStore.blueConfigNetworkMachine = device;
	router.push({
		path: '/bluetooth/machine/wifi'
	});
};

const $q = useQuasar();
</script>

<style scoped lang="scss">
.bluetooth-local-machine-root {
	width: 100%;
	height: 100%;

	.local-machine-scroll {
		width: 100%;
		height: calc(100% - 56px - 48px - 48px);
		.machines-list {
			padding-left: 20px;
			padding-right: 20px;
			padding-bottom: 20px;
			width: 100%;
		}
	}
	.bottom-view {
		width: 100%;
		padding-bottom: 48px;
		padding-left: 20px;
		padding-right: 20px;
	}
}
</style>
