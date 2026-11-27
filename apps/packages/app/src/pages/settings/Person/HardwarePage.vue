<template>
	<page-title-component :show-back="true" :title="t('My hardware')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div
			class="q-mt-lg q-mb-lg column items-center justify-center"
			v-if="terminusDStore.olaresInfo"
		>
			<q-img :src="deviceLogo(terminusDStore.olaresInfo)" width="100px" />
			<div class="text-h4 text-ink-1 q-mt-sm">
				{{ terminusDStore.olaresInfo?.host_name }}
			</div>
		</div>
		<bt-list v-if="terminusDStore.olaresInfo">
			<!-- <bt-form-item
				:title="t('Model')"
				:data="
					terminusDStore.olaresInfo.mode || terminusDStore.olaresInfo.host_name
				"
			></bt-form-item> -->
			<bt-form-item :title="t('Device status')">
				<olares-status :status="terminusDStore.olaresInfo" :node-size="8" />
			</bt-form-item>
			<bt-form-item
				:title="t('Device identifier')"
				:data="terminusStore.olares_device_id"
			/>
			<bt-form-item
				:title="t('CPU')"
				:data="terminusDStore.olaresInfo.cpu_info || '--'"
			/>
			<bt-form-item
				:title="t('GPU')"
				:data="terminusDStore.olaresInfo.gpu_info || '--'"
				:width-separator="isOlareOne"
			/>
			<!--			<bt-form-item-->
			<!--				:title="t('Memory')"-->
			<!--				:data="terminusDStore.olaresInfo.memory || '&#45;&#45;'"-->
			<!--			/>-->
			<!--			<bt-form-item-->
			<!--				:title="t('DISK')"-->
			<!--				:data="terminusDStore.olaresInfo.disk || '&#45;&#45;'"-->
			<!--				:widthSeparator="isOlareOne"-->
			<!--			/>-->
			<bt-form-item
				v-if="isOlareOne"
				:title="t('Power mode')"
				:widthSeparator="true"
			>
				<bt-select
					v-model="oneWorkerMode"
					:options="oneWorkerModeOptions()"
					@update:modelValue="updateWorkerMode"
				/>
			</bt-form-item>
			<bt-form-item
				v-if="isOlareOne"
				:title="t('CPU frequency limit')"
				:widthSeparator="true"
			>
				<bt-switch
					size="sm"
					truthy-track-color="blue-default"
					v-model="cpuFreqLimitEnabled"
					@update:model-value="updateCpuFreqLimit"
				/>
			</bt-form-item>
			<bt-form-item
				v-if="isOlareOne"
				:title="t('Auto power on')"
				:description="
					autoPowerLoading || autoPowerSupported
						? ''
						: t('Auto power is not supported by the current device')
				"
				:widthSeparator="false"
			>
				<bt-switch
					size="sm"
					truthy-track-color="blue-default"
					v-model="autoPowerOn"
					:disable="autoPowerLoading || !autoPowerSupported"
					@update:model-value="updateAutoPower"
				/>
			</bt-form-item>
		</bt-list>
		<div
			class="row justify-end q-my-lg"
			v-if="terminusDStore.olaresInfo && meStore.isAdmin"
		>
			<CustomButton
				v-if="meStore.isOwner && isOlaresDevice"
				:label="t('Reset SSH Password')"
				icon="sym_r_terminal"
				class="q-mr-md"
				@click="resetSSHPasswordOlares"
			/>

			<CustomButton
				v-if="meStore.isOwner"
				:label="t('Shutdown')"
				icon="sym_r_power_settings_new"
				class="q-mr-md"
				@click="shutdownOlares"
			/>
			<CustomButton
				v-if="meStore.isOwner"
				:label="t('Restart')"
				icon="sym_r_reset_tv"
				@click="restartOlares"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from '../../../components/settings/PageTitleComponent.vue';
import { useI18n } from 'vue-i18n';
import BtList from '../../../components/settings/base/BtList.vue';
import BtFormItem from '../../../components/settings/base/BtFormItem.vue';
import {
	deviceLogo,
	getSourceId,
	isOlaresGlobalDevice,
	OlaresSourceType,
	OneWorkerMode,
	oneWorkerModeOptions,
	getSettingsServerMdnsRequestApi,
	MdnsApiEmum
} from '../../../services/abstractions/mdns/service';
import OlaresStatus from '../../../components/base/OlaresStatus.vue';
import { useTerminusDStore } from '../../../stores/settings/terminusd';
import { computed, onMounted, ref, watch } from 'vue';
import { useMeStore } from '../../../stores/settings/me';
import { useTerminusStore } from '../../../stores/terminus';
import CustomButton from '../../../components/settings/CustomButton.vue';
import BtSelect from '../../../components/settings/base/BtSelect.vue';
import { useRouter } from 'vue-router';
import UpdateSSHPassworDialog from './dialog/UpdateSSHPassworDialog.vue';
import { useQuasar } from 'quasar';
import { useTokenStore } from 'src/stores/settings/token';
import axios from 'axios';
import { notifyFailed } from 'src/utils/settings/btNotify';

const { t } = useI18n();

const terminusDStore = useTerminusDStore();

const router = useRouter();
const quasar = useQuasar();

const meStore = useMeStore();
const terminusStore = useTerminusStore();

onMounted(async () => {
	await terminusDStore.system_status();
});

const shutdownOlares = () => {
	terminusDStore.commandData = {
		username: terminusStore.olaresId.split('@')[0]
	};
	router.push('/hardware/shutdown');
};

const restartOlares = () => {
	terminusDStore.commandData = {
		username: terminusStore.olaresId.split('@')[0]
	};
	router.push('/hardware/reboot');
};

const resetSSHPasswordOlares = () => {
	quasar
		.dialog({
			component: UpdateSSHPassworDialog
		})
		.onOk((password: string) => {
			terminusDStore.commandData = {
				password
			};
			router.push('/hardware/ssh-password');
		});
};

const isOlareOne = computed(() => {
	if (!terminusDStore.olaresInfo) {
		return false;
	}
	return getSourceId(terminusDStore.olaresInfo) == OlaresSourceType.ONE;
});

const isOlaresDevice = computed(() => {
	if (!terminusDStore.olaresInfo) {
		return false;
	}
	return isOlaresGlobalDevice(terminusDStore.olaresInfo);
});

const tokenStore = useTokenStore();

const updateWorkerMode = async () => {
	const url = getOneGpuUrl();
	try {
		await axios.post(url, {
			mode: oneWorkerMode.value
		});
	} catch (error) {
		notifyFailed(error);
	}
};

const oneWorkerMode = ref(OneWorkerMode.Quite);

const cpuFreqLimitEnabled = ref(false);
const autoPowerOn = ref(false);
const autoPowerSupported = ref(false);
const autoPowerLoading = ref(false);

let cpuFreqReqId = 0;
let autoPowerReqId = 0;

const getOneUrl = (api: MdnsApiEmum) => {
	return tokenStore.url + getSettingsServerMdnsRequestApi(api);
};

watch(
	() => terminusDStore.olaresInfo,
	async () => {
		if (terminusDStore.olaresInfo && isOlareOne.value) {
			loadWorkerMode();
			loadCpuFreqLimit();
			loadAutoPower();
		}
	}
);

const loadWorkerMode = async () => {
	try {
		const result: { mode: OneWorkerMode } = await axios.get(getOneGpuUrl());
		oneWorkerMode.value = result.mode;
	} catch (error) {
		/* empty */
	}
};

const loadCpuFreqLimit = async () => {
	const id = ++cpuFreqReqId;
	try {
		const result: { enabled: boolean } = await axios.get(
			getOneUrl(MdnsApiEmum.ONE_CPU_FREQ_LIMIT_STATUS)
		);
		if (id !== cpuFreqReqId) return;
		cpuFreqLimitEnabled.value = !!result.enabled;
	} catch (error) {
		/* empty */
	}
};

const loadAutoPower = async () => {
	const id = ++autoPowerReqId;
	autoPowerLoading.value = true;
	try {
		const result: { auto_power_on: boolean } = await axios.get(
			getOneUrl(MdnsApiEmum.ONE_GET_AUTO_POWER)
		);
		if (id !== autoPowerReqId) return;
		autoPowerSupported.value = true;
		autoPowerOn.value = !!result.auto_power_on;
	} catch (error: any) {
		if (id !== autoPowerReqId) return;
		autoPowerSupported.value = false;
		if (error?.response?.status !== 501) {
			notifyFailed(error);
		}
	} finally {
		if (id === autoPowerReqId) autoPowerLoading.value = false;
	}
};

const updateCpuFreqLimit = async (value: boolean) => {
	++cpuFreqReqId;
	try {
		await axios.post(
			getOneUrl(
				value
					? MdnsApiEmum.ONE_CPU_FREQ_LIMIT_ON
					: MdnsApiEmum.ONE_CPU_FREQ_LIMIT_OFF
			)
		);
	} catch (error) {
		cpuFreqLimitEnabled.value = !value;
		notifyFailed(error);
	}
};

const updateAutoPower = async (value: boolean) => {
	++autoPowerReqId;
	try {
		await axios.post(getOneUrl(MdnsApiEmum.ONE_SET_AUTO_POWER), {
			auto_power_mode: value ? 1 : 0
		});
	} catch (error: any) {
		autoPowerOn.value = !value;
		if (error?.response?.status === 501) {
			autoPowerSupported.value = false;
		} else {
			notifyFailed(error);
		}
	}
};

const getOneGpuUrl = () => {
	return getOneUrl(MdnsApiEmum.ONE_GPU_MODE);
};
</script>

<style scoped lang="scss"></style>
