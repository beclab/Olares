<template>
	<div class="empty column items-center justify-center">
		<div class="text-h3 text-ink-1 q-mb-sm">{{ t('home_welcome') }}</div>
		<div class="text-subtitle1 text-ink-3">
			{{ t('home_desc') }}
		</div>
		<img class="empty-image" src="../assets/empty.svg" />
		<div class="deploy column">
			<div
				class="text-subtitle1 text-teal-pressed deploy-text row items-center justify-start"
				@click="createHelloWorld"
			>
				<q-icon class="q-mr-sm" name="sym_r_developer_mode_tv" size="24px" />
				<span>{{ t('docker.create_default_app') }}</span>
			</div>

			<div
				class="text-subtitle1 text-teal-pressed deploy-text row items-center justify-start"
				@click="createApp"
			>
				<q-icon class="q-mr-sm" name="sym_r_deployed_code" size="24px" />
				<span>{{ t('docker.deploy_container') }}</span>
			</div>
			<!--
			<div
				class="text-subtitle1 text-teal-pressed deploy-text row items-center justify-start deploy-disabled"
				disabled

			>
				<q-icon class="q-mr-sm" name="sym_r_upload_file" size="24px" />
				<span>{{ t('docker.transplant_app') }}</span>
			</div> -->

			<div
				class="text-subtitle1 text-teal-pressed deploy-text row items-center justify-start"
				@click="codingInOlares"
			>
				<q-icon class="q-mr-sm" name="sym_r_box_edit" size="24px" />
				<span>{{ t('docker.coding') }}</span>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useDockerStore } from './../stores/docker';
import { pushToSystem } from './../utils/utils';
import CreateApp from '../components/dialog/CreateApp.vue';
import CodingInOlares from '../components/dialog/CodingInOlares.vue';
import { useDevelopingApps } from '@apps/studio/src/stores/app';

const { t } = useI18n();
const $q = useQuasar();
const route = useRoute();
const router = useRouter();

const dockerStore = useDockerStore();
const apps = useDevelopingApps();

const createHelloWorld = async () => {
	$q.dialog({
		component: CreateApp,
		componentProps: {
			defaultValues: {
				image: 'beclab/studio-app:1.0.0',
				startCmdArgs: 'hello world',
				port: '80',
				requiredCpu: '100m',
				requiredMemory: '128Mi'
			},
			disabledFields: {
				image: true,
				startCmd: true,
				startCmdArgs: true,
				port: true
			},
			hiddenFields: {
				startCmd: true
			}
		}
	});
};

const createApp = async () => {
	$q.dialog({
		component: CreateApp
	});
};

const codingInOlares = () => {
	$q.dialog({
		component: CodingInOlares,
		componentProps: {
			minMemory: 256
		}
	});
};

const initApp = async () => {
	try {
		await dockerStore.create_app(route.params.id as string);

		$q.loading.show({
			message: t('installing')
		});

		await dockerStore.install_app(route.params.id as string);
		await apps.getApps();
		await updateStatus();
	} catch (error) {
		await updateStatus();
	}
};

const updateStatus = async () => {
	await pushToSystem(route.params.id, router);
	await dockerStore.get_app_status(route.params.id as string);
	await dockerStore.get_app_install_state(route.params.id as string);
	$q.loading.hide();
};
</script>
<style lang="scss" scoped>
.empty {
	height: 100vh;
	.empty-image {
		margin: 56px 0 32px;
	}

	.deploy-text {
		margin-bottom: 24px;
		cursor: pointer;
		&:hover {
			opacity: 0.8;
		}

		&.deploy-disabled {
			opacity: 0.5;
		}
	}
}
</style>
../stores/docker../utils/utils
