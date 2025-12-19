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
				:title="t('Device Identifier')"
				:data="adminStore.olares_device_id"
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
			<!--				:title="t('Memroy')"-->
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
				:widthSeparator="false"
			>
				<bt-select
					v-model="oneWorkerMode"
					:options="oneWorkerModeOptions()"
					@update:modelValue="updateWorkerMode"
				/>
			</bt-form-item>
		</bt-list>
		<div
			class="row justify-end q-my-lg"
			v-if="terminusDStore.olaresInfo && adminStore.isAdmin"
		>
			<CustomButton
				v-if="adminStore.isOwner && isOlaresDevice"
				:label="t('Reset SSH Password')"
				icon="sym_r_terminal"
				class="q-mr-md"
				@click="resetSSHPasswordOlares"
			/>

			<CustomButton
				:label="t('Shutdown')"
				icon="sym_r_power_settings_new"
				class="q-mr-md"
				@click="shutdownOlares"
			/>
			<CustomButton
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
import { useAdminStore } from '../../../stores/settings/admin';
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

const adminStore = useAdminStore();

onMounted(async () => {
	await terminusDStore.system_status();
});

const shutdownOlares = () => {
	terminusDStore.commandData = {
		username: adminStore.olaresId.split('@')[0]
	};
	router.push('/hardware/shutdown');
};

const restartOlares = () => {
	terminusDStore.commandData = {
		username: adminStore.olaresId.split('@')[0]
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

watch(
	() => terminusDStore.olaresInfo,
	async () => {
		if (terminusDStore.olaresInfo && isOlareOne.value) {
			const url = getOneGpuUrl();
			try {
				const result: { mode: OneWorkerMode } = await axios.get(url);
				oneWorkerMode.value = result.mode;
			} catch (error) {
				/* empty */
			}
		}
	}
);

const getOneGpuUrl = () => {
	return (
		tokenStore.url + getSettingsServerMdnsRequestApi(MdnsApiEmum.ONE_GPU_MODE)
	);
};
</script>

<style scoped lang="scss"></style>
