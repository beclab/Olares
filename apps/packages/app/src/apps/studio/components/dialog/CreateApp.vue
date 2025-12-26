<template>
	<q-dialog
		class="create-app"
		position="right"
		full-height
		v-model="show"
		ref="dialogCreateRef"
		tabindex="1"
	>
		<div class="create-container">
			<div class="row items-center justify-between create-header">
				<span class="text-h6 text-ink-1">{{
					t('docker.create_app_title')
				}}</span>
				<q-btn
					dense
					flat
					icon="close"
					color="ink-3"
					v-close-popup
					@click="onDialogCancel"
				>
					<q-tooltip>{{ t('close') }}</q-tooltip>
				</q-btn>
			</div>
			<div style="height: calc(100% - 164px)">
				<BtScrollArea style="height: 100%; width: 100%">
					<image-deployer
						ref="imageDeployerRef"
						:default-values="imageDefaultValues"
						:disabled-fields="disabledFields"
						:hidden-fields="hiddenFields"
						@update-container="updateContainer"
					/>
					<instance-config
						ref="instanceRef"
						:default-values="instanceDefaultValues"
						@update-instance="updateInstance"
					/>
					<environment-config @update-env="updateEnv" />
					<mounts-address @update-mounts="updateMounts" />
				</BtScrollArea>
			</div>

			<div class="footer row items-center justify-end q-pr-lg">
				<TerminusFormFooter
					:ok-text="t('create')"
					:cancel-text="t('cancel')"
					:loading="loading"
					@submit="submit"
					@cancel="onDialogCancel"
				/>
			</div>
		</div>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { useDialogPluginComponent } from 'quasar';
import { useDockerStore } from '../../stores/docker';
import { CreateWithOneDockerConfig, VENDOR } from '@apps/studio/src/types/core';
import { pushToSystem } from '../../utils/utils';
import { useDevelopingApps } from '@apps/studio/src/stores/app';

import ImageDeployer from '../config/ImageDeployer.vue';
import InstanceConfig from '../config/InstanceConfig.vue';
import EnvironmentConfig from '../config/EnvironmentConfig.vue';
import MountsAddress from '../config/MountsAddress.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

interface Props {
	defaultValues?: {
		image?: string;
		startCmd?: string;
		startCmdArgs?: string;
		port?: string;
		requiredCpu?: string;
		requiredMemory?: string;
	};
	disabledFields?: {
		image?: boolean;
		startCmd?: boolean;
		startCmdArgs?: boolean;
		port?: boolean;
	};
	hiddenFields?: {
		image?: boolean;
		startCmd?: boolean;
		startCmdArgs?: boolean;
		port?: boolean;
	};
}

const props = withDefaults(defineProps<Props>(), {
	defaultValues: () => ({}),
	disabledFields: () => ({}),
	hiddenFields: () => ({})
});

const {
	dialogRef: dialogCreateRef,
	onDialogCancel,
	onDialogOK
} = useDialogPluginComponent();

const $q = useQuasar();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const dockerStore = useDockerStore();
const apps = useDevelopingApps();

const show = ref(true);
const imageDeployerRef = ref();
const instanceRef = ref();
const loading = ref(false);

const imageDefaultValues = computed(() => ({
	image: props.defaultValues?.image,
	startCmd: props.defaultValues?.startCmd,
	startCmdArgs: props.defaultValues?.startCmdArgs,
	port: props.defaultValues?.port
}));

const instanceDefaultValues = computed(() => ({
	requiredCpu: props.defaultValues?.requiredCpu,
	requiredMemory: props.defaultValues?.requiredMemory
}));

const config = reactive<CreateWithOneDockerConfig>({
	name: route.params.id as string,
	container: {
		image: '',
		port: 80,
		startCmd: '',
		startCmdArgs: ''
	},
	requiredCpu: '',
	limitedCpu: '',
	requiredMemory: '',
	limitedMemory: '',
	requiredGpu: false,
	needPg: false,
	needRedis: false,
	gpuVendor: VENDOR.NVIDIA,
	env: {},
	mounts: {}
});

const updateContainer = (value) => {
	config.container = { ...value, port: Number(value.port) };
};

const updateInstance = (value) => {
	Object.assign(config, value);
};

const updateEnv = (value) => {
	config.env = value;
};

const updateMounts = (value) => {
	config.mounts = value;
};

const submit = async () => {
	const imageValidate = imageDeployerRef.value.validate();
	const instanceValidate = instanceRef.value.validate();

	if (imageValidate && instanceValidate) {
		try {
			loading.value = true;
			await dockerStore.config_app(config);

			$q.loading.show({
				message: t('installing')
			});

			await dockerStore.install_app(config.name);
			await apps.getApps();

			await updateStatus();
		} catch (error) {
			await updateStatus();
		}

		onDialogOK();
	}
};

const updateStatus = async () => {
	$q.loading.hide();
	pushToSystem(route.params.id, router);
	await dockerStore.get_app_status(route.params.id as string);
	loading.value = false;
};
</script>

<style lang="scss" scoped>
.create-app {
	padding: 0 !important;
	::v-deep(.q-dialog__inner--minimized) {
		padding: 0;
	}

	.create-container {
		width: calc(100vw * 0.6);
		max-width: calc(100vw * 0.6);
		height: 100vh !important;
		background-color: $background-3;

		.create-header {
			height: 64px;
			background-color: $background-1;
			padding: 0 20px;
		}

		.footer {
			width: 100%;
			height: 80px;
			background-color: $background-1;
			position: absolute;
			bottom: 0;
			right: 0;
			border-top: 1px solid $input-stroke;
		}
	}
}
</style>
