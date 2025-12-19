<template>
	<div class="q-mt-lg machine-item q-pa-md column items-center justify-center">
		<div class="row items-center justify-between" style="width: 100%">
			<q-img :src="deviceLogo(machine.status)" width="80px" height="80px" />
			<div class="machine-info" v-if="machine.status">
				<div class="text-h6 text-ink-1 q-pb-xs">
					{{ machine.status.device_name }}
				</div>
				<div
					class="row items-center q-pb-xs"
					:class="installDisplayStatus(machine.status).textClass"
				>
					<q-icon
						:name="installDisplayStatus(machine.status).icon"
						size="16px"
					/>
					<div class="text-body3 q-ml-xs">
						{{ installDisplayStatus(machine.status).status }}
					</div>
				</div>
				<div class="text-body3 text-ink-2 q-pb-xs">
					{{ t('System version') + ':' + machine.status.terminusVersion }}
				</div>
				<div class="text-body3 text-ink-2">
					{{ `${t('IP')}: ${machine.status.hostIp || '--'}` }}
				</div>
			</div>
		</div>
		<template v-if="isBluetooth">
			<q-btn
				v-if="machine.status && !disableOperate(machine.status)"
				class="confirm row items-center justify-center q-mt-md"
				@click="bluetoothConfigNetwork(machine)"
				flat
				no-caps
				dense
			>
				<div class="text-white">{{ t('Network setup') }}</div>
			</q-btn>
		</template>

		<template v-else>
			<progress-button
				v-if="machine.status && isInstalling(machine.status)"
				:buttonText="
					t('Installing...') + ' ' + machine.status.installingProgress
				"
				textClass="text-subtitle2"
				:progress="
					machine.status.installingProgress
						? machine.status.installingProgress.split('%')[0]
						: '0'
				"
				:defaultTextColor="processBarColor"
				:progress-bar-Color="processBarColor"
				covered-text-color="#fff"
				progress-bar-class="process-bar-class"
				style="border-radius: 8px; overflow: hidden; height: 32px"
				:backgroundColor="defaultBGColor"
				class="q-mt-md"
			/>
			<q-btn
				v-if="machine.status && canInstall(machine.status)"
				class="confirm row items-center justify-center q-mt-md"
				@click="installAction(machine)"
				flat
				no-caps
				dense
			>
				<div class="text-white">{{ t('Install Now') }}</div>
			</q-btn>

			<q-btn
				v-if="
					machine.status &&
					canActive(machine.status) &&
					(!machine.status.terminusName ||
						currentUserName == machine.status.terminusName)
				"
				class="confirm row items-center justify-center q-mt-md"
				@click="activeLocal(machine)"
				flat
				no-caps
				dense
			>
				<div class="text-white">{{ t('Activate Now') }}</div>
			</q-btn>

			<q-btn
				v-if="machine.status && canUnInstall(machine.status)"
				class="confirm row items-center justify-center q-mt-md"
				@click="uninstall(machine)"
				flat
				no-caps
				dense
			>
				<div class="text-white">{{ t('Uninstall') }}</div>
			</q-btn>
		</template>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import {
	TerminusServiceInfo,
	canInstall,
	isInstalling,
	canActive,
	installDisplayStatus,
	canUnInstall,
	deviceLogo,
	disableOperate
} from '../../services/abstractions/mdns/service';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../stores/user';
import ProgressButton from '../../components/common/ProgressButton.vue';
import { useColor } from '@bytetrade/ui';

defineProps({
	machine: {
		type: Object as PropType<TerminusServiceInfo>,
		required: true
	},
	isBluetooth: {
		type: Boolean,
		required: false,
		default: false
	}
});

const { t } = useI18n();

const userStore = useUserStore();

const currentUserName = userStore.current_user?.name;

const installAction = async (machine: TerminusServiceInfo) => {
	emits('installAction', machine);
};

const activeLocal = async (machine: TerminusServiceInfo) => {
	emits('activeAction', machine);
};

const uninstall = async (machine: TerminusServiceInfo) => {
	emits('uninstallAction', machine);
};

const bluetoothConfigNetwork = async (machine: TerminusServiceInfo) => {
	emits('bluetoothConfigNetwork', machine);
};

const emits = defineEmits([
	'installAction',
	'activeAction',
	'uninstallAction',
	'bluetoothConfigNetwork'
]);

const { color: processBarColor } = useColor('light-blue-default');

const { color: defaultBGColor } = useColor('light-blue-soft');
</script>

<style scoped lang="scss">
.machine-item {
	border: 1px solid $separator;
	border-radius: 8px;
	width: 100%;
	.confirm {
		width: 100%;
		height: 32px;
		background: $light-blue-default;
		border-radius: 8px;

		&:before {
			box-shadow: none;
		}
	}

	.machine-info {
		width: calc(100% - 100px);
	}

	.process-bar-class {
		border-top-left-radius: 8px;
		border-bottom-left-radius: 8px;
	}
}
</style>
