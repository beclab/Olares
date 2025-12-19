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
				<span class="text-h6 text-ink-1">{{ t('docker.coding') }}</span>
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
					<create-container
						ref="imageDeployerRef"
						:min-memory="props.minMemory"
						@update-image="updateImage"
					/>
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
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { useDialogPluginComponent } from 'quasar';
import { useDockerStore } from '../../stores/docker';
import { CreateWithOneImageConfig } from '@apps/studio/src/types/core';
import { pushToSystem } from '../../utils/utils';

import CreateContainer from '../config/CreateContainer.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';
import { useDevelopingApps } from '@apps/studio/src/stores/app';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

interface Props {
	minMemory?: number;
}

const props = withDefaults(defineProps<Props>(), {
	minMemory: 128
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
const apps = useDevelopingApps();
const dockerStore = useDockerStore();

const show = ref(true);
const imageDeployerRef = ref();
const loading = ref(false);

const config = ref<CreateWithOneImageConfig>({
	name: route.params.id as string,
	devEnv: '',
	requiredCpu: '',
	requiredMemory: '',
	requiredDisk: ''
});

const updateImage = (value) => {
	console.log('value', value);
	console.log('config', config.value);

	config.value = { ...config.value, ...value };
};

const submit = async () => {
	const imageValidate = imageDeployerRef.value.validate();

	if (imageValidate) {
		try {
			loading.value = true;
			await dockerStore.create_app_code_in_olares(config.value);

			$q.loading.show({
				message: t('installing')
			});

			await dockerStore.install_app(config.value.name);
			await dockerStore.get_app_status(route.params.id as string);
			await dockerStore.get_app_install_state(route.params.id as string);
			await apps.getApps();
			$q.loading.hide();
			loading.value = false;
			pushToSystem(route.params.id, router);
		} catch (error) {
			$q.loading.hide();
			loading.value = false;
		}

		onDialogOK();
	}
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
